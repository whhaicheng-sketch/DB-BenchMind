#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> Backend release-gate tests"

go test ./internal/app/usecase \
  ./internal/domain/config \
  ./internal/domain/execution \
  ./internal/domain/report \
  ./internal/domain/template \
  ./internal/infra/adapter \
  ./internal/infra/database \
  ./internal/infra/database/repository \
  ./internal/infra/keyring \
  ./internal/infra/report \
  ./internal/infra/tool \
  ./internal/transportwails \
  ./internal/transportwails/collector \
  -count=1
