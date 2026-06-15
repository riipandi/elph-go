package renderer

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/stretchr/testify/require"
)

func TestFormatModelsSyncResult(t *testing.T) {
	require.Equal(t, "Model metadata updated: openai.json, anthropic.json", formatModelsSyncResult(modelsSyncDoneMsg{
		result: provider.UpdateModelsResult{Updated: []string{"openai.json", "anthropic.json"}},
	}))
	require.Equal(t, "Model metadata is up to date.", formatModelsSyncResult(modelsSyncDoneMsg{
		result: provider.UpdateModelsResult{},
	}))
	require.Contains(t, formatModelsSyncResult(modelsSyncDoneMsg{
		err: errors.New("network down"),
	}), "Model metadata update failed:")
}

func TestModelsSyncFormDescription(t *testing.T) {
	desc := formatModelsSyncDescription(60)
	require.Contains(t, desc, "updates are available")
	require.NotContains(t, desc, "models.dev")
}

func TestResolveModelsSyncConfirmAccept(t *testing.T) {
	m := New()
	m.ready = true
	m.width = 100
	m.height = 40
	m.content.SetWidth(100)
	m.content.SetHeight(20)
	m.modelsSyncForm = newModelsSyncForm([]string{"openai.json"}, 60)

	updated, cmd := m.resolveModelsSyncConfirm(true)
	require.NotNil(t, cmd)
	require.Nil(t, updated.modelsSyncForm)
	require.True(t, updated.modelsSyncing)
}

func TestResolveModelsSyncConfirmDecline(t *testing.T) {
	m := New()
	m.ready = true
	m.width = 100
	m.height = 40
	m.modelsSyncForm = newModelsSyncForm([]string{"openai.json"}, 60)

	updated, cmd := m.resolveModelsSyncConfirm(false)
	require.Nil(t, cmd)
	require.Nil(t, updated.modelsSyncForm)
	require.Contains(t, stripANSI(updated.messages[len(updated.messages)-1].text), "skipped")
}

func TestOfferModelsSyncOpensHuhForm(t *testing.T) {
	m := New()
	m.ready = true
	m.width = 100
	m.height = 40

	updated, cmd := m.offerModelsSync([]string{"openai.json"})
	require.NotNil(t, cmd)
	require.NotNil(t, updated.modelsSyncForm)
	require.Equal(t, huh.StateNormal, updated.modelsSyncForm.State)
}

func TestModelsSyncDialogMatchesAskUserLayout(t *testing.T) {
	form := newModelsSyncForm([]string{"openai.json", "anthropic.json"}, 60)
	if updated, _ := form.Update(tea.WindowSizeMsg{Width: 100, Height: 40}); updated != nil {
		if f, ok := updated.(*huh.Form); ok {
			form = f
		}
	}

	m := testInputModel(t)
	m.modelsSyncForm = form

	view := stripANSI(m.modelsSyncChromeView())
	require.Contains(t, view, modelsSyncDialogLabel)
	require.Contains(t, view, "Apply model metadata updates")
	require.Contains(t, view, "updates are available")
	require.NotContains(t, view, "models.dev has updates")
	require.Contains(t, view, "Update")
	require.Contains(t, view, "Skip")
	require.Contains(t, view, "y update")
	require.Contains(t, view, "↑/↓")
	require.NotContains(t, view, "↑ up")
}

func TestStartModelsSyncShowsStatusMessage(t *testing.T) {
	m := New()
	m.ready = true
	m.width = 100
	m.height = 40
	m.content.SetWidth(100)
	m.content.SetHeight(20)

	updated, cmd := m.startModelsSync()
	require.NotNil(t, cmd)
	require.True(t, updated.modelsSyncing)
	require.Len(t, updated.messages, 1)
	require.Contains(t, stripANSI(updated.messages[0].text), modelsSyncUpdatingLabel)
}
