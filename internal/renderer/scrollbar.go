package renderer

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/riipandi/elph/internal/constants"
)

const scrollBarWidth = 1

var (
	scrollTrackStyle = lipgloss.NewStyle().Foreground(constants.DimText)
	scrollThumbStyle = lipgloss.NewStyle().Foreground(constants.Gray)
)

func (m Model) contentScrollable() bool {
	if !m.ready || m.content.Height <= 0 {
		return false
	}
	return m.content.TotalLineCount() > m.content.Height
}

func (m Model) contentAreaWidth() int {
	if m.content.Width > 0 {
		return m.content.Width
	}
	return m.width
}

// chromeOuterWidth is the edge-to-edge width shared by the banner, stream
// messages, and input box.
func (m Model) chromeOuterWidth() int {
	return max(m.contentAreaWidth(), 1)
}

// borderedChromeWidth is the lipgloss Width for rounded-border chrome whose
// rendered outer size equals chromeOuterWidth.
func borderedChromeWidth(outer int) int {
	return max(outer-2, 1)
}

func (m Model) contentAreaView() string {
	vp := m.content.View()
	if !m.contentScrollable() {
		return vp
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, vp, m.scrollBarView())
}

func (m Model) scrollBarView() string {
	h := m.content.Height
	if h <= 0 || !m.contentScrollable() {
		return ""
	}

	total := m.content.TotalLineCount()
	if total <= h {
		return ""
	}

	thumbH := max(1, (h*h)/total)
	if thumbH > h {
		thumbH = h
	}

	thumbStart := int(m.content.ScrollPercent() * float64(h-thumbH))
	if thumbStart < 0 {
		thumbStart = 0
	}
	if thumbStart > h-thumbH {
		thumbStart = h - thumbH
	}

	lines := make([]string, h)
	for i := range h {
		if i >= thumbStart && i < thumbStart+thumbH {
			lines[i] = scrollThumbStyle.Render("█")
		} else {
			lines[i] = scrollTrackStyle.Render("░")
		}
	}
	return strings.Join(lines, "\n")
}