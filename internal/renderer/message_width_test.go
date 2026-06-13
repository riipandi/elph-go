package renderer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/riipandi/elph/internal/constants"
)

func TestUserMessageLinesFitContentViewportWithScrollbar(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 14
	m.ready = true
	for i := range 25 {
		m.messages = append(m.messages, message{
			text: fmt.Sprintf("user message %d with enough text to matter", i),
			kind: constants.MessageUser,
		})
	}
	m.contentDirty = true
	m = m.syncLayout(false)

	if !m.contentScrollable() {
		t.Fatal("expected scrollable content")
	}

	content := m.contentView()
	maxLineW := 0
	for _, line := range strings.Split(content, "\n") {
		if w := lipgloss.Width(line); w > maxLineW {
			maxLineW = w
		}
	}
	if maxLineW > m.content.Width {
		t.Fatalf("content line width %d exceeds viewport %d", maxLineW, m.content.Width)
	}

	userW := lipgloss.Width(m.renderMessage(m.messages[len(m.messages)-1]))
	msgW := m.messageAreaWidth()
	if userW != msgW {
		t.Fatalf("user message width %d != messageAreaWidth %d", userW, msgW)
	}
	if userW >= m.content.Width {
		t.Fatalf("user message width %d should be narrower than viewport %d", userW, m.content.Width)
	}
	if lipgloss.Width(m.contentAreaView()) > m.width {
		t.Fatalf("content area %d exceeds terminal %d", lipgloss.Width(m.contentAreaView()), m.width)
	}
}