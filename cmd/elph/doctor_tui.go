package main

import (
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/stopwatch"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/riipandi/elph/internal/constants"
)

type doctorStep struct {
	label string
	run   func(*doctorReport, string) error
}

var doctorCheckSteps = []doctorStep{
	{label: "environment", run: func(r *doctorReport, _ string) error {
		r.checkEnvironment()
		return nil
	}},
	{label: "settings", run: func(r *doctorReport, workDir string) error {
		return r.checkSettings(workDir)
	}},
	{label: "version metadata", run: func(r *doctorReport, _ string) error {
		return r.checkVersion()
	}},
	{label: "providers", run: func(r *doctorReport, _ string) error {
		return r.checkProviders()
	}},
	{label: "active model", run: func(r *doctorReport, _ string) error {
		r.checkActiveModel()
		return nil
	}},
}

func isInteractiveStdout() bool {
	return term.IsTerminal(os.Stdout.Fd())
}

func runDoctorChecks(workDir string) (doctorReport, error) {
	report := doctorReport{}
	for _, step := range doctorCheckSteps {
		if err := step.run(&report, workDir); err != nil {
			return report, err
		}
	}
	return report, nil
}

func runDoctorUI(workDir string) (doctorReport, error) {
	if isInteractiveStdout() {
		return runDoctorInteractive(workDir)
	}
	report, err := runDoctorChecks(workDir)
	if err != nil {
		return report, err
	}
	report.render(os.Stdout)
	return report, nil
}

func runDoctorInteractive(workDir string) (doctorReport, error) {
	spin := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(constants.Yellow)),
	)
	clock := stopwatch.New(stopwatch.WithInterval(100 * time.Millisecond))
	clock = applyStopwatchCmd(clock, clock.Start())

	spinInterval := spin.Spinner.FPS
	if spinInterval <= 0 {
		spinInterval = time.Second / 12
	}
	spinTicker := time.NewTicker(spinInterval)
	clockTicker := time.NewTicker(100 * time.Millisecond)
	defer spinTicker.Stop()
	defer clockTicker.Stop()

	report := doctorReport{}
	drawn := false
	viewLines := 0

	redraw := func(activeStep string) {
		if drawn && viewLines > 0 {
			clearDoctorView(viewLines)
		}
		content := renderDoctorCard(report, activeStep, spin.View(), clock.Elapsed()) + "\n"
		viewLines = doctorViewLineCount(content)
		fmt.Print(content)
		drawn = true
	}

	for _, step := range doctorCheckSteps {
		done := make(chan error, 1)
		go func(s doctorStep) {
			done <- s.run(&report, workDir)
		}(step)

		active := step.label
		running := true
		for running {
			select {
			case <-spinTicker.C:
				spin, _ = spin.Update(spin.Tick())
				redraw(active)
			case <-clockTicker.C:
				clock, _ = clock.Update(stopwatch.TickMsg{ID: clock.ID()})
				redraw(active)
			case err := <-done:
				if err != nil {
					redraw("")
					return report, err
				}
				redraw("")
				running = false
			}
		}
	}

	clock = applyStopwatchCmd(clock, clock.Stop())
	redraw("")
	return report, nil
}
