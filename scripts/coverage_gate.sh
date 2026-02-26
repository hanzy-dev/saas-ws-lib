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

# Convert "X.YZ" -> integer basis points
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

for gate in "${GATES[@]}"; do
  pkg="${gate% *}"
  min="${gate#* }"

  tmp="$(mktemp -t coverpkg.XXXXXX.out)"

  # Run tests only for the target package tree and produce a dedicated coverprofile.
  # -count=1 avoids caching skewing the report.
  if ! go test "${pkg}/..." -count=1 -covermode=atomic -coverprofile="${tmp}" >/dev/null; then
    echo "ERROR: tests failed for ${pkg}"
    rm -f "${tmp}"
    fail=1
    continue
  fi

  # Extract total coverage percentage
  pct="$(go tool cover -func="${tmp}" | awk '/^total:/{print $3}' | tr -d '%')"
  rm -f "${tmp}"

  if [[ -z "${pct}" ]]; then
    echo "ERROR: could not read coverage for ${pkg}"
    fail=1
    continue
  fi

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