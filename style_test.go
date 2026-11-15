package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestApplySgr(t *testing.T) {
	base := tcell.StyleDefault.Bold(true).Foreground(tcell.PaletteColor(1))
	tests := []struct {
		name   string
		params string
		want   tcell.Style
	}{
		{name: "empty resets", params: "", want: tcell.StyleDefault},
		{name: "reset", params: "0", want: tcell.StyleDefault},
		{name: "bold off", params: "22", want: tcell.StyleDefault.Foreground(tcell.PaletteColor(1))},
		{name: "foreground", params: "32", want: base.Foreground(tcell.PaletteColor(2))},
		{name: "bright background", params: "101", want: base.Background(tcell.PaletteColor(9))},
		{name: "default foreground", params: "39", want: tcell.StyleDefault.Bold(true)},
		{name: "256 color", params: "38;5;208", want: base.Foreground(tcell.PaletteColor(208))},
		{name: "rgb color", params: "48;2;1;2;3", want: base.Background(tcell.NewRGBColor(1, 2, 3))},
		{name: "multiple", params: "0;4;34", want: tcell.StyleDefault.Underline(true).Foreground(tcell.PaletteColor(4))},
		{name: "after extended color", params: "38;5;208;3", want: base.Foreground(tcell.PaletteColor(208)).Italic(true)},
		{name: "invalid extended color", params: "38;9;1", want: base},
		{name: "unknown", params: "1:2;x", want: base},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applySgr(base, tt.params); got != tt.want {
				t.Errorf("applySGR(%q) = %v, want %v", tt.params, got, tt.want)
			}
		})
	}
}
