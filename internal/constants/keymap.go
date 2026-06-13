package constants

import "github.com/charmbracelet/bubbletea"

// KeyBinding defines a single keybinding.
type KeyBinding struct {
	Type   tea.KeyType // Key type (e.g. tea.KeyCtrlC)
	Rune   rune        // For letter keys (e.g. 'a' for Ctrl+A)
	Action KeyAction
	Label  string
}

// KeyAction represents the action triggered by a keybinding.
type KeyAction string

const (
	ActionQuit       KeyAction = "quit"
	ActionExit       KeyAction = "exit"
	ActionSwitchMode KeyAction = "switch_mode"
	ActionCycleThink KeyAction = "cycle_thinking"
	ActionSubmit     KeyAction = "submit"
	ActionNewline    KeyAction = "newline"
	ActionClearInput KeyAction = "clear_input"
	ActionCopy       KeyAction = "copy"
	ActionExport     KeyAction = "export"
)

// DefaultKeyBindings defines all keybindings for the TUI.
var DefaultKeyBindings = []KeyBinding{
	{Type: tea.KeyCtrlC, Action: ActionQuit, Label: "Cancel / Quit"},
	{Type: tea.KeyCtrlX, Action: ActionQuit, Label: "Cancel / Quit"},
	{Type: tea.KeyCtrlD, Action: ActionExit, Label: "Exit application"},
	{Type: tea.KeyCtrlA, Action: ActionSwitchMode, Label: "Switch agent mode"},
	{Type: tea.KeyShiftTab, Action: ActionCycleThink, Label: "Cycle thinking level"},
	{Type: tea.KeyEnter, Action: ActionSubmit, Label: "Send message"},
	{Type: tea.KeyCtrlJ, Action: ActionNewline, Label: "Insert newline in input"},
	{Type: tea.KeyCtrlY, Action: ActionCopy, Label: "Copy last message"},
}

// KeyBindingsByAction returns a map of action to keybinding for quick lookup.
func KeyBindingsByAction() map[KeyAction]KeyBinding {
	result := make(map[KeyAction]KeyBinding, len(DefaultKeyBindings))
	for _, kb := range DefaultKeyBindings {
		if _, exists := result[kb.Action]; !exists {
			result[kb.Action] = kb
		}
	}
	return result
}

// KeyBindingLabels returns a list of "key: description" strings for display.
func KeyBindingLabels() []string {
	labels := make([]string, 0, len(DefaultKeyBindings))
	for _, kb := range DefaultKeyBindings {
		labels = append(labels, kb.Label)
	}
	return labels
}
