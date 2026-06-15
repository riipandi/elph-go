package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate elph configuration",
	Long: `Check settings, provider files, and the active model selection.

Runs staged checks with a live spinner in interactive terminals and prints a
styled report with sectioned findings. Exits with status 1 when any error is
found.`,
	RunE: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	report, err := runDoctorUI(workDir)
	if err != nil {
		return err
	}

	if report.hasFailures() {
		return fmt.Errorf("configuration check failed")
	}
	return nil
}
