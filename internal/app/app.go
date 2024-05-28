package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Raimguhinov/dav-go/configs"
	"github.com/Raimguhinov/dav-go/pkg/httpserver"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
)

func Run(cfg *configs.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(context.TODO(), cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - repo.New: %w", err))
	}
	defer pg.Close()

	//authProvider, err := auth.NewFromURL(authURL)
	//if err != nil {
	//	l.Error(fmt.Errorf("failed to load auth provider: %w", err))
	//}

	// HTTP Server
	router := SetupRouter(l, pg, cfg)
	httpServer := httpserver.New(router, httpserver.Port(cfg.HTTP.Port))

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
