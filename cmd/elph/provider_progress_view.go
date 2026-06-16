package main

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/riipandi/elph/pkg/ai/provider"
)

const (
	minimalBarWidth           = 24
	maxCompletedProviderLines = 4
)

var (
	progressDimStyle    = lipgloss.NewStyle().Foreground(uiconst.DimText)
	progressMutedStyle  = lipgloss.NewStyle().Foreground(uiconst.DimText)
	progressActiveStyle = lipgloss.NewStyle().Foreground(uiconst.BrightText)
	progressOkStyle     = lipgloss.NewStyle().Foreground(uiconst.Green)
	progressSkipStyle   = lipgloss.NewStyle().Foreground(uiconst.Gray)
)

func formatCompactElapsed(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		secs := d.Seconds()
		if secs < 10 {
			return fmt.Sprintf("%.1fs", secs)
		}
		return fmt.Sprintf("%ds", int(secs+0.5))
	default:
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%02ds", mins, secs)
	}
}

func shortPhaseLabel(phase string) string {
	switch phase {
	case "Connecting providers":
		return "connect"
	case "Syncing model catalogs":
		return "sync"
	default:
		return phase
	}
}

func renderMinimalBar(pct float64) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(float64(minimalBarWidth)*pct + 0.5)
	if filled > minimalBarWidth {
		filled = minimalBarWidth
	}
	bar := strings.Repeat("━", filled) + strings.Repeat("─", minimalBarWidth-filled)
	return progressDimStyle.Render(bar) + progressMutedStyle.Render(fmt.Sprintf(" %3.0f%%", pct*100))
}

func formatCompletedProvider(evt provider.ProviderProgressEvent) string {
	id := evt.ProviderID
	if id == "" {
		return ""
	}
	status := completedStatusLabel(evt)
	style := progressSkipStyle
	if isPositiveProviderAction(evt.Action) {
		style = progressOkStyle
	}
	return progressDimStyle.Render("  ") +
		progressMutedStyle.Render(fmt.Sprintf("%-12s", id)) +
		style.Render(status)
}

func completedStatusLabel(evt provider.ProviderProgressEvent) string {
	switch evt.Action {
	case provider.ProviderProgressCreated:
		return "created"
	case provider.ProviderProgressBackfill:
		return "thinking"
	case provider.ProviderProgressSynced:
		return "synced"
	case provider.ProviderProgressUnchanged:
		return "unchanged"
	case provider.ProviderProgressSkipped:
		if detail := strings.TrimSpace(evt.Detail); detail != "" {
			return detail
		}
		return "skipped"
	default:
		return ""
	}
}

func isPositiveProviderAction(action provider.ProviderProgressAction) bool {
	switch action {
	case provider.ProviderProgressCreated,
		provider.ProviderProgressBackfill,
		provider.ProviderProgressSynced:
		return true
	default:
		return false
	}
}

func currentProviderName(providerID, label string) string {
	name := strings.TrimSpace(label)
	if name == "" {
		name = strings.TrimSpace(providerID)
	}
	return name
}

func tailCompletedLines(lines []string, max int) []string {
	if len(lines) <= max {
		return lines
	}
	return lines[len(lines)-max:]
}

func appendCompletedLine(lines []string, line string) []string {
	if line == "" {
		return lines
	}
	lines = append(lines, line)
	if len(lines) > maxCompletedProviderLines {
		return lines[len(lines)-maxCompletedProviderLines:]
	}
	return lines
}
