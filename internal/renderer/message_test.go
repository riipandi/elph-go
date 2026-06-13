package renderer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/riipandi/elph/internal/constants"
)

func testModel() Model {
	m := New()
	m.width = 80
	m.content.Width = 80
	return m
}

func TestMessageKindsNoPipePrefix(t *testing.T) {
	m := testModel()
	kinds := []constants.MessageKind{
		constants.MessageUser,
		constants.MessageAI,
		constants.MessageSystem,
		constants.MessageTool,
		constants.MessageThinking,
	}
	for _, kind := range kinds {
		rendered := m.renderMessage(message{text: "sample text", kind: kind})
		if strings.Contains(stripANSI(rendered), "| ") {
			t.Fatalf("kind %d should not use pipe prefix: %q", kind, stripANSI(rendered))
		}
	}
}

func TestMessageKindsNoChevronPrefix(t *testing.T) {
	m := testModel()
	rendered := m.renderMessage(message{text: "Copied to clipboard", kind: constants.MessageSystem})
	if strings.HasPrefix(strings.TrimSpace(stripANSI(rendered)), ">") {
		t.Fatalf("system message should not use > prefix: %q", stripANSI(rendered))
	}
}

func TestUserMessageStyled(t *testing.T) {
	m := testModel()
	rendered := m.renderMessage(message{text: "hello from user", kind: constants.MessageUser})
	if !strings.Contains(rendered, "hello from user") {
		t.Fatalf("missing message text: %q", rendered)
	}
}

func TestAIMessageRendersText(t *testing.T) {
	m := testModel()
	rendered := m.renderMessage(message{text: "response from agent", kind: constants.MessageAI})
	if !strings.Contains(rendered, "response from agent") {
		t.Fatalf("missing message text: %q", rendered)
	}
}

func TestUserMessageVerticalSpacing(t *testing.T) {
	m := testModel()
	m.messages = []message{
		{text: "from agent", kind: constants.MessageAI},
		{text: "from user", kind: constants.MessageUser},
		{text: "reply", kind: constants.MessageAI},
	}
	content := stripANSI(m.contentView())
	agentEnd := strings.Index(content, "from agent")
	userStart := strings.Index(content, "from user")
	replyStart := strings.Index(content, "reply")
	if agentEnd < 0 || userStart < 0 || replyStart < 0 {
		t.Fatalf("missing messages in content:\n%q", content)
	}
	if !strings.Contains(content[agentEnd:userStart], "\n\n") {
		t.Fatal("expected blank line between AI and user message")
	}
	if !strings.Contains(content[userStart:replyStart], "\n\n") {
		t.Fatal("expected blank line between user and AI message")
	}
}

func TestUserMessageMultiline(t *testing.T) {
	m := testModel()
	rendered := m.renderMessage(message{text: "line one\nline two", kind: constants.MessageUser})
	if lipgloss.Height(rendered) < 4 {
		t.Fatalf("multiline user message should include vertical padding: h=%d", lipgloss.Height(rendered))
	}
	plain := stripANSI(rendered)
	if !strings.Contains(plain, "line one") || !strings.Contains(plain, "line two") {
		t.Fatalf("missing line text: %q", plain)
	}
}

func TestUserMessageWidthMatchesChrome(t *testing.T) {
	m := testModel()
	m.messages = []message{{text: "hello", kind: constants.MessageUser}}
	assertChromeWidthsMatch(t, m)
}

func TestUserMessageWidthMatchesChromeWithScrollbar(t *testing.T) {
	m := testModel()
	m.height = 12
	m.ready = true
	for i := range 30 {
		m.messages = append(m.messages, message{
			text: fmt.Sprintf("message %d", i),
			kind: constants.MessageUser,
		})
	}
	m = m.syncLayout(false)
	assertChromeWidthsMatch(t, m)
}

func assertChromeWidthsMatch(t *testing.T, m Model) {
	t.Helper()
	userW := lipgloss.Width(m.renderMessage(m.messages[len(m.messages)-1]))
	bannerW := lipgloss.Width(m.bannerView())
	inputW := lipgloss.Width(m.inputView())
	msgW := m.messageAreaWidth()
	if userW != msgW {
		t.Fatalf("user message width %d != messageAreaWidth %d", userW, msgW)
	}
	if bannerW != m.chromeOuterWidth() {
		t.Fatalf("banner width %d != chromeOuterWidth %d", bannerW, m.chromeOuterWidth())
	}
	if inputW != bannerW {
		t.Fatalf("input width %d != banner width %d", inputW, bannerW)
	}
}

func TestMessageWidthUsesContentAreaWidth(t *testing.T) {
	m := testModel()
	m.height = 12
	m.ready = true
	for i := range 30 {
		m.messages = append(m.messages, message{
			text: fmt.Sprintf("message %d", i),
			kind: constants.MessageUser,
		})
	}
	m = m.syncLayout(false)

	areaW := m.contentAreaWidth()
	msgW := m.messageAreaWidth()
	if areaW != m.width-scrollBarWidth {
		t.Fatalf("contentAreaWidth %d, want %d", areaW, m.width-scrollBarWidth)
	}
	if msgW != areaW-messageScrollInset {
		t.Fatalf("messageAreaWidth %d, want %d", msgW, areaW-messageScrollInset)
	}
	for _, kind := range []constants.MessageKind{
		constants.MessageUser,
		constants.MessageAI,
		constants.MessageSystem,
	} {
		renderedW := lipgloss.Width(m.renderMessage(message{text: "hello", kind: kind}))
		if renderedW != msgW {
			t.Fatalf("kind %d width %d != messageAreaWidth %d", kind, renderedW, msgW)
		}
	}
}

func TestUserMessageHorizontalPadding(t *testing.T) {
	m := testModel()
	rendered := stripANSI(m.renderMessage(message{text: "hello", kind: constants.MessageUser}))
	if !strings.HasPrefix(rendered, "  ") {
		t.Fatalf("user message should have horizontal padding: %q", rendered)
	}
}

func TestUserMsgBgConstant(t *testing.T) {
	if constants.UserMsgBg == constants.DimText {
		t.Fatal("user message background should differ from dim text")
	}
	_ = lipgloss.NewStyle().Background(constants.UserMsgBg).Render("x")
}

// stripANSI is a minimal helper for tests; lipgloss output includes sequences.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}