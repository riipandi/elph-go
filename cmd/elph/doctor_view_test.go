package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRenderDoctorCardGroupsSections(t *testing.T) {
	report := doctorReport{}
	report.add(doctorOK, "Environment", "no ELPH_* overrides")
	report.add(doctorOK, "Settings", "~/.elph/settings.json")
	report.add(doctorWarn, "Settings", "syncInterval invalid")
	report.add(doctorFail, "Providers", "missing directory")

	out := renderDoctorCard(report, "", "", 0)
	require.Contains(t, out, "Elph doctor")
	require.Contains(t, out, "Environment")
	require.Contains(t, out, "Settings")
	require.Contains(t, out, "Providers")
	require.Contains(t, out, "2 passed")
	require.Contains(t, out, "1 warning")
	require.Contains(t, out, "1 failed")
}

func TestRenderDoctorCardShowsActiveStep(t *testing.T) {
	report := doctorReport{}
	out := renderDoctorCard(report, "providers", "⠋", 250*time.Millisecond)
	require.Contains(t, strings.ToLower(out), "checking")
	require.Contains(t, out, "providers")
}
