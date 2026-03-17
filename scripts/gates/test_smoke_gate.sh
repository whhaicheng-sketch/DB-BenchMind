#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../.."

log_file="$(mktemp)"
cleanup() {
    rm -f "$log_file"
}
trap cleanup EXIT

echo "==> Smoke gate: start application"
if ! timeout 10s ./scripts/entry/start >"$log_file" 2>&1; then
    status=$?
    if [ "$status" -ne 124 ]; then
        cat "$log_file"
        exit "$status"
    fi
fi

cat "$log_file"

grep -q "Wails App: startup called" "$log_file"
grep -q "BenchmarkBinding context injected" "$log_file"
grep -q "MonitorBinding context injected" "$log_file"
grep -q "TaskBinding context injected" "$log_file"
grep -q "System monitoring started" "$log_file"

if grep -q "Wails applications will not build without the correct build tags." "$log_file"; then
    echo "smoke gate detected invalid Go-built Wails binary" >&2
    exit 1
fi

echo "smoke gate passed"
