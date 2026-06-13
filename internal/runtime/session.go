package runtime

import (
	tea "charm.land/bubbletea/v2"
	"github.com/riipandi/elph/internal/prompt"
	"github.com/riipandi/elph/pkg/core/agent"
	"go.jetify.com/typeid/v2"
)

// Session binds a coding-agent runtime to a single interactive session.
type Session struct {
	ID              typeid.TypeID
	WorkDir         string
	SystemPrompt    string
	LogPath         string
	RequestsLogPath string
}

// NewSession creates a session with a generated typeid and assembled system prompt.
func NewSession(workDir string) Session {
	id := typeid.MustGenerate("sess")
	logPath, _ := OpenLog(workDir, id)

	return Session{
		ID:              id,
		WorkDir:         workDir,
		SystemPrompt:    prompt.Build(prompt.Options{WorkDir: workDir}),
		LogPath:         logPath,
		RequestsLogPath: RequestsLogPath(workDir, id),
	}
}

// AppendLog records an event in the session log file.
func (s Session) AppendLog(kind, text string) {
	_ = AppendLog(s.LogPath, kind, text)
}

// RunTurn starts an agent turn for the given user prompt.
func (s Session) RunTurn(userPrompt string) tea.Cmd {
	return agent.RunTurn(userPrompt)
}
