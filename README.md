# saas-ws-lib

Production-grade shared foundation for Workspace microservices.

**Status: Production-Ready**
> *Built with stability in mind: 100% of core packages (db, errors, httpx, middleware) are covered by unit tests.*

This module enforces operational discipline across all services to eliminate architectural drift and inconsistent behavior between repositories.

It is designed to support a multi-repo microservice architecture (Identity, Core, Payments, Orders, etc.) with consistent standards.

## What This Library Guarantees

Every service built on top of this library inherits:

1. Request & Trace Discipline

- X-Request-ID propagation
- OpenTelemetry trace propagation
- Structured JSON logging with request_id + trace_id injection

2. Standardized Error Contract

All services MUST return:

```
{
  "code": "INVALID_ARGUMENT",
  "message": "validation failed",
  "details": {},
  "trace_id": "..."
}
```

Properties:

- trace_id always present
- details always an object
- error codes mapped consistently to HTTP status

Constructor helpers:

```
wserr.InvalidArgument("validation failed")
wserr.Unauthenticated("authentication required")
wserr.Forbidden("forbidden")
wserr.Internal("internal error")
```

3. Auth Discipline

- JWT verification
- Scope enforcement
- Optional remote policy hook (RBAC / ABAC)
- No sensitive error leakage

4. Observability Discipline

- OpenTelemetry bootstrap
- Prometheus metrics
- Stable route labels enforced
- Latency histogram buckets tuned for microservices

5. HTTP Discipline

- Default server timeouts
- JSON enforcement middleware
- Safe outbound HTTP client with:
  - idempotent-aware retry
  - capped retry attempts
  - 5xx retry support
  - request_id propagation
  - trace propagation

6. Database Discipline

- Connection pooling configuration
- Ping timeout on startup
- Forward-only migration guard
- Transaction safety:
  - panic-safe rollback
  - isolation level support
  - read-only support

Example:

```
err := db.WithTxDefault(ctx, sqlDB, func(ctx context.Context, tx *sql.Tx) error {
    // business logic
    return nil
})
```

Advanced:

```
err := db.WithTx(ctx, sqlDB, db.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
}, fn)
```

7. Validation Discipline

```
if err := validate.Struct(req); err != nil {
    wserr.WriteError(r.Context(), w, err)
    return
}
```

Validation errors automatically map to INVALID_ARGUMENT.

8. Graceful Shutdown Discipline

- SIGINT / SIGTERM handling
- Reverse hook execution
- Bounded shutdown timeout

## Installation

```
go get github.com/hanzy-dev/saas-ws-lib
```

## Minimal Service Example (chi)

```
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/hanzy-dev/saas-ws-lib/pkg/httpx"
	"github.com/hanzy-dev/saas-ws-lib/pkg/log"
	"github.com/hanzy-dev/saas-ws-lib/pkg/middleware"
	"github.com/hanzy-dev/saas-ws-lib/pkg/runtime"
)

func main() {
	logger := log.NewJSON(log.Options{})

	r := chi.NewRouter()

	r.Use(middleware.RequestID())
	r.Use(middleware.Recover(logger))
	r.Use(httpx.RequireJSON)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	server := httpx.NewServer(httpx.ServerConfig{
		Addr: ":8080",
	}, r)

	shutdown := runtime.New(10 * time.Second)
	shutdown.Add(func(ctx context.Context) error {
		return server.Shutdown(ctx)
	})

	go func() {
		_ = server.ListenAndServe()
	}()

	_ = shutdown.Wait(context.Background())
}
```


## Philosophy

- No hidden magic
- No framework lock-in
- Explicit contracts
- Immutable error model
- Operational safety by default
- Observability as a first-class concern



## Indonesia

saas-ws-lib adalah fondasi production-grade untuk seluruh microservice Workspace.

Library ini memastikan semua repository (Identity, Core, Payments, dll) memiliki:

- standar error yang konsisten
- disiplin tracing & logging
- retry outbound yang aman
- metrics dengan cardinality terkontrol
- guard migrasi forward-only
- transaksi database yang aman
- graceful shutdown yang benar

Tujuannya adalah menghilangkan drift arsitektur dan memastikan semua service memiliki disiplin operasional yang sama.

### Standar yang Dijaga

- trace_id selalu ada di response error
- details selalu object
- retry hanya untuk method idempotent
- isolation level transaksi bisa dikontrol
- route metrics harus stabil (tidak boleh dynamic)