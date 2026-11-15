package main

import (
	"strings"

	"github.com/cheonjaeung/chai/internal/ansi"
	"github.com/cheonjaeung/chai/internal/diff"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

var (
	addedBackground   = tcell.NewRGBColor(78, 91, 87)
	deletedBackground = tcell.NewRGBColor(94, 72, 86)

	addedNumberStyle   = tcell.StyleDefault.Foreground(tcell.PaletteColor(2))
	deletedNumberStyle = tcell.StyleDefault.Foreground(tcell.PaletteColor(1))
	contextNumberStyle = tcell.StyleDefault.Foreground(tcell.PaletteColor(246))
	separatorStyle     = tcell.StyleDefault.Foreground(tcell.PaletteColor(244))
	hunkHeaderStyle    = tcell.StyleDefault.Foreground(tcell.PaletteColor(6))
	fileHeaderStyle    = tcell.StyleDefault.Bold(true)
)

// Viewer is an interactive diff viewer in the terminal.
type Viewer struct {
	// screen is the terminal screen the viewer draws on.
	screen tcell.Screen
	// top is the index in screenRows of the first visible line.
	top int
	// rows are logical lines to display.
	rows []row
	// screenRows are rows wrapped to the screen width.
	screenRows []screenRow
	// border is true when files are drawn in boxes.
	border bool
}

// screenRow represents a single line of content in the screen.
type screenRow struct {
	// row is the index in rows that this line belongs to.
	row int
	// text is the content of this line.
	text string
	// softWrap is true when this line continues the previous one.
	softWrap bool
	// style is the style at the start of text, used only by rowText.
	style tcell.Style
	// border is true when this line is a horizontal border of a file box.
	border bool
}

// NewViewer returns a new interactive viewer.
func NewViewer(d diff.Diff) (*Viewer, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	viewer := &Viewer{screen: screen, rows: formatDiff(d, false), border: true}
	viewer.layout()
	return viewer, nil
}

// Loop draws the viewer and handles events until the viewer finished.
func (v *Viewer) Loop() error {
	defer v.screen.Fini()
	for {
		v.draw()
		if !v.handle(v.screen.PollEvent()) {
			return nil
		}
	}
}

// handle processes ev and reports whether the viewer should keep running.
func (v *Viewer) handle(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case nil:
		return false
	case *tcell.EventResize:
		v.layout()
		v.screen.Sync()
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlC:
			return false
		case tcell.KeyUp:
			v.scroll(-1)
		case tcell.KeyDown:
			v.scroll(1)
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q':
				return false
			}
		}
	}
	return true
}

// layout measures and lays out rows into screenRows for the screen width.
func (v *Viewer) layout() {
	anchor := 0

	// Keep current scroll position
	if v.top < len(v.screenRows) {
		anchor = v.screenRows[v.top].row
	}

	width, _ := v.screen.Size()
	v.screenRows = nil
	for i, r := range v.rows {
		boxed := v.border && r.kind != rowText
		if boxed && r.kind == rowFileHeader {
			v.screenRows = append(v.screenRows, screenRow{row: i, text: borderLine("╭", "╮", width), border: true})
		} else if boxed && v.rows[i-1].kind == rowFileHeader {
			v.screenRows = append(v.screenRows, screenRow{row: i, text: borderLine("├", "┤", width), border: true})
		}

		contentWidth := width
		if boxed {
			contentWidth -= 4
		}
		if r.hasPrefix() {
			contentWidth -= uniseg.StringWidth(r.prefix())
		}
		style := tcell.StyleDefault
		for j, text := range wrap(r.text, contentWidth) {
			v.screenRows = append(v.screenRows, screenRow{row: i, text: text, softWrap: j > 0, style: style})
			style = endStyle(text, style)
		}

		if boxed && (i+1 == len(v.rows) || v.rows[i+1].kind == rowText || v.rows[i+1].kind == rowFileHeader) {
			v.screenRows = append(v.screenRows, screenRow{row: i, text: borderLine("╰", "╯", width), border: true})
		}
	}

	v.top = 0
	for v.top < len(v.screenRows) && v.screenRows[v.top].row < anchor {
		v.top++
	}
	v.scroll(0)
}

