package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetrics_Registers(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	m := NewMetrics(MetricsConfig{Registry: reg, Namespace: "x", Subsystem: "y"})
	if m == nil || m.RequestsTotal == nil || m.RequestDuration == nil {
		t.Fatalf("expected metrics initialized")
	}
}

func TestMetricsHandler(t *testing.T) {
	t.Parallel()

	// nil registry uses default handler
	h := Handler(nil)
	if h == nil {
		t.Fatalf("expected handler")
	}

	reg := prometheus.NewRegistry()
	h2 := Handler(reg)
	if h2 == nil {
		t.Fatalf("expected handler")
	}
}

func TestMetrics_Instrument_Panics(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	m := NewMetrics(MetricsConfig{Registry: reg})

	tests := []struct {
		name  string
		route string
	}{
		{"empty", ""},
		{"has params", "GET /v1/x/{id}"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic")
				}
			}()
			_ = m.Instrument(tt.route)
		})
	}

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic")
			}
		}()
		var m2 *Metrics
		_ = m2.Instrument("GET /x")
	})
}

func TestFormatStatus(t *testing.T) {
	t.Parallel()

	if got := formatStatus(200, true); got != "2xx" {
		t.Fatalf("got=%q", got)
	}
	if got := formatStatus(404, true); got != "4xx" {
		t.Fatalf("got=%q", got)
	}
	if got := formatStatus(503, true); got != "5xx" {
		t.Fatalf("got=%q", got)
	}
	if got := formatStatus(123, true); got != "xxx" {
		t.Fatalf("got=%q", got)
	}
	if got := formatStatus(201, false); got != "201" {
		t.Fatalf("got=%q", got)
	}
}

func TestMetrics_Instrument_Records(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	m := NewMetrics(MetricsConfig{Registry: reg, GroupStatus: true})

	h := m.Instrument("GET /x")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != 201 {
		t.Fatalf("status=%d", rr.Code)
	}

	// scrape metrics to ensure they were registered and updated
	scrape := Handler(reg)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	scrape.ServeHTTP(rr2, req2)

	b, _ := io.ReadAll(rr2.Body)
	if len(b) == 0 {
		t.Fatalf("expected metrics output")
	}
}
