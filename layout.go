package main

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	pagePadX      = 1
	paneGap       = 2
	stackGap      = 1
	boxPadY       = 1
	kvLabelW      = 11
	kvGap         = 2
	minSplitWidth = 64
	minStackPane  = 6
	minDetailsH   = 8
)

func useStackedLayout(width, height int) bool {
	return width < minSplitWidth || height > width
}

func boxHeightFor(contentLines int) int {
	return max(3, contentLines+boxPadY+1+2)
}

func stackPaneHeights(avail, tunnelContent int) (tunnelsH, detailsH int) {
	inner := avail - stackGap
	if inner < minStackPane*2 {
		top := max(3, inner/2)
		bot := max(3, inner-top)
		return top, bot
	}
	want := boxHeightFor(tunnelContent)
	if want < minStackPane {
		want = minStackPane
	}
	maxTop := inner - minDetailsH
	if maxTop < minStackPane {
		maxTop = inner / 2
	}
	if want > maxTop {
		want = maxTop
	}
	return want, inner - want
}

func pageInnerWidth(termWidth int) int {
	if termWidth < pagePadX*2+24 {
		return termWidth
	}
	return termWidth - pagePadX*2
}

func boxPadXFor(innerWidth int) int {
	switch {
	case innerWidth >= 36:
		return 3
	case innerWidth >= 16:
		return 2
	case innerWidth >= 8:
		return 1
	default:
		return 0
	}
}

func boxInnerWidth(width int) int {
	inW := max(1, width-2)
	return max(1, inW-boxPadXFor(inW)*2)
}

func fitCell(s string, width int) string {
	s = truncate(s, width)
	n := width - lipgloss.Width(s)
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)+"…") > width {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

func padStyled(s string, width int) string {
	n := width - lipgloss.Width(s)
	if n > 0 {
		return s + strings.Repeat(" ", n)
	}
	if n < 0 {
		return truncate(s, width)
	}
	return s
}

func indentBlock(s string, pad int) string {
	if pad <= 0 || s == "" {
		return s
	}
	prefix := strings.Repeat(" ", pad)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

// titledBox draws a rounded pane of exactly width x height.
// Title sits in the top edge. Content is padded; we never set
// lipgloss Width on a bordered style.
func titledBox(title string, lines []string, width, height int) string {
	inW := max(1, width-2)
	inH := max(1, height-2)
	padX := boxPadXFor(inW)
	contentW := max(1, inW-padX*2)
	side := strings.Repeat(" ", padX)

	b := lipgloss.RoundedBorder()
	edge := headerRuleStyle

	label := "─ " + truncate(title, max(1, inW-4)) + " "
	tw := lipgloss.Width(label)
	if tw > inW {
		label = truncate(strings.TrimSpace(title), inW)
		tw = lipgloss.Width(label)
	}
	rest := max(0, inW-tw)
	top := edge.Render(b.TopLeft) + edge.Render(label) + edge.Render(strings.Repeat(b.Top, rest)+b.TopRight)

	body := make([]string, inH)
	for i := range body {
		cell := strings.Repeat(" ", inW)
		src := i - boxPadY
		if src >= 0 && src < len(lines) {
			cell = side + padStyled(lines[src], contentW) + side
		}
		body[i] = edge.Render(b.Left) + cell + edge.Render(b.Right)
	}
	bot := edge.Render(b.BottomLeft + strings.Repeat(b.Bottom, inW) + b.BottomRight)

	out := make([]string, 0, inH+2)
	out = append(out, top)
	out = append(out, body...)
	out = append(out, bot)
	return strings.Join(out, "\n")
}

func formatKVRow(label, value string, width int) string {
	if width < kvLabelW+kvGap+1 {
		return fitCell(label+" "+dash(value), width)
	}
	return dimStyle.Render(fitCell(label, kvLabelW)) +
		strings.Repeat(" ", kvGap) +
		valueStyle.Render(fitCell(dash(value), width-kvLabelW-kvGap))
}
