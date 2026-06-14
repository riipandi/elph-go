package renderer

import (
	"testing"

	"github.com/riipandi/elph/internal/constants"
	"github.com/stretchr/testify/require"
)

func TestStreamPrefixCacheInvalidatesOnDetailExpand(t *testing.T) {
	m := testModel()
	m.messages = []message{
		{kind: constants.MessageUser, text: "hello"},
		{
			kind:        constants.MessageDetail,
			detailLabel: "Prompt",
			text:        "alpha\nbeta\ngamma",
		},
		{kind: constants.MessageAI, text: "streaming"},
	}
	m.agent.Busy = true
	m.agent.ResponseMsgID = 2

	m = m.refreshStreamPrefixCache()
	collapsed := m.layout.StreamPrefix

	m.messages[1].detailExpanded = true
	m.messages[1].renderCache = messageRenderCache{}
	m = m.refreshStreamPrefixCache()
	expanded := m.layout.StreamPrefix

	require.NotEqual(t, collapsed, expanded)
	require.Contains(t, expanded, "beta")
}
