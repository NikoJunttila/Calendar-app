package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// StructuredLogger is a middleware that logs the start and end of each request using slog.
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()
		defer func() {
			slog.Info("request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", ww.Status()),
				slog.Duration("duration", time.Since(t1)),
				slog.Int("bytes", ww.BytesWritten()),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
