package renderer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/riipandi/elph/internal/constants"
)

func TestScrollBarHiddenWhenContentFits(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 40
	m.ready = true
	m = m.syncLayout(false)

	if m.contentScrollable() {
		t.Fatal("banner-only content should not be scrollable in tall terminal")
	}
	if m.contentScrollBarView() != "" {
		t.Fatal("scrollbar should be hidden when content fits")
	}
	if m.contentAreaView() != lipgloss.NewStyle().Width(m.width).MaxWidth(m.width).Render(m.content.View()) {
		t.Fatal("content area should equal viewport without gutter")
	}
}

func TestScrollBarVisibleWhenOverflow(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 16
	m.ready = true

	for i := range 30 {
		m.messages = append(m.messages, message{
			text: fmt.Sprintf("message line %d with some extra text to wrap nicely", i),
			kind: constants.MessageUser,
		})
	}
	m.contentDirty = true
	m = m.syncLayout(false)

	if !m.contentScrollable() {
		t.Fatal("expected scrollable content")
	}
	if m.contentAreaWidth() != m.width-scrollBarWidth {
		t.Fatalf("contentAreaWidth %d, want %d", m.contentAreaWidth(), m.width-scrollBarWidth)
	}
	if m.content.Width != m.contentAreaWidth() {
		t.Fatalf("viewport width %d != contentAreaWidth %d", m.content.Width, m.contentAreaWidth())
	}
	if lipgloss.Width(m.contentAreaView()) > m.width {
		t.Fatalf("content area wider than terminal: %d > %d", lipgloss.Width(m.contentAreaView()), m.width)
	}

	bar := m.contentScrollBarView()
	if lipgloss.Height(bar) != m.content.Height {
		t.Fatalf("scrollbar height %d, want %d", lipgloss.Height(bar), m.content.Height)
	}
	if !strings.Contains(bar, "█") {
		t.Fatal("scrollbar should contain thumb")
	}
	if !strings.Contains(m.contentAreaView(), "█") {
		t.Fatal("content area should include scrollbar")
	}
}

func TestScrollBarThumbMovesDown(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 16
	m.ready = true
	for i := range 40 {
		m.messages = append(m.messages, message{text: fmt.Sprintf("msg %d", i), kind: constants.MessageUser})
	}
	m.contentDirty = true
	m = m.syncLayout(false)

	m.content.GotoTop()
	topBar := m.contentScrollBarView()
	topOffset := m.content.YOffset

	m.content.GotoBottom()
	bottomBar := m.contentScrollBarView()
	bottomOffset := m.content.YOffset

	if topOffset >= bottomOffset {
		t.Fatalf("expected scroll offset to increase: top=%d bottom=%d", topOffset, bottomOffset)
	}
	if topBar == bottomBar {
		t.Fatal("scrollbar thumb should move when scrolled")
	}
}

func TestContentAreaWidthMatchesChromeWhenScrollable(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 14
	m.ready = true
	for i := range 25 {
		m.messages = append(m.messages, message{text: fmt.Sprintf("overflow %d", i), kind: constants.MessageAI})
	}
	m.contentDirty = true
	m = m.syncLayout(false)

	if lipgloss.Width(m.contentAreaView()) > m.width {
		t.Fatalf("content area %d exceeds terminal %d", lipgloss.Width(m.contentAreaView()), m.width)
	}
	if m.chromeOuterWidth() != m.content.Width {
		t.Fatalf("chrome width %d != viewport width %d", m.chromeOuterWidth(), m.content.Width)
	}
}

func TestInputScrollBarVisibleWhenOverflow(t *testing.T) {
	m := testInputModel(t)
	lines := make([]string, maxInputLines+2)
	for i := range lines {
		lines[i] = fmt.Sprintf("input line %d", i+1)
	}
	m.input.SetValue(strings.Join(lines, "\n"))
	m = m.syncInputWidth()

	if !m.inputScrollable() {
		t.Fatal("expected scrollable input")
	}
	if m.inputScrollBarView() == "" {
		t.Fatal("input scrollbar should be visible")
	}
	if !strings.Contains(m.inputView(), "█") {
		t.Fatal("input view should include scrollbar thumb")
	}
}

func TestInputScrollBarHiddenWhenFits(t *testing.T) {
	m := testInputModel(t)
	m.input.SetValue("short")
	m = m.syncInputWidth()

	if m.inputScrollBarView() != "" {
		t.Fatal("input scrollbar should be hidden for short text")
	}
}