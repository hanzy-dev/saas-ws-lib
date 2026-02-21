package httpx

import (
	"context"
	"net/http"
	"time"
)

type CheckFunc func(ctx context.Context) error

type Health struct {
	Checks  []CheckFunc
	Timeout time.Duration
}

func NewHealth(timeout time.Duration, checks ...CheckFunc) *Health {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Health{
		Checks:  checks,
		Timeout: timeout,
	}
}

// Liveness: process is running.
func (h *Health) Healthz(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// Readiness: dependencies are ready.
func (h *Health) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.Timeout)
	defer cancel()

	for _, check := range h.Checks {
		if err := check(ctx); err != nil {
			JSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not_ready",
			})
			return
		}
	}

	JSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}
