package renderer

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/riipandi/elph/internal/constants"
)

func scrollableTestModel(t *testing.T) Model {
	t.Helper()
	m := New()
	m.width = 80
	m.height = 14
	m.ready = true
	for i := range 30 {
		m.messages = append(m.messages, message{
			text: fmt.Sprintf("scroll test line %d", i),
			kind: constants.MessageUser,
		})
	}
	m.contentDirty = true
	return m.syncLayout(false)
}

func TestPlainArrowDoesNotScrollContent(t *testing.T) {
	m := scrollableTestModel(t)
	m.input.Focus()
	offset := m.content.YOffset

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)

	if m.content.YOffset != offset {
		t.Fatalf("plain down arrow scrolled content: %d -> %d", offset, m.content.YOffset)
	}
}

func TestPlainJKHLDoesNotScrollContent(t *testing.T) {
	m := scrollableTestModel(t)
	m.input.Focus()
	offset := m.content.YOffset

	for _, r := range []rune{'j', 'k', 'h', 'l'} {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	if m.content.YOffset != offset {
		t.Fatalf("j/k/h/l scrolled content: %d -> %d", offset, m.content.YOffset)
	}
	if m.input.Value() != "jkhl" {
		t.Fatalf("input value %q, want jkhl", m.input.Value())
	}
}

func TestShiftArrowScrollsContent(t *testing.T) {
	m := scrollableTestModel(t)
	m.content.GotoTop()
	m.input.Focus()
	m.input.SetValue("typed")
	offset := m.content.YOffset

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	m = updated.(Model)

	if m.content.YOffset <= offset {
		t.Fatalf("shift+down should scroll content: %d -> %d", offset, m.content.YOffset)
	}
	if m.input.Value() != "typed" {
		t.Fatalf("shift+down changed input: %q", m.input.Value())
	}
}

func TestShiftArrowScrollsUp(t *testing.T) {
	m := scrollableTestModel(t)
	m.content.GotoBottom()
	m.input.Focus()
	offset := m.content.YOffset

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftUp})
	m = updated.(Model)

	if m.content.YOffset >= offset {
		t.Fatalf("shift+up should scroll content up: %d -> %d", offset, m.content.YOffset)
	}
}