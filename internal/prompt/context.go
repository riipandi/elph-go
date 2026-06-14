package prompt

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func formatProjectContextSection(content, path string) string {
	content = strings.TrimSpace(content)
	if content == "" || strings.TrimSpace(path) == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString("<project_context>\n\n")
	b.WriteString("Project-specific instructions and guidelines:\n\n")
	fmt.Fprintf(&b, "<project_instructions path=%q>\n", path)
	b.WriteString(content)
	b.WriteString("\n</project_instructions>\n\n</project_context>")
	return b.String()
}

func formatRuntimeContextSection(date, workDir string) string {
	workDir = strings.TrimSpace(workDir)
	if workDir != "" {
		if abs, err := filepath.Abs(workDir); err == nil {
			workDir = abs
		}
	}
	if strings.TrimSpace(date) == "" {
		date = time.Now().Format("2006-01-02")
	}
	if workDir == "" {
		return fmt.Sprintf("Current date: %s", date)
	}
	return fmt.Sprintf("Current date: %s\nCurrent working directory: %s", date, workDir)
}

func formatSessionStateSection(mode string) string {
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = "build"
	}
	return fmt.Sprintf("<session_state>\n\n<session_mode>%s</session_mode>\n\n</session_state>", escapeXML(mode))
}
