package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

type MetricsConfig struct {
	Namespace string
	Subsystem string
	Registry  prometheus.Registerer
}

func NewMetrics(cfg MetricsConfig) *Metrics {
	ns := cfg.Namespace
	sub := cfg.Subsystem

	reg := cfg.Registry
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Subsystem: sub,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests processed.",
			},
			[]string{"method", "route", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: ns,
				Subsystem: sub,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
			},
			[]string{"method", "route", "status"},
		),
	}

	prometheus.MustRegister(m.RequestsTotal, m.RequestDuration)
	if reg != prometheus.DefaultRegisterer {
		reg.MustRegister(m.RequestsTotal, m.RequestDuration)
	}

	return m
}

func Handler(reg *prometheus.Registry) http.Handler {
	if reg == nil {
		return promhttp.Handler()
	}
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

func (m *Metrics) Instrument(route string) func(http.Handler) http.Handler {
	if m == nil {
		panic("middleware.Metrics.Instrument requires non-nil Metrics")
	}
	if route == "" {
		route = "unknown"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			next.ServeHTTP(sw, r)

			status := strconv.Itoa(sw.status)
			m.RequestsTotal.WithLabelValues(r.Method, route, status).Inc()
			m.RequestDuration.WithLabelValues(r.Method, route, status).Observe(time.Since(start).Seconds())
		})
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
