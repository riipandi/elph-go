package renderer

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/stretchr/testify/require"
)

func TestCollapsibleHeaderRendersLabel(t *testing.T) {
	m := testModel()
	rendered := stripANSI(m.renderMessage(message{
		kind:        uiconst.MessageDetail,
		detailLabel: "Prompt",
		text:        "body",
	}))
	require.Contains(t, rendered, "▸")
	require.Contains(t, rendered, "Prompt")
}

func TestCollapsibleHeaderIsNotFullWidth(t *testing.T) {
	m := testModel()
	m.width = 80
	rendered := m.renderMessage(message{
		kind:        uiconst.MessageDetail,
		detailLabel: "Prompt",
		text:        "body",
	})
	firstLine := strings.SplitN(rendered, "\n", 2)[0]
	require.Less(t, lipgloss.Width(firstLine), m.messageAreaWidth())
}

func TestCollapsibleLayoutHasHintGapAfterContent(t *testing.T) {
	m := testModel()
	rendered := stripANSI(m.renderMessage(message{
		kind:        uiconst.MessageDetail,
		detailLabel: "Prompt",
		text:        "preview line",
	}))
	hintIdx := strings.Index(rendered, "click or ctrl+o to expand")
	previewIdx := strings.Index(rendered, "preview line")
	require.Greater(t, hintIdx, previewIdx)
	require.Greater(t, strings.Count(rendered[previewIdx:hintIdx], "\n"), 1)
}
