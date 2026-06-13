package renderer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/riipandi/elph/internal/constants"
)

func TestActivityViewHiddenWhenIdle(t *testing.T) {
	m := New()
	m.width = 80
	if m.activityView() != "" {
		t.Fatal("activity view should be empty when idle")
	}
}

func TestActivityViewShowsLabel(t *testing.T) {
	m := New()
	m.width = 80
	m.activity = constants.ActivityWriting
	m.spinnerFrame = 0

	view := m.activityView()
	if !strings.Contains(view, "Writing") {
		t.Fatalf("expected Writing label, got %q", view)
	}
	if lipgloss.Height(view) != 2 {
		t.Fatalf("activity view should be 2 lines (margin + label), got height %d", lipgloss.Height(view))
	}
}

func TestSubmitStartsAgentActivity(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 24
	m.ready = true
	m = m.syncLayout(false)

	m.input.SetValue("hello")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("expected agent turn command after submit")
	}
	if !m.busy {
		t.Fatal("expected busy after submit")
	}
	if m.activity != constants.ActivityConnecting {
		t.Fatalf("activity %q, want Connecting", m.activity)
	}
	if m.activityView() == "" {
		t.Fatal("activity indicator should be visible")
	}
}

func TestActivityProgression(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 24
	m.ready = true
	m = m.beginAgentTurn()

	updated, _ := m.Update(ActivityMsg{Activity: constants.ActivityReading})
	m = updated.(Model)
	if m.activity != constants.ActivityReading {
		t.Fatalf("activity %q, want Reading", m.activity)
	}

	updated, _ = m.Update(AgentDoneMsg{Response: "done"})
	m = updated.(Model)
	if m.busy {
		t.Fatal("expected not busy after agent done")
	}
	if m.activity != constants.ActivityIdle {
		t.Fatalf("activity %q, want idle", m.activity)
	}
	if len(m.messages) != 1 || m.messages[0].kind != msgAI {
		t.Fatal("expected AI response message")
	}
}

func TestBeginAgentTurnIncreasesChrome(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 24
	m.ready = true
	idle := m.syncLayout(false)
	idleChrome := idle.chromeH
	idleVP := idle.content.Height

	busy := idle.beginAgentTurn().syncLayout(true)

	if busy.chromeH <= idleChrome {
		t.Fatalf("chrome after beginAgentTurn %d should exceed idle %d", busy.chromeH, idleChrome)
	}
	if busy.content.Height >= idleVP {
		t.Fatalf("viewport should shrink from %d to make room for activity, got %d", idleVP, busy.content.Height)
	}
}

func TestAgentPhaseDelaysAreOrdered(t *testing.T) {
	if agentPhaseDelay <= 0 {
		t.Fatal("agent phase delay must be positive")
	}
	if spinnerInterval >= agentPhaseDelay {
		t.Fatal("spinner should animate faster than phase changes")
	}
	if len(constants.AgentTurnPhases) < 2 {
		t.Fatal("expected multiple activity phases")
	}
}