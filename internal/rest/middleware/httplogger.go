package middleware

import (
	"be-wedding/pkg/logger"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func HTTPLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := newLoggingResponseWriter(w)

		defer func() {
			lvl := httpCodeToZerologLevel(lrw.statusCode)
			ctxLogger := logger.FromContext(r.Context()).WithLevel(lvl)

			ctxLogger.
				Str("method", r.Method).
				Str("url", r.URL.RequestURI()).
				Str("user_agent", r.UserAgent()).
				Dur("elapsed_ms", time.Since(start)).
				Int("status_code", lrw.statusCode).
				Send()
		}()

		next.ServeHTTP(lrw, r)
	}
	return http.HandlerFunc(fn)
}

func httpCodeToZerologLevel(code int) zerolog.Level {
	var lvl zerolog.Level

	if code >= 200 && code <= 299 {
		lvl = zerolog.InfoLevel
	} else {
		lvl = zerolog.ErrorLevel
	}
	return lvl
}
