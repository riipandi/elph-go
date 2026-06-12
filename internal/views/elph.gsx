package views

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grindlemire/go-tui"
	"github.com/riipandi/elph/internal/config"
)

type elphApp struct {
	app           *tui.App
	textarea      *tui.TextArea
	termW         int
	gitBranch     string
	gitHash       string
	sessionID     string
	modelName     string
	workDir       string
	welcomeOK     bool
	lastCtrlCTime time.Time
}

func MainApplication(width int) *elphApp {
	wd, _ := os.Getwd()
	p := &elphApp{
		termW:     width,
		gitBranch: "main",
		gitHash:   config.BuildHash,
		sessionID: "019ebc71-19ec-7cf7-947e-5a10e78e01ba",
		modelName: "DeepSeek V4 Flash Free [opencode]",
		workDir:   wd,
		welcomeOK: false,
	}
	p.textarea = tui.NewTextArea(
		tui.WithTextAreaAutoFocus(true),
		tui.WithTextAreaPlaceholder("Type a message or /command..."),
		tui.WithTextAreaWidth(p.getW()-4),
		tui.WithTextAreaBorder(tui.BorderNone),
		tui.WithTextAreaOnSubmit(p.onSubmit),
	)
	return p
}

func (p *elphApp) BindApp(app *tui.App) {
	p.app = app
	p.textarea.BindApp(app)
}

func (p *elphApp) getW() int {
	if p.termW <= 0 {
		return 80
	}
	return p.termW
}

func (p *elphApp) emitWelcome() {
	p.welcomeOK = true
	p.app.PrintAboveln("    %s    Welcome to Elph %s", ElphLogo1, config.AppVersion)
	p.app.PrintAboveln("    %s    Send /changelog to show version history.", ElphLogo2)
	p.app.PrintAboveln("")
	p.app.PrintAboveln("  %-11s %s", "Directory:", p.workDir)
	p.app.PrintAboveln("  %-11s %s", "Model:", fmt.Sprintf("%s (000 available)", p.modelName))
	p.app.PrintAboveln("  %-11s %s", "Stats:", "00 ext, 00 commands, 00 skills, 00 tools")
	p.app.PrintAboveln("")
	p.app.PrintAboveln("Tip: Use --no-session for ephemeral mode — no session file is saved, useful for")
	p.app.PrintAboveln("one-off queries.")
	p.app.PrintAboveln("")
	p.app.PrintAboveln("MCP: 0 servers connected (000 tools)")
	p.app.PrintAboveln("")
}

func (p *elphApp) onSubmit(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	if text == "/q" || text == "/quit" {
		if p.app != nil {
			p.app.Stop()
		}
		return
	}

	p.textarea.Clear()

	if p.app == nil {
		return
	}

	if !p.welcomeOK {
		p.emitWelcome()
	}

	p.app.PrintAboveln("> You: %s", text)
	p.app.PrintAboveln("> Elph: %s", "You said: "+text)
	p.app.PrintAboveln("")
}

func (p *elphApp) KeyMap() tui.KeyMap {
	km := p.textarea.KeyMap()
	km = append(km,
		tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.KeyCtrlQ, func(ke tui.KeyEvent) {
			if time.Since(p.lastCtrlCTime) < 2*time.Second {
				ke.App().Stop()
			} else {
				p.lastCtrlCTime = time.Now()
				p.app.PrintAboveln("Press Ctrl+Q again to quit.")
			}
		}),
	)
	return km
}

func (p *elphApp) Watchers() []tui.Watcher {
	return p.textarea.Watchers()
}

templ (p *elphApp) Render() {
	<div class="flex-col" width={p.getW()}>
		// Input prompt
		<div class="flex items-start gap-1" width={p.getW() - 2}>
			<span class="text-green font-bold">{">"}</span>
			@p.textarea
		</div>

		@StatusBars(p.getW(), p.modelName, p.gitBranch, p.sessionID, p.workDir)
	</div>
}
