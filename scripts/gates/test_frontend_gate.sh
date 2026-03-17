#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "==> Frontend release-gate tests"
npm --prefix frontend run test

echo "==> Frontend production build"
npm --prefix frontend run build
