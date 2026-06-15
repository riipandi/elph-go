package toolinteract

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/riipandi/elph/pkg/tools"
)

const (
	ApprovalChoiceOnce    = "once"
	ApprovalChoiceSession = "session"
	ApprovalChoiceDeny    = "deny"
	DialogChoiceCancel    = "cancel"

	MaxApprovalDescriptionLines = 6

	approvalChoiceOnce    = ApprovalChoiceOnce
	approvalChoiceSession = ApprovalChoiceSession
	approvalChoiceDeny    = ApprovalChoiceDeny
	dialogChoiceCancel    = DialogChoiceCancel

	maxApprovalDescriptionLines = MaxApprovalDescriptionLines
)

// ApprovalPromptText returns the approval dialog headline.
func ApprovalPromptText(name string) string {
	return fmt.Sprintf("Allow %s?", name)
}

// FormatApprovalDescription renders tool args for the approval dialog body.
func FormatApprovalDescription(name string, args map[string]any, width int) string {
	var b strings.Builder
	switch name {
	case tools.Bash:
		if cmd, ok := StringArg(args, "command"); ok {
			b.WriteString(cmd)
		}
		if desc, ok := StringArg(args, "description"); ok {
			if b.Len() > 0 {
				b.WriteString("\n\n")
			}
			b.WriteString(desc)
		}
	default:
		for _, key := range SortedArgKeys(args) {
			if val, ok := StringArg(args, key); ok {
				if b.Len() > 0 {
					b.WriteString("\n")
				}
				b.WriteString(key)
				b.WriteString(": ")
				b.WriteString(val)
			}
		}
	}
	return ClampMultilineText(strings.TrimSpace(b.String()), width, maxApprovalDescriptionLines)
}

// ClampMultilineText wraps and truncates multiline text to a line budget.
func ClampMultilineText(text string, width, maxLines int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if maxLines <= 0 {
		maxLines = 1
	}

	paragraphs := strings.Split(text, "\n")
	var lines []string
	for pi, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			if len(lines) > 0 && lines[len(lines)-1] != "" {
				lines = append(lines, "")
			}
			continue
		}
		wrapped := para
		if width > 0 {
			wrapped = ansi.Hardwrap(ansi.Wordwrap(para, width, ""), width, false)
		}
		wrappedLines := strings.Split(wrapped, "\n")
		for i, line := range wrappedLines {
			lines = append(lines, line)
			if len(lines) >= maxLines {
				more := i < len(wrappedLines)-1 || pi < len(paragraphs)-1
				return truncateApprovalLines(lines, maxLines, width, more)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func truncateApprovalLines(lines []string, maxLines, width int, more bool) string {
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		more = true
	}
	if !more {
		return strings.Join(lines, "\n")
	}
	last := lines[maxLines-1]
	if width > 0 {
		last = ansi.Truncate(last, max(1, width-1), "…")
	} else if !strings.HasSuffix(last, "…") {
		last += "…"
	}
	lines[maxLines-1] = last
	return strings.Join(lines, "\n")
}

// ApprovalSignature builds a stable key for approval denial caching.
func ApprovalSignature(req agent.ToolInteractRequest) string {
	name, ok := tools.ResolveName(req.Name)
	if !ok {
		name = req.Name
	}
	var b strings.Builder
	b.WriteString(name)
	if name == tools.Bash {
		if cmd, ok := BashCommandArg(req.Args); ok {
			b.WriteByte(0)
			b.WriteString(cmd)
		}
		return b.String()
	}
	for _, key := range SortedArgKeys(req.Args) {
		if val, ok := StringArg(req.Args, key); ok {
			b.WriteByte(0)
			b.WriteString(key)
			b.WriteByte('=')
			b.WriteString(val)
		}
	}
	return b.String()
}

// RecordApprovalDenial caches a denied approval so it is not re-prompted in the same turn.
func RecordApprovalDenial(b *Bridge, resp agent.ToolInteractResponse, req agent.ToolInteractRequest) {
	if b == nil || req.Kind != agent.ToolInteractApproval || resp.Approved || resp.Cancelled {
		return
	}
	if b.DeniedApprovals == nil {
		b.DeniedApprovals = make(map[string]struct{})
	}
	b.DeniedApprovals[ApprovalSignature(req)] = struct{}{}
}

// NormalizeApprovalChoice maps form labels to canonical approval choices.
func NormalizeApprovalChoice(raw string) string {
	return normalizeApprovalChoice(raw)
}

func normalizeApprovalChoice(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case approvalChoiceOnce, "allow once":
		return approvalChoiceOnce
	case approvalChoiceSession, "allow for session":
		return approvalChoiceSession
	case approvalChoiceDeny:
		return approvalChoiceDeny
	case dialogChoiceCancel:
		return dialogChoiceCancel
	default:
		if raw == "" {
			return approvalChoiceOnce
		}
		return raw
	}
}

// IsDialogCancelChoice reports whether a form value selects cancel.
func IsDialogCancelChoice(raw string) bool {
	return normalizeDialogChoice(raw) == dialogChoiceCancel
}

func normalizeDialogChoice(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), dialogChoiceCancel) {
		return dialogChoiceCancel
	}
	return strings.TrimSpace(raw)
}