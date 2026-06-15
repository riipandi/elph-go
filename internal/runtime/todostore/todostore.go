package todostore

import "context"

type sessionKey struct{}

// Session holds workDir/sessionID for TodoList snapshot persistence.
type Session struct {
	WorkDir   string
	SessionID string
}

// WithSession stores TodoList persistence metadata on ctx.
func WithSession(ctx context.Context, workDir, sessionID string) context.Context {
	if workDir == "" && sessionID == "" {
		return ctx
	}
	return context.WithValue(ctx, sessionKey{}, Session{
		WorkDir:   workDir,
		SessionID: sessionID,
	})
}

// FromSession returns workDir and sessionID stored for TodoList persistence.
func FromSession(ctx context.Context) (workDir, sessionID string, ok bool) {
	if ctx == nil {
		return "", "", false
	}
	sess, ok := ctx.Value(sessionKey{}).(Session)
	if !ok {
		return "", "", false
	}
	return sess.WorkDir, sess.SessionID, true
}
