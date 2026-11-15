package main

import (
	"strconv"
	"strings"

	"github.com/cheonjaeung/chai/internal/ansi"
	"github.com/gdamore/tcell/v2"
)

// putAnsi draws s at (x, y), starting with style and changing it at each SGR sequence in s.
func (v *Viewer) putAnsi(x, y int, s string, style tcell.Style) {
	start := 0
	for i := 0; i < len(s); i++ {
		if n := ansi.CsiLen(s[i:]); n > 0 {
			x = v.put(x, y, s[start:i], style)
			style = applySgr(style, s[i+2:i+n-1])
			i += n - 1
			start = i + 1
		}
	}
	v.put(x, y, s[start:], style)
}

// endStyle returns style after applying all SGR sequences in s.
func endStyle(s string, style tcell.Style) tcell.Style {
	for i := 0; i < len(s); i++ {
		if n := ansi.CsiLen(s[i:]); n > 0 {
			style = applySgr(style, s[i+2:i+n-1])
			i += n - 1
		}
	}
	return style
}

// applySgr returns style updated by SGR parameters, such as "1;31" of "\x1b[1;31m".
// Unknown parameters are ignored.
func applySgr(style tcell.Style, params string) tcell.Style {
	codes := strings.Split(params, ";")
	for i := 0; i < len(codes); i++ {
		code := 0
		if codes[i] != "" {
			var err error
			if code, err = strconv.Atoi(codes[i]); err != nil {
				continue
			}
		}
		switch {
		case code == 0:
			style = tcell.StyleDefault
		case code == 1:
			style = style.Bold(true)
		case code == 2:
			style = style.Dim(true)
		case code == 3:
			style = style.Italic(true)
		case code == 4:
			style = style.Underline(true)
		case code == 5:
			style = style.Blink(true)
		case code == 7:
			style = style.Reverse(true)
		case code == 9:
			style = style.StrikeThrough(true)
		case code == 22:
			style = style.Bold(false).Dim(false)
		case code == 23:
			style = style.Italic(false)
		case code == 24:
			style = style.Underline(false)
		case code == 25:
			style = style.Blink(false)
		case code == 27:
			style = style.Reverse(false)
		case code == 29:
			style = style.StrikeThrough(false)
		case code >= 30 && code <= 37:
			style = style.Foreground(tcell.PaletteColor(code - 30))
		case code == 38:
			color, n, ok := extendedColor(codes[i+1:])
			if ok {
				style = style.Foreground(color)
			}
			i += n
		case code == 39:
			style = style.Foreground(tcell.ColorDefault)
		case code >= 40 && code <= 47:
			style = style.Background(tcell.PaletteColor(code - 40))
		case code == 48:
			color, n, ok := extendedColor(codes[i+1:])
			if ok {
				style = style.Background(color)
			}
			i += n
		case code == 49:
			style = style.Background(tcell.ColorDefault)
		case code >= 90 && code <= 97:
			style = style.Foreground(tcell.PaletteColor(code - 90 + 8))
		case code >= 100 && code <= 107:
			style = style.Background(tcell.PaletteColor(code - 100 + 8))
		}
	}
	return style
}

// extendedColor parses the color codes after 38 or 48, such as "5;n" or "2;r;g;b".
// It returns the color and the number of codes it used.
func extendedColor(codes []string) (color tcell.Color, n int, ok bool) {
	if len(codes) >= 2 && codes[0] == "5" {
		index, err := strconv.Atoi(codes[1])
		return tcell.PaletteColor(index), 2, err == nil && index >= 0 && index <= 255
	}
	if len(codes) >= 4 && codes[0] == "2" {
		var rgb [3]int
		for k := range rgb {
			c, err := strconv.Atoi(codes[k+1])
			if err != nil || c < 0 || c > 255 {
				return tcell.ColorDefault, 4, false
			}
			rgb[k] = c
		}
		return tcell.NewRGBColor(int32(rgb[0]), int32(rgb[1]), int32(rgb[2])), 4, true
	}
	// Skips the remaining codes because their meaning is unknown.
	return tcell.ColorDefault, len(codes), false
}

// sgr returns the SGR sequence of style, such as "\x1b[0;1;38;5;6m".
// It starts with a reset, so the previous style does not remain.
func sgr(style tcell.Style) string {
	fg, bg, attrs := style.Decompose()
	codes := []string{"0"}
	if attrs&tcell.AttrBold != 0 {
		codes = append(codes, "1")
	}
	if attrs&tcell.AttrDim != 0 {
		codes = append(codes, "2")
	}
	codes = append(codes, sgrColor(fg, "38")...)
	codes = append(codes, sgrColor(bg, "48")...)
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// sgrColor returns the SGR codes of color, with the code of 38 or 48.
// It returns nil if color is the default color.
func sgrColor(color tcell.Color, code string) []string {
	if !color.Valid() {
		return nil
	}
	if color.IsRGB() {
		r, g, b := color.RGB()
		return []string{code, "2", strconv.Itoa(int(r)), strconv.Itoa(int(g)), strconv.Itoa(int(b))}
	}
	return []string{code, "5", strconv.Itoa(int(color &^ tcell.ColorValid))}
}
