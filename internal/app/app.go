package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Raimguzhinov/dav-go/internal/auth"
	"github.com/Raimguzhinov/dav-go/internal/config"
	"github.com/Raimguzhinov/dav-go/pkg/logger"
	"github.com/Raimguzhinov/dav-go/pkg/postgres"
)

func Run(cfg *config.Config) {
	log := logger.New(cfg.Log.Level, cfg.App.Env)

	// Repository
	pg, err := postgres.New(context.TODO(), log, cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		log.Error("app.Run", logger.Err(err))
		os.Exit(1)
	}
	defer pg.Close()

	authProvider, err := auth.NewFromURL(cfg, "basic://")
	if err != nil {
		log.Error(fmt.Sprintf("failed to load auth provider: %v", err))
	}

	// HTTP Server
	router := SetupRouter(log, pg, cfg, authProvider)
	httpServer := http.NewServer(router, http.Port(cfg.HTTP.Port))
	httpServer.Start()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app.Run", slog.String("signal", s.String()))
	case err = <-httpServer.Notify():
		log.Error("app.Run", logger.Err(err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		log.Error("app.Run", logger.Err(err))
	}
	log.Info("Gracefully stopped")
}
