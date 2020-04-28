package web

import "context"

type ctxKey int

const (
	userCtxKey ctxKey = iota
)

type UserData struct {
	Email  string
	Scopes []string
}

func ContextWithUser(ctx context.Context, user *UserData) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func UserFromContext(ctx context.Context) (*UserData, bool) {
	user, ok := ctx.Value(userCtxKey).(*UserData)
	if user == nil || !ok {
		return nil, false
	}
	return user, true
}
