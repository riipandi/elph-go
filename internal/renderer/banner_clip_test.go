package renderer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestViewHeightFitsTerminal(t *testing.T) {
	for _, h := range []int{20, 24, 30, 40} {
		for _, w := range []int{60, 80, 120} {
			m := New()
			m.width = w
			m.height = h
			m.ready = true
			m = m.syncLayout(false)

			if lipgloss.Height(m.View()) > h {
				t.Fatalf("w=%d h=%d view height %d exceeds terminal (chrome=%d vp=%d)",
					w, h, lipgloss.Height(m.View()), m.chromeH, m.content.Height)
			}
		}
	}
}

func TestBannerTopVisibleAtStart(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 24
	m.ready = true
	m = m.syncLayout(false)

	if m.content.YOffset != 0 {
		t.Fatalf("YOffset %d, want 0 so banner starts at top", m.content.YOffset)
	}

	vp := m.content.View()
	if !strings.Contains(vp, "Welcome to") {
		t.Fatalf("banner header not visible in viewport")
	}
}

func TestViewOmitsEmptyActivityLayer(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 24
	m.ready = true
	m = m.syncLayout(false)

	parts := m.viewParts()
	if len(parts) != 3 {
		t.Fatalf("expected 3 view parts without activity, got %d", len(parts))
	}
}