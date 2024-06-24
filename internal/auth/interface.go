package auth

import (
	"context"
	"net/http"
)

type contextKey string

var authCtxKey contextKey = "auth"

type AuthContext struct {
	AuthMethod string
	UserName   string
	// TODO more?
}

func NewContext(ctx context.Context, a *AuthContext) context.Context {
	return context.WithValue(ctx, authCtxKey, a)
}

func FromContext(ctx context.Context) (*AuthContext, bool) {
	a, ok := ctx.Value(authCtxKey).(*AuthContext)
	return a, ok
}

// Abstracts the authentication backend for the server.
type AuthProvider interface {
	// Returns HTTP middleware for performing authentication.
	Middleware() func(http.Handler) http.Handler
}
