package users

import (
	"context"
)

type contextKey struct{}

var ctxKey contextKey

func NewContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, ctxKey, user)
}

func FromContext(ctx context.Context) (User, bool) {
	v := ctx.Value(ctxKey)
	if user, ok := v.(User); ok {
		return user, true
	}
	return User{}, false
}
