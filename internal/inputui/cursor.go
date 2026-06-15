package inputui

import "strings"

// CursorByteCol converts a rune column index to a byte offset within a line.
func CursorByteCol(line string, col int) int {
	runes := []rune(line)
	if col > len(runes) {
		col = len(runes)
	}
	if col <= 0 {
		return 0
	}
	return len(string(runes[:col]))
}

// CursorOffset returns the byte offset of a cursor position in a textarea value.
func CursorOffset(val string, line, col int) int {
	lines := strings.Split(val, "\n")
	if line < 0 {
		line = 0
	}
	if line >= len(lines) {
		line = max(len(lines)-1, 0)
	}

	offset := 0
	for i := 0; i < line; i++ {
		offset += len(lines[i]) + 1
	}
	offset += CursorByteCol(lines[line], col)
	if offset > len(val) {
		offset = len(val)
	}
	return offset
}