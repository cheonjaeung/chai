package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/cheonjaeung/chai/internal/ansi"
	"github.com/cheonjaeung/chai/internal/diff"
	"github.com/cheonjaeung/chai/internal/parser"
	"github.com/gdamore/tcell/v2"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

var version = "0.0.0"

const usage = `chai is an interactive diff tool.

Usage: chai [OPTIONS] [FILE1] [FILE2]
       <COMMAND> | chai [OPTIONS]
       <COMMAND> | chai --git-diff-filter

Arguments:
  [FILE1]: The first file to compare.
  [FILE2]: The second file to compare.

chai reads stdin when there are no files.

Options:
      --git-diff-filter  Print diff for interactive.diffFilter of git.
  -h, --help             Print help message.
  -V, --version          Print version.
`

func main() {
	os.Exit(runMain())
}

func runMain() int {
	var gitDiffFilter, help, showVersion bool
	pflag.BoolVar(&gitDiffFilter, "git-diff-filter", false, "")
	pflag.BoolVarP(&help, "help", "h", false, "")
	pflag.BoolVarP(&showVersion, "version", "V", false, "")
	pflag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	pflag.Parse()

	if help {
		fmt.Print(usage)
		return 0
	}
	if showVersion {
		fmt.Println("chai version", version)
		return 0
	}

	args := pflag.Args()
	var d diff.Diff
	var err error

	if gitDiffFilter && len(args) > 0 {
		fmt.Fprintf(os.Stderr, "error: --git-diff-filter is not compatible with file arguments\n\n")
		pflag.Usage()
		return 2
	} else if len(args) == 2 {
		d, err = diffFiles(args[0], args[1])
	} else if len(args) == 0 && !term.IsTerminal(int(os.Stdin.Fd())) {
		d, err = parseStdin()
	} else {
		fmt.Fprintf(os.Stderr, "error: expected 2 files or a diff from stdin, got %d files\n\n", len(args))
		pflag.Usage()
		return 2
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	if len(d.Entries) == 0 {
		return 0
	}

	if gitDiffFilter {
		err = runGitDiffFilterPrinter(os.Stdout, d)
	} else if term.IsTerminal(int(os.Stdout.Fd())) {
		err = runInteractiveViewer(d)
	} else {
		err = runPrinter(os.Stdout, d)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}
	return 0
}

func diffFiles(path1, path2 string) (diff.Diff, error) {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return diff.Diff{}, err
	}
	content2, err := os.ReadFile(path2)
	if err != nil {
		return diff.Diff{}, err
	}

	file := diff.MyersDiff(string(content1), string(content2))
	file.OldPath = path1
	file.NewPath = path2
	return diff.Diff{Entries: []diff.Entry{{File: &file}}}, nil
}

func parseStdin() (diff.Diff, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return diff.Diff{}, err
	}
	return parser.ParseUnified(string(input))
}

func runInteractiveViewer(d diff.Diff) error {
	viewer, err := NewViewer(d)
	if err != nil {
		return err
	}
	return viewer.Loop()
}

func runPrinter(writer io.Writer, d diff.Diff) error {
	bufWriter := bufio.NewWriter(writer)
	for _, r := range formatDiff(d, false) {
		bufWriter.WriteString(ansi.Strip(r.String()))
		bufWriter.WriteByte('\n')
	}
	return bufWriter.Flush()
}

// runGitDiffFilterPrinter writes diff to writer for --git-diff-filter option.
//
// Git requires that custom diff filter maintain a one-to-one line correspondence
// between original and custom output.
func runGitDiffFilterPrinter(writer io.Writer, d diff.Diff) error {
	bufWriter := bufio.NewWriter(writer)
	for _, r := range formatDiff(d, true) {
		switch r.kind {
		case rowLine, rowNoNewline:
			numberStyle, lineStyle := lineStyles(r)
			bufWriter.WriteString(sgr(numberStyle) + r.oldNumber)
			bufWriter.WriteString(sgr(separatorStyle) + "⋮")
			bufWriter.WriteString(sgr(numberStyle) + r.newNumber)
			bufWriter.WriteString(sgr(tcell.StyleDefault) + " ")
			bufWriter.WriteString(sgr(lineStyle) + r.marker)
			for _, s := range highlightSegments(r.text, 0, r.highlights) {
				if s.highlighted {
					bufWriter.WriteString(sgr(highlightStyle(r)) + s.text + sgr(lineStyle))
				} else {
					bufWriter.WriteString(s.text)
				}
			}
			// Erases to the end of the line to fill the background like the viewer.
			bufWriter.WriteString("\x1b[K")
		case rowHunkHeader:
			bufWriter.WriteString(sgr(hunkHeaderStyle) + r.text)
		default:
			bufWriter.WriteString(r.text)
		}
		bufWriter.WriteString("\x1b[m\n")
	}
	return bufWriter.Flush()
}
