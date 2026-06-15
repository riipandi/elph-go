package session

import (
	"context"

	"github.com/riipandi/elph/internal/runtime/todostore"
)

func withTodoSession(ctx context.Context, workDir, sessionID string) context.Context {
	return todostore.WithSession(ctx, workDir, sessionID)
}
