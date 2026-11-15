package utils

import (
	"slices"
	"testing"
)

func TestSplitLines(t *testing.T) {
	txt := "a\nb\nc"

	split := SplitLines(txt)
	expect := []string{"a\n", "b\n", "c"}

	if !slices.Equal(expect, split) {
		t.Errorf("expected=%q, actual=%q", expect, split)
	}
}

func TestSplitLinesTrailingNewline(t *testing.T) {
	txt := "a\nb\nc\n"

	split := SplitLines(txt)
	expect := []string{"a\n", "b\n", "c\n"}

	if !slices.Equal(expect, split) {
		t.Errorf("expected=%q, actual=%q", expect, split)
	}
}

func TestSplitLinesEmpty(t *testing.T) {
	txt := ""

	split := SplitLines(txt)
	expect := []string{}

	if !slices.Equal(expect, split) {
		t.Errorf("expected=%q, actual=%q", expect, split)
	}
}
