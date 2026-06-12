package views

// message represents a chat entry.
type message struct {
	role string // "user" or "assistant"
	text string
}

// Elph logo in Braille art.
var (
	ElphLogo1 = "\u28FF\u28FF\u285F\u28FF\u285F\u28FF\u28FF"
	ElphLogo2 = "\u28FF\u28FF\u28FF\u28FF\u28FF\u28FF\u28FF"
)

// shortHash truncates a git hash to 7 characters.
func shortHash(h string) string {
	if len(h) < 7 {
		return h
	}
	return h[:7]
}

// shortSession truncates a session ID to 8 characters.
func shortSession(s string) string {
	if len(s) < 8 {
		return s
	}
	return s[:8]
}
