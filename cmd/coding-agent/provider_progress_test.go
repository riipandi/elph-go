package main

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/stopwatch"
	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/stretchr/testify/require"
)

func TestProviderProgressModelView(t *testing.T) {
	m := newProviderProgressModel(providerProgressUpdate, false)
	m.providerIndex = 2
	m.providerTotal = 6
	m.currentAction = provider.ProviderProgressWorking
	m.currentLabel = "Anthropic"
	view := m.View()
	require.Contains(t, view, "connect")
	require.Contains(t, view, "2/6")
	require.Contains(t, view, "Anthropic")
	require.Contains(t, view, "━━")
	require.Contains(t, view, "17%")
	require.NotContains(t, view, "█")
}

func TestProviderProgressPercentComplete(t *testing.T) {
	m := newProviderProgressModel(providerProgressUpdate, false)
	m.providerIndex = 6
	m.providerTotal = 6
	m.currentAction = provider.ProviderProgressSynced
	require.InDelta(t, 1.0, m.progressPercent(), 0.001)
	view := m.View()
	require.Contains(t, view, "100%")
}

func TestProviderProgressPercentWhileWorking(t *testing.T) {
	m := newProviderProgressModel(providerProgressUpdate, false)
	m.providerIndex = 3
	m.providerTotal = 6
	m.currentAction = provider.ProviderProgressWorking
	require.InDelta(t, 2.0/6.0, m.progressPercent(), 0.001)
}

func TestProviderProgressConnectLayout(t *testing.T) {
	m := newProviderProgressModel(providerProgressConnect, false)
	view := m.View()
	require.Contains(t, view, "connect")
	require.Contains(t, view, "starting")
	require.Contains(t, view, "0ms")
	require.Contains(t, view, "0%")
	require.NotContains(t, view, "(0/0)")
}

func TestProviderProgressBeginWorkOnce(t *testing.T) {
	m := newProviderProgressModel(providerProgressConnect, false)
	require.False(t, m.workStarted)
	m.beginWork()
	require.True(t, m.workStarted)
	m.beginWork()
	require.True(t, m.workStarted)
}

func TestIsInteractiveStderrFalseInTests(t *testing.T) {
	require.False(t, isInteractiveStderr())
}

func TestFormatCompletedProviderMinimal(t *testing.T) {
	line := formatCompletedProvider(provider.ProviderProgressEvent{
		ProviderID: "openai",
		Action:     provider.ProviderProgressSynced,
	})
	require.Contains(t, line, "openai")
	require.Contains(t, line, "synced")
	require.NotContains(t, line, "✓")
}

func TestApplyProgressEventTracksCompleted(t *testing.T) {
	m := newProviderProgressModel(providerProgressUpdate, false)
	m.applyProgressEvent(provider.ProviderProgressEvent{
		Phase:      provider.ProviderProgressConnect,
		ProviderID: "openai",
		Label:      "OpenAI",
		Index:      1,
		Total:      6,
		Action:     provider.ProviderProgressCreated,
	})
	require.Len(t, m.completed, 1)
	require.Contains(t, m.completed[0], "openai")
	require.Contains(t, m.completed[0], "created")
}

func TestRenderMinimalBar(t *testing.T) {
	bar := renderMinimalBar(0.5)
	require.Contains(t, bar, "50%")
	require.Contains(t, bar, "━")
	require.Contains(t, bar, "─")
}

func TestFormatCompactElapsed(t *testing.T) {
	require.Equal(t, "0ms", formatCompactElapsed(0))
	require.Equal(t, "450ms", formatCompactElapsed(450*time.Millisecond))
	require.Equal(t, "4.2s", formatCompactElapsed(4200*time.Millisecond))
	require.Equal(t, "12s", formatCompactElapsed(12*time.Second))
	require.Equal(t, "1m05s", formatCompactElapsed(65*time.Second))
}

func TestProviderProgressShowsElapsed(t *testing.T) {
	m := newProviderProgressModel(providerProgressConnect, false)
	for _, msg := range drainTeaCmd(m.stopwatch.Start()) {
		m.stopwatch, _ = m.stopwatch.Update(msg)
	}
	m.stopwatch, _ = m.stopwatch.Update(stopwatch.TickMsg{ID: m.stopwatch.ID()})
	require.Greater(t, m.stopwatch.Elapsed(), time.Duration(0))
	require.Contains(t, m.View(), formatCompactElapsed(m.stopwatch.Elapsed()))
}
