#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> Wails build gate"
./scripts/entry/build

if [ ! -x ./bin/db-benchmind ]; then
    echo "missing built binary: ./bin/db-benchmind" >&2
    exit 1
fi

echo "build artifact verified: ./bin/db-benchmind"
