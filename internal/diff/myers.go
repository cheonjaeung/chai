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

// MyersDiff compares two files, oldText and newText, using the myers diff
// algorithm.
//
// The result has a single hunk that contains all lines of both files, or no
// hunk if both files are empty.
func MyersDiff(oldText, newText string) File {
	a := utils.SplitLines(oldText)
	b := utils.SplitLines(newText)
	lines := diffLines(a, b)
	if len(lines) == 0 {
		return File{}
	}
	return File{Hunks: []Hunk{newHunk(lines)}}
}

func diffLines(a []string, b []string) []Line {
	n := len(a)
	m := len(b)
	offset := (n+m+1)/2 + 1

	var script []Line

	vf := make([]int, 2*offset+1)
	vb := make([]int, 2*offset+1)

	diffRange(&script, a, b, 0, 0, vf, vb)

	return script
}

func diffRange(script *[]Line, a []string, b []string, aStart int, bStart int, vf []int, vb []int) {
	if len(a) == 0 {
		for i, t := range b {
			*script = append(*script, newLine(LineAdded, t, 0, bStart+i+1))
		}
		return
	}
	if len(b) == 0 {
		for i, t := range a {
			*script = append(*script, newLine(LineDeleted, t, aStart+i+1, 0))
		}
		return
	}

	mid, d := middleSnake(a, b, vf, vb)

	if d > 1 {
		diffRange(script, a[:mid.startX], b[:mid.startY], aStart, bStart, vf, vb)
		for i, t := range a[mid.startX:mid.endX] {
			*script = append(*script, newLine(LineContext, t, aStart+mid.startX+i+1, bStart+mid.startY+i+1))
		}
		diffRange(script, a[mid.endX:], b[mid.endY:], aStart+mid.endX, bStart+mid.endY, vf, vb)
		return
	}

	x, y := 0, 0
	for x < len(a) && y < len(b) && a[x] == b[y] {
		*script = append(*script, newLine(LineContext, a[x], aStart+x+1, bStart+y+1))
		x++
		y++
	}
	if len(a) > len(b) {
		*script = append(*script, newLine(LineDeleted, a[x], aStart+x+1, 0))
		x++
	} else if len(b) > len(a) {
		*script = append(*script, newLine(LineAdded, b[y], 0, bStart+y+1))
		y++
	}
	for x < len(a) {
		*script = append(*script, newLine(LineContext, a[x], aStart+x+1, bStart+y+1))
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
