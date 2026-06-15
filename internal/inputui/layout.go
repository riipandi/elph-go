package inputui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	MaxInputLines    = 8
	MinViewportRows  = 6
	InputChromeSlack = 2
)

// ContentWidth is the textarea width inside the input border and padding.
func ContentWidth(outer int) int {
	return max(outer-4, 1)
}

// OverlayScrollBar replaces the last column of each textarea line with the scrollbar.
func OverlayScrollBar(body, bar string, targetWidth int) string {
	bodyLines := strings.Split(body, "\n")
	barLines := strings.Split(bar, "\n")
	if len(bodyLines) > 0 && bodyLines[len(bodyLines)-1] == "" {
		bodyLines = bodyLines[:len(bodyLines)-1]
	}
	if len(barLines) > 0 && barLines[len(barLines)-1] == "" {
		barLines = barLines[:len(barLines)-1]
	}

	textW := max(targetWidth-1, 0)

	out := make([]string, len(bodyLines))
	for i, line := range bodyLines {
		if i >= len(barLines) || barLines[i] == "" {
			out[i] = line
			continue
		}
		truncated := ansi.Truncate(line, textW, "")
		pad := textW - lipgloss.Width(truncated)
		if pad < 0 {
			pad = 0
		}
		out[i] = truncated + strings.Repeat(" ", pad) + barLines[i]
	}
	return strings.Join(out, "\n")
}

// WrappedRows returns how many display rows a line occupies at the given width.
func WrappedRows(line string, width int) int {
	if width < 1 {
		width = 1
	}
	if line == "" {
		return 1
	}
	wrapped := ansi.Hardwrap(ansi.Wordwrap(line, width, ""), width, false)
	return max(1, strings.Count(wrapped, "\n")+1)
}

// DisplayRows returns total display rows for a value including paste tokens.
func DisplayRows(val string, pastes map[int]string, width int) int {
	val = DisplayValue(val, pastes)
	if val == "" {
		return 1
	}
	w := max(width, 1)
	rows := 0
	for _, line := range strings.Split(val, "\n") {
		rows += WrappedRows(line, w)
	}
	return max(rows, 1)
}

// NormalizeForSubmit trims trailing whitespace from each line before submit.
func NormalizeForSubmit(s string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.Trim(strings.Join(lines, "\n"), " \t\n")
}

// IsSlashCommand reports whether input starts with a slash command.
func IsSlashCommand(s string) bool {
	return strings.HasPrefix(strings.TrimLeft(s, " \t"), "/")
}