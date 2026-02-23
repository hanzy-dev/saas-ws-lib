# saas-ws-lib

[![Go Report Card](https://goreportcard.com/badge/github.com/hanzy-dev/saas-ws-lib)](https://goreportcard.com/report/github.com/hanzy-dev/saas-ws-lib)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hanzy-dev/saas-ws-lib)]()
[![License](https://img.shields.io/github/license/hanzy-dev/saas-ws-lib)]()

Production-grade shared foundation for Workspace microservices.

**Status: Production-Ready**

> *Built with stability in mind: 100% of core packages (db, errors, httpx, middleware) are covered by unit tests.*

This module enforces consistent architecture and operational standards across all Workspace services (Identity, Core, Payments, Orders, etc.), eliminating repository drift and inconsistent patterns.

## What This Library Guarantees

Every service built on top of this library inherits strict engineering discipline.

1. Request & Trace Discipline

- X-Request-ID propagation
- OpenTelemetry trace propagation
- Structured JSON logging
- Automatic request_id + trace_id injection into logs

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

### Invariants

- trace_id always present
- details always an object
- Error codes consistently mapped to HTTP status
- No sensitive internal error leakage

### Constructor Helpers

```
wserr.InvalidArgument("validation failed")
wserr.Unauthenticated("authentication required")
wserr.Forbidden("forbidden")
wserr.Internal("internal error")
wserr.ResourceExhausted("payload too large")
```

3. Authentication Discipline

- JWT verification
- Scope enforcement
- Optional remote policy hook (RBAC / ABAC)
- No token validation detail leakage
- Deterministic error mapping

4. Observability Discipline

- OpenTelemetry bootstrap
- Prometheus metrics middleware
- Stable route labels (no cardinality explosion)
- Latency histogram instrumentation

5. HTTP Discipline

- Secure default server timeouts
- JSON enforcement middleware
- Safe outbound HTTP client:
  - Idempotent-aware retry
  - Capped retry attempts
  - Retry on transient 5xx
  - Request ID propagation
  - Trace propagation
  - Context-aware backoff

6. Database Discipline

- Configurable connection pooling
- Startup ping timeout
- Forward-only migration guard
- Transaction safety:
  - Panic-safe rollback
  - Isolation level support
  - Read-only support

### Example

```
err := db.WithTxDefault(ctx, sqlDB, func(ctx context.Context, tx *sql.Tx) error {
    return nil
})
```

### Advanced

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
- Deterministic resource cleanup

## Architectural Guarantees

This library enforces the following invariants across all Workspace services:

- Error responses are immutable and traceable.
- Transactions are panic-safe.
- Database migrations are forward-only.
- HTTP clients never retry unsafe methods.
- Metrics labels are stable and bounded.
- Authentication never exposes verification details.
- Observability is first-class, not optional.

All Workspace services rely on these guarantees.

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

## Versioning Policy

- Semantic Versioning (MAJOR.MINOR.PATCH)
- Backward-compatible changes only in MINOR
- Breaking changes only in MAJOR
- Forward-only migration policy enforced
- Services must pin minor versions explicitly

## Roadmap

 - [x] Standardized error discipline
 - [x] Safe retry-aware HTTP client
 - [x] Transaction isolation support
 - [x] Forward-only migration guard
 - [ ] DB metrics instrumentation
 - [ ] OpenTelemetry exporter auto-bootstrap helper
 - [ ] Kubernetes production example

## Philosophy

- No hidden magic
- No framework lock-in
- Explicit contracts over implicit behavior
- Immutable error model
- Operational safety by default
- Observability as a first-class concern

This is not a helper library.
This is the foundation of a multi-repository microservice ecosystem.

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