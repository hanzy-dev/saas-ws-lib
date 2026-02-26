package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type OTelConfig struct {
	// Optional: customize span naming. If nil, otelhttp default is used.
	SpanNameFormatter func(operation string, r *http.Request) string
}

// OTel instruments inbound HTTP requests with OpenTelemetry spans.
// Tracer/provider configuration (resource service.name, exporters, sampling) is handled elsewhere.
func OTel(cfg OTelConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		opts := []otelhttp.Option{}
		if cfg.SpanNameFormatter != nil {
			opts = append(opts, otelhttp.WithSpanNameFormatter(cfg.SpanNameFormatter))
		}
		return otelhttp.NewHandler(next, "http.server", opts...)
	}
}
