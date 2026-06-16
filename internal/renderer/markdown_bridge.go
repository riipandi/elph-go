package renderer

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/riipandi/elph/internal/rendermd"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/riipandi/elph/pkg/core/agent"
)

type markdownRenderMsg struct {
	index  int
	width  int
	source string
	output string
}

func renderAISourcePreview(blockWidth int, text string) string {
	body := rendermd.RenderSourcePreview(aiContentWidth(blockWidth), text)
	return renderAIBlock(blockWidth, body, false)
}

func renderAIMessagePlain(blockWidth int, text string, streaming bool) string {
	body := rendermd.RenderPlainBody(aiContentWidth(blockWidth), text, !streaming)
	return renderAIBlock(blockWidth, body, false)
}

func renderAIMarkdown(blockWidth int, text string) string {
	contentW := aiContentWidth(blockWidth)
	rendered, err := rendermd.RenderGlamour(contentW, text)
	if err != nil || rendered == "" {
		return renderAIMessagePlain(blockWidth, text, false)
	}
	return renderAIPreformattedBlock(blockWidth, rendered, false)
}

func renderAIMessage(blockWidth int, text string, streaming, markdownPending bool) string {
	text = agent.SanitizeAssistantDisplay(text)
	if strings.TrimSpace(text) == "" {
		return ""
	}
	switch {
	case markdownPending:
		return renderAISourcePreview(blockWidth, text)
	case streaming || !rendermd.LooksLikeMarkdown(text):
		if !streaming && !rendermd.LooksLikeMarkdown(text) {
			text = rendermd.NormalizeProseSeparators(text)
		}
		return renderAIMessagePlain(blockWidth, text, streaming)
	default:
		return renderAIMarkdown(blockWidth, text)
	}
}

func markdownRenderCmd(index, width int, source string) tea.Cmd {
	return func() tea.Msg {
		return markdownRenderMsg{
			index:  index,
			width:  width,
			source: source,
			output: renderAIMarkdown(width, source),
		}
	}
}

func (m Model) handleMarkdownRenderMsg(msg markdownRenderMsg) (Model, tea.Cmd) {
	if msg.index < 0 || msg.index >= len(m.messages) {
		return m, nil
	}
	if m.messages[msg.index].text != msg.source {
		return m, nil
	}
	m.messages[msg.index].markdownPending = false
	m.messages[msg.index].renderCache = messageRenderCache{
		width:     msg.width,
		sourceLen: len(msg.source),
		streaming: false,
		output:    msg.output,
	}
	m.layout.ContentDirty = true
	m = m.syncLayout(m.content.AtBottom())
	return m, nil
}

func (m Model) scheduleMarkdownRender(index int) (Model, tea.Cmd) {
	if index < 0 || index >= len(m.messages) {
		return m, nil
	}
	msg := m.messages[index]
	if msg.kind != uiconst.MessageAI || !rendermd.LooksLikeMarkdown(msg.text) || len(msg.text) < rendermd.AsyncMinLen {
		return m, nil
	}
	width := m.messageAreaWidth()
	m.messages[index].markdownPending = true
	m.layout.ContentDirty = true
	return m, markdownRenderCmd(index, width, msg.text)
}
