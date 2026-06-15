package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/spf13/cobra"
)

var providerEnableCmd = &cobra.Command{
	Use:   "enable <provider>",
	Short: "Enable a provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runProviderEnable,
}

var providerDisableCmd = &cobra.Command{
	Use:   "disable <provider>",
	Short: "Disable a provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runProviderDisable,
}

var providerModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage individual models within a provider",
}

var providerModelListCmd = &cobra.Command{
	Use:   "list <provider>",
	Short: "List models for a provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runProviderModelList,
}

var providerModelEnableCmd = &cobra.Command{
	Use:   "enable <provider> <model>",
	Short: "Enable a model",
	Args:  cobra.ExactArgs(2),
	RunE:  runProviderModelEnable,
}

var providerModelDisableCmd = &cobra.Command{
	Use:   "disable <provider> <model>",
	Short: "Disable a model",
	Args:  cobra.ExactArgs(2),
	RunE:  runProviderModelDisable,
}

func runProviderEnable(_ *cobra.Command, args []string) error {
	id := args[0]
	if err := provider.SetProviderEnabled(id, true); err != nil {
		return err
	}
	fmt.Printf("Enabled provider %s\n", id)
	return nil
}

func runProviderDisable(_ *cobra.Command, args []string) error {
	id := args[0]
	if err := provider.SetProviderEnabled(id, false); err != nil {
		return err
	}
	fmt.Printf("Disabled provider %s\n", id)
	return nil
}

func runProviderModelEnable(_ *cobra.Command, args []string) error {
	providerID, modelID := args[0], args[1]
	if err := provider.SetModelEnabled(providerID, modelID, true); err != nil {
		return err
	}
	fmt.Printf("Enabled model %s/%s\n", providerID, modelID)
	return nil
}

func runProviderModelDisable(_ *cobra.Command, args []string) error {
	providerID, modelID := args[0], args[1]
	if err := provider.SetModelEnabled(providerID, modelID, false); err != nil {
		return err
	}
	fmt.Printf("Disabled model %s/%s\n", providerID, modelID)
	return nil
}

func runProviderModelList(_ *cobra.Command, args []string) error {
	catalog, err := provider.LoadCatalog("")
	if err != nil {
		return err
	}

	providerID := args[0]
	reg, ok := catalog.Provider(providerID)
	if !ok {
		return fmt.Errorf("provider %q not found", providerID)
	}

	name := strings.TrimSpace(reg.Config.Name)
	if name == "" {
		name = reg.ID
	}
	fmt.Printf("Models for %s (%s)\n\n", name, reg.ID)

	if len(reg.Models) == 0 {
		fmt.Println("No models configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MODEL\tNAME\tSTATUS")
	for _, model := range reg.Models {
		label := strings.TrimSpace(model.Name)
		if label == "" {
			label = model.ID
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", model.ID, label, enabledStatus(model.Enabled))
	}
	return w.Flush()
}

func enabledStatus(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

func providerStatus(reg provider.RegisteredProvider) string {
	if !provider.ProviderConfigEnabled(reg.Config) {
		return "disabled"
	}
	return "enabled"
}

func initProviderEnableCommands() {
	providerModelCmd.AddCommand(providerModelListCmd)
	providerModelCmd.AddCommand(providerModelEnableCmd)
	providerModelCmd.AddCommand(providerModelDisableCmd)

	providerCmd.AddCommand(providerEnableCmd)
	providerCmd.AddCommand(providerDisableCmd)
	providerCmd.AddCommand(providerModelCmd)
}
