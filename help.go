package main

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Toggle    key.Binding
	Import    key.Binding
	Rename    key.Binding
	Delete    key.Binding
	Help      key.Binding
	Quit      key.Binding
	ForceQuit key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "move down"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "toggle"),
		),
		Import: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "import"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename config"),
		),
		Delete: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete config"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.Import, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Toggle},
		{k.Import, k.Rename, k.Delete},
		{k.Help, k.Quit, k.ForceQuit},
	}
}

func renderKeyFooter(h help.Model, bindings []key.Binding, width int) string {
	sep := h.Styles.ShortSeparator.Inline(true).Render(h.ShortSeparator)
	sepW := lipgloss.Width(sep)
	var lines []string
	var cur string
	var curW int
	for _, kb := range bindings {
		if !kb.Enabled() {
			continue
		}
		item := h.Styles.ShortKey.Inline(true).Render(kb.Help().Key) + " " +
			h.Styles.ShortDesc.Inline(true).Render(kb.Help().Desc)
		w := lipgloss.Width(item)
		need := w
		if curW > 0 {
			need += sepW
		}
		if width > 0 && curW > 0 && curW+need > width {
			lines = append(lines, cur)
			cur, curW = item, w
			continue
		}
		if curW > 0 {
			cur += sep
			curW += sepW
		}
		cur += item
		curW += w
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return strings.Join(lines, "\n")
}

func newHelp() help.Model {
	h := help.New()
	s := help.DefaultDarkStyles()
	s.ShortKey = lipgloss.NewStyle().Foreground(accent)
	s.ShortDesc = lipgloss.NewStyle().Foreground(dimCol)
	s.ShortSeparator = lipgloss.NewStyle().Foreground(borderCol)
	s.FullKey = lipgloss.NewStyle().Foreground(accent)
	s.FullDesc = lipgloss.NewStyle().Foreground(textCol)
	s.FullSeparator = lipgloss.NewStyle().Foreground(borderCol)
	h.Styles = s
	return h
}
