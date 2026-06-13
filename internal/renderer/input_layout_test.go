package renderer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/riipandi/elph/internal/constants"
)

func TestInputScrollbarOverlaysLastColumn(t *testing.T) {
	m := testInputModel(t)
	lines := make([]string, maxInputLines+2)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	m.input.SetValue(strings.Join(lines, "\n"))
	m = m.syncInputWidth()

	if !m.inputScrollable() {
		t.Fatal("expected scrollable input")
	}

	body := m.input.View()
	inner := m.inputBodyView()
	if lipgloss.Width(inner) != lipgloss.Width(body) {
		t.Fatalf("overlay changed width: body=%d inner=%d", lipgloss.Width(body), lipgloss.Width(inner))
	}
	if !strings.Contains(inner, "█") && !strings.Contains(inner, "░") {
		t.Fatal("overlay should include scrollbar glyphs")
	}
	sideBySide := lipgloss.JoinHorizontal(lipgloss.Top, body, m.inputScrollBarView())
	if lipgloss.Width(sideBySide) == lipgloss.Width(inner) {
		t.Fatalf("side-by-side layout should be wider than overlay: side=%d overlay=%d", lipgloss.Width(sideBySide), lipgloss.Width(inner))
	}
}

func TestInputScrollbarSitsFlushToRightPadding(t *testing.T) {
	m := testInputModel(t)
	lines := make([]string, maxInputLines+2)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	m.input.SetValue(strings.Join(lines, "\n"))
	m = m.syncInputWidth()

	inner := m.inputBodyView()
	prefixW := 0
	if m.showPromptPrefix {
		prefix := lipgloss.NewStyle().Foreground(constants.White).Bold(true).Render(m.promptChar + " ")
		prefixW = lipgloss.Width(prefix)
	}
	wantInnerW := inputContentWidth(m.chromeOuterWidth()) - prefixW
	if got := lipgloss.Width(inner); got != wantInnerW {
		t.Fatalf("inner width %d, want %d", got, wantInnerW)
	}

	boxW := borderedChromeWidth(m.chromeOuterWidth())
	border := cachedInputBorder(m.mode)
	rendered := border.Width(boxW).Render(inner)
	if got := lipgloss.Width(rendered); got != m.chromeOuterWidth() {
		t.Fatalf("input outer width %d, want %d", got, m.chromeOuterWidth())
	}
}

func TestOverlayInputScrollBarShortLine(t *testing.T) {
	body := "hi" + strings.Repeat(" ", 10)
	bar := scrollBarFor(1, 2, 0)
	got := overlayInputScrollBar(body, bar)
	if lipgloss.Width(got) != lipgloss.Width(body) {
		t.Fatalf("width %d, want %d", lipgloss.Width(got), lipgloss.Width(body))
	}
	if !strings.HasSuffix(stripANSI(got), "░") && !strings.HasSuffix(stripANSI(got), "█") {
		t.Fatalf("last column should be scrollbar: %q", stripANSI(got))
	}
}