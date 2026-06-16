package exec

import (
	"context"

	"github.com/riipandi/elph/internal/runtime/log"
	"github.com/riipandi/elph/internal/runtime/todostore"
	"github.com/riipandi/elph/internal/runtime/toolresult"
	"github.com/riipandi/elph/pkg/tools/todolist"
)

func executeTodoList(ctx context.Context, workDir string, args map[string]any) toolresult.ToolResult {
	raw, present := args["todos"]
	store := todolist.StoreFrom(ctx)
	var before []todolist.Todo
	if store != nil {
		before = append([]todolist.Todo(nil), *store...)
	}

	out, err := todolist.Apply(ctx, raw, present)
	if err != nil {
		return toolresult.ToolResult{Err: err}
	}

	if store != nil && !log.TodosEqual(before, *store) {
		sessWorkDir, sessionID, ok := todostore.FromSession(ctx)
		if workDir == "" {
			workDir = sessWorkDir
		}
		if ok {
			_ = log.SaveTodosSnapshot(workDir, sessionID, *store)
		}
	}
	return toolresult.ToolResult{Output: out}
}
