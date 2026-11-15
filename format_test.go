package main

import (
	"reflect"
	"testing"

	"github.com/cheonjaeung/chai/internal/diff"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		keepSGR bool
		want    string
	}{
		{name: "plain", text: "abc", want: "abc"},
		{name: "trailing carriage return", text: "abc\r", want: "abc"},
		{name: "carriage return in middle", text: "a\rb", want: "a^Mb"},
		{name: "tab at start", text: "\tb", want: "    b"},
		{name: "tab to next stop", text: "ab\tc", want: "ab  c"},
		{name: "tab after full stop", text: "abcd\te", want: "abcd    e"},
		{name: "tab after wide characters", text: "한\tb", want: "한  b"},
		{name: "escape", text: "\x1b[31mred", want: "^[[31mred"},
		{name: "DEL", text: "a\x7f", want: "a^?"},
		{name: "invalid UTF-8", text: "a\xffb", want: "a�b"},
		{name: "keep SGR", text: "\x1b[31mred\x1b[m", keepSGR: true, want: "\x1b[31mred\x1b[m"},
		{name: "remove non-SGR CSI", text: "a\x1b[Kb", keepSGR: true, want: "ab"},
		{name: "tab after SGR", text: "\x1b[1ma\tb", keepSGR: true, want: "\x1b[1ma   b"},
		{name: "incomplete CSI", text: "a\x1b[31", keepSGR: true, want: "a^[[31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitize(tt.text, tt.keepSGR); got != tt.want {
				t.Errorf("sanitize(%q, %v) = %q, want %q", tt.text, tt.keepSGR, got, tt.want)
			}
		})
	}
}

func TestFormatDiff(t *testing.T) {
	f := diff.MyersDiff("a\n", "b\n")
	f.OldPath, f.NewPath = "a.txt", "b.txt"
	d := diff.Diff{Entries: []diff.Entry{
		{Text: "commit\t1"},
		{File: &f},
		{Text: "end"},
	}}

	var got []string
	for _, r := range formatDiff(d, false) {
		got = append(got, r.String())
	}
	want := []string{
		"commit  1",
		"a.txt → b.txt",
		"@@ -1,1 +1,1 @@",
		"1⋮  -a",
		" ⋮1 +b",
		"end",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("formatDiff()\ngot:  %q\nwant: %q", got, want)
	}
}

func TestWordHighlightsPaired(t *testing.T) {
	lines := []diff.Line{
		{Kind: diff.LineDeleted, Content: "let x = 1"},
		{Kind: diff.LineAdded, Content: "let x = 2"},
	}
	highlights := wordHighlights(lines, contents(lines))
	expect := [][]diff.Range{
		{{Start: 8, End: 9}},
		{{Start: 8, End: 9}},
	}

	if !reflect.DeepEqual(expect, highlights) {
		t.Errorf("expected=%+v, actual=%+v", expect, highlights)
	}
}

func TestWordHighlightsMultipleLines(t *testing.T) {
	lines := []diff.Line{
		{Kind: diff.LineContext, Content: "x"},
		{Kind: diff.LineDeleted, Content: "foo(1)"},
		{Kind: diff.LineDeleted, Content: "bar(1)"},
		{Kind: diff.LineAdded, Content: "foo(2)"},
		{Kind: diff.LineAdded, Content: "bar(2)"},
	}
	highlights := wordHighlights(lines, contents(lines))
	expect := [][]diff.Range{
		nil,
		{{Start: 4, End: 5}},
		{{Start: 4, End: 5}},
		{{Start: 4, End: 5}},
		{{Start: 4, End: 5}},
	}

	if !reflect.DeepEqual(expect, highlights) {
		t.Errorf("expected=%+v, actual=%+v", expect, highlights)
	}
}

func TestWordHighlightsUnequalBlocks(t *testing.T) {
	lines := []diff.Line{
		{Kind: diff.LineDeleted, Content: "let x = 1"},
		{Kind: diff.LineAdded, Content: "let x = 2"},
		{Kind: diff.LineAdded, Content: "let y = 3"},
	}
	highlights := wordHighlights(lines, contents(lines))
	expect := [][]diff.Range{nil, nil, nil}

	if !reflect.DeepEqual(expect, highlights) {
		t.Errorf("expected=%+v, actual=%+v", expect, highlights)
	}
}

