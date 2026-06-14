package renderer

import (
	"errors"
	"testing"

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
	desc := modelsSyncFormDescription([]string{"openai.json", "anthropic.json"})
	require.Contains(t, desc, "openai.json, anthropic.json")
	require.Contains(t, desc, "models.dev")
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
