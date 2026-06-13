package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.jetify.com/typeid/v2"
)

func TestOpenAppendAndReadLog(t *testing.T) {
	dir := t.TempDir()
	id := typeid.MustGenerate("sess")

	path, err := OpenLog(dir, id)
	require.NoError(t, err)
	require.FileExists(t, path)

	require.NoError(t, AppendLog(path, "user", "hello"))
	require.NoError(t, AppendLog(path, "system", "world"))

	content, err := ReadLogTail(path, 4096)
	require.NoError(t, err)
	require.Contains(t, content, "[user] hello")
	require.Contains(t, content, "[system] world")
}

func TestReadLogTailTruncates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tail.log")
	require.NoError(t, os.WriteFile(path, []byte(strings.Repeat("x", 5000)), 0o644))

	content, err := ReadLogTail(path, 100)
	require.NoError(t, err)
	require.LessOrEqual(t, len(content), 100)
}