// draw renders the screenRows to the screen.
func (v *Viewer) draw() {
	v.screen.Clear()
	width, height := v.screen.Size()
	for y := 0; y < height && v.top+y < len(v.screenRows); y++ {
		sr := v.screenRows[v.top+y]
		if sr.border {
			v.put(0, y, sr.text, separatorStyle)
			continue
		}

		r := v.rows[sr.row]
		x, right := 0, width
		if v.border && r.kind != rowText {
			v.put(0, y, "│", separatorStyle)
			v.put(width-1, y, "│", separatorStyle)
			x, right = 2, width-2
		}

		switch r.kind {
		case rowLine, rowNoNewline:
			v.drawLine(x, y, right, r, sr)
		case rowFileHeader:
			v.screen.PutStrStyled(x, y, sr.text, fileHeaderStyle)
		case rowHunkHeader:
			v.screen.PutStrStyled(x, y, sr.text, hunkHeaderStyle)
		case rowText:
			v.putAnsi(x, y, sr.text, sr.style)
		default:
			v.screen.PutStrStyled(x, y, sr.text, tcell.StyleDefault)
		}
	}
	v.screen.Show()
}

// scroll moves top position by the given screen row amount, clamping it
// within the scrollable range.
func (v *Viewer) scroll(n int) {
	_, height := v.screen.Size()
	v.top = max(min(v.top+n, len(v.screenRows)-height), 0)
}

// lineStyles returns the styles of the line numbers and the line content of r.
func lineStyles(r row) (numberStyle, lineStyle tcell.Style) {
	numberStyle = contextNumberStyle
	background := tcell.ColorDefault
	if r.kind == rowLine {
		switch r.lineKind {
		case diff.LineAdded:
			numberStyle = addedNumberStyle
			background = addedBackground
		case diff.LineDeleted:
			numberStyle = deletedNumberStyle
			background = deletedBackground
		}
	}
	return numberStyle, tcell.StyleDefault.Background(background)
}

// drawLine renders a sr to the screen from x with line numbers and diff marker,
// filling the background until right.
// If the sr is longer than screen width, it soft-wraps the row with blank
// line number and a soft-wrap marker.
func (v *Viewer) drawLine(x, y, right int, r row, sr screenRow) {
	numberStyle, lineStyle := lineStyles(r)

	oldNumber, newNumber := r.oldNumber, r.newNumber
	marker, markerStyle := r.marker, lineStyle
	if sr.softWrap {
		oldNumber = strings.Repeat(" ", len(oldNumber))
		newNumber = strings.Repeat(" ", len(newNumber))
		_, background, _ := lineStyle.Decompose()
		marker, markerStyle = "↪", separatorStyle.Background(background)
	}

	x = v.put(x, y, oldNumber, numberStyle)
	x = v.put(x, y, "⋮", separatorStyle)
	x = v.put(x, y, newNumber, numberStyle)
	x = v.put(x, y, " ", tcell.StyleDefault)
	x = v.put(x, y, marker, markerStyle)
	x = v.put(x, y, sr.text, lineStyle)
	for ; x < right; x++ {
		v.screen.SetContent(x, y, ' ', nil, lineStyle)
	}
}

// borderLine returns a horizontal border line of width, between left and right corners.
func borderLine(left, right string, width int) string {
	return left + strings.Repeat("─", max(width-2, 0)) + right
}

// put draws s at (x, y) and returns the column after it.
func (v *Viewer) put(x, y int, s string, style tcell.Style) int {
	v.screen.PutStrStyled(x, y, s, style)
	return x + uniseg.StringWidth(s)
}

// wrap splits text into multiple lines if the text width is larger than width,
// without breaking a grapheme cluster. If the text is smaller than width, it
// returns text as is.
func wrap(text string, width int) []string {
	width = max(width, 1)
	var parts []string
	start, lineWidth := 0, 0
	rest := text
	state := -1
	for rest != "" {
		if n := ansi.CsiLen(rest); n > 0 {
			rest = rest[n:]
			state = -1
			continue
		}
		var cluster string
		var w int
		cluster, rest, w, state = uniseg.FirstGraphemeClusterInString(rest, state)
		pos := len(text) - len(rest) - len(cluster)
		if lineWidth+w > width && pos > start {
			parts = append(parts, text[start:pos])
			start, lineWidth = pos, 0
		}
		lineWidth += w
	}
	return append(parts, text[start:])
}
