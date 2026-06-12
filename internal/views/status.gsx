package views

import (
	"fmt"
	"path/filepath"

	"github.com/grindlemire/go-tui"
)

templ MCPStatus(w int) {
	<div class="flex" width={w - 2}>
		<span class="font-dim">MCP: 0 servers connected (000 tools)</span>
	</div>
}

// StatusBars renders the two-line bottom status bar.
templ StatusBars(w int, modelName string, branch string, sessionID string, workDir string) {
	<div class="flex justify-between" width={w - 2}>
		<span class="font-dim">{fmt.Sprintf("%s | opencode | T: high | IMG", modelName)}</span>
		<span class="font-dim">$0.00 | 0.0% (262k)</span>
	</div>
	<div class="flex justify-between" width={w - 2}>
		<span class="font-dim">{fmt.Sprintf("%s [%s]", filepath.Base(workDir), shortSession(sessionID))}</span>
		<span class="font-dim">{fmt.Sprintf("turn: 0 | %s [+00 -00]", branch)}</span>
	</div>
}
