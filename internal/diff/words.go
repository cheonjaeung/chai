package diff

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Range is a byte range of a text line. Start is inclusive and End is exclusive.
type Range struct {
	Start int
	End   int
}

// WordDiff compares two string lines word by word and returns the changed
// byte range of each line.
//
// It returns no ranges if less than half of the lines are the same.
func WordDiff(oldText string, newText string) (oldRanges []Range, newRanges []Range) {
	a := tokenize(oldText)
	b := tokenize(newText)
	aStarts := tokenStarts(a)
	bStarts := tokenStarts(b)

	same := 0
	for _, e := range editScript(a, b) {
		switch e.kind {
		case LineContext:
			same += len(a[e.oldIndex])
		case LineDeleted:
			oldRanges = addRange(oldRanges, aStarts[e.oldIndex], aStarts[e.oldIndex+1])
		case LineAdded:
			newRanges = addRange(newRanges, bStarts[e.newIndex], bStarts[e.newIndex+1])
		}
	}

	// Same as checking similarity, 2*same / (len(oldText)+len(newText)), is less than 0.5.
	if 4*same < len(oldText)+len(newText) {
		return nil, nil
	}
	return oldRanges, newRanges
}

// tokenize splits text line into tokens.
//
// Token is a base unit of the word diff. It can be a single word, whitespaces,
// or a special character.
func tokenize(text string) []string {
	var tokens []string
	for text != "" {
		r, end := utf8.DecodeRuneInString(text)
		switch {
		case isUnicodeWord(r):
			end = len(text) - len(strings.TrimLeftFunc(text, isUnicodeWord))
		case unicode.IsSpace(r):
			end = len(text) - len(strings.TrimLeftFunc(text, unicode.IsSpace))
		}
		tokens = append(tokens, text[:end])
		text = text[end:]
	}
	return tokens
}

// isUnicodeWord returns true if the given rune is a letter, digit or a mark in
// the unicode, or an underscore.
func isUnicodeWord(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || r == '_'
}

// tokenStarts returns the start offset of each token.
//
// The length of the returned offsets is len(tokens)+1. The offset at i is the
// start of the i-th token, so the first offset is always 0, and the last offset
// is the end of the last token.
func tokenStarts(tokens []string) []int {
	starts := make([]int, len(tokens)+1)
	starts[0] = 0
	for i, t := range tokens {
		starts[i+1] = starts[i] + len(t)
	}
	return starts
}

// addRange appends a Range to ranges.
//
// If the new range is right next to the last range, merging it into the last item.
func addRange(ranges []Range, start int, end int) []Range {
	if len(ranges) > 0 && ranges[len(ranges)-1].End == start {
		ranges[len(ranges)-1].End = end
		return ranges
	}
	return append(ranges, Range{Start: start, End: end})
}
