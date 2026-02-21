package observability

import "github.com/prometheus/client_golang/prometheus"

type Prometheus struct {
	Registry *prometheus.Registry
}

func NewPrometheus() *Prometheus {
	return &Prometheus{Registry: prometheus.NewRegistry()}
}