func TestWordHighlightsAddedBeforeDeleted(t *testing.T) {
	lines := []diff.Line{
		{Kind: diff.LineAdded, Content: "let x = 2"},
		{Kind: diff.LineDeleted, Content: "let x = 1"},
	}
	highlights := wordHighlights(lines, contents(lines))
	expect := [][]diff.Range{nil, nil}

	if !reflect.DeepEqual(expect, highlights) {
		t.Errorf("expected=%+v, actual=%+v", expect, highlights)
	}
}

func TestFormatFileHighlightsAfterSanitize(t *testing.T) {
	f := diff.File{Hunks: []diff.Hunk{{
		Lines: []diff.Line{
			{Kind: diff.LineDeleted, Content: "\tx = 1", OldNumber: 1},
			{Kind: diff.LineAdded, Content: "\tx = 2", NewNumber: 1},
		},
	}}}

	var highlights [][]diff.Range
	for _, r := range formatFile(f, false) {
		if r.kind == rowLine {
			highlights = append(highlights, r.highlights)
		}
	}
	// The tab is expanded to 4 spaces, so "1" and "2" are at 8, not 5.
	expect := [][]diff.Range{
		{{Start: 8, End: 9}},
		{{Start: 8, End: 9}},
	}

	if !reflect.DeepEqual(expect, highlights) {
		t.Errorf("expected=%+v, actual=%+v", expect, highlights)
	}
}

func contents(lines []diff.Line) []string {
	texts := make([]string, len(lines))
	for i, l := range lines {
		texts[i] = l.Content
	}
	return texts
}

func TestHighlightSegmentsNoHighlights(t *testing.T) {
	segments := highlightSegments("let x = 1;", 0, nil)
	expect := []segment{
		{text: "let x = 1;"},
	}

	if !reflect.DeepEqual(expect, segments) {
		t.Errorf("expected=%+v, actual=%+v", expect, segments)
	}
}

func TestHighlightSegmentsMiddle(t *testing.T) {
	highlights := []diff.Range{{Start: 8, End: 9}}
	segments := highlightSegments("let x = 1;", 0, highlights)
	expect := []segment{
		{text: "let x = "},
		{text: "1", highlighted: true},
		{text: ";"},
	}

	if !reflect.DeepEqual(expect, segments) {
		t.Errorf("expected=%+v, actual=%+v", expect, segments)
	}
}

func TestHighlightSegmentsWholeText(t *testing.T) {
	highlights := []diff.Range{{Start: 0, End: 3}}
	segments := highlightSegments("abc", 0, highlights)
	expect := []segment{
		{text: "abc", highlighted: true},
	}

	if !reflect.DeepEqual(expect, segments) {
		t.Errorf("expected=%+v, actual=%+v", expect, segments)
	}
}

func TestHighlightSegmentsWithOffset(t *testing.T) {
	// "x = 1;" is the soft-wrapped part of "let x = 1;" from 4.
	highlights := []diff.Range{{Start: 8, End: 9}}
	segments := highlightSegments("x = 1;", 4, highlights)
	expect := []segment{
		{text: "x = "},
		{text: "1", highlighted: true},
		{text: ";"},
	}

	if !reflect.DeepEqual(expect, segments) {
		t.Errorf("expected=%+v, actual=%+v", expect, segments)
	}
}

func TestHighlightSegmentsOutsideText(t *testing.T) {
	// "let " is the first part of "let x = 1;", before the highlight.
	highlights := []diff.Range{{Start: 8, End: 9}}
	segments := highlightSegments("let ", 0, highlights)
	expect := []segment{
		{text: "let "},
	}

	if !reflect.DeepEqual(expect, segments) {
		t.Errorf("expected=%+v, actual=%+v", expect, segments)
	}
}

func TestHighlightSegmentsCutBySoftWrap(t *testing.T) {
	// "abcdef" is soft-wrapped into "abc" and "def", and "cde" is highlighted.
	highlights := []diff.Range{{Start: 2, End: 5}}
	first := highlightSegments("abc", 0, highlights)
	second := highlightSegments("def", 3, highlights)
	expectFirst := []segment{
		{text: "ab"},
		{text: "c", highlighted: true},
	}
	expectSecond := []segment{
		{text: "de", highlighted: true},
		{text: "f"},
	}

	if !reflect.DeepEqual(expectFirst, first) {
		t.Errorf("first: expected=%+v, actual=%+v", expectFirst, first)
	}
	if !reflect.DeepEqual(expectSecond, second) {
		t.Errorf("second: expected=%+v, actual=%+v", expectSecond, second)
	}
}
