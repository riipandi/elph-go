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
	if m.scrollBarView() != "" {
		t.Fatal("scrollbar should be hidden when content fits")
	}
	if m.contentAreaView() != m.content.View() {
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
	if m.content.Width != m.width-scrollBarWidth {
		t.Fatalf("viewport width %d, want %d", m.content.Width, m.width-scrollBarWidth)
	}

	bar := m.scrollBarView()
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
	topBar := m.scrollBarView()
	topOffset := m.content.YOffset

	m.content.GotoBottom()
	bottomBar := m.scrollBarView()
	bottomOffset := m.content.YOffset

	if topOffset >= bottomOffset {
		t.Fatalf("expected scroll offset to increase: top=%d bottom=%d", topOffset, bottomOffset)
	}
	if topBar == bottomBar {
		t.Fatal("scrollbar thumb should move when scrolled")
	}
}