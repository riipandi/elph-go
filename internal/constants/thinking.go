package constants

import "github.com/charmbracelet/lipgloss"

// ─── Thinking Level ──────────────────────────────────────────────────────────

type ThinkingLevel string

const (
	ThinkingOff     ThinkingLevel = "off"
	ThinkingMinimal ThinkingLevel = "minimal"
	ThinkingLow     ThinkingLevel = "low"
	ThinkingMedium  ThinkingLevel = "medium"
	ThinkingHigh    ThinkingLevel = "high"
	ThinkingXHigh   ThinkingLevel = "xhigh"
)

// ThinkingColor returns the lipgloss color for a given thinking level.
func ThinkingColor(lvl ThinkingLevel) lipgloss.Color {
	switch lvl {
	case ThinkingOff:
		return lipgloss.Color("#6B7280") // gray
	case ThinkingMinimal:
		return lipgloss.Color("#06B6D4") // cyan
	case ThinkingLow:
		return lipgloss.Color("#22C55E") // green
	case ThinkingMedium:
		return lipgloss.Color("#EAB308") // yellow
	case ThinkingHigh:
		return lipgloss.Color("#F97316") // orange
	case ThinkingXHigh:
		return lipgloss.Color("#EF4444") // red
	default:
		return lipgloss.Color("#6B7280")
	}
}

var thinkingLevels = []ThinkingLevel{
	ThinkingOff,
	ThinkingMinimal,
	ThinkingLow,
	ThinkingMedium,
	ThinkingHigh,
	ThinkingXHigh,
}

func NextThinkingLevel(lvl ThinkingLevel) ThinkingLevel {
	for i, l := range thinkingLevels {
		if l == lvl {
			return thinkingLevels[(i+1)%len(thinkingLevels)]
		}
	}
	return thinkingLevels[0]
}

func PrevThinkingLevel(lvl ThinkingLevel) ThinkingLevel {
	for i, l := range thinkingLevels {
		if l == lvl {
			p := i - 1
			if p < 0 {
				p = len(thinkingLevels) - 1
			}
			return thinkingLevels[p]
		}
	}
	return thinkingLevels[len(thinkingLevels)-1]
}
