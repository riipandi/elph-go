package renderer

import (
	"github.com/riipandi/elph/internal/constants"
	"github.com/riipandi/elph/internal/runtime"
)

func shellDetailStatus(result *runtime.ShellResult, running bool) constants.DetailStatus {
	if running || result == nil {
		return constants.DetailStatusRunning
	}
	if result.Cancelled {
		return constants.DetailStatusWarning
	}
	if result.Err != nil || result.ExitCode != 0 {
		return constants.DetailStatusError
	}
	return constants.DetailStatusSuccess
}
