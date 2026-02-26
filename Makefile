SHELL := /usr/bin/env bash
.SHELLFLAGS := -euo pipefail -c

GO ?= go

.PHONY: test race coverage gate lint vuln fmt tidy ci test-integration

test:
	$(GO) test ./... -count=1

race:
	$(GO) test ./... -race -count=1

coverage:
	$(GO) test ./... -covermode=atomic -coverprofile=coverage.out

gate: coverage
	bash scripts/coverage_gate.sh coverage.out

lint:
	golangci-lint run

vuln:
	govulncheck -scan=module

fmt:
	gofmt -w .

tidy:
	$(GO) mod tidy

test-integration:
	$(GO) test ./... -tags=integration -count=1

ci: test race gate lint vuln