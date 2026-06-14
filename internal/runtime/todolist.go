package runtime

import (
	"context"

	"github.com/riipandi/elph/pkg/tools/todolist"
)

func executeTodoList(ctx context.Context, workDir string, args map[string]any) ToolResult {
	raw, present := args["todos"]
	store := todolist.StoreFrom(ctx)
	var before []todolist.Todo
	if store != nil {
		before = append([]todolist.Todo(nil), *store...)
	}

	out, err := todolist.Apply(ctx, raw, present)
	if err != nil {
		return ToolResult{Err: err}
	}

	if store != nil && !todosEqual(before, *store) {
		sess, _ := todoSessionFrom(ctx)
		if workDir == "" {
			workDir = sess.WorkDir
		}
		_ = SaveTodosSnapshot(workDir, sess.SessionID, *store)
	}
	return ToolResult{Output: out}
}
