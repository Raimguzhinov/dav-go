package auth

import (
	"fmt"
	"net/url"

	"github.com/Raimguzhinov/dav-go/internal/config"
)

func NewFromURL(cfg *config.Config, authURL string) (AuthProvider, error) {
	u, err := url.Parse(authURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing auth URL: %s", err.Error())
	}

	switch u.Scheme {
	case "basic":
		return NewBasicAuth(cfg.App.Name, cfg.HTTP.User, cfg.HTTP.Password)
	case "http", "https":
		if u.User == nil {
			return nil, fmt.Errorf("missing client ID for http Basic auth")
		}
		//clientID := u.User.Username()
		//clientSecret, _ := u.User.Password()
		//u.User = nil
		//return NewOAuth2(u.String(), clientID, clientSecret)
		return nil, fmt.Errorf("http OAuth2 auth is not supported")
	default:
		return nil, fmt.Errorf("no auth provider found for %s:// URL", u.Scheme)
	}
}
