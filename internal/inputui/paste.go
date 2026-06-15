package inputui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	PasteCollapseMinLines = 4
	PasteCollapseMinRunes = 400
)

var PasteTokenRe = regexp.MustCompile(`\[\[paste:(\d+)\]\]`)

// PasteLineCount returns the number of lines in text.
func PasteLineCount(text string) int {
	if text == "" {
		return 0
	}
	return strings.Count(text, "\n") + 1
}

// ShouldCollapsePaste reports whether pasted text should be collapsed into a token.
func ShouldCollapsePaste(text string) bool {
	if PasteLineCount(text) >= PasteCollapseMinLines {
		return true
	}
	return len([]rune(text)) >= PasteCollapseMinRunes
}

// PasteToken returns the storage token for a paste id.
func PasteToken(id int) string {
	return fmt.Sprintf("[[paste:%d]]", id)
}

// PasteDisplayToken returns the visible collapsed paste label.
func PasteDisplayToken(id int, lines int, pastes map[int]string) string {
	if pastes != nil {
		if text, ok := pastes[id]; ok {
			lines = PasteLineCount(text)
		}
	}
	return fmt.Sprintf("[Pasted: %d lines]", lines)
}

// OverlayPasteTokens replaces paste tokens in a rendered view with display labels.
func OverlayPasteTokens(view, val string, pastes map[int]string) string {
	if len(pastes) == 0 || view == "" {
		return view
	}
	out := view
	for _, loc := range PasteTokenRe.FindAllStringSubmatchIndex(val, -1) {
		if len(loc) < 4 {
			continue
		}
		token := val[loc[0]:loc[1]]
		id, err := strconv.Atoi(val[loc[2]:loc[3]])
		if err != nil {
			continue
		}
		display := PasteDisplayToken(id, 0, pastes)
		out = strings.ReplaceAll(out, token, display)
	}
	return out
}

// DisplayValue replaces paste tokens with display labels.
func DisplayValue(val string, pastes map[int]string) string {
	return PasteTokenRe.ReplaceAllStringFunc(val, func(match string) string {
		sub := PasteTokenRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		id, err := strconv.Atoi(sub[1])
		if err != nil {
			return match
		}
		return PasteDisplayToken(id, 0, pastes)
	})
}

// ExpandPastes substitutes paste tokens with stored text.
func ExpandPastes(val string, pastes map[int]string) string {
	return PasteTokenRe.ReplaceAllStringFunc(val, func(match string) string {
		sub := PasteTokenRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		id, err := strconv.Atoi(sub[1])
		if err != nil {
			return match
		}
		if text, ok := pastes[id]; ok {
			return text
		}
		return match
	})
}

// PasteIDAtOffset returns the paste id when the cursor is on a token.
func PasteIDAtOffset(val string, offset int) (int, bool) {
	for _, loc := range PasteTokenRe.FindAllStringSubmatchIndex(val, -1) {
		if len(loc) < 4 {
			continue
		}
		start, end := loc[0], loc[1]
		if offset >= start && offset <= end {
			id, err := strconv.Atoi(val[loc[2]:loc[3]])
			if err != nil {
				return 0, false
			}
			return id, true
		}
	}
	return 0, false
}

// PasteIDsInValue returns all paste ids referenced in a value.
func PasteIDsInValue(val string) []int {
	var ids []int
	for _, loc := range PasteTokenRe.FindAllStringSubmatchIndex(val, -1) {
		if len(loc) < 4 {
			continue
		}
		id, err := strconv.Atoi(val[loc[2]:loc[3]])
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

// PasteIDOnLine returns the paste id when a line contains only a token.
func PasteIDOnLine(val string, lineIdx int) (int, bool) {
	lines := strings.Split(val, "\n")
	if lineIdx < 0 || lineIdx >= len(lines) {
		return 0, false
	}
	sub := PasteTokenRe.FindStringSubmatch(lines[lineIdx])
	if len(sub) < 2 {
		return 0, false
	}
	id, err := strconv.Atoi(sub[1])
	return id, err == nil
}

// PrunePastes removes stored pastes no longer referenced in the value.
func PrunePastes(val string, pastes map[int]string) {
	if len(pastes) == 0 {
		return
	}
	seen := make(map[int]struct{})
	for _, loc := range PasteTokenRe.FindAllStringSubmatchIndex(val, -1) {
		if len(loc) < 4 {
			continue
		}
		id, err := strconv.Atoi(val[loc[2]:loc[3]])
		if err != nil {
			continue
		}
		seen[id] = struct{}{}
	}
	for id := range pastes {
		if _, ok := seen[id]; !ok {
			delete(pastes, id)
		}
	}
}

// ReplacePasteToken ensures a paste token remains in the value after editing.
func ReplacePasteToken(val string, id int, _ string) (string, bool) {
	token := PasteToken(id)
	replaced := false
	out := PasteTokenRe.ReplaceAllStringFunc(val, func(match string) string {
		sub := PasteTokenRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		matchID, err := strconv.Atoi(sub[1])
		if err != nil || matchID != id {
			return match
		}
		replaced = true
		return token
	})
	return out, replaced
}