package runtime

import (
	"context"
	"testing"

	"github.com/riipandi/elph/pkg/tools/todolist"
	"github.com/stretchr/testify/require"
)

func TestSessionTodoStoreSurvivesStructCopy(t *testing.T) {
	original := NewSession(t.TempDir())
	copied := original

	ctx := todolist.WithStore(context.Background(), copied.todoStore)
	_, err := todolist.Apply(ctx, []any{
		map[string]any{"title": "ship feature", "status": "in_progress"},
	}, true)
	require.NoError(t, err)

	require.Len(t, original.Todos(), 1)
	require.Equal(t, "ship feature", original.Todos()[0].Title)
	require.Len(t, copied.Todos(), 1)
}
