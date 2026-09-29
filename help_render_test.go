package main

import (
	"strings"
	"testing"
)

func TestKeyFooterWraps(t *testing.T) {
	initColors()
	initStyles()
	h := newHelp()
	got := renderKeyFooter(h, newKeyMap().ShortHelp(), 36)
	if strings.Count(got, "\n") < 1 {
		t.Fatalf("expected wrapped footer, got %q", got)
	}
	for _, want := range []string{"enter", "toggle", "import", "?", "quit"} {
		if !strings.Contains(got, want) {
			t.Errorf("wrapped footer missing %q in %q", want, got)
		}
	}
}

func TestShortHelpContents(t *testing.T) {
	initColors()
	initStyles()
	h := newHelp()
	h.SetWidth(200)
	view := h.ShortHelpView(newKeyMap().ShortHelp())
	for _, want := range []string{"enter", "toggle", "import", "?"} {
		if !strings.Contains(view, want) {
			t.Errorf("short help missing %q in %q", want, view)
		}
	}
}
