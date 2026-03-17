#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/../.."

script_path="scripts/entry/start"

require_pattern() {
    local pattern=$1
    if ! grep -Eq "$pattern" "$script_path"; then
        echo "missing pattern: $pattern" >&2
        exit 1
    fi
}

require_pattern 'pkill -f "bin/db-benchmind"'
require_pattern 'LauncherBootstrap.*oewizard'
require_pattern 'LauncherBootstrap.*charbench'
require_pattern 'com\\.dom\\.benchmarking\\.swingbench\\.wizards\\.Wizard'
require_pattern 'com\\.dom\\.benchmarking\\.swingbench\\.CharBench'
require_pattern 'pkill -TERM'
require_pattern 'pkill -KILL'
require_pattern 'sysbench'
require_pattern 'hammerdbcli'

echo "start cleanup patterns look correct"
