package adapter

import (
	"context"
	"strings"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// TestPostgreSQLCommandGeneration tests PostgreSQL Sysbench command generation
func TestPostgreSQLCommandGeneration(t *testing.T) {
	ctx := context.Background()
	adapter := NewSysbenchAdapter()

	conn := &connection.PostgreSQLConnection{
		BaseConnection: connection.BaseConnection{
			Name: "Test PG",
		},
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "secret",
		SSLMode:  "prefer",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"threads": 8,
			"time":    60,
		},
		WorkDir: "/tmp",
	}

	cmd, err := adapter.BuildRunCommand(ctx, config)
	if err != nil {
		t.Fatalf("BuildRunCommand() failed: %v", err)
	}

	// Verify command
	checks := map[string]bool{
		"contains sysbench":               strings.Contains(cmd.CmdLine, "sysbench"),
		"contains pgsql":                 strings.Contains(cmd.CmdLine, "pgsql"),
		"contains --pgsql-host=localhost": strings.Contains(cmd.CmdLine, "--pgsql-host=localhost"),
		"contains --pgsql-port=5432":     strings.Contains(cmd.CmdLine, "--pgsql-port=5432"),
		"contains --pgsql-user=postgres": strings.Contains(cmd.CmdLine, "--pgsql-user=postgres"),
		"contains --pgsql-db=testdb":     strings.Contains(cmd.CmdLine, "--pgsql-db=testdb"),
		"NO password in command":         !strings.Contains(cmd.CmdLine, "secret"),
	}

	for check, passed := range checks {
		if !passed {
			t.Errorf("Check failed: %s", check)
		}
	}

	// Check environment
	hasPGPASSWORD := false
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "PGPASSWORD=") {
			hasPGPASSWORD = true
			break
		}
	}
	if !hasPGPASSWORD {
		t.Error("PGPASSWORD environment variable not set")
	}
}
