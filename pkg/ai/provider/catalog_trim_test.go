package provider

import (
	"testing"

	"github.com/riipandi/elph/internal/constants"
	"github.com/stretchr/testify/require"
)

func TestTrimCatalogForRuntimeSlimsInactiveModels(t *testing.T) {
	enabled := true
	full := Catalog{Providers: []RegisteredProvider{{
		ID: "openai",
		Config: FileConfig{
			Enabled: &enabled,
			Name:    "OpenAI",
		},
		Models: []ResolvedModel{
			{
				ID:         "gpt-4o",
				Enabled:    true,
				Name:       "GPT-4o",
				ProviderID: "openai",
				ThinkingLevelMap: map[constants.ThinkingLevel]ThinkingMapValue{
					constants.ThinkingHigh: {},
				},
				Headers: map[string]string{"x-test": "1"},
			},
			{
				ID:      "gpt-4o-mini",
				Enabled: true,
				Name:    "Mini",
			},
		},
	}}}

	trimmed := TrimCatalogForRuntime(full, "openai", "gpt-4o")
	require.Len(t, trimmed.Providers, 1)
	require.Len(t, trimmed.Providers[0].Models, 2)

	active := trimmed.Providers[0].Models[0]
	require.NotNil(t, active.Headers)
	require.NotEmpty(t, active.ThinkingLevelMap)

	inactive := trimmed.Providers[0].Models[1]
	require.Nil(t, inactive.Headers)
	require.Empty(t, inactive.ThinkingLevelMap)
}

func TestTotalEnabledModels(t *testing.T) {
	enabled := true
	disabled := false
	catalog := Catalog{Providers: []RegisteredProvider{
		{
			Config: FileConfig{Enabled: &enabled},
			Models: []ResolvedModel{
				{Enabled: true},
				{Enabled: false},
			},
		},
		{
			Config: FileConfig{Enabled: &disabled},
			Models: []ResolvedModel{{Enabled: true}},
		},
	}}
	require.Equal(t, 1, catalog.TotalEnabledModels())
}
