package middleware

import (
	"be-wedding/pkg/tracer"
	"net/http"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func RequestID(zlogger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, r *http.Request) {
			reqID := tracer.TraceIDFromContext(r.Context())
			if reqID == "" {
				reqID = xid.New().String()
			}

			ctx := tracer.UpdateContextWithRequestID(r.Context(), reqID)
			ctxLogger := zlogger.With().Str("request_id", reqID).Logger()
			ctx = ctxLogger.WithContext(r.Context())

			span := trace.SpanFromContext(r.Context())
			span.SetAttributes(attribute.String("request_id", reqID))

			next.ServeHTTP(rw, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
