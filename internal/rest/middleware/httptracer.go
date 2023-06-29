package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func HTTPTracer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		otelhttp.NewHandler(otelhttp.WithRouteTag(r.URL.RequestURI(),
			next), "example-operation").ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}