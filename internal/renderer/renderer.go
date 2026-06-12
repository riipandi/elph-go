package renderer

import (
	"fmt"
	"os"

	"github.com/grindlemire/go-tui"
	"github.com/riipandi/elph/internal/views"
	"golang.org/x/term"
)

func Render() {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
	}
	if width > 120 {
		width = 120
	}

	app, err := tui.NewApp(
		tui.WithInlineHeight(5),
		tui.WithRootComponent(views.MainApplication(width)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
