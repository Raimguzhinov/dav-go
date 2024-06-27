package app

import (
	"log/slog"
	"net/http"

	"github.com/Raimguhinov/dav-go/internal/auth"
	"github.com/Raimguhinov/dav-go/internal/config"
	mwlogger "github.com/Raimguhinov/dav-go/internal/delivery/http/middleware/logger"
	"github.com/Raimguhinov/dav-go/internal/usecase"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/ceres919/go-webdav/caldav"
	"github.com/ceres919/go-webdav/carddav"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func SetupRouter(log *logger.Logger, pg *postgres.Postgres, cfg *config.Config, auth auth.AuthProvider) http.Handler {
	log.With(
		slog.Any("AllowedMethods", cfg.HTTP.CORS.AllowedMethods),
		slog.Any("AllowedOrigins", cfg.HTTP.CORS.AllowedOrigins),
		slog.Bool("AllowCredentials", cfg.HTTP.CORS.AllowCredentials),
		slog.Any("AllowedHeaders", cfg.HTTP.CORS.AllowedHeaders),
		slog.Bool("OptionsPassthrough", cfg.HTTP.CORS.OptionsPassthrough),
		slog.Any("ExposedHeaders", cfg.HTTP.CORS.ExposedHeaders),
		slog.Bool("Debug", cfg.HTTP.CORS.Debug),
	).Info("CORS initializing")

	for _, method := range cfg.HTTP.CORS.AllowedMethods {
		chi.RegisterMethod(method)
	}

	s := chi.NewRouter()
	s.Use(middleware.RequestID)
	s.Use(mwlogger.New(log))
	s.Use(cors.New(cors.Options{
		AllowedMethods:     cfg.HTTP.CORS.AllowedMethods,
		AllowedOrigins:     cfg.HTTP.CORS.AllowedOrigins,
		AllowCredentials:   cfg.HTTP.CORS.AllowCredentials,
		AllowedHeaders:     cfg.HTTP.CORS.AllowedHeaders,
		OptionsPassthrough: cfg.HTTP.CORS.OptionsPassthrough,
		ExposedHeaders:     cfg.HTTP.CORS.ExposedHeaders,
		Debug:              cfg.HTTP.CORS.Debug,
		Logger:             log,
	}).Handler)
	s.Use(auth.Middleware())
	s.Use(middleware.Recoverer)

	upBackend := &userPrincipalBackend{}
	url := usecase.NewURL(cfg.PG.URL, cfg.App.CalDAVPrefix, cfg.App.CardDAVPrefix, upBackend)

	caldavBackend, carddavBackend, err := usecase.NewFromURL(url, pg, log)
	if err != nil {
		log.Error("app.SetupRouter", logger.Err(err))
	}

	carddavHandler := carddav.Handler{Backend: carddavBackend}
	caldavHandler := caldav.Handler{Backend: caldavBackend}
	handler := davHandler{
		authBackend:    auth,
		upBackend:      upBackend,
		caldavBackend:  caldavBackend,
		carddavBackend: carddavBackend,
	}

	s.Mount("/", &handler)
	s.Mount("/.well-known/caldav", &caldavHandler)
	s.Mount("/.well-known/carddav", &carddavHandler)
	s.Mount("/{user}/"+cfg.App.CardDAVPrefix, &carddavHandler)
	s.Mount("/{user}/"+cfg.App.CalDAVPrefix, &caldavHandler)

	return s
}
