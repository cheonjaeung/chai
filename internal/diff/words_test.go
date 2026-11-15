package diff

import (
	"reflect"
	"testing"
)

func TestWordDiffChangedWord(t *testing.T) {
	o := "let total = price * count;"
	n := "let total = price * amount;"
	oldRanges, newRanges := WordDiff(o, n)
	expectOld := []Range{{Start: 20, End: 25}}
	expectNew := []Range{{Start: 20, End: 26}}

	testRanges(t, expectOld, expectNew, oldRanges, newRanges)
}

func TestWordDiffMergeAdjacentTokens(t *testing.T) {
	o := "value.x(1)"
	n := "value->y(1)"
	oldRanges, newRanges := WordDiff(o, n)
	expectOld := []Range{{Start: 5, End: 7}}
	expectNew := []Range{{Start: 5, End: 8}}

	testRanges(t, expectOld, expectNew, oldRanges, newRanges)
}

func TestWordDiffInsertionOnly(t *testing.T) {
	o := "f(a)"
	n := "f(a, b)"
	oldRanges, newRanges := WordDiff(o, n)
	var expectOld []Range
	expectNew := []Range{{Start: 3, End: 6}}

	testRanges(t, expectOld, expectNew, oldRanges, newRanges)
}

func TestWordDiffKorean(t *testing.T) {
	o := "안녕 세계"
	n := "안녕 우주"
	oldRanges, newRanges := WordDiff(o, n)
	expectOld := []Range{{Start: 7, End: 13}}
	expectNew := []Range{{Start: 7, End: 13}}

	testRanges(t, expectOld, expectNew, oldRanges, newRanges)
}

func TestWordDiffIdentical(t *testing.T) {
	o := "same"
	n := "same"
	oldRanges, newRanges := WordDiff(o, n)

	testRanges(t, nil, nil, oldRanges, newRanges)
}

func TestWordDiffTooDifferent(t *testing.T) {
	o := "foo(bar)"
	n := "baz = qux"
	oldRanges, newRanges := WordDiff(o, n)

	testRanges(t, nil, nil, oldRanges, newRanges)
}

func TestWordDiffBothEmpty(t *testing.T) {
	o := ""
	n := ""
	oldRanges, newRanges := WordDiff(o, n)

	testRanges(t, nil, nil, oldRanges, newRanges)
}

func TestTokenizeEmpty(t *testing.T) {
	tokens := tokenize("")
	var expect []string

	if !reflect.DeepEqual(expect, tokens) {
		t.Errorf("expected=%q, actual=%q", expect, tokens)
	}
}

func TestTokenizeWordsAndSymbols(t *testing.T) {
	tokens := tokenize("foo_bar1  (x)")
	expect := []string{"foo_bar1", "  ", "(", "x", ")"}

	if !reflect.DeepEqual(expect, tokens) {
		t.Errorf("expected=%q, actual=%q", expect, tokens)
	}
}

func TestTokenizeConsecutiveSymbols(t *testing.T) {
	tokens := tokenize("a->b")
	expect := []string{"a", "-", ">", "b"}

	if !reflect.DeepEqual(expect, tokens) {
		t.Errorf("expected=%q, actual=%q", expect, tokens)
	}
}

func TestTokenizeKorean(t *testing.T) {
	tokens := tokenize("안녕 하세요")
	expect := []string{"안녕", " ", "하세요"}

	if !reflect.DeepEqual(expect, tokens) {
		t.Errorf("expected=%q, actual=%q", expect, tokens)
	}
}

func TestTokenizeCombiningMark(t *testing.T) {
	tokens := tokenize("cafe\u0301!") // "café!"
	expect := []string{"cafe\u0301", "!"}

	if !reflect.DeepEqual(expect, tokens) {
		t.Errorf("expected=%q, actual=%q", expect, tokens)
	}
}

func testRanges(t *testing.T, expectOld []Range, expectNew []Range, oldRanges []Range, newRanges []Range) {
	t.Helper()
	if !reflect.DeepEqual(expectOld, oldRanges) {
		t.Errorf("old: expected=%+v, actual=%+v", expectOld, oldRanges)
	}
	if !reflect.DeepEqual(expectNew, newRanges) {
		t.Errorf("new: expected=%+v, actual=%+v", expectNew, newRanges)
	}
}
