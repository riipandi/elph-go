package command

import "strings"

// RequiresArgs reports whether a slash command expects user-provided arguments.
func RequiresArgs(cmd SlashCommand, ctx Context) bool {
	if len(EffectiveArgs(cmd, ctx)) > 0 {
		return true
	}
	return strings.TrimSpace(cmd.ArgumentHint) != ""
}

// CommandExactMatch reports whether input is an exact /command with no arguments.
func CommandExactMatch(input string, ctx Context) bool {
	trimmed := strings.TrimLeft(input, " \t")
	if !strings.HasPrefix(trimmed, "/") {
		return false
	}
	withoutSlash := strings.TrimPrefix(trimmed, "/")
	if strings.Contains(withoutSlash, " ") {
		return false
	}
	body := strings.TrimSpace(withoutSlash)
	if body == "" {
		return false
	}
	_, ok := Get(strings.ToLower(body), ctx)
	return ok
}

// InputPlaceholderHint returns the ghost hint shown after a fully typed command.
func InputPlaceholderHint(cmd SlashCommand, ctx Context) string {
	if hint := ArgsHint(EffectiveArgs(cmd, ctx)); hint != "" {
		return hint
	}
	return strings.TrimSpace(cmd.ArgumentHint)
}
