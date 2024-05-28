package app

import (
	"fmt"
	"net/http"

	"github.com/Raimguhinov/dav-go/configs"
	"github.com/Raimguhinov/dav-go/internal/usecase"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(l *logger.Logger, pg *postgres.Postgres, cfg *configs.Config) http.Handler {
	for _, method := range []string{
		"PROPFIND",
		"PROPPATCH",
		"REPORT",
		"MKCOL",
		"COPY",
		"MOVE",
		"OPTIONS",
	} {
		chi.RegisterMethod(method)
	}

	s := chi.NewRouter()
	s.Use(middleware.Logger)
	s.Use(corsMiddleware)
	//s.Use(authProvider.Middleware())

	upBackend := &userPrincipalBackend{}
	url := usecase.NewURL(cfg.PG.URL, "/calendars/", "/contacts/", upBackend)

	caldavBackend, carddavBackend, err := usecase.NewFromURL(url, pg, l)
	if err != nil {
		l.Error(fmt.Errorf("failed to load storage backend: %w", err))
	}

	carddavHandler := carddav.Handler{Backend: carddavBackend}
	caldavHandler := caldav.Handler{Backend: caldavBackend}
	handler := davHandler{
		upBackend:      upBackend,
		caldavBackend:  caldavBackend,
		carddavBackend: carddavBackend,
	}

	s.Mount("/", &handler)
	s.Mount("/.well-known/caldav", &caldavHandler)
	s.Mount("/.well-known/carddav", &carddavHandler)
	s.Mount("/{user}/contacts", &carddavHandler)
	s.Mount("/{user}/calendars", &caldavHandler)

	return s
}
