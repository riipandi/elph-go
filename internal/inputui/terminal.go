package inputui

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	KittyModShift = 1
	KittyModCtrl  = 4
	XtermModShift = 2
	XtermModCtrl  = 5
)

var (
	kittyEnterModsRe   = regexp.MustCompile(`13;(\d+)u`)
	xtermEnterModsRe   = regexp.MustCompile(`27;(\d+);13`)
	kittyCtrlJRe       = regexp.MustCompile(`^(?:10|106);(\d+)u$`)
	xtermCtrlJRe       = regexp.MustCompile(`^27;(\d+);(?:10|106)~?$`)
	legacyShiftEnterRe = regexp.MustCompile(`13;2~`)
	csiByteListRe      = regexp.MustCompile(`^\[([0-9]+(?: [0-9]+)*)\]$`)
)

type TermFeaturesMsg struct{}

// ActivateTerminalFeaturesSync enables enhanced key reporting before the program starts.
func ActivateTerminalFeaturesSync() {
	_, _ = fmt.Fprint(os.Stdout, ansi.SetModifyOtherKeys2)
}

// EnableTerminalFeatures requests enhanced key reporting for Shift+Enter detection.
func EnableTerminalFeatures() tea.Cmd {
	return func() tea.Msg {
		ActivateTerminalFeaturesSync()
		return TermFeaturesMsg{}
	}
}

// DisableTerminalFeatures restores the terminal modifyOtherKeys state.
func DisableTerminalFeatures() tea.Cmd {
	return func() tea.Msg {
		_, _ = fmt.Fprint(os.Stdout, ansi.ResetModifyOtherKeys)
		return nil
	}
}

// RawCSIPayload extracts the CSI payload from a raw byte slice message.
func RawCSIPayload(msg tea.Msg) (string, bool) {
	v := reflect.ValueOf(msg)
	if v.Kind() != reflect.Slice || v.Type().Elem().Kind() != reflect.Uint8 {
		return "", false
	}
	b := v.Bytes()
	if len(b) < 3 || b[0] != '\x1b' || b[1] != '[' {
		return "", false
	}
	return string(b[2:]), true
}

// CSIPayloadFromString extracts CSI payload from a String() representation.
func CSIPayloadFromString(text string) string {
	i := strings.Index(text, "CSI")
	if i < 0 {
		return ""
	}
	payload := strings.TrimSuffix(text[i+3:], "?")
	if m := csiByteListRe.FindStringSubmatch(payload); len(m) == 2 {
		return decodeCSIByteList(m[1])
	}
	return payload
}

// DecodeCSIByteList decodes a space-separated list of CSI byte values.
func DecodeCSIByteList(list string) string {
	return decodeCSIByteList(list)
}

func decodeCSIByteList(list string) string {
	parts := strings.Fields(list)
	buf := make([]byte, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return ""
		}
		buf = append(buf, byte(n))
	}
	return string(buf)
}

// CSIPayload returns the CSI payload from a tea message.
func CSIPayload(msg tea.Msg) string {
	if payload, ok := RawCSIPayload(msg); ok {
		return payload
	}
	if s, ok := msg.(interface{ String() string }); ok {
		return CSIPayloadFromString(s.String())
	}
	return ""
}

// IsShiftEnterPayload reports xterm/Kitty CSI payloads for Shift+Enter.
func IsShiftEnterPayload(payload string) bool {
	return isShiftEnterPayload(payload)
}

func isShiftEnterPayload(payload string) bool {
	if payload == "" {
		return false
	}
	if m := kittyEnterModsRe.FindStringSubmatch(payload); len(m) == 2 {
		mods, err := strconv.Atoi(m[1])
		if err == nil && mods&KittyModShift != 0 {
			return true
		}
		if mods == 2 {
			return true
		}
	}
	if m := xtermEnterModsRe.FindStringSubmatch(payload); len(m) == 2 {
		mod, err := strconv.Atoi(m[1])
		if err == nil && (mod == XtermModShift || mod == 4) {
			return true
		}
	}
	return legacyShiftEnterRe.MatchString(payload)
}

// IsCtrlJPayload reports xterm/Kitty CSI payloads for Ctrl+J newline.
func IsCtrlJPayload(payload string) bool {
	return isCtrlJPayload(payload)
}

func isCtrlJPayload(payload string) bool {
	if payload == "" {
		return false
	}
	payload = strings.TrimSuffix(payload, "~")
	if m := kittyCtrlJRe.FindStringSubmatch(payload); len(m) == 2 {
		mods, err := strconv.Atoi(m[1])
		if err == nil && mods&KittyModCtrl != 0 {
			return true
		}
	}
	if m := xtermCtrlJRe.FindStringSubmatch(payload); len(m) == 2 {
		mod, err := strconv.Atoi(m[1])
		if err == nil && mod == XtermModCtrl {
			return true
		}
	}
	return false
}

func isNewlineCSIPayload(payload string) bool {
	return isShiftEnterPayload(payload) || isCtrlJPayload(payload)
}

// IsShiftEnterKeyMsg reports Bubble Tea key messages for Shift+Enter.
func IsShiftEnterKeyMsg(msg tea.KeyPressMsg) bool {
	return msg.String() == "shift+enter"
}

// IsLiteralNewlineKeyMsg matches terminals that inject a newline rune for Shift+Enter.
func IsLiteralNewlineKeyMsg(msg tea.KeyPressMsg) bool {
	return len(msg.Text) == 1 && msg.Text[0] == '\n'
}

// IsInputNewlineKey reports Enter variants that should insert a newline.
func IsInputNewlineKey(msg tea.KeyPressMsg) bool {
	if msg.String() == "ctrl+j" || (msg.Code == 'j' && msg.Mod.Contains(tea.ModCtrl)) {
		return true
	}
	return IsShiftEnterKeyMsg(msg) || IsLiteralNewlineKeyMsg(msg)
}

// IsShiftEnterMsg reports newline CSI sequences for Shift+Enter.
func IsShiftEnterMsg(msg tea.Msg) bool {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		return IsShiftEnterKeyMsg(k) || IsLiteralNewlineKeyMsg(k)
	}
	return isShiftEnterPayload(CSIPayload(msg))
}

// IsNewlineInputMsg reports messages that should insert a newline in the input.
func IsNewlineInputMsg(msg tea.Msg) bool {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		return IsInputNewlineKey(k)
	}
	return isNewlineCSIPayload(CSIPayload(msg))
}
