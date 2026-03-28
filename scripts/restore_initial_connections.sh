#!/bin/bash
# restore_initial_connections.sh
# Restores the 4 initial baseline database connections if they are missing.
# This script is idempotent - running it multiple times will not create duplicates.
#
# Usage:
#   ./scripts/restore_initial_connections.sh
#
# What it does:
#   - Reads the current connections from data/db-benchmind.db
#   - Checks if each of the 4 initial connections (MySQL, Oracle, PostgreSQL, SQL Server) exists
#   - Only inserts missing connections
#   - Does NOT modify or delete any existing connections
#
# Recovery scope: MySQL, Oracle, PostgreSQL, SQL Server connections

set -euo pipefail

DB_PATH="./data/db-benchmind.db"

if [ ! -f "$DB_PATH" ]; then
    echo "ERROR: Database not found at $DB_PATH"
    echo "Please run the application first to initialize the database."
    exit 1
fi

echo "=== Connection Recovery Script ==="
echo "Database: $DB_PATH"
echo ""

# Define the 4 initial connections with their expected IDs
declare -A EXPECTED_CONNECTIONS
EXPECTED_CONNECTIONS=(
  ["ce4daee0-f1dc-4296-a6ce-51eb7b806759"]="MySQL"
  ["3f48caae-2a75-4fc3-aae3-e44c8e930992"]="Oracle"
  ["50710b65-8793-4670-8304-69c6f1053afd"]="PostgreSQL"
  ["085716d4-2e41-4d05-9c7a-28fa52bf0055"]="SQLServer"
)

restored=0
already_exists=0

for conn_id in "${!EXPECTED_CONNECTIONS[@]}"; do
    conn_name="${EXPECTED_CONNECTIONS[$conn_id]}"

    # Check if connection exists
    count=$(echo "SELECT COUNT(*) FROM connections WHERE id = '$conn_id';" | \
        python3 -c "import sqlite3,sys; db=sqlite3.connect('$DB_PATH'); print(db.execute(sys.stdin.read().strip()).fetchone()[0])" 2>/dev/null || echo "0")

    if [ "$count" -gt 0 ]; then
        echo "[OK] $conn_name ($conn_id) - already exists"
        already_exists=$((already_exists + 1))
        continue
    fi

    echo "[MISSING] $conn_name ($conn_id) - attempting restore..."

    # Restore connection based on type
    case "$conn_name" in
        "MySQL")
            config_json='{"ai_assistants":[],"created_at":"2026-03-27T13:20:36Z","database":"","host":"192.168.134.129","id":"'"$conn_id"'","name":"MySQL","port":3357,"ssh":{"enabled":true,"host":"192.168.134.129","local_port":0,"port":22,"username":"root"},"ssl_mode":"","type":"mysql","updated_at":"2026-03-27T13:20:36Z","username":"admin"}'
            db_type="mysql"
            ;;
        "Oracle")
            config_json='{"ai_assistants":[],"connect_as":"normal","connect_type":"basic","created_at":"2026-03-27T13:20:45Z","host":"192.168.134.129","id":"'"$conn_id"'","identifier_type":"sid","name":"Oracle","port":1521,"service_name":"","sid":"orcl","ssh":{"enabled":true,"host":"192.168.134.129","local_port":0,"port":22,"username":"root"},"tns_name":"","type":"oracle","updated_at":"2026-03-27T13:20:45Z","username":"system"}'
            db_type="oracle"
            ;;
        "PostgreSQL")
            config_json='{"ai_assistants":[],"created_at":"2026-03-27T13:17:08Z","database":"postgres","host":"192.168.134.129","id":"'"$conn_id"'","name":"PostgreSQL","port":5432,"ssh":{"enabled":true,"host":"192.168.134.129","local_port":0,"port":22,"username":"root"},"ssl_mode":"","type":"postgresql","updated_at":"2026-03-27T13:17:08Z","username":"admin"}'
            db_type="postgresql"
            ;;
        "SQLServer")
            config_json='{"ai_assistants":[],"created_at":"2026-03-27T13:19:56Z","database":"","host":"192.168.134.129","id":"'"$conn_id"'","name":"SQLServer","port":1433,"ssh":{"enabled":true,"host":"192.168.134.129","local_port":0,"port":22,"username":"root"},"trust_server_certificate":false,"type":"sqlserver","updated_at":"2026-03-27T13:19:56Z","username":"sa"}'
            db_type="sqlserver"
            ;;
    esac

    now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Use python3 to insert (sqlite3 CLI may not be available)
    python3 -c "
import sqlite3, sys, json
db = sqlite3.connect('$DB_PATH')
db.execute('''INSERT OR IGNORE INTO connections (id, name, db_type, config_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)''',
    ('$conn_id', '$conn_name', '$db_type', json.dumps(json.loads('''$config_json''')), '$now', '$now'))
db.commit()
db.close()
print('[RESTORED] $conn_name ($conn_id)')
"
    restored=$((restored + 1))
done

echo ""
echo "=== Summary ==="
echo "Already existing: $already_exists"
echo "Restored: $restored"
echo "Total initial connections: $((already_exists + restored))"

if [ $((already_exists + restored)) -eq 4 ]; then
    echo "All 4 initial connections are present."
else
    echo "WARNING: Not all 4 initial connections could be verified."
fi
