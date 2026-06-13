package constants

import "github.com/charmbracelet/lipgloss"

// MessageKind identifies a stream message in the content area.
type MessageKind int

const (
	MessageUser MessageKind = iota
	MessageAI
	MessageSystem
	MessageTool
	MessageThinking
)

// Stream message colors — foreground + background only (no prefixes).
var (
	UserMsgFg = BrightText
	UserMsgBg = lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#2A2A2A"}

	AIMsgFg = BrightText
	AIMsgBg = lipgloss.NoColor{}

	SystemMsgFg = DimText
	SystemMsgBg = lipgloss.AdaptiveColor{Light: "#F3F4F6", Dark: "#252525"}

	ToolMsgFg = lipgloss.AdaptiveColor{Light: "#0E7490", Dark: "#67E8F9"}
	ToolMsgBg = lipgloss.AdaptiveColor{Light: "#ECFEFF", Dark: "#0C1A1D"}

	ThinkingMsgFg = DimText
	ThinkingMsgBg = lipgloss.NoColor{}
)

// MessageStyle returns the lipgloss style for a stream message kind.
func MessageStyle(kind MessageKind) lipgloss.Style {
	switch kind {
	case MessageUser:
		return lipgloss.NewStyle().Foreground(UserMsgFg).Background(UserMsgBg)
	case MessageAI:
		return lipgloss.NewStyle().Foreground(AIMsgFg).Background(AIMsgBg)
	case MessageSystem:
		return lipgloss.NewStyle().Foreground(SystemMsgFg).Background(SystemMsgBg)
	case MessageTool:
		return lipgloss.NewStyle().Foreground(ToolMsgFg).Background(ToolMsgBg)
	case MessageThinking:
		return lipgloss.NewStyle().Foreground(ThinkingMsgFg).Background(ThinkingMsgBg).Italic(true)
	default:
		return lipgloss.NewStyle().Foreground(AIMsgFg)
	}
}