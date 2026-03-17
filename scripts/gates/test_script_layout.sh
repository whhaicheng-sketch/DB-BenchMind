#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../.."

require_file() {
    local path=$1
    if [ ! -f "$path" ]; then
        echo "missing file: $path" >&2
        exit 1
    fi
}

for path in \
    scripts/entry/build \
    scripts/entry/start \
    scripts/entry/regression \
    scripts/gates/release_gate.sh \
    scripts/gates/test_build_gate.sh \
    scripts/gates/test_backend_gate.sh \
    scripts/gates/test_frontend_gate.sh \
    scripts/gates/test_start_cleanup.sh \
    scripts/gates/test_smoke_gate.sh \
    scripts/frontend/verify-tasks-monitor-layout.mjs \
    scripts/frontend/verify-template-db-tool-linkage.mjs \
    scripts/frontend/verify-template-test-coverage.mjs \
    scripts/manual/hammerdb_quick_commands.sh
do
    require_file "$path"
done

for path in \
    frontend/check_connections.js \
    frontend/check_horizontal_visible.js \
    frontend/create_fail_connection.js \
    frontend/create_fail_connection2.js \
    frontend/final_verify.js \
    frontend/quick_verify.js \
    frontend/test_edit_full.js \
    frontend/test_edit_modal.js \
    frontend/test_full.js \
    frontend/test_horizontal.js \
    frontend/test_list_feedback.js \
    frontend/test_list_test.js \
    frontend/test_modal.js \
    frontend/test_multi_connection.js \
    frontend/test_password_real.js \
    frontend/test_password_real2.js \
    frontend/test_regression.js \
    frontend/test_scrollbar.js \
    frontend/test_ssh_autofill.js \
    frontend/test_ssh_autofill2.js \
    frontend/test_ssh_edit.js \
    frontend/verify_all.js \
    frontend/verify_fixes.js \
    frontend/verify_independence.js \
    frontend/verify_scrollbar.js \
    playwright_test.js
do
    if [ -e "$path" ]; then
        echo "unexpected legacy script remains: $path" >&2
        exit 1
    fi
done

if [ -d frontend/scripts ]; then
    echo "unexpected frontend/scripts directory remains" >&2
    exit 1
fi

for path in \
    scripts/build \
    scripts/start \
    scripts/regression \
    scripts/release_gate.sh \
    scripts/test_build_gate.sh \
    scripts/test_backend_gate.sh \
    scripts/test_frontend_gate.sh \
    scripts/test_start_cleanup.sh \
    scripts/test_smoke_gate.sh \
    scripts/hammerdb_quick_commands.sh
do
    if [ -e "$path" ]; then
        echo "legacy path still present: $path" >&2
        exit 1
    fi
done

echo "script layout looks correct"
