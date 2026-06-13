package renderer

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
)

func contentViewportKeyMap() viewport.KeyMap {
	return viewport.KeyMap{
		Up:         key.NewBinding(key.WithKeys("shift+up")),
		Down:       key.NewBinding(key.WithKeys("shift+down")),
		Left:       key.NewBinding(key.WithKeys("shift+left")),
		Right:      key.NewBinding(key.WithKeys("shift+right")),
		PageUp:     key.NewBinding(key.WithKeys("pgup")),
		PageDown:   key.NewBinding(key.WithKeys("pgdown")),
		HalfPageUp: key.NewBinding(),
		HalfPageDown: key.NewBinding(),
	}
}

func isContentScrollKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyShiftUp, tea.KeyShiftDown, tea.KeyShiftLeft, tea.KeyShiftRight,
		tea.KeyPgUp, tea.KeyPgDown:
		return true
	}
	return false
}