// Package ansi provides ANSI escape sequences helpers.
package ansi

import "strings"

// CsiLen returns the length of the CSI sequence, such as "\x1b[31m".
// It returns 0 if the given string does not start with CSI sequence.
func CsiLen(str string) int {
	if !strings.HasPrefix(str, "\x1b[") {
		return 0
	}
	for i := 2; i < len(str); i++ {
		switch c := str[i]; {
		case c >= 0x40 && c <= 0x7e:
			return i + 1
		case c < 0x20 || c > 0x3f:
			return 0
		}
	}
	return 0
}

// Strip removes CSI sequences from the given string.
func Strip(str string) string {
	if !strings.Contains(str, "\x1b[") {
		return str
	}
	var sb strings.Builder
	for i := 0; i < len(str); i++ {
		if n := CsiLen(str[i:]); n > 0 {
			i += n - 1
			continue
		}
		sb.WriteByte(str[i])
	}
	return sb.String()
}
