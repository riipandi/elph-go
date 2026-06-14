package runtime

import "context"

type todoSessionKey struct{}

type todoSession struct {
	WorkDir   string
	SessionID string
}

func withTodoSession(ctx context.Context, workDir, sessionID string) context.Context {
	if workDir == "" && sessionID == "" {
		return ctx
	}
	return context.WithValue(ctx, todoSessionKey{}, todoSession{
		WorkDir:   workDir,
		SessionID: sessionID,
	})
}

func todoSessionFrom(ctx context.Context) (todoSession, bool) {
	if ctx == nil {
		return todoSession{}, false
	}
	sess, ok := ctx.Value(todoSessionKey{}).(todoSession)
	return sess, ok
}
