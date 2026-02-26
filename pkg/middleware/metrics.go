package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec

	groupStatus bool
}

type MetricsConfig struct {
	Namespace string
	Subsystem string
	Registry  prometheus.Registerer

	// If true, status will be grouped (2xx, 4xx, 5xx) instead of exact code.
	GroupStatus bool
}

// NewMetrics creates and registers HTTP metrics into cfg.Registry (or DefaultRegisterer if nil).
func NewMetrics(cfg MetricsConfig) *Metrics {
	reg := cfg.Registry
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests processed.",
			},
			[]string{"method", "route", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: cfg.Namespace,
				Subsystem: cfg.Subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				Buckets: []float64{
					0.005, 0.01, 0.02, 0.05,
					0.1, 0.2, 0.3, 0.5,
					0.75, 1, 1.5, 2,
					3, 5,
				},
			},
			[]string{"method", "route", "status"},
		),
		groupStatus: cfg.GroupStatus,
	}

	reg.MustRegister(m.RequestsTotal, m.RequestDuration)
	return m
}

// Handler returns a Prometheus scrape handler.
// If reg is nil, it uses the global default gatherer.
func Handler(reg *prometheus.Registry) http.Handler {
	if reg == nil {
		return promhttp.Handler()
	}
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

// Instrument records request count and duration.
// route MUST be a stable identifier (no path params). Example: "POST /v1/auth/login".
func (m *Metrics) Instrument(route string) func(http.Handler) http.Handler {
	if m == nil {
		panic("middleware.Metrics.Instrument requires non-nil Metrics")
	}
	if route == "" {
		panic("metrics route must be a stable identifier, e.g. 'POST /v1/auth/login'")
	}
	if strings.Contains(route, "{") || strings.Contains(route, "}") {
		panic("metrics route must not contain dynamic path parameters")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			next.ServeHTTP(sw, r)

			status := formatStatus(sw.status, m.groupStatus)

			m.RequestsTotal.WithLabelValues(r.Method, route, status).Inc()
			m.RequestDuration.WithLabelValues(r.Method, route, status).Observe(time.Since(start).Seconds())
		})
	}
}

func formatStatus(code int, grouped bool) string {
	if !grouped {
		return strconv.Itoa(code)
	}
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500 && code < 600:
		return "5xx"
	default:
		return "xxx"
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
