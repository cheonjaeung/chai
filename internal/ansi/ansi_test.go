package ansi

import "testing"

func TestCsiLen(t *testing.T) {
	testSet := []struct {
		str    string
		csiLen int
	}{
		{"abc", 0},
		{"\x1b[m", 3},
		{"\x1b", 0},
		{"\x1b[1;3mabc", 6},
	}

	for _, s := range testSet {
		if actual := CsiLen(s.str); actual != s.csiLen {
			t.Errorf("%q: expected=%d, actual=%d", s.str, s.csiLen, actual)
		}
	}
}

func TestStrip(t *testing.T) {
	testSet := []struct {
		str      string
		stripped string
	}{
		{"abc\n", "abc\n"},
		{"\x1b[31m-a\x1b[m\n", "-a\n"},
		{"\x1b[1m\x1b[31mH\x1b[0mi", "Hi"},
		{"a\x1b[Kb", "ab"},
		{"a\x1b[31", "a\x1b[31"},
	}

	for _, s := range testSet {
		if actual := Strip(s.str); actual != s.stripped {
			t.Errorf("%q: expected=%q, actual=%q", s.str, s.stripped, actual)
		}
	}
}
