package inputui

import "strings"

// Attachment is a pending image attachment in the input area.
type Attachment struct {
	AbsPath string
	RelPath string
	MIME    string
	Name    string
}

// DisplaySuffix returns the user-visible attachment summary appended on submit.
func DisplaySuffix(atts []Attachment) string {
	if len(atts) == 0 {
		return ""
	}
	names := make([]string, len(atts))
	for i, att := range atts {
		names[i] = att.Name
	}
	return "\n[images: " + strings.Join(names, ", ") + "]"
}

// RuntimeMediaNote appends attachment paths for non-vision models.
func RuntimeMediaNote(atts []Attachment) string {
	if len(atts) == 0 {
		return ""
	}
	paths := make([]string, len(atts))
	for i, att := range atts {
		paths[i] = att.RelPath
	}
	var b strings.Builder
	b.WriteString("\n\nAttached images (use ReadMediaFile to view):")
	for _, p := range paths {
		b.WriteString("\n- ")
		b.WriteString(p)
	}
	return b.String()
}