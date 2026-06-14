package renderer

import (
	"image/color"
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

type collapsibleRenderOpts struct {
	showStatusPreview bool
	spinnerFrame      int
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

func collapsibleActiveLabel(kind constants.MessageKind, status constants.DetailStatus) string {
	switch kind {
	case constants.MessageThinking:
		return "Thinking…"
	case constants.MessageDetail:
		return constants.DetailStatusPreviewLabel(status)
	default:
		return ""
	}
}

func foregroundOnBox(box lipgloss.Style, fg color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(fg).Background(box.GetBackground())
}

func collapsibleStatusPreview(kind constants.MessageKind, status constants.DetailStatus, box lipgloss.Style, spinnerFrame, maxWidth int) string {
	label := collapsibleActiveLabel(kind, status)
	if label == "" {
		return ""
	}
	frame := spinnerFrames[spinnerFrame%len(spinnerFrames)]
	var spinnerStyle, labelStyle lipgloss.Style
	switch kind {
	case constants.MessageThinking:
		spinnerStyle = foregroundOnBox(box, constants.Yellow)
		labelStyle = foregroundOnBox(box, constants.DimText)
	default:
		spinnerStyle = foregroundOnBox(box, constants.DetailStatusAccent(status).GetForeground())
		labelStyle = foregroundOnBox(box, constants.DetailStatusBodyStyle(status).GetForeground())
	}
	line := spinnerStyle.Render(frame) + labelStyle.Render(" "+label)
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

func collapsibleBodyBox(style lipgloss.Style, kind constants.MessageKind, status constants.DetailStatus, blockWidth, innerW, vPad, hPad int, body string, expanded bool, opts collapsibleRenderOpts) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" && !opts.showStatusPreview {
		return ""
	}
	var content string
	switch {
	case expanded:
		content = trimmed
		if kind == constants.MessageThinking {
			content = dimStyle.Render(trimmed)
		}
	case opts.showStatusPreview:
		content = collapsibleStatusPreview(kind, status, style, opts.spinnerFrame, innerW)
	case trimmed != "":
		if preview := collapsiblePreview(trimmed, innerW); preview != "" {
			content = dimStyle.Render(preview)
		}
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

func renderThinkingCollapsible(blockWidth int, label, body string, expanded bool, opts collapsibleRenderOpts) string {
	style := constants.MessageStyle(constants.MessageThinking).Italic(true)
	vPad, hPad := messageBlockPadding(constants.MessageThinking)
	innerW := max(blockWidth-2*hPad, 1)

	var b strings.Builder
	b.WriteString(collapsibleHeaderChip(style, constants.MessageThinking, label, expanded))
	if box := collapsibleBodyBox(style, constants.MessageThinking, constants.DetailStatusNeutral, blockWidth, innerW, vPad, hPad, body, expanded, opts); box != "" {
		b.WriteString("\n\n")
		b.WriteString(box)
	}
	b.WriteString("\n\n")
	b.WriteString(collapsibleHintLine(hPad, expanded))
	return b.String()
}

func renderDetailCollapsible(blockWidth int, label, body string, expanded bool, status constants.DetailStatus, opts collapsibleRenderOpts) string {
	style := constants.DetailStatusStyle(status)
	vPad, hPad := messageBlockPadding(constants.MessageDetail)
	innerW := max(blockWidth-2*hPad, 1)

	var b strings.Builder
	b.WriteString(collapsibleDetailTitleLine(hPad, status, label, expanded))
	if box := collapsibleBodyBox(style, constants.MessageDetail, status, blockWidth, innerW, vPad, hPad, body, expanded, opts); box != "" {
		b.WriteString("\n\n")
		b.WriteString(box)
	}
	b.WriteString("\n\n")
	b.WriteString(collapsibleHintLine(hPad, expanded))
	return b.String()
}

func renderDetailMessage(blockWidth int, label, body string, expanded bool, status constants.DetailStatus, opts collapsibleRenderOpts) string {
	return renderDetailCollapsible(blockWidth, label, body, expanded, status, opts)
}

func renderThinkingMessage(blockWidth int, label, body string, expanded bool, opts collapsibleRenderOpts) string {
	return renderThinkingCollapsible(blockWidth, label, body, expanded, opts)
}

func shellDetailLabel(command string) string {
	return "$ " + command
}
