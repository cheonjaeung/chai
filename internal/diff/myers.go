package diff

import (
	"strings"

	"github.com/cheonjaeung/chai/internal/utils"
)

type snake struct {
	startX int
	startY int
	endX   int
	endY   int
}

// edit is a single step to transform old into new.
//
// oldIndex is the index of this step at the old sequence, and newIndex is the
// index at the new sequence.
type edit struct {
	kind     LineKind
	oldIndex int
	newIndex int
}

// MyersDiff compares two files, oldText and newText, using the myers diff
// algorithm.
//
// The result has a single hunk that contains all lines of both files, or no
// hunk if both files are empty.
func MyersDiff(oldText string, newText string) File {
	a := utils.SplitLines(oldText)
	b := utils.SplitLines(newText)
	lines := diffLines(a, b)
	if len(lines) == 0 {
		return File{}
	}
	return File{Hunks: []Hunk{newHunk(lines)}}
}

func diffLines(a []string, b []string) []Line {
	var lines []Line
	for _, e := range editScript(a, b) {
		switch e.kind {
		case LineContext:
			lines = append(lines, newLine(LineContext, a[e.oldIndex], e.oldIndex+1, e.newIndex+1))
		case LineDeleted:
			lines = append(lines, newLine(LineDeleted, a[e.oldIndex], e.oldIndex+1, 0))
		case LineAdded:
			lines = append(lines, newLine(LineAdded, b[e.newIndex], 0, e.newIndex+1))
		}
	}
	return lines
}

func editScript(a []string, b []string) []edit {
	n := len(a)
	m := len(b)
	offset := (n+m+1)/2 + 1

	var script []edit

	vf := make([]int, 2*offset+1)
	vb := make([]int, 2*offset+1)

	diffRange(&script, a, b, 0, 0, vf, vb)

	return script
}

func diffRange(script *[]edit, a []string, b []string, aStart int, bStart int, vf []int, vb []int) {
	if len(a) == 0 {
		for i := range b {
			*script = append(*script, edit{kind: LineAdded, newIndex: bStart + i})
		}
		return
	}
	if len(b) == 0 {
		for i := range a {
			*script = append(*script, edit{kind: LineDeleted, oldIndex: aStart + i})
		}
		return
	}

	mid, d := middleSnake(a, b, vf, vb)

	if d > 1 {
		diffRange(script, a[:mid.startX], b[:mid.startY], aStart, bStart, vf, vb)
		for i := range mid.endX - mid.startX {
			*script = append(*script, edit{kind: LineContext, oldIndex: aStart + mid.startX + i, newIndex: bStart + mid.startY + i})
		}
		diffRange(script, a[mid.endX:], b[mid.endY:], aStart+mid.endX, bStart+mid.endY, vf, vb)
		return
	}

	x, y := 0, 0
	for x < len(a) && y < len(b) && a[x] == b[y] {
		*script = append(*script, edit{kind: LineContext, oldIndex: aStart + x, newIndex: bStart + y})
		x++
		y++
	}
	if len(a) > len(b) {
		*script = append(*script, edit{kind: LineDeleted, oldIndex: aStart + x})
		x++
	} else if len(b) > len(a) {
		*script = append(*script, edit{kind: LineAdded, newIndex: bStart + y})
		y++
	}
	for x < len(a) {
		*script = append(*script, edit{kind: LineContext, oldIndex: aStart + x, newIndex: bStart + y})
		x++
		y++
	}
}

func middleSnake(a []string, b []string, vf []int, vb []int) (snake, int) {
	n := len(a)
	m := len(b)
	delta := n - m
	isOdd := delta%2 != 0
	max := (n + m + 1) / 2
	offset := max + 1

	vf[offset+1] = 0
	vb[offset+1] = 0
	for d := 0; d <= max; d++ {
		// Search from forward.
		for k := d; k >= -d; k -= 2 {
			var x int
			if k == -d || (k != d && vf[offset+k-1] < vf[offset+k+1]) {
				x = vf[offset+k+1]
			} else {
				x = vf[offset+k-1] + 1
			}
			y := x - k
			x0, y0 := x, y
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			vf[offset+k] = x
			c := delta - k
			if isOdd && -(d-1) <= c && c <= d-1 && x+vb[offset+c] >= n {
				return snake{x0, y0, x, y}, 2*d - 1
			}
		}

		// Search from backward.
		for c := -d; c <= d; c += 2 {
			var u int
			if c == -d || (c != d && vb[offset+c-1] < vb[offset+c+1]) {
				u = vb[offset+c+1]
			} else {
				u = vb[offset+c-1] + 1
			}
			w := u - c
			u0, w0 := u, w
			for u < n && w < m && a[n-u-1] == b[m-w-1] {
				u++
				w++
			}
			vb[offset+c] = u
			k := delta - c
			if !isOdd && -d <= k && k <= d && u+vf[offset+k] >= n {
				return snake{n - u, m - w, n - u0, m - w0}, 2 * d
			}
		}
	}

	panic("unreachable")
}

func newLine(kind LineKind, raw string, oldNumber int, newNumber int) Line {
	content, hasNewline := strings.CutSuffix(raw, "\n")
	return Line{
		Kind:           kind,
		Content:        content,
		OldNumber:      oldNumber,
		NewNumber:      newNumber,
		NoNewlineAtEOF: !hasNewline,
	}
}

func newHunk(lines []Line) Hunk {
	h := Hunk{Lines: lines}
	for _, l := range lines {
		if l.OldNumber != 0 {
			if h.OldStart == 0 {
				h.OldStart = l.OldNumber
			}
			h.OldLines++
		}
		if l.NewNumber != 0 {
			if h.NewStart == 0 {
				h.NewStart = l.NewNumber
			}
			h.NewLines++
		}
	}
	return h
}
