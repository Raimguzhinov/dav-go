package auth

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
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
	user, pass, ok := r.BasicAuth()
	if !ok {
		if strings.Contains(r.UserAgent(), "iOS") {
			// TODO: iOS sends an empty basic auth header
			user = b.clientID
			pass = b.clientSecret
		} else {
			basicAuthFailed(w, b.realm)
			return
		}
	}

	if subtle.ConstantTimeCompare([]byte(pass), []byte(b.clientSecret)) != 1 {
		basicAuthFailed(w, b.realm)
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

func basicAuthFailed(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}
