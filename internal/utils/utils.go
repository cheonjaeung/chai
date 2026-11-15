// Package utils provides common utilities for other packages.
package utils

import "strings"

// SplitLines separates text into array by newline character.
//
// Each line keeps its trailing newline character. But if a line has no
// trailing newline, the line will have no trailing newline.
//
// Unlike strings.SplitAfter, it doesn't return empty last element when the
// text ends with a newline character.
func SplitLines(text string) []string {
	lines := strings.SplitAfter(text, "\n")
	// Drop the last element.
	// strings.SplitAfter returns one more element if the text ends with '\n'.
	// In this case, the last element is not actual content.
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
