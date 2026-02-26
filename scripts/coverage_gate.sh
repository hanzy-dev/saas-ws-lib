#!/usr/bin/env bash
set -euo pipefail

# Package coverage gates (v0.1 usable)
# Format: "import/path minPercent"
GATES=(
  "./pkg/errors 90"
  "./pkg/httpx 80"
  "./pkg/middleware 80"
  "./pkg/db 80"
  "./pkg/auth 80"
  "./pkg/observability 70"
)

PROFILE="${1:-coverage.out}"

if [[ ! -f "${PROFILE}" ]]; then
  echo "coverage profile not found: ${PROFILE}" >&2
  exit 1
fi

# Convert "X.YZ" -> integer basis points to avoid float issues
to_bp() {
  local p="$1" # e.g. "82.35"
  if [[ "$p" != *.* ]]; then
    printf "%d" "$((10#$p * 100))"
    return
  fi
  local int="${p%.*}"
  local frac="${p#*.}"
  frac="${frac}00"
  frac="${frac:0:2}"
  printf "%d" "$((10#$int * 100 + 10#$frac))"
}

fail=0

echo "== Coverage gates (per package) =="

# We compute per-package coverage using `go test` with coverpkg scoped to that package.
# This is slower but deterministic and matches real coverage for that package.
for gate in "${GATES[@]}"; do
  pkg="${gate% *}"
  min="${gate#* }"

  # Run tests for all packages but only measure coverage for target package.
  # Use -count=1 to avoid caching.
  out="$(go test ./... -count=1 -covermode=atomic -coverpkg="${pkg}/..." 2>/dev/null | tail -n 1 || true)"

  # Expect: "coverage: XX.Y% of statements"
  if [[ "${out}" != coverage:* ]]; then
    echo "ERROR: could not compute coverage for ${pkg} (got: '${out}')" >&2
    fail=1
    continue
  fi

  pct="$(echo "${out}" | sed -E 's/.*coverage:[[:space:]]*([0-9]+(\.[0-9]+)?)%.*/\1/')"
  got_bp="$(to_bp "${pct}")"
  min_bp="$(to_bp "${min}.00")"

  if (( got_bp < min_bp )); then
    echo "FAIL ${pkg}: ${pct}% < ${min}%"
    fail=1
  else
    echo "PASS ${pkg}: ${pct}% >= ${min}%"
  fi
done

if (( fail != 0 )); then
  echo
  echo "Coverage gate FAILED. Increase tests or adjust gates (only if justified)."
  exit 1
fi

echo
echo "Coverage gate PASSED."