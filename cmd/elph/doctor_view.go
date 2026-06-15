package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/constants"
)

var (
	doctorTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(constants.PrimaryText)
	doctorSectionStyle = lipgloss.NewStyle().Bold(true).Foreground(constants.BrightText)
	doctorDetailStyle  = lipgloss.NewStyle().Foreground(constants.BrightText)
	doctorDimStyle     = lipgloss.NewStyle().Foreground(constants.DimText)
	doctorOkStyle      = lipgloss.NewStyle().Foreground(constants.Green).Bold(true)
	doctorWarnStyle    = lipgloss.NewStyle().Foreground(constants.Yellow).Bold(true)
	doctorFailStyle    = lipgloss.NewStyle().Foreground(constants.Red).Bold(true)
	doctorBoxStyle     = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(constants.DimText).
				Padding(1, 2)
	doctorSummaryOkStyle   = lipgloss.NewStyle().Foreground(constants.Green)
	doctorSummaryWarnStyle = lipgloss.NewStyle().Foreground(constants.Yellow)
	doctorSummaryFailStyle = lipgloss.NewStyle().Foreground(constants.Red)
)

func doctorStatusIcon(status doctorStatus) string {
	switch status {
	case doctorOK:
		return doctorOkStyle.Render("✓")
	case doctorWarn:
		return doctorWarnStyle.Render("!")
	default:
		return doctorFailStyle.Render("✗")
	}
}

func renderDoctorFinding(f doctorFinding) string {
	icon := doctorStatusIcon(f.Status)
	gap := strings.Repeat(" ", max(2-lipgloss.Width(icon), 0))
	return doctorDimStyle.Render("  ") + icon + gap + doctorDetailStyle.Render(f.Detail)
}

func renderDoctorSections(findings []doctorFinding) string {
	if len(findings) == 0 {
		return doctorDimStyle.Render("  No checks yet")
	}

	var blocks []string
	current := ""
	var lines []string

	flush := func() {
		if current == "" || len(lines) == 0 {
			return
		}
		block := doctorSectionStyle.Render(current) + "\n" + strings.Join(lines, "\n")
		blocks = append(blocks, block)
		lines = nil
	}

	for _, f := range findings {
		if f.Label != current {
			flush()
			current = f.Label
		}
		lines = append(lines, renderDoctorFinding(f))
	}
	flush()

	return strings.Join(blocks, "\n\n")
}

func renderDoctorSummary(report doctorReport) string {
	ok, warn, fail := report.counts()
	parts := make([]string, 0, 3)
	if ok > 0 {
		parts = append(parts, doctorSummaryOkStyle.Render(fmt.Sprintf("✓ %d passed", ok)))
	}
	if warn > 0 {
		parts = append(parts, doctorSummaryWarnStyle.Render(fmt.Sprintf("! %d warning%s", warn, plural(warn))))
	}
	if fail > 0 {
		parts = append(parts, doctorSummaryFailStyle.Render(fmt.Sprintf("✗ %d failed", fail)))
	}
	if len(parts) == 0 {
		return doctorDimStyle.Render("No checks recorded")
	}
	return strings.Join(parts, doctorDimStyle.Render("  ·  "))
}

func renderDoctorActiveLine(spinnerView, stepLabel string) string {
	if strings.TrimSpace(stepLabel) == "" {
		return ""
	}
	label := doctorDimStyle.Render("Checking ")
	label += doctorDetailStyle.Render(stepLabel)
	label += doctorDimStyle.Render("…")
	if strings.TrimSpace(spinnerView) != "" {
		return doctorDimStyle.Render("  ") + spinnerView + " " + label
	}
	return doctorDimStyle.Render("  ") + label
}

func renderDoctorCard(report doctorReport, activeStep, spinnerView string, elapsed time.Duration) string {
	title := doctorTitleStyle.Render("Elph doctor")
	if elapsed > 0 {
		title += doctorDimStyle.Render(" · " + formatCompactElapsed(elapsed))
	}

	body := renderDoctorSections(report.Findings)
	if active := renderDoctorActiveLine(spinnerView, activeStep); active != "" {
		if body != "" {
			body += "\n\n" + active
		} else {
			body = active
		}
	}

	summary := renderDoctorSummary(report)
	content := strings.Join(compactDoctorBlocks(title, body, summary), "\n\n")
	return doctorBoxStyle.Render(content)
}

func compactDoctorBlocks(blocks ...string) []string {
	out := make([]string, 0, len(blocks))
	for _, block := range blocks {
		if strings.TrimSpace(block) != "" {
			out = append(out, block)
		}
	}
	return out
}

func doctorViewLineCount(content string) int {
	if content == "" {
		return 0
	}
	return lipgloss.Height(content)
}

func clearDoctorView(lines int) {
	if lines <= 0 {
		return
	}
	fmt.Fprintf(os.Stdout, "\033[%dF\r\033[J", lines)
}

func (r doctorReport) render(w io.Writer) {
	fmt.Fprintln(w, renderDoctorCard(r, "", "", 0))
}
