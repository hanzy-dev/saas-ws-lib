package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type OTelConfig struct {
	ServiceName string
	// Optional: customize span naming. If nil, otelhttp default is used.
	SpanNameFormatter func(operation string, r *http.Request) string
}

func OTel(cfg OTelConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		opts := []otelhttp.Option{}

		if cfg.SpanNameFormatter != nil {
			opts = append(opts, otelhttp.WithSpanNameFormatter(cfg.SpanNameFormatter))
		}

		// ServiceName isn't directly an otelhttp option; it's set via resource/service.name
		// in the tracer provider. Kept here to make config explicit at call site.

		h := otelhttp.NewHandler(next, "http.server", opts...)
		return h
	}
}
