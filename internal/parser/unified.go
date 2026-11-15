// Package parser provides parsers that convert diff text into diff models.
package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cheonjaeung/chai/internal/ansi"
	"github.com/cheonjaeung/chai/internal/diff"
	"github.com/cheonjaeung/chai/internal/utils"
)

// gitExtendedHeaders are prefixes of the lines between "diff --git" and the file paths.
var gitExtendedHeaders = []string{
	"old mode ",
	"new mode ",
	"deleted file mode ",
	"new file mode ",
	"copy from ",
	"copy to ",
	"rename from ",
	"rename to ",
	"similarity index ",
	"dissimilarity index ",
	"index ",
	"Binary files ",
}

// ParseUnified parses a unified diff text format such as output of the
// 'git diff' or 'git show' command.
//
// It keeps lines outside of file diff as raw text to display as is.
//
// ANSI escape sequences are removed from diff, but kept in raw text.
func ParseUnified(text string) (diff.Diff, error) {
	raws := utils.SplitLines(text)
	lines := make([]string, len(raws))
	for i, raw := range raws {
		lines[i] = ansi.Strip(raw)
	}

	p := parser{lines: lines, rawLines: raws, pos: 0}
	var d diff.Diff
	for p.pos < len(p.lines) {
		if p.isGitDiffHeader() || p.isFileHeader() {
			file, err := p.parseFile()
			if err != nil {
				return diff.Diff{}, err
			}
			d.Entries = append(d.Entries, diff.Entry{File: &file})
			continue
		}
		d.Entries = append(d.Entries, diff.Entry{Text: strings.TrimSuffix(p.rawLines[p.pos], "\n")})
		p.pos++
	}

	return d, nil
}

type parser struct {
	// rawLines are the input lines as is.
	rawLines []string
	// lines are the input lines without ANSI escape sequences.
	lines []string
	// pos is the current line position to parse.
	pos int
}

func (p *parser) parseFile() (diff.File, error) {
	var f diff.File
	start := p.pos

	isGit := p.isGitDiffHeader()
	if isGit {
		f.OldPath, f.NewPath = parseGitPaths(p.lines[p.pos])
		p.pos++
		for p.pos < len(p.lines) && isGitExtendedHeader(p.lines[p.pos]) {
			if p.isBinaryFile() {
				f.Binary = true
			}
			p.pos++
		}
	}

	if p.isFileHeader() {
		if isGit {
			f.OldPath = strings.TrimPrefix(parseFilePath(p.lines[p.pos], true), "a/")
			f.NewPath = strings.TrimPrefix(parseFilePath(p.lines[p.pos+1], false), "b/")
		} else {
			f.OldPath = parseFilePath(p.lines[p.pos], true)
			f.NewPath = parseFilePath(p.lines[p.pos+1], false)
		}
		p.pos += 2
	}

	for _, raw := range p.rawLines[start:p.pos] {
		f.RawHeader = append(f.RawHeader, strings.TrimSuffix(raw, "\n"))
	}

	for p.isHunkHeader() {
		h, err := p.parseHunk()
		if err != nil {
			return diff.File{}, err
		}
		f.Hunks = append(f.Hunks, h)
	}

	return f, nil
}

// isGitDiffHeader returns true if the current line starts with git diff prefix.
func (p *parser) isGitDiffHeader() bool {
	return p.pos < len(p.lines) && strings.HasPrefix(p.lines[p.pos], "diff --git ")
}

// isFileHeader returns true if the current line starts with file header
// such as "--- a/file\n+++ b/file".
func (p *parser) isFileHeader() bool {
	return p.pos < len(p.lines) && strings.HasPrefix(p.lines[p.pos], "--- ") &&
		p.pos+1 < len(p.lines) && strings.HasPrefix(p.lines[p.pos+1], "+++ ")
}

func (p *parser) isBinaryFile() bool {
	return p.pos < len(p.lines) && strings.HasPrefix(p.lines[p.pos], "Binary files ")
}

func (p *parser) isHunkHeader() bool {
	return p.pos < len(p.lines) && strings.HasPrefix(p.lines[p.pos], "@@ ")
}

func isGitExtendedHeader(line string) bool {
	return slices.ContainsFunc(gitExtendedHeaders, func(prefix string) bool {
		return strings.HasPrefix(line, prefix)
	})
}

// parseGitPaths parses the paths of a line such as "diff --git a/old b/new".
func parseGitPaths(line string) (oldPath, newPath string) {
	paths := strings.TrimSuffix(strings.TrimPrefix(line, "diff --git "), "\n")
	oldPath, newPath, _ = strings.Cut(paths, " b/")
	return strings.TrimPrefix(oldPath, "a/"), newPath
}

// parseFilePath parses the path from a file header such as "--- a/file" or "+++ b/file".
func parseFilePath(text string, isOld bool) string {
	text = strings.TrimSuffix(text, "\n")
	if isOld {
		return strings.TrimPrefix(text, "--- ")
	} else {
		return strings.TrimPrefix(text, "+++ ")
	}
}

func (p *parser) parseHunk() (diff.Hunk, error) {
	var h diff.Hunk
	var parsed bool

	h.OldStart, h.OldLines, h.NewStart, h.NewLines, parsed = parseHunkHeader(p.lines[p.pos])
	if !parsed {
		return diff.Hunk{}, fmt.Errorf("invalid hunk header at pos %d", p.pos)
	}
	p.pos++

	oldLeft := h.OldLines
	newLeft := h.NewLines
	oldNumber := h.OldStart
	newNumber := h.NewStart
	for {
		if p.pos < len(p.lines) && strings.HasPrefix(p.lines[p.pos], "\\") && len(h.Lines) > 0 {
			h.Lines[len(h.Lines)-1].NoNewlineAtEOF = true
			p.pos++
			continue
		}

		if oldLeft == 0 && newLeft == 0 {
			return h, nil
		}

		if p.pos >= len(p.lines) {
			return diff.Hunk{}, fmt.Errorf("unexpected end of file in hunk at pos %d", p.pos)
		}

		line := p.lines[p.pos]
		if line == "" {
			return diff.Hunk{}, fmt.Errorf("empty hunk line at pos %d", p.pos)
		}

		content := strings.TrimSuffix(line[1:], "\n")
		switch {
		case line[0] == ' ' && oldLeft > 0 && newLeft > 0:
			h.Lines = append(h.Lines, diff.Line{Kind: diff.LineContext, Content: content, OldNumber: oldNumber, NewNumber: newNumber})
			oldNumber++
			newNumber++
			oldLeft--
			newLeft--
		case line[0] == '-' && oldLeft > 0:
			h.Lines = append(h.Lines, diff.Line{Kind: diff.LineDeleted, Content: content, OldNumber: oldNumber})
			oldNumber++
			oldLeft--
		case line[0] == '+' && newLeft > 0:
			h.Lines = append(h.Lines, diff.Line{Kind: diff.LineAdded, Content: content, NewNumber: newNumber})
			newNumber++
			newLeft--
		default:
			return diff.Hunk{}, fmt.Errorf("invalid hunk line format at pos %d: %q", p.pos, line)
		}
		p.pos++
	}
}

// parseHunkHeader parses a line such as "@@ -1,3 +1,4 @@ func main() {".
func parseHunkHeader(line string) (oldStart int, oldLines int, newStart int, newLines int, parsed bool) {
	rest, parsed := strings.CutPrefix(line, "@@ -")
	if !parsed {
		return 0, 0, 0, 0, false
	}
	ranges, _, parsed := strings.Cut(rest, " @@")
	if !parsed {
		return 0, 0, 0, 0, false
	}
	oldRange, newRange, parsed := strings.Cut(ranges, " +")
	if !parsed {
		return 0, 0, 0, 0, false
	}
	oldStart, oldLines, parsed = parseRange(oldRange)
	if !parsed {
		return 0, 0, 0, 0, false
	}
	newStart, newLines, parsed = parseRange(newRange)
	if !parsed {
		return 0, 0, 0, 0, false
	}
	return oldStart, oldLines, newStart, newLines, true
}

// parseRange parses start and line count from a hunk range. Default count is 1.
func parseRange(s string) (start, count int, ok bool) {
	startText, countText, hasCount := strings.Cut(s, ",")
	start, err := strconv.Atoi(startText)
	if err != nil {
		return 0, 0, false
	}
	if !hasCount {
		return start, 1, true
	}
	count, err = strconv.Atoi(countText)
	return start, count, err == nil
}
