package renderer

import (
	"errors"
	"testing"

	"github.com/riipandi/elph/internal/constants"
	"github.com/riipandi/elph/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestShellDetailStatusTransitions(t *testing.T) {
	require.Equal(t, constants.DetailStatusRunning, shellDetailStatus(nil, true))
	require.Equal(t, constants.DetailStatusSuccess, shellDetailStatus(&runtime.ShellResult{ExitCode: 0}, false))
	require.Equal(t, constants.DetailStatusError, shellDetailStatus(&runtime.ShellResult{ExitCode: 2}, false))
	require.Equal(t, constants.DetailStatusWarning, shellDetailStatus(&runtime.ShellResult{Cancelled: true}, false))
	require.Equal(t, constants.DetailStatusError, shellDetailStatus(&runtime.ShellResult{Err: errors.New("boom")}, false))
}

func TestShellDetailMessageUsesDynamicStatusColors(t *testing.T) {
	m := testModel()
	m.messages = []message{{
		kind:         constants.MessageDetail,
		detailLabel:  "$ ls",
		text:         "file.txt",
		detailStatus: constants.DetailStatusSuccess,
	}}
	success := m.renderMessageAt(0)
	m.messages[0].detailStatus = constants.DetailStatusError
	m.messages[0].renderCache = messageRenderCache{}
	errorState := m.renderMessageAt(0)
	require.NotEqual(t, success, errorState)
}
