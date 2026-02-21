# saas-ws-lib

Shared library for Workspace microservices.

This module provides consistent foundations across all services:
- structured JSON logging (slog)
- request/trace context propagation
- standardized error model
- JWT auth + scope enforcement
- optional policy hook (RBAC/ABAC)
- OpenTelemetry tracing
- Prometheus metrics
- HTTP server/client defaults
- DB pooling + forward-only migration guard
- validation helpers
- graceful shutdown utilities
- test helpers

The goal is to eliminate drift between repositories.

---

## Installation

```bash
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

	// middleware chain
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

## Error Format

All services must return:

```
{
  "code": "INVALID_ARGUMENT",
  "message": "validation failed",
  "details": {},
  "trace_id": "..."
}
```

## Health Endpoints

```
health := httpx.NewHealth(2*time.Second,
	func(ctx context.Context) error {
		return db.PingContext(ctx)
	},
)

r.Get("/healthz", health.Healthz)
r.Get("/readyz", health.Readyz)
```

## Validation Example

```
type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
}

if err := validate.Struct(req); err != nil {
	wserr.WriteError(w, err.WithTrace(r.Context()))
	return
}
```

## HTTP Client with Retry + Trace Propagation

```
client := httpx.NewClient(httpx.ClientConfig{
	Timeout:    5 * time.Second,
	MaxRetries: 1,
})

req, _ := http.NewRequest(http.MethodGet, "http://service", nil)
resp, err := httpx.Do(ctx, client, req, 1)
```

## Forward-Only Migration Guard

```
_ = db.EnsureSchemaVersionTable(ctx, sqlDB)
_ = db.EnforceForwardOnly(ctx, sqlDB, newVersion)
```

## Graceful Shutdown

```
shutdown := runtime.New(10 * time.Second)
shutdown.Add(func(ctx context.Context) error {
	return server.Shutdown(ctx)
})
_ = shutdown.Wait(context.Background())
```

## Philosophy

- No hidden magic.
- No framework lock-in.
- Explicit over implicit.
- Operational discipline by default.


# Indonesia

Library bersama untuk seluruh microservice Workspace.

Modul ini menyediakan fondasi yang konsisten di semua service:

- logging terstruktur JSON (slog)
- propagasi request_id dan trace_id
- standar format error
- autentikasi JWT + enforcement scope
- hook policy opsional (RBAC/ABAC)
- tracing OpenTelemetry
- metrics Prometheus
- standar HTTP server & client
- pooling database + guard migrasi forward-only
- helper validasi request
- utilitas graceful shutdown
- helper untuk testing

Tujuannya adalah menghilangkan drift antar repository.

## Instalasi
```bash
go get github.com/hanzy-dev/saas-ws-lib
```

## Contoh Service Minimal (chi)

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

	// chain middleware standar
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

## Format Error

Semua service wajib mengembalikan format berikut:

```
{
  "code": "INVALID_ARGUMENT",
  "message": "validation failed",
  "details": {},
  "trace_id": "..."
}
```

Format ini memastikan konsistensi response lintas service.

## Endpoint Health

```
health := httpx.NewHealth(2*time.Second,
	func(ctx context.Context) error {
		return db.PingContext(ctx)
	},
)

r.Get("/healthz", health.Healthz)
r.Get("/readyz", health.Readyz)
```

/healthz → memastikan service hidup

/readyz → memastikan dependency siap (misalnya database)

## Contoh Validasi

```
type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email"`
}

if err := validate.Struct(req); err != nil {
	wserr.WriteError(w, err.WithTrace(r.Context()))
	return
}
```

Validasi otomatis menghasilkan error dengan format standar.

## HTTP Client dengan Retry + Propagasi Trace

```
client := httpx.NewClient(httpx.ClientConfig{
	Timeout:    5 * time.Second,
	MaxRetries: 1,
})

req, _ := http.NewRequest(http.MethodGet, "http://service", nil)
resp, err := httpx.Do(ctx, client, req, 1)
```

Client ini:

- menerapkan timeout default
- retry terbatas hanya untuk method idempotent
- mempropagasi trace dan request_id

## Guard Migrasi Forward-Only

```
_ = db.EnsureSchemaVersionTable(ctx, sqlDB)
_ = db.EnforceForwardOnly(ctx, sqlDB, newVersion)
```

Mencegah rollback migrasi secara tidak sengaja di environment produksi.

## Graceful Shutdown

```
shutdown := runtime.New(10 * time.Second)
shutdown.Add(func(ctx context.Context) error {
	return server.Shutdown(ctx)
})
_ = shutdown.Wait(context.Background())
```

Menjamin service berhenti dengan aman saat menerima SIGTERM/SIGINT.

## Filosofi

- Tidak ada magic tersembunyi.
- Tidak terkunci ke framework tertentu.
- Eksplisit lebih baik daripada implisit.
- Disiplin operasional sebagai standar bawaan.