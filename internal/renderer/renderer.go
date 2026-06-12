package renderer

import (
	"fmt"
	"os"

	"github.com/riipandi/elph/internal/tui"
)

// Render starts the TUI application using Bubble Tea.
func Render() {
	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
