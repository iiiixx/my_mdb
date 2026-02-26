package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

func Logger(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			next.ServeHTTP(ww, r)

			fields := logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     ww.Status(),
				"size":       ww.BytesWritten(),
				"durationMs": time.Since(start).Milliseconds(),
			}

			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				fields["request_id"] = reqID
			}

			log.WithFields(fields).Info("http request")
		})
	}
}
