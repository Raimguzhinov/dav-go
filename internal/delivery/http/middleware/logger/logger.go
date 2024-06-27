package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5/middleware"
)

func New(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/logger"),
		)

		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}

				var status string
				switch {
				case ww.Status() < 200:
					status = color.New(color.FgBlue).Sprintf("%03d", ww.Status())
				case ww.Status() < 300:
					status = color.New(color.FgGreen).Sprintf("%03d", ww.Status())
				case ww.Status() < 400:
					status = color.New(color.FgCyan).Sprintf("%03d", ww.Status())
				case ww.Status() < 500:
					status = color.New(color.FgYellow).Sprintf("%03d", ww.Status())
				default:
					status = color.New(color.FgRed).Sprintf("%03d", ww.Status())
				}

				entry.Info(fmt.Sprintf("%s %s://%s%s - %s", r.Method, scheme, r.Host, r.RequestURI, status),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
