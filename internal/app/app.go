package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Raimguhinov/dav-go/config"
	backend "github.com/Raimguhinov/dav-go/internal/caldav"
	"github.com/Raimguhinov/dav-go/pkg/httpserver"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguzhinov/go-webdav/caldav"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)
	b := backend.New()
	var err error

	// HTTP Server
	for _, method := range []string{
		"PROPFIND",
		"PROPPATCH",
		"REPORT",
		"MKCOL",
		"COPY",
		"MOVE",
	} {
		chi.RegisterMethod(method)
	}

	s := chi.NewRouter()
	s.Use(middleware.Logger)

	handler := caldav.Handler{Backend: b, Prefix: "/dav"}
	s.Mount("/", &handler)

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
