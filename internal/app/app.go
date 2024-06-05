package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Raimguhinov/dav-go/internal/config"
	"github.com/Raimguhinov/dav-go/pkg/httpserver"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
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

	//authProvider, err := auth.NewFromURL(authURL)
	//if err != nil {
	//	log.Error(fmt.Errorf("failed to load auth provider: %w", err))
	//}

	// HTTP Server
	router := SetupRouter(log, pg, cfg)
	httpServer := httpserver.New(router, httpserver.Port(cfg.HTTP.Port))

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
