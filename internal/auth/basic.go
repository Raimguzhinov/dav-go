package auth

import (
	"fmt"
	"net/http"
)

type BasicAuth struct {
	realm        string
	clientID     string
	clientSecret string
}

func NewBasicAuth(realm, username, password string) (AuthProvider, error) {
	if username == "" {
		return nil, fmt.Errorf("missing username")
	}
	if password == "" {
		return nil, fmt.Errorf("missing password")
	}
	return &BasicAuth{realm: realm, clientID: username, clientSecret: password}, nil
}

func (b *BasicAuth) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b.basicAuth(next, w, r)
		})
	}
}

func (b *BasicAuth) basicAuth(next http.Handler, w http.ResponseWriter, r *http.Request) {
	user, password, ok := r.BasicAuth()
	if !ok || user != b.clientID || password != b.clientSecret {
		w.Header().Add("WWW-Authenticate", `Basic realm="Please provide your system credentials", charset="UTF-8"`)
		http.Error(w, "HTTP Basic auth is required", http.StatusUnauthorized)
		return
	}
	authCtx := AuthContext{
		AuthMethod: "basic",
		UserName:   user,
	}
	ctx := NewContext(r.Context(), &authCtx)
	r = r.WithContext(ctx)
	next.ServeHTTP(w, r)
}
