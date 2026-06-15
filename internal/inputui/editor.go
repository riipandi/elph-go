package inputui

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
)

// PasteEditorState tracks an in-progress paste preview/editor overlay.
type PasteEditorState struct {
	Active           bool
	PasteID          int
	Input            textarea.Model
	SavedInputLine   int
	SavedInputColumn int
}

// NewPasteEditor builds a focused textarea for editing collapsed paste content.
func NewPasteEditor(text string, width, maxHeight int, styles func() textarea.Styles) textarea.Model {
	ta := textarea.New()
	ta.SetValue(text)
	ta.Prompt = ""
	ta.Placeholder = ""
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetStyles(styles())
	ta.KeyMap.InsertNewline.SetKeys("ctrl+j", "shift+enter")
	ConfigureKeyMap(&ta)
	ta.SetWidth(max(width, 1))
	lines := PasteLineCount(text)
	h := min(max(lines, 1), max(maxHeight, 1))
	ta.SetHeight(h)
	ta.Focus()
	return ta
}

// PasteEditorRows returns display rows for the paste editor content.
func PasteEditorRows(val string, width int) int {
	if val == "" {
		return 1
	}
	w := max(width, 1)
	rows := 0
	for _, line := range strings.Split(val, "\n") {
		rows += WrappedRows(line, w)
	}
	return max(rows, 1)
}