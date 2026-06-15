package session

import (
	"context"
	"os"
	"testing"

	"github.com/riipandi/elph/internal/projectdir"
	"github.com/riipandi/elph/internal/runtime/exec"
	"github.com/riipandi/elph/internal/runtime/log"
	"github.com/riipandi/elph/internal/runtime/todostore"
	"github.com/riipandi/elph/pkg/tools/todolist"
	"github.com/stretchr/testify/require"
)

func TestNewSessionDoesNotLoadPreviousTodos(t *testing.T) {
	workDir := t.TempDir()
	require.NoError(t, log.SaveTodosSnapshot(workDir, "sess_prev", []todolist.Todo{
		{Title: "resume work", Status: todolist.StatusInProgress},
	}))

	s := NewSession(workDir)
	require.Empty(t, s.Todos())
}

func TestExecuteTodoListWritesJSONL(t *testing.T) {
	workDir := t.TempDir()
	s := NewSession(workDir)
	ctx := todolist.WithStore(context.Background(), s.todoStore)
	ctx = todostore.WithSession(ctx, workDir, s.ID.String())

	result := exec.ExecuteTool(ctx, workDir, "TodoList", map[string]any{
		"todos": []any{
			map[string]any{"title": "ship", "status": "pending"},
		},
	})
	require.NoError(t, result.Err)

	path := projectdir.SessionTodosPath(workDir, s.ID.String())
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"title":"ship"`)

	s2 := NewSession(workDir)
	require.Empty(t, s2.Todos())
}

func TestSessionTodosMutateInPlace(t *testing.T) {
	s := NewSession(t.TempDir())
	ctx := todolist.WithStore(context.Background(), s.todoStore)
	ctx = todostore.WithSession(ctx, s.WorkDir, s.ID.String())

	result := exec.ExecuteTool(ctx, s.WorkDir, "TodoList", map[string]any{
		"todos": []any{
			map[string]any{"title": "updated", "status": "done"},
		},
	})
	require.NoError(t, result.Err)
	todos := s.Todos()
	require.Len(t, todos, 1)
	require.Equal(t, todolist.StatusDone, todos[0].Status)
}
