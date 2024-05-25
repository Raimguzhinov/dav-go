package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Raimguhinov/dav-go/configs"
	"github.com/Raimguhinov/dav-go/internal/usecase"
	"github.com/Raimguhinov/dav-go/pkg/httpserver"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Run(cfg *configs.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(context.TODO(), cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - repo.New: %w", err))
	}
	defer pg.Close()

	// HTTP Server
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

	//authProvider, err := auth.NewFromURL(authURL)
	//if err != nil {
	//	l.Error(fmt.Errorf("failed to load auth provider: %w", err))
	//}
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
		upBackend: upBackend,
		//authBackend:    authProvider,
		caldavBackend:  caldavBackend,
		carddavBackend: carddavBackend,
	}

	s.Mount("/", &handler)
	s.Mount("/.well-known/caldav", &caldavHandler)
	s.Mount("/.well-known/carddav", &carddavHandler)
	s.Mount("/{user}/contacts", &carddavHandler)
	s.Mount("/{user}/calendars", &caldavHandler)

	httpServer := httpserver.New(s, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
