package runtime

import (
	"context"
	"testing"

	"github.com/riipandi/elph/pkg/tools/todolist"
	"github.com/stretchr/testify/require"
)

func TestExecuteTodoListPersistsAcrossCalls(t *testing.T) {
	ctx := context.Background()
	todos := []todolist.Todo{}
	ctx = todolist.WithStore(ctx, &todos)

	set := ExecuteTool(ctx, "", "TodoList", map[string]any{
		"todos": []any{
			map[string]any{"title": "read file", "status": "in_progress"},
			map[string]any{"title": "write tests", "status": "pending"},
		},
	})
	require.NoError(t, set.Err)
	require.Contains(t, set.Output, "[in_progress] read file")
	require.Len(t, todos, 2)

	query := ExecuteTool(ctx, "", "TodoList", map[string]any{})
	require.NoError(t, query.Err)
	require.Contains(t, query.Output, "[pending] write tests")

	clear := ExecuteTool(ctx, "", "TodoList", map[string]any{"todos": []any{}})
	require.NoError(t, clear.Err)
	require.Equal(t, "Todo list cleared.", clear.Output)
	require.Empty(t, todos)
}

func TestSessionTodosMutateInPlace(t *testing.T) {
	s := NewSession(t.TempDir())
	ctx := todolist.WithStore(context.Background(), s.todoStore)

	result := executeTodoList(ctx, s.WorkDir, map[string]any{
		"todos": []any{
			map[string]any{"title": "updated", "status": "done"},
		},
	})
	require.NoError(t, result.Err)
	todos := s.Todos()
	require.Len(t, todos, 1)
	require.Equal(t, todolist.StatusDone, todos[0].Status)
}
