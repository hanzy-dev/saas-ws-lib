# saas-ws-lib

[![CI](https://github.com/hanzy-dev/saas-ws-lib/actions/workflows/ci.yml/badge.svg)](https://github.com/hanzy-dev/saas-ws-lib/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hanzy-dev/saas-ws-lib)](https://goreportcard.com/report/github.com/hanzy-dev/saas-ws-lib)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hanzy-dev/saas-ws-lib)](https://github.com/hanzy-dev/saas-ws-lib/blob/main/go.mod)
[![License](https://img.shields.io/github/license/hanzy-dev/saas-ws-lib)](https://github.com/hanzy-dev/saas-ws-lib/blob/main/LICENSE)
[![Codecov](https://codecov.io/gh/hanzy-dev/saas-ws-lib/branch/main/graph/badge.svg?flag=unittests&token=0)](https://app.codecov.io/github/hanzy-dev/saas-ws-lib)

Production-grade shared foundation for Workspace microservices.

**Status: v0.1 usable**

This library enforces consistent architecture and operational standards across all Workspace services (Identity, Core, Payments, Orders, etc.), eliminating repository drift and inconsistent patterns.

## Proof (CI-enforced)

CI runs: unit tests (+race), govulncheck, golangci-lint, Codecov upload, and **coverage gates per package**.

Coverage gates (v0.1):

- `pkg/errors >= 90%`
- `pkg/httpx >= 80%`
- `pkg/middleware >= 80%`
- `pkg/db >= 80%`
- `pkg/auth >= 80%`
- `pkg/observability >= 70%`

Integration (DB) tests are optional and run with build tag:

- `go test ./... -tags=integration`

## Compatibility

- Go ≥ 1.24
- OpenTelemetry SDK ≥ 1.40
- Prometheus client ≥ 1.19

## Installation

```bash
go get github.com/hanzy-dev/saas-ws-lib@v0.1.0
```

## Local development

```
go test ./... -count=1
go test ./... -race
go test ./... -covermode=atomic -coverprofile=coverage.out
bash scripts/coverage_gate.sh coverage.out
golangci-lint run
govulncheck -scan=module
```

## Quickstart (chi)

```
r := chi.NewRouter()
logger := log.NewJSON(log.Options{})

r.Use(middleware.RequestID())
r.Use(middleware.Recover(logger))
r.Use(httpx.RequireJSON)

h := httpx.NewHealth(2*time.Second)
r.Get("/healthz", h.Healthz)
r.Get("/readyz", h.Readyz)

r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
})
```

## What this library guarantees

Every service built on top of this library inherits strict engineering discipline.

1) Request & trace discipline

- X-Request-ID propagation (request_id)
- OpenTelemetry trace propagation (trace_id)
- Structured JSON logging
- request_id + trace_id are injected into logs

2) Standardized error contract

All services MUST return:

```
{
  "code": "INVALID_ARGUMENT",
  "message": "validation failed",
  "details": {},
  "trace_id": "..."
}
```

### Invariants:

- details is always an object
- error codes map deterministically to HTTP status
- no sensitive internal error leakage

### Helpers:

```
wserr.InvalidArgument("validation failed")
wserr.Unauthenticated("authentication required")
wserr.Forbidden("forbidden")
wserr.Internal("internal error")
wserr.ResourceExhausted("payload too large")
```

3) Authentication discipline

- JWT verification (configurable issuer/audience)
- scope helpers (Has, HasAll, HasAny)
- optional remote policy hook (RBAC/ABAC) via PolicyChecker
- no token validation detail leakage
- deterministic error mapping

4) Observability discipline

- OTel bootstrap helper (tracer provider + W3C propagators)
- Prometheus registry bootstrap
- HTTP metrics middleware with stable route labels (no cardinality explosion)

5) HTTP discipline

- secure default server timeouts
- JSON enforcement middleware
- outbound HTTP client:
  - idempotent-aware retry
  - capped retries
  - retry only on transient failures / retryable upstream status
  - request_id propagation
  - trace propagation
  - context-aware backoff

6) Database discipline

- configurable connection pooling
- startup ping timeout
- forward-only migration guard
- transaction safety:
  - panic-safe rollback
  - isolation level + read-only options

### Example:

```
err := db.WithTxDefault(ctx, sqlDB, func(ctx context.Context, tx *sql.Tx) error {
	return nil
})
```

### Advanced:

```
err := db.WithTx(ctx, sqlDB, db.TxOptions{
	Isolation: sql.LevelSerializable,
	ReadOnly:  false,
}, fn)
```

7) Validation discipline

```
if err := validate.Struct(req); err != nil {
	wserr.WriteError(r.Context(), w, err)
	return
}
```

Validation errors map to INVALID_ARGUMENT.

8) Graceful shutdown discipline

- SIGINT/SIGTERM handling
- reverse hook execution
- bounded shutdown timeout
- deterministic resource cleanup

## Versioning policy

- Semantic Versioning (MAJOR.MINOR.PATCH)
- backward-compatible changes only in MINOR
- breaking changes only in MAJOR
- services should pin minor versions

## Roadmap

 - [x] Standardized error discipline
 - [x] Safe retry-aware HTTP client
 - [x] Transaction isolation support
 - [x] Forward-only migration guard
 - [x] Coverage gates in CI
 - [ ] DB metrics instrumentation
 - [ ] OTel exporter auto-bootstrap helper
 - [ ] Kubernetes production example

## Indonesia

saas-ws-lib adalah fondasi untuk seluruh microservice Workspace.

Library ini memastikan setiap service punya disiplin yang konsisten:

- error contract standar
- tracing & logging disiplin
- retry outbound aman
- metrics dengan cardinality terkontrol
- guard migrasi forward-only
- transaksi database yang aman
- graceful shutdown yang benar

Tujuannya: menghilangkan drift arsitektur dan memastikan semua service punya disiplin operasional yang sama.