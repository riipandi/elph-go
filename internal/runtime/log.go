package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.jetify.com/typeid/v2"
)

const defaultLogTailBytes = 24 * 1024

// OpenLog opens or creates the session log file for workDir and sessionID.
func OpenLog(workDir string, sessionID typeid.TypeID) (string, error) {
	dir := filepath.Join(workDir, ".elph", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, sessionID.String()+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return path, nil
}

// AppendLog appends a timestamped line to the session log.
func AppendLog(path, kind, text string) error {
	if path == "" {
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	ts := time.Now().UTC().Format(time.RFC3339)
	body := strings.TrimRight(text, "\n")
	_, err = fmt.Fprintf(f, "%s [%s] %s\n", ts, kind, body)
	return err
}

// ReadLogTail returns up to maxBytes from the end of the log file.
func ReadLogTail(path string, maxBytes int) (string, error) {
	if path == "" {
		return "", fmt.Errorf("log path is empty")
	}
	if maxBytes <= 0 {
		maxBytes = defaultLogTailBytes
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(data) <= maxBytes {
		return string(data), nil
	}

	truncated := data[len(data)-maxBytes:]
	if idx := strings.IndexByte(string(truncated), '\n'); idx >= 0 && idx < len(truncated)-1 {
		truncated = truncated[idx+1:]
	}
	return string(truncated), nil
}
