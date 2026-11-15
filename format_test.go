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
