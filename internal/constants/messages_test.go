package constants

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageStyleKindsDiffer(t *testing.T) {
	user := MessageStyle(MessageUser).GetForeground()
	tool := MessageStyle(MessageTool).GetForeground()
	require.NotEqual(t, user, tool, "user and tool foreground colors should differ")
}

func TestThinkingUsesDimText(t *testing.T) {
	require.Equal(t, DimText, MessageStyle(MessageThinking).GetForeground(),
		"thinking messages should use dim gray foreground")
}

func TestThinkingBackgroundDiffersFromDetail(t *testing.T) {
	thinking := MessageStyle(MessageThinking).GetBackground()
	detail := MessageStyle(MessageDetail).GetBackground()
	require.NotEqual(t, thinking, detail)
}

func TestStickyUserUsesNeutralBackground(t *testing.T) {
	user := MessageStyle(MessageUser).GetBackground()
	sticky := StickyUserStyle().GetBackground()
	require.NotEqual(t, user, sticky)
	require.Equal(t, UserStickyMsgBg, sticky)
}

func TestStickyUserTitleUsesNeutralForeground(t *testing.T) {
	title := StickyUserTitleStyle().GetForeground()
	require.Equal(t, DimText, title)
	require.NotEqual(t, UserStickyTimestampFg, title)
}

func TestStickyUserTimestampUsesSoftGreenForeground(t *testing.T) {
	ts := StickyUserTimestampStyle().GetForeground()
	require.Equal(t, UserStickyTimestampFg, ts)
	require.NotEqual(t, DimText, ts)
	require.NotEqual(t, BrightText, ts)
}

func TestDetailNeutralUsesSoftPalette(t *testing.T) {
	detail := MessageStyle(MessageDetail)
	neutral := DetailStatusStyle(DetailStatusNeutral)
	require.Equal(t, neutral.GetForeground(), detail.GetForeground())
	require.Equal(t, neutral.GetBackground(), detail.GetBackground())
}
