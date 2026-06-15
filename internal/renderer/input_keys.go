package renderer

import (
	tea "charm.land/bubbletea/v2"
	"github.com/riipandi/elph/internal/inputui"
)

func (m Model) handleInputWordDelete(msg tea.Msg) (Model, bool) {
	if !m.input.Focused() {
		m.inputPendingEsc = false
		return m, false
	}

	if payload := csiPayload(msg); payload != "" {
		if msg, ok := inputui.WordDeleteMsgFromCSI(payload); ok {
			m.inputPendingEsc = false
			m.input, _ = m.input.Update(msg)
			return m, true
		}
	}

	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, false
	}

	if inputui.IsInputEscapeKey(key) && !m.shell.Running && !m.modelSelectorActive() {
		m.inputPendingEsc = true
		return m, true
	}
	if m.inputPendingEsc && inputui.IsBackspaceKey(key) {
		m.inputPendingEsc = false
		m.input, _ = m.input.Update(inputui.DeleteWordBackwardKeyMsg())
		return m, true
	}
	m.inputPendingEsc = false

	if msg, ok := inputui.WordDeleteMsgFromKey(key); ok {
		m.input, _ = m.input.Update(msg)
		return m, true
	}

	return m, false
}

func isInputEditingKey(msg tea.Msg) bool {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return false
	}
	return !isContentScrollKey(key)
}