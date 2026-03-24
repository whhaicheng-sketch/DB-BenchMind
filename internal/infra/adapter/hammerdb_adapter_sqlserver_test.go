// Package adapter provides HammerDB benchmark tool adapter.
// Tests for SQL Server specific command building.
package adapter

import (
	"context"
	"encoding/base64"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// extractBase64Script extracts and decodes the base64 encoded script from a command line
func extractBase64Script(t *testing.T, cmdLine string) string {
	// Match: echo <base64> | base64 -d
	re := regexp.MustCompile(`echo ([A-Za-z0-9+/=]+) \| base64 -d`)
	matches := re.FindStringSubmatch(cmdLine)
	if len(matches) < 2 {
		t.Fatalf("Failed to extract base64 script from command: %s", cmdLine)
	}
	decoded, err := base64.StdEncoding.DecodeString(matches[1])
	if err != nil {
		t.Fatalf("Failed to decode base64 script: %v", err)
	}
	return string(decoded)
}

// TestHammerDBAdapter_SQLServer_CleanupCommand verifies the cleanup command
// uses HammerDB's deleteschema command following the standard workflow.
// Uses HammerDB 4.10+ format: diset connection mssqls_* instead of legacy disconn
func TestHammerDBAdapter_SQLServer_CleanupCommand(t *testing.T) {
	// Setup SQL Server connection
	now := time.Now()
	conn := &connection.SQLServerConnection{
		BaseConnection: connection.BaseConnection{
			ID:        "test-id",
			Name:      "Test SQL Server",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Host:     "192.168.134.129",
		Port:     1433,
		Username: "sa",
		Password: "SqlServer@2019",
		Database: "tpcc",
	}

	config := &Config{
		Connection: conn,
		WorkDir:    "/tmp",
		Parameters: map[string]interface{}{
			"database_name": "tpcc",
		},
	}

	adapter := NewHammerDBAdapter()

	cmd, err := adapter.BuildCleanupCommand(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to build cleanup command: %v", err)
	}

	// Verify the command uses hammerdbcli
	if !strings.Contains(cmd.CmdLine, "hammerdbcli") {
		t.Errorf("Cleanup command should use hammerdbcli, got:\n%s", cmd.CmdLine)
	}

	// Extract and decode the base64 script to verify its content
	script := extractBase64Script(t, cmd.CmdLine)
	t.Logf("Decoded script:\n%s", script)

	// Verify HammerDB 4.10+ Linux connection format: diset connection mssqls_linux_server
	if !strings.Contains(script, "dbset db mssqls") {
		t.Errorf("Cleanup script should use 'dbset db mssqls', got:\n%s", script)
	}
	if !strings.Contains(script, "diset connection mssqls_linux_server") {
		t.Errorf("Cleanup script should use 'diset connection mssqls_linux_server', got:\n%s", script)
	}

	// Verify TCP is enabled (required for Linux)
	if !strings.Contains(script, "diset connection mssqls_tcp true") {
		t.Errorf("Cleanup script should enable TCP with 'diset connection mssqls_tcp true', got:\n%s", script)
	}

	// Verify port is set correctly
	if !strings.Contains(script, "diset connection mssqls_port 1433") {
		t.Errorf("Cleanup script should set port with 'diset connection mssqls_port 1433', got:\n%s", script)
	}

	// Verify username is set
	if !strings.Contains(script, "diset connection mssqls_uid") {
		t.Errorf("Cleanup script should set uid with 'diset connection mssqls_uid', got:\n%s", script)
	}

	// Verify password is set
	if !strings.Contains(script, "diset connection mssqls_pass") {
		t.Errorf("Cleanup script should set password with 'diset connection mssqls_pass', got:\n%s", script)
	}

	// Verify the command sets the database name
	if !strings.Contains(script, "diset tpcc mssqls_dbase") {
		t.Errorf("Cleanup script should set database name with diset tpcc mssqls_dbase, got:\n%s", script)
	}

	// Verify the command uses deleteschema
	if !strings.Contains(script, "deleteschema") {
		t.Errorf("Cleanup script should use deleteschema, got:\n%s", script)
	}

	// Verify waittocomplete is called after deleteschema
	if !strings.Contains(script, "waittocomplete") {
		t.Errorf("Cleanup script should call waittocomplete, got:\n%s", script)
	}

	// Verify legacy disconn is NOT used (HammerDB 4.10+ does not support it)
	if strings.Contains(script, "disconn") {
		t.Errorf("Cleanup script should NOT use legacy 'disconn' format (HammerDB 4.10+), got:\n%s", script)
	}

	t.Logf("Generated cleanup command:\n%s", cmd.CmdLine)
}

// TestHammerDBAdapter_SQLServer_RunCommand_Script verifies the generated TCL script
// contains the required HammerDB commands.
func TestHammerDBAdapter_SQLServer_RunCommand_Script(t *testing.T) {
	now := time.Now()
	conn := &connection.SQLServerConnection{
		BaseConnection: connection.BaseConnection{
			ID:        "test-id",
			Name:      "Test SQL Server",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Host:     "10.5.54.87",
		Port:     1433,
		Username: "sa",
		Password: "Abcd_1234",
		Database: "tpcc",
	}

	config := &Config{
		Connection: conn,
		WorkDir:    "/tmp",
		PrepareThreads: 4,
		Parameters: map[string]interface{}{
			"database_name": "tpcc",
			"virtual_users": 10,
			"warehouses":    10,
			"build_users":   10,
			"rampup":        2,
			"duration":      10,
			"driver":        "timed",
			"all_warehouse": "true",
		},
	}

	adapter := NewHammerDBAdapter()

	t.Run("Run command builds successfully", func(t *testing.T) {
		cmd, err := adapter.BuildRunCommand(context.Background(), config)
		if err != nil {
			t.Fatalf("Failed to build run command: %v", err)
		}

		// Verify command contains hammerdb
		if !strings.Contains(cmd.CmdLine, "hammerdb") {
			t.Errorf("Run command should contain 'hammerdb', got: %s", cmd.CmdLine)
		}

		runScript := extractBase64Script(t, cmd.CmdLine)
		if !strings.Contains(runScript, "dbset db mssqls") {
			t.Errorf("Run script should use 'dbset db mssqls', got:\n%s", runScript)
		}
		if !strings.Contains(runScript, "diset tpcc mssqls_dbase tpcc") {
			t.Errorf("Run script should set database name, got:\n%s", runScript)
		}

		t.Logf("Generated run command:\n%s", cmd.CmdLine)
	})

	t.Run("Prepare command builds successfully", func(t *testing.T) {
		cmd, err := adapter.BuildPrepareCommand(context.Background(), config)
		if err != nil {
			t.Fatalf("Failed to build prepare command: %v", err)
		}

		// Verify it returns a command sequence
		if len(cmd.Commands) != 2 {
			t.Errorf("Prepare command should return 2 steps (deleteschema + buildschema), got %d", len(cmd.Commands))
		}

		// Verify step 2 (buildschema) uses correct format - extract and decode base64 script
		buildschemaCmd := cmd.Commands[1]
		buildschemaScript := extractBase64Script(t, buildschemaCmd.CmdLine)
		t.Logf("Decoded buildschema script:\n%s", buildschemaScript)

		// Verify connection settings
		if !strings.Contains(buildschemaScript, "dbset db mssqls") {
			t.Errorf("Buildschema script should use 'dbset db mssqls', got:\n%s", buildschemaScript)
		}
		if !strings.Contains(buildschemaScript, "diset connection mssqls_linux_server") {
			t.Errorf("Buildschema script should use 'diset connection mssqls_linux_server', got:\n%s", buildschemaScript)
		}
		if !strings.Contains(buildschemaScript, "diset connection mssqls_tcp true") {
			t.Errorf("Buildschema script should enable TCP, got:\n%s", buildschemaScript)
		}

		// Verify TPC-C settings
		if !strings.Contains(buildschemaScript, "diset tpcc mssqls_count_ware") {
			t.Errorf("Buildschema script should set warehouses, got:\n%s", buildschemaScript)
		}
		if !strings.Contains(buildschemaScript, "diset tpcc mssqls_num_vu 4") {
			t.Errorf("Buildschema script should use prepare threads as build users, got:\n%s", buildschemaScript)
		}
		if !strings.Contains(buildschemaScript, "diset tpcc mssqls_dbase tpcc") {
			t.Errorf("Buildschema script should set database name, got:\n%s", buildschemaScript)
		}

		// Verify manual-script-compatible buildschema flow
		if !strings.Contains(buildschemaScript, "vudestroy") {
			t.Errorf("Buildschema script should reset virtual users before buildschema, got:\n%s", buildschemaScript)
		}
		if !strings.Contains(buildschemaScript, "buildschema") {
			t.Errorf("Buildschema script should contain 'buildschema', got:\n%s", buildschemaScript)
		}
		if strings.Contains(buildschemaScript, "tcldb") {
			t.Errorf("Buildschema script should not contain invalid tcldb command, got:\n%s", buildschemaScript)
		}

		// Verify uses yes pipe for auto-confirmation
		if !strings.Contains(buildschemaCmd.CmdLine, "yes |") {
			t.Errorf("Buildschema command should use 'yes |' for auto-confirmation, got:\n%s", buildschemaCmd.CmdLine)
		}

		t.Logf("Generated buildschema command:\n%s", buildschemaCmd.CmdLine)
	})

	t.Run("Prepare command caps build users to warehouse count", func(t *testing.T) {
		cappedConfig := &Config{
			Connection: conn,
			WorkDir:    "/tmp",
			PrepareThreads: 2,
			Parameters: map[string]interface{}{
				"database_name": "tpcc",
				"warehouses":    1,
			},
		}

		cmd, err := adapter.BuildPrepareCommand(context.Background(), cappedConfig)
		if err != nil {
			t.Fatalf("Failed to build prepare command: %v", err)
		}
		if len(cmd.Commands) != 2 {
			t.Fatalf("Prepare command should return 2 steps, got %d", len(cmd.Commands))
		}

		buildschemaScript := extractBase64Script(t, cmd.Commands[1].CmdLine)
		if !strings.Contains(buildschemaScript, "diset tpcc mssqls_num_vu 1") {
			t.Fatalf("Buildschema script should cap build users to warehouse count, got:\n%s", buildschemaScript)
		}
	})
}
