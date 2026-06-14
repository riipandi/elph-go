package command

import (
	"testing"

	"github.com/riipandi/elph/internal/prompttemplate"
	"github.com/stretchr/testify/require"
)

func TestRequiresArgs(t *testing.T) {
	ctx := Context{
		PromptTemplates: []prompttemplate.Template{{
			Name:         "identify",
			ArgumentHint: "<focus-area>",
		}},
	}

	openLog, ok := Get(DiagnosticOpenLog, Context{})
	require.True(t, ok)
	require.True(t, RequiresArgs(openLog, Context{}))

	identify, ok := Get("identify", ctx)
	require.True(t, ok)
	require.True(t, RequiresArgs(identify, ctx))

	help, ok := Get("help", Context{})
	require.True(t, ok)
	require.False(t, RequiresArgs(help, Context{}))
}

func TestCompleteInputAddsSpaceForArgumentHint(t *testing.T) {
	ctx := Context{
		PromptTemplates: []prompttemplate.Template{{
			Name:         "identify",
			ArgumentHint: "<focus-area>",
		}},
	}
	cmd, ok := Get("identify", ctx)
	require.True(t, ok)
	require.Equal(t, "/identify ", CompleteInput(cmd, ctx))
}
