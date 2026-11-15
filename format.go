package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cheonjaeung/chai/internal/ansi"
	"github.com/cheonjaeung/chai/internal/diff"
	"github.com/rivo/uniseg"
)

const tabWidth = 4

type rowKind int

const (
	rowFileHeader rowKind = iota
	rowHunkHeader
	rowBinary
	rowLine
	rowNoNewline
	rowText
)

// row is a single logical row shared by the viewer and printers.
type row struct {
	kind rowKind
	// lineKind, oldNumber, newNumber and marker are used only by rowLine and rowNoNewline.
	lineKind   diff.LineKind
	oldNumber  string
	newNumber  string
	marker     string
	text       string
	highlights []diff.Range
}

func (r row) hasPrefix() bool {
	return r.kind == rowLine || r.kind == rowNoNewline
}

func (r row) prefix() string {
	return r.oldNumber + "⋮" + r.newNumber + " " + r.marker
}

func (r row) String() string {
	if r.hasPrefix() {
		return r.prefix() + r.text
	}
	return r.text
}

// formatDiff formats d into rows. If rawHeader is true, file headers are
// kept as is instead of the summarized one, to keep lines one-to-one.
func formatDiff(d diff.Diff, rawHeader bool) []row {
	var rows []row
	for _, e := range d.Entries {
		if e.File == nil {
			rows = append(rows, row{kind: rowText, text: sanitize(e.Text, true)})
			continue
		}
		rows = append(rows, formatFile(*e.File, rawHeader)...)
	}
	return rows
}

func formatFile(f diff.File, rawHeader bool) []row {
	var rows []row
	if rawHeader {
		for _, h := range f.RawHeader {
			rows = append(rows, row{kind: rowText, text: sanitize(h, true)})
		}
	} else {
		if f.OldPath == f.NewPath {
			rows = append(rows, row{kind: rowFileHeader, text: sanitize(f.OldPath, false)})
		} else {
			rows = append(rows, row{kind: rowFileHeader, text: sanitize(f.OldPath, false) + " → " + sanitize(f.NewPath, false)})
		}
		if f.Binary {
			return append(rows, row{kind: rowBinary, text: "Binary file"})
		}
	}

	width := numberWidth(f)
	blank := strings.Repeat(" ", width)
	for _, h := range f.Hunks {
		rows = append(rows, row{
			kind: rowHunkHeader,
			text: fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldLines, h.NewStart, h.NewLines),
		})
		texts := make([]string, len(h.Lines))
		for i, l := range h.Lines {
			texts[i] = sanitize(l.Content, false)
		}
		highlights := wordHighlights(h.Lines, texts)
		for i, l := range h.Lines {
			rows = append(rows, row{
				kind:       rowLine,
				lineKind:   l.Kind,
				oldNumber:  formatNumber(l.OldNumber, width),
				newNumber:  formatNumber(l.NewNumber, width),
				marker:     marker(l.Kind),
				text:       texts[i],
				highlights: highlights[i],
			})
			if l.NoNewlineAtEOF {
				rows = append(rows, row{
					kind:      rowNoNewline,
					oldNumber: blank,
					newNumber: blank,
					marker:    " ",
					text:      `\ No newline at end of file`,
				})
			}
		}
	}
	return rows
}

// wordHighlights returns the changed byte ranges of each text of lines.
//
// It pairs a block of deleted lines with the block of added lines right after
// it, only if both blocks have the same number of lines.
func wordHighlights(lines []diff.Line, texts []string) [][]diff.Range {
	highlights := make([][]diff.Range, len(lines))
	for i := 0; i < len(lines); {
		deleted := countKind(lines[i:], diff.LineDeleted)
		if deleted == 0 {
			i++
			continue
		}
		added := countKind(lines[i+deleted:], diff.LineAdded)
		// Highlight words only if the number of deleted and added lines is the same,
		// because pairing unequal lines may show a wrong highlight.
		if deleted == added {
			for j := range deleted {
				o, n := i+j, i+deleted+j
				highlights[o], highlights[n] = diff.WordDiff(texts[o], texts[n])
			}
		}
		i += deleted + added
	}
	return highlights
}

// countKind returns the number of consecutive lines of kind from the start of lines.
func countKind(lines []diff.Line, kind diff.LineKind) int {
	n := 0
	for n < len(lines) && lines[n].Kind == kind {
		n++
	}
	return n
}

// numberWidth returns the number of digits of the largest line number in f.
func numberWidth(f diff.File) int {
	largest := 0
	for _, h := range f.Hunks {
		for _, l := range h.Lines {
			largest = max(largest, l.OldNumber, l.NewNumber)
		}
	}
	return len(strconv.Itoa(largest))
}

func formatNumber(n, width int) string {
	if n == 0 {
		return strings.Repeat(" ", width)
	}
	return fmt.Sprintf("%*d", width, n)
}

func marker(kind diff.LineKind) string {
	switch kind {
	case diff.LineAdded:
		return "+"
	case diff.LineDeleted:
		return "-"
	default:
		return " "
	}
}

// sanitize makes text safe to display on a terminal.
// It removes a trailing carriage return, expands tabs to the next tab stop,
// replaces control characters with caret notation and invalid UTF-8 with U+FFFD.
// If keepSGR is true, SGR sequences such as "\x1b[31m" are kept and other CSI
// sequences are removed.
func sanitize(text string, keepSGR bool) string {
	text = strings.ToValidUTF8(strings.TrimSuffix(text, "\r"), "�")

	var sb strings.Builder
	column := 0
	state := -1
	for text != "" {
		if n := ansi.CsiLen(text); keepSGR && n > 0 {
			if text[n-1] == 'm' {
				sb.WriteString(text[:n])
			}
			text = text[n:]
			state = -1
			continue
		}
		var cluster string
		var width int
		cluster, text, width, state = uniseg.FirstGraphemeClusterInString(text, state)
		switch {
		case cluster == "\t":
			n := tabWidth - column%tabWidth
			sb.WriteString(strings.Repeat(" ", n))
			column += n
		case len(cluster) == 1 && (cluster[0] < 0x20 || cluster[0] == 0x7f):
			sb.WriteByte('^')
			sb.WriteByte(cluster[0] ^ 0x40)
			column += 2
		default:
			sb.WriteString(cluster)
			column += width
		}
	}
	return sb.String()
}
