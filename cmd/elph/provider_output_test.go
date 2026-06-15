package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/stretchr/testify/require"
)

func captureOutput(fn func()) (stdout, stderr string) {
	oldOut := os.Stdout
	oldErr := os.Stderr
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	os.Stdout = outW
	os.Stderr = errW

	doneOut := make(chan []byte)
	doneErr := make(chan []byte)
	go func() { b, _ := io.ReadAll(outR); doneOut <- b }()
	go func() { b, _ := io.ReadAll(errR); doneErr <- b }()

	fn()

	outW.Close()
	errW.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	return string(<-doneOut), string(<-doneErr)
}

func TestPrintProviderUpdateResultShowsSteps(t *testing.T) {
	stdout, stderr := captureOutput(func() {
		printProviderUpdateResult(
			provider.BootstrapResult{
				Dir:        "/tmp/providers",
				Backfilled: []string{"openai.json"},
				Skipped:    []string{"anthropic.json"},
			},
			provider.UpdateModelsResult{
				Updated: []string{"openai.json"},
				Skipped: []string{"anthropic.json: already up to date"},
			},
		)
	})

	require.Contains(t, stdout, "[1/2] Provider files")
	require.Contains(t, stdout, "[2/2] Model catalogs")
	require.Contains(t, stdout, "+ Synced openai")
	require.Contains(t, stdout, "· Up to date anthropic")
	require.Contains(t, stdout, "Done.")
	require.Empty(t, stderr)
}

func TestPrintProviderUpdateResultHumanizesAPIKeyWarning(t *testing.T) {
	_, stderr := captureOutput(func() {
		printProviderUpdateResult(
			provider.BootstrapResult{Dir: "/tmp/providers"},
			provider.UpdateModelsResult{
				Updated: []string{"kimi.json"},
				Warnings: []string{
					"kimi.json: API key unavailable for live /models (env.MOONSHOT_API_KEY); using models.dev catalog only",
				},
			},
		)
	})

	require.Contains(t, stderr, "kimi: no API key")
	require.Contains(t, stderr, "models.dev only")
}

func TestPrintConnectResultSuggestsNextStep(t *testing.T) {
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printConnectResult(provider.BootstrapResult{
		Dir:     "/tmp/providers",
		Created: []string{"openai.json"},
	})

	w.Close()
	os.Stdout = old
	_, _ = buf.ReadFrom(r)
	stdout := buf.String()

	require.Contains(t, stdout, "Setting up providers")
	require.Contains(t, stdout, "+ Created openai")
	require.Contains(t, stdout, "elph provider update")
}

func TestSplitProviderLogEntry(t *testing.T) {
	file, reason := splitProviderLogEntry("anthropic.json: already up to date")
	require.Equal(t, "anthropic.json", file)
	require.Equal(t, "already up to date", reason)

	file, reason = splitProviderLogEntry("custom.json: provider not in models.dev catalog")
	require.Equal(t, "custom.json", file)
	require.Equal(t, "provider not in models.dev catalog", reason)
}

func TestJoinProviderNamesStripsJSONSuffix(t *testing.T) {
	require.Equal(t, "openai, anthropic", joinProviderNames([]string{"openai.json", "anthropic.json"}))
	require.True(t, strings.Contains(joinProviderNames([]string{"kimi.json"}), "kimi"))
}
