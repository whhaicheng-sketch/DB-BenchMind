#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../.."

echo "========================================="
echo "DB-BenchMind Release Gate"
echo "========================================="

bash scripts/gates/test_script_layout.sh
bash scripts/gates/test_build_gate.sh
bash scripts/gates/test_backend_gate.sh
bash scripts/gates/test_frontend_gate.sh
bash scripts/gates/test_start_cleanup.sh
bash scripts/gates/test_smoke_gate.sh

echo "========================================="
echo "Release gate passed"
echo "========================================="
