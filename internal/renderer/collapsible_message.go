package renderer

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/riipandi/elph/internal/constants"
)

func isCollapsibleKind(kind constants.MessageKind) bool {
	return kind == constants.MessageDetail || kind == constants.MessageThinking
}

func collapsibleLabel(msg message) string {
	if label := strings.TrimSpace(msg.detailLabel); label != "" {
		return label
	}
	switch msg.kind {
	case constants.MessageThinking:
		return "Thinking"
	default:
		return "Details"
	}
}

func collapsibleExpandHint(expanded bool) string {
	if expanded {
		return "click or ctrl+o to collapse"
	}
	return "click or ctrl+o to expand"
}

func detailChevron(expanded bool) string {
	if expanded {
		return "▾"
	}
	return "▸"
}

func collapsiblePreview(body string, maxWidth int) string {
	line := firstDetailLine(body)
	if line == "" {
		return ""
	}
	if maxWidth <= 0 {
		return line
	}
	return ansi.Truncate(line, maxWidth, "…")
}

func firstDetailLine(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return strings.TrimSpace(body)
}

func collapsibleHeaderChip(style lipgloss.Style, kind constants.MessageKind, label string, expanded bool) string {
	chevron := lipgloss.NewStyle().Foreground(constants.DimText).Render(detailChevron(expanded))
	title := lipgloss.NewStyle().Bold(true).Render(label)
	return style.Padding(0, 1).Render(chevron + " " + title)
}

func collapsibleDetailTitleLine(hPad int, status constants.DetailStatus, label string, expanded bool) string {
	chevron := constants.DetailStatusAccent(status).Render(detailChevron(expanded))
	title := lipgloss.NewStyle().Bold(true).Foreground(constants.DimText).Render(label)
	return strings.Repeat(" ", hPad) + chevron + " " + title
}

func collapsibleBodyBox(style lipgloss.Style, kind constants.MessageKind, blockWidth, innerW, vPad, hPad int, body string, expanded bool) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	var content string
	if expanded {
		content = trimmed
		if kind == constants.MessageThinking {
			content = dimStyle.Render(trimmed)
		}
	} else if preview := collapsiblePreview(trimmed, innerW); preview != "" {
		content = dimStyle.Render(preview)
	}
	if content == "" {
		return ""
	}
	return style.Padding(vPad, hPad).Width(blockWidth).Render(content)
}

func collapsibleHintLine(hPad int, expanded bool) string {
	return lipgloss.NewStyle().
		Foreground(constants.DimText).
		Italic(true).
		Background(lipgloss.NoColor{}).
		PaddingLeft(hPad).
		Render(collapsibleExpandHint(expanded))
}

func renderThinkingCollapsible(blockWidth int, label, body string, expanded bool) string {
	style := constants.MessageStyle(constants.MessageThinking).Italic(true)
	vPad, hPad := messageBlockPadding(constants.MessageThinking)
	innerW := max(blockWidth-2*hPad, 1)

	var b strings.Builder
	b.WriteString(collapsibleHeaderChip(style, constants.MessageThinking, label, expanded))
	if box := collapsibleBodyBox(style, constants.MessageThinking, blockWidth, innerW, vPad, hPad, body, expanded); box != "" {
		b.WriteString("\n\n")
		b.WriteString(box)
	}
	b.WriteString("\n\n")
	b.WriteString(collapsibleHintLine(hPad, expanded))
	return b.String()
}

func renderDetailCollapsible(blockWidth int, label, body string, expanded bool, status constants.DetailStatus) string {
	style := constants.DetailStatusStyle(status)
	vPad, hPad := messageBlockPadding(constants.MessageDetail)
	innerW := max(blockWidth-2*hPad, 1)

	var b strings.Builder
	b.WriteString(collapsibleDetailTitleLine(hPad, status, label, expanded))
	if box := collapsibleBodyBox(style, constants.MessageDetail, blockWidth, innerW, vPad, hPad, body, expanded); box != "" {
		b.WriteString("\n\n")
		b.WriteString(box)
	}
	b.WriteString("\n\n")
	b.WriteString(collapsibleHintLine(hPad, expanded))
	return b.String()
}

func renderDetailMessage(blockWidth int, label, body string, expanded bool, status constants.DetailStatus) string {
	return renderDetailCollapsible(blockWidth, label, body, expanded, status)
}

func renderThinkingMessage(blockWidth int, label, body string, expanded bool) string {
	return renderThinkingCollapsible(blockWidth, label, body, expanded)
}

func shellDetailLabel(command string) string {
	return "$ " + command
}
