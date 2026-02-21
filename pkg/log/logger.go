package log

import (
	"context"
	"io"
	"log/slog"
	"os"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"

	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	base *slog.Logger
}

type Options struct {
	Level slog.Level
	Out   io.Writer
}

func NewJSON(opts Options) *Logger {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}
	h := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: opts.Level,
	})
	return &Logger{base: slog.New(h)}
}

func (l *Logger) Base() *slog.Logger {
	return l.base
}

func (l *Logger) With(ctx context.Context) *slog.Logger {
	var attrs []slog.Attr

	if rid := wsctx.RequestID(ctx); rid != "" {
		attrs = append(attrs, slog.String("request_id", rid))
	}
	if tid := wsctx.TenantID(ctx); tid != "" {
		attrs = append(attrs, slog.String("tenant_id", tid))
	}

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		attrs = append(attrs, slog.String("trace_id", sc.TraceID().String()))
	}

	if len(attrs) == 0 {
		return l.base
	}

	args := make([]any, 0, len(attrs))
	for i := range attrs {
		args = append(args, attrs[i])
	}

	return l.base.With(slog.Group("ctx", args...))
}
