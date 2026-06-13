package renderer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBannerMetadataLineClamp(t *testing.T) {
	m := New()
	m.width = 50
	m.workDir = strings.Repeat("x", 80)

	banner := m.bannerView()
	if lipgloss.Width(banner) > m.width {
		t.Fatalf("banner wider than terminal: banner=%d terminal=%d", lipgloss.Width(banner), m.width)
	}

	if strings.Contains(banner, strings.Repeat("x", 80)) {
		t.Fatal("directory value was not line-clamped")
	}
	if !strings.Contains(banner, "Directory:") {
		t.Fatal("directory metadata line not found")
	}
}

func TestBannerTipWraps(t *testing.T) {
	m := New()
	m.width = 40
	m.tip = strings.Repeat("word ", 30)

	banner := m.bannerView()
	if lipgloss.Height(banner) < 12 {
		t.Fatalf("expected tip to wrap to multiple lines, height=%d", lipgloss.Height(banner))
	}
}

func TestFooterLineClamp(t *testing.T) {
	m := New()
	m.width = 42
	m.modelName = "Claude Sonnet 4.6 Extended Edition"

	footer := m.footerView()
	if lipgloss.Width(footer) > m.width {
		t.Fatalf("footer wider than terminal: footer=%d terminal=%d", lipgloss.Width(footer), m.width)
	}

	lines := strings.Split(strings.TrimSpace(footer), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 footer lines, got %d", len(lines))
	}
	if lipgloss.Width(lines[0]) > footerContentWidth(m.width)+1 {
		t.Fatalf("footer line 1 exceeds content width: %d > %d", lipgloss.Width(lines[0]), footerContentWidth(m.width)+1)
	}
}

func TestFooterRowSpacing(t *testing.T) {
	left := "Claude Sonnet 4.6 | anthropic | T: high | IMG"
	right := "$0.00 | 0.0% (262k)"
	row := footerRow(60, left, right)

	if strings.Contains(row, "IMG$0.00") {
		t.Fatalf("left and right collapsed: %q", row)
	}
	if !strings.HasSuffix(row, "(262k)") {
		t.Fatalf("right segment not flush to edge: %q", row)
	}
	if lipgloss.Width(row) != 60 {
		t.Fatalf("row width %d, want 60", lipgloss.Width(row))
	}
}

func TestFooterRightSegmentFlush(t *testing.T) {
	m := New()
	m.width = 80
	m.workDir = "/Users/dev/renderer"

	lines := strings.Split(strings.TrimSpace(m.footerView()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 footer lines, got %d", len(lines))
	}

	maxLineW := footerContentWidth(m.width) + 1 // includes 1-char left padding per row
	for i, line := range lines {
		if lipgloss.Width(line) > maxLineW {
			t.Fatalf("footer line %d width %d exceeds %d: %q", i+1, lipgloss.Width(line), maxLineW, line)
		}
	}
	if !strings.HasSuffix(lines[0], "(262k)") {
		t.Fatalf("line 1 right segment not flush to edge: %q", lines[0])
	}
	if !strings.HasSuffix(strings.TrimRight(lines[1], " "), "[-]") {
		t.Fatalf("line 2 right segment not flush to edge: %q", lines[1])
	}
	if strings.Contains(lines[1], "buildturn") {
		t.Fatalf("line 2 left/right segments collapsed: %q", lines[1])
	}
}