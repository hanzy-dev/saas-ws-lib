package observability

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type OTel struct {
	tp *sdktrace.TracerProvider
}

type OTelConfig struct {
	ServiceName    string
	ServiceVersion string

	Exporter sdktrace.SpanExporter

	// Default sampler is ParentBased(TraceIDRatioBased(1.0)) if unset.
	Sampler sdktrace.Sampler
}

func InitOTel(cfg OTelConfig) (*OTel, error) {
	if cfg.ServiceName == "" {
		return nil, errors.New("otel: missing service name")
	}
	if cfg.Exporter == nil {
		return nil, errors.New("otel: missing span exporter")
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	sampler := cfg.Sampler
	if sampler == nil {
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(cfg.Exporter),
	)

	otel.SetTracerProvider(tp)

	// W3C TraceContext + Baggage propagator (standard)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return &OTel{tp: tp}, nil
}

func (o *OTel) Shutdown(ctx context.Context) error {
	if o == nil || o.tp == nil {
		return nil
	}
	// enforce bounded shutdown to avoid hanging on exit
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return o.tp.Shutdown(cctx)
}
