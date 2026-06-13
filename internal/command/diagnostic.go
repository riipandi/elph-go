package command

import (
	"fmt"
	"strings"

	"github.com/riipandi/elph/internal/runtime"
	"github.com/riipandi/elph/internal/tools"
	"github.com/riipandi/elph/pkg/tool"
)

const (
	DiagnosticListTools    = "diagnostic:list-tools"
	DiagnosticSystemPrompt = "diagnostic:system-prompt"
	DiagnosticOpenLog      = "diagnostic:open-log"
	DiagnosticDebug        = "diagnostic:debug"
)

func diagnosticListTools(Context, string) string {
	var b strings.Builder
	b.WriteString("Available tools:\n")

	for _, def := range tool.All() {
		fmt.Fprintf(&b, "  %s (%s) — %s\n", def.Name, def.DefaultApproval, def.Description)
	}

	for _, def := range tools.Diagnostic() {
		fmt.Fprintf(&b, "  %s (%s) — %s\n", def.Name, def.DefaultApproval, def.Description)
	}

	return strings.TrimRight(b.String(), "\n")
}

func diagnosticSystemPrompt(ctx Context, _ string) string {
	if strings.TrimSpace(ctx.SystemPrompt) == "" {
		return fmt.Sprintf("/%s: not yet implemented", DiagnosticSystemPrompt)
	}

	var b strings.Builder
	b.WriteString("System prompt:\n\n")
	b.WriteString(ctx.SystemPrompt)
	return b.String()
}

func diagnosticOpenLog(ctx Context, _ string) string {
	if ctx.LogPath == "" {
		return fmt.Sprintf("/%s: session log not available", DiagnosticOpenLog)
	}

	content, err := runtime.ReadLogTail(ctx.LogPath, 0)
	if err != nil {
		return fmt.Sprintf("/%s: %v", DiagnosticOpenLog, err)
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Sprintf("Session log (%s) is empty.", ctx.LogPath)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Session log: %s\n\n", ctx.LogPath)
	b.WriteString(content)
	return b.String()
}

func diagnosticDebug(Context, string) string {
	return notImplemented(DiagnosticDebug)(Context{}, "")
}
