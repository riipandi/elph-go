package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/spf13/cobra"
)

var providerForce bool

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage AI provider definitions",
}

var providerConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Create or backfill provider definitions",
	Long: strings.TrimSpace(`
Write starter provider files for OpenAI, Anthropic, OpenCode Zen, OpenCode Go,
DeepSeek, and Kimi under ~/.elph/providers.

New files include per-model reasoning flags, thinkingLevelMap entries, and compat
settings needed for thinking-level controls in the TUI.

Existing files are left untouched unless --force is passed. Without --force,
missing reasoning/thinkingLevelMap/compat fields are backfilled from built-in
templates without overwriting values you already configured.

Set API keys via environment variables referenced in the JSON files:
  OPENAI_API_KEY, ANTHROPIC_API_KEY, OPENCODE_API_KEY,
  DEEPSEEK_API_KEY, MOONSHOT_API_KEY
`),
	RunE: runProviderConnect,
}

var providerUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Refresh provider definitions and model catalogs",
	Long: strings.TrimSpace(`
Update ~/.elph/providers in two steps:

1. Connect/backfill provider files (same as "elph provider connect").
2. Refresh model metadata from models.dev and live /models endpoints.

OpenCode Zen, OpenCode Go, DeepSeek, and Kimi model lists come from live /models
endpoints when API keys are available; models.dev supplies context window,
pricing, and other metadata.

Provider credentials, headers, temperature, thinkingLevelMap, compat, and other
per-model overrides are preserved. New or incomplete models receive reasoning and
thinking metadata from built-in templates when available.

Records the sync time in ~/.elph/settings.json. The TUI checks this timestamp
on startup and auto-syncs when models.syncInterval has elapsed (default 24h).
`),
	RunE: runProviderUpdate,
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured providers and models",
	Long: strings.TrimSpace(`
Show provider files under ~/.elph/providers with API key status and model counts.

API key status is based on whether a credential is configured in the provider
file (env references count as configured even when the variable is unset).
`),
	RunE: runProviderList,
}

func runProviderConnect(cmd *cobra.Command, args []string) error {
	result, err := runConnectWithProgress(providerForce)
	if err != nil {
		return err
	}
	printConnectResult(result)
	return nil
}

func runProviderUpdate(cmd *cobra.Command, args []string) error {
	bootstrap, sync, err := runUpdateWithProgress(providerForce)
	if err != nil {
		return err
	}
	printProviderUpdateResult(bootstrap, sync)
	return nil
}

func runProviderList(cmd *cobra.Command, args []string) error {
	catalog, err := provider.LoadCatalog("")
	if err != nil {
		return err
	}

	dir := catalog.Dir
	if dir == "" {
		var dirErr error
		dir, dirErr = provider.ProvidersDir()
		if dirErr != nil {
			return dirErr
		}
	}
	fmt.Printf("Providers in %s\n\n", dir)

	if len(catalog.Providers) == 0 {
		fmt.Fprintln(os.Stderr, "No providers found. Run: elph provider connect")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROVIDER\tNAME\tSTATUS\tAPI KEY\tMODELS")
	for _, reg := range catalog.Providers {
		name := strings.TrimSpace(reg.Config.Name)
		if name == "" {
			name = reg.ID
		}
		keyStatus := "not set"
		if provider.IsConfigured(reg.Config.APIKey) {
			keyStatus = "configured"
		}
		enabled := provider.EnabledModelCount(reg)
		models := fmt.Sprintf("%d/%d", enabled, len(reg.Models))
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", reg.ID, name, providerStatus(reg), keyStatus, models)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	for _, err := range catalog.Errors {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	return nil
}

func init() {
	providerConnectCmd.Flags().BoolVar(&providerForce, "force", false, "Overwrite existing provider files")
	providerUpdateCmd.Flags().BoolVar(&providerForce, "force", false, "Overwrite existing provider files")

	providerCmd.AddCommand(providerConnectCmd)
	providerCmd.AddCommand(providerUpdateCmd)
	providerCmd.AddCommand(providerListCmd)
	initProviderEnableCommands()
}
