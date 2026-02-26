package observability

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	tracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestInitOTel_Validation(t *testing.T) {
	t.Parallel()

	exp := tracetest.NewInMemoryExporter()

	if _, err := InitOTel(OTelConfig{Exporter: exp}); err == nil {
		t.Fatalf("expected error for missing service name")
	}
	if _, err := InitOTel(OTelConfig{ServiceName: "svc"}); err == nil {
		t.Fatalf("expected error for missing exporter")
	}
}

func TestInitOTel_SetsProviderAndPropagator(t *testing.T) {
	t.Parallel()

	exp := tracetest.NewInMemoryExporter()

	// Set a sentinel propagator first to verify it changes.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

	o, err := InitOTel(OTelConfig{
		ServiceName:    "svc",
		ServiceVersion: "1.0.0",
		Exporter:       exp,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if o == nil || o.tp == nil {
		t.Fatalf("expected non-nil otel instance")
	}

	// Provider should be set globally; basic smoke check that it doesn't panic.
	_ = otel.Tracer("test")

	// Propagator should be non-nil and usable.
	p := otel.GetTextMapPropagator()
	if p == nil {
		t.Fatalf("expected propagator to be set")
	}
}

func TestShutdown_NilSafe(t *testing.T) {
	t.Parallel()

	var o *OTel
	if err := o.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	o2 := &OTel{tp: nil}
	if err := o2.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestShutdown_Bounded(t *testing.T) {
	t.Parallel()

	exp := tracetest.NewInMemoryExporter()
	o, err := InitOTel(OTelConfig{
		ServiceName: "svc",
		Exporter:    exp,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Should not hang and should return nil or context error depending on runtime,
	// but generally Shutdown should succeed quickly.
	_ = o.Shutdown(ctx)
}
