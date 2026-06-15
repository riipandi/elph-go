package toolinteract

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/riipandi/elph/pkg/core/agent"
)

// ShortcutResponse maps keyboard shortcuts to a tool-interact response.
func ShortcutResponse(req agent.ToolInteractRequest, msg tea.KeyPressMsg) (agent.ToolInteractResponse, bool) {
	switch req.Kind {
	case agent.ToolInteractApproval:
		switch strings.ToLower(msg.String()) {
		case "y":
			return agent.ToolInteractResponse{Approved: true}, true
		case "a":
			return agent.ToolInteractResponse{Approved: true, AllowSession: true}, true
		case "n":
			return agent.ToolInteractResponse{Approved: false}, true
		case "1":
			return agent.ToolInteractResponse{Approved: true}, true
		case "2":
			return agent.ToolInteractResponse{Approved: true, AllowSession: true}, true
		case "3":
			return agent.ToolInteractResponse{Approved: false}, true
		case "4", "c":
			return agent.ToolInteractResponse{Cancelled: true}, true
		}
	case agent.ToolInteractAskUser:
		fields := ParseAskUserArgs(req.Args)
		opts := fields.Options
		if len(opts) > 0 && len(msg.Text) == 1 {
			if n, err := strconv.Atoi(msg.Text); err == nil && n >= 1 && n <= len(opts)+1 {
				if n == len(opts)+1 {
					return agent.ToolInteractResponse{Cancelled: true}, true
				}
				return agent.ToolInteractResponse{Answer: opts[n-1]}, true
			}
		}
		if strings.ToLower(msg.String()) == "c" && len(opts) > 0 {
			return agent.ToolInteractResponse{Cancelled: true}, true
		}
	}
	return agent.ToolInteractResponse{}, false
}

// AskUserChoiceEnterResponse submits the highlighted option when Enter is pressed
// on the choice select. Huh would otherwise advance to the custom input field.
func AskUserChoiceEnterResponse(form *huh.Form, req agent.ToolInteractRequest, msg tea.KeyPressMsg) (agent.ToolInteractResponse, bool) {
	if form == nil || req.Kind != agent.ToolInteractAskUser {
		return agent.ToolInteractResponse{}, false
	}
	fields := ParseAskUserArgs(req.Args)
	if len(fields.Options) == 0 || !fields.AllowCustom {
		return agent.ToolInteractResponse{}, false
	}
	if msg.Code != tea.KeyEnter && msg.String() != "enter" {
		return agent.ToolInteractResponse{}, false
	}
	focused := form.GetFocusedField()
	if focused == nil || focused.GetKey() != "choice" {
		return agent.ToolInteractResponse{}, false
	}
	choice := askUserChoiceSelection(form, fields.Options)
	if choice == "" {
		return agent.ToolInteractResponse{}, false
	}
	if IsDialogCancelChoice(choice) {
		return agent.ToolInteractResponse{Cancelled: true}, true
	}
	return agent.ToolInteractResponse{Answer: choice}, true
}

func askUserChoiceSelection(form *huh.Form, options []string) string {
	if choice := formFieldString(form, "choice"); choice != "" {
		return choice
	}
	type choiceHovered interface {
		GetKey() string
		Hovered() (string, bool)
	}
	if focused := form.GetFocusedField(); focused != nil {
		if sel, ok := focused.(choiceHovered); ok && sel.GetKey() == "choice" {
			if hovered, ok := sel.Hovered(); ok && strings.TrimSpace(hovered) != "" {
				return hovered
			}
		}
	}
	if len(options) > 0 {
		return options[0]
	}
	return ""
}