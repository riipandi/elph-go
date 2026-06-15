package renderer

import (
	"fmt"
	"strings"

	"github.com/riipandi/elph/internal/inputui"
)

func (m Model) pasteIDForEdit() (int, bool) {
	val := m.input.Value()
	if id, ok := pasteIDAtOffset(val, m.inputCursorOffset()); ok {
		if _, ok := m.inputPastes[id]; ok {
			return id, true
		}
	}
	if id, ok := pasteIDOnLine(val, m.input.Line()); ok {
		if _, ok := m.inputPastes[id]; ok {
			return id, true
		}
	}
	ids := pasteIDsInValue(val)
	if len(ids) == 1 {
		if _, ok := m.inputPastes[ids[0]]; ok {
			return ids[0], true
		}
	}
	return 0, false
}

func (m Model) restoreInputCursorLineCol(line, col int) Model {
	lines := splitInputLines(m.input.Value())
	if len(lines) == 0 {
		m.input.MoveToBegin()
		return m
	}
	if line < 0 {
		line = 0
	}
	if line >= len(lines) {
		line = len(lines) - 1
	}
	maxCol := len([]rune(lines[line]))
	if col > maxCol {
		col = maxCol
	}
	if col < 0 {
		col = 0
	}

	m.input.MoveToBegin()
	for m.input.Line() < line {
		m.input.CursorDown()
	}
	m.input.SetCursorColumn(col)
	return m
}

func (m Model) setInputCursorByteOffset(off int) Model {
	val := m.input.Value()
	if len(val) == 0 {
		return m
	}
	off = max(0, min(off, len(val)))
	targetLine := countNewlinesBefore(val, off)
	lineStart := lastNewlineBefore(val, off) + 1
	targetCol := len([]rune(val[lineStart:off]))
	return m.restoreInputCursorLineCol(targetLine, targetCol)
}

func (m Model) placeCursorOnPasteToken(id int) Model {
	token := pasteToken(id)
	idx := strings.Index(m.input.Value(), token)
	if idx < 0 {
		return m
	}
	return m.setInputCursorByteOffset(idx + len(token))
}

func (m Model) pruneInputPastes() Model {
	inputui.PrunePastes(m.input.Value(), m.inputPastes)
	return m
}

func (m Model) insertTextAtCursor(text string) Model {
	val := m.input.Value()
	offset := m.inputCursorOffset()
	if offset < 0 {
		offset = 0
	}
	if offset > len(val) {
		offset = len(val)
	}
	m.input.SetValue(val[:offset] + text + val[offset:])
	return m
}

func (m Model) insertCollapsedPaste(text string) Model {
	if m.inputPastes == nil {
		m.inputPastes = make(map[int]string)
	}
	id := m.nextPasteID
	m.nextPasteID++
	m.inputPastes[id] = text
	token := pasteToken(id)
	m = m.insertTextAtCursor(token)
	return m.placeCursorOnPasteToken(id)
}

func (m Model) replacePasteToken(id int, text string) Model {
	out, replaced := inputui.ReplacePasteToken(m.input.Value(), id, text)
	if replaced {
		m.input.SetValue(out)
	}
	return m
}

func (m Model) clearInputPastes() Model {
	m.inputPastes = nil
	m.nextPasteID = 0
	m.pasteEditor = pasteEditorState{}
	return m
}

func (m Model) pasteHintView() string {
	if m.pasteEditorActive() {
		return ""
	}
	id, ok := m.pasteIDForEdit()
	if !ok {
		return ""
	}
	lines := pasteLineCount(m.inputPastes[id])
	return dimStyle.Render(fmt.Sprintf("Pasted block · %d lines · ctrl+o to preview/edit", lines))
}

func splitInputLines(s string) []string {
	return strings.Split(s, "\n")
}

func countNewlinesBefore(s string, off int) int {
	n := 0
	for i := 0; i < off && i < len(s); i++ {
		if s[i] == '\n' {
			n++
		}
	}
	return n
}

func lastNewlineBefore(s string, off int) int {
	for i := off - 1; i >= 0; i-- {
		if s[i] == '\n' {
			return i
		}
	}
	return -1
}

