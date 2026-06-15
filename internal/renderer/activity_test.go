package renderer

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/internal/constants"
	"github.com/riipandi/elph/internal/settings"
	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/stretchr/testify/require"
)

func TestActivityViewHiddenWhenIdle(t *testing.T) {
	m := New()
	m.width = 80
	require.Empty(t, m.activityView())
}

func TestInputHasTopMarginWhenIdle(t *testing.T) {
	m := testInputModel(t)
	require.Equal(t, agent.ActivityIdle, m.agent.Activity)
	require.Greater(t, lipgloss.Height(m.inputView()), lipgloss.Height(m.inputBodyView())+1)

	m = m.beginAgentTurn()
	require.GreaterOrEqual(t, lipgloss.Height(m.inputView()), lipgloss.Height(m.inputBodyView()))
}

func TestActivityViewShowsLabel(t *testing.T) {
	m := New()
	m.width = 80
	m.agent.Busy = true
	m.agent.Activity = agent.ActivityWriting
	m.agent.SpinnerFrame = 0

	view := m.activityView()
	require.Contains(t, view, "Writing")
	require.Equal(t, 2, lipgloss.Height(view), "activity view should include top gap")
}

func TestInputStaysFocusedDuringAgentTurn(t *testing.T) {
	m := testInputModel(t)
	m.input.SetValue("hello")
	updated, _ := m.Update(keyEnter())
	m = updated.(Model)

	require.True(t, m.agent.Busy)
	require.True(t, m.input.Focused())
}

func TestSubmitStartsAgentActivity(t *testing.T) {
	m := withActiveTestModel(New())
	m.width = 80
	m.height = 24
	m.ready = true
	m = m.syncLayout(false)

	m.input.SetValue("hello")
	updated, cmd := m.Update(keyEnter())
	m = updated.(Model)

	require.NotNil(t, cmd)
	require.True(t, m.agent.Busy)
	require.Equal(t, agent.ActivityConnecting, m.agent.Activity)
	require.NotEmpty(t, m.activityView())
}

func TestActivityProgression(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 24
	m.ready = true
	m = m.beginAgentTurn()

	updated, _ := m.Update(agentEventMsg{event: agent.ActivityEvent(agent.ActivityReading)})
	m = updated.(Model)
	require.Equal(t, agent.ActivityReading, m.agent.Activity)

	updated, _ = m.Update(agentEventMsg{event: agent.TurnDoneEvent(provider.TurnResult{Content: "done"})})
	m = updated.(Model)
	require.False(t, m.agent.Busy)
	require.Equal(t, agent.ActivityIdle, m.agent.Activity)
	require.Len(t, m.messages, 1)
	require.Equal(t, constants.MessageAI, m.messages[0].kind)
}

func TestBeginAgentTurnSwapsInputMarginForActivity(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 24
	m.ready = true
	idle := m.syncLayout(false)
	idleChrome := idle.layout.ChromeH
	idleVP := idle.content.Height()

	busy := idle.beginAgentTurn().syncLayout(true)

	require.NotEmpty(t, busy.activityView())
	require.Equal(t, idleChrome+1, busy.layout.ChromeH, "activity line adds one chrome row below content gap")
	require.Equal(t, idleVP-1, busy.content.Height())
}

func TestAgentThinkingDeltaHiddenWhenDisabled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	disabled := false
	require.NoError(t, settings.Save(settings.Settings{
		SyncInterval: "24h",
		ShowThinking: &disabled,
	}))

	m := New()
	m = m.beginAgentTurn()
	updated, _ := m.Update(agentEventMsg{event: agent.ThinkingDeltaEvent("hidden")})
	m = updated.(Model)
	require.Empty(t, m.messages)
}

func TestAgentThinkingDeltaRendersDimmed(t *testing.T) {
	m := New()
	m.width = 80
	m = m.beginAgentTurn()

	updated, _ := m.Update(agentEventMsg{event: agent.ThinkingDeltaEvent("reasoning chunk")})
	m = updated.(Model)
	require.Len(t, m.messages, 1)
	require.Equal(t, constants.MessageThinking, m.messages[0].kind)
	require.Equal(t, "reasoning chunk", m.messages[0].text)

	updated, _ = m.Update(agentEventMsg{event: agent.ThinkingDeltaEvent(" more")})
	m = updated.(Model)
	require.Equal(t, "reasoning chunk more", m.messages[0].text)

	updated, _ = m.Update(agentEventMsg{event: agent.ResponseDeltaEvent("answer")})
	m = updated.(Model)
	require.Len(t, m.messages, 2)
	require.Equal(t, constants.MessageAI, m.messages[1].kind)
}

func TestAgentPhaseDelaysAreOrdered(t *testing.T) {
	require.Positive(t, agent.PhaseDelay)
	require.Less(t, spinnerInterval, agent.PhaseDelay)
	require.GreaterOrEqual(t, len(agent.TurnPhases), 2)
}
