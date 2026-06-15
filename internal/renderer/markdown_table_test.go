package renderer

import (
	"strings"
	"testing"

	"github.com/riipandi/elph/internal/uiconst"
	"github.com/stretchr/testify/require"
)

func TestWideTableKeepsSingleRowPerLine(t *testing.T) {
	m := testModel()
	table := "| Trigger | Handler |\n|-------------------------|------------------------------------------------------------------------|\n| Normal chat input | session.Session.StartTurn → agent.RunTurn |"
	plain := stripANSI(m.renderMessage(message{text: table, kind: uiconst.MessageAI}))
	require.Contains(t, plain, "Trigger")
	require.Contains(t, plain, "Handler")
	require.Contains(t, plain, "│")
	require.NotContains(t, plain, "StartTurn\n")
	lines := strings.Split(plain, "\n")
	dataRows := 0
	for _, line := range lines {
		if strings.Contains(line, "Normal chat input") {
			dataRows++
		}
	}
	require.Equal(t, 1, dataRows, "table data row should stay on one line")
}
