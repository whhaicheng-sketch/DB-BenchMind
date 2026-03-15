// Package adapter provides unit tests for Swingbench adapter.
package adapter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// TestSwingbenchAdapter_Type tests the Type method.
func TestSwingbenchAdapter_Type(t *testing.T) {
	adapter := NewSwingbenchAdapter()
	assert.Equal(t, AdapterTypeSwingbench, adapter.Type())
}

// TestSwingbenchAdapter_BuildPrepareCommand tests building prepare command.
func TestSwingbenchAdapter_BuildPrepareCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-1",
			Name: "Test Oracle",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "testuser",
		Password:    "testpass",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"scale":        0.1, // Test with float scale (100MB)
			"threads":      32,
			"dba_username": "sys as sysdba",
			"dba_password": "testpass",
		},
		WorkDir: "/tmp/test",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, config)
	require.NoError(t, err)

	// Should return a command sequence with 5 steps
	assert.Equal(t, "oracle_prepare_sequence", cmd.CmdLine)
	assert.Len(t, cmd.Commands, 5, "Prepare should have 5 steps")

	// Step 1: Connection probe (using DBA user)
	step1 := cmd.Commands[0]
	assert.Equal(t, "Connection Probe", step1.StepName)
	assert.Contains(t, step1.CmdLine, "sqlplus")
	assert.Contains(t, step1.CmdLine, "sys as sysdba") // Uses DBA credentials

	// Step 2: Schema existence check
	step2 := cmd.Commands[1]
	assert.Equal(t, "Schema Check", step2.StepName)
	assert.Contains(t, step2.CmdLine, "sqlplus")
	assert.Contains(t, step2.CmdLine, "SOE_SCHEMA_EXISTS")

	// Step 3: Create tablespaces (with datafile paths)
	step3 := cmd.Commands[2]
	assert.Equal(t, "Create Tablespaces", step3.StepName)
	assert.Contains(t, step3.CmdLine, "sqlplus")
	assert.Contains(t, step3.CmdLine, "/tmp/db-benchmind-creatablespace-")

	// Step 4: oewizard -create
	step4 := cmd.Commands[3]
	assert.Equal(t, "Initialize Data", step4.StepName)
	assert.Contains(t, step4.CmdLine, "oewizard")
	assert.Contains(t, step4.CmdLine, "-c oewizard.xml")
	assert.Contains(t, step4.CmdLine, "-cl")
	assert.Contains(t, step4.CmdLine, "-create")
	assert.Contains(t, step4.CmdLine, "-version")
	assert.Contains(t, step4.CmdLine, "-scale 0.1")
	assert.Contains(t, step4.CmdLine, "-tc")
	assert.Contains(t, step4.CmdLine, "//localhost:1521/ORCL")

	step5 := cmd.Commands[4]
	assert.Equal(t, "Post-Schema Setup", step5.StepName)
	assert.Contains(t, step5.CmdLine, "sqlplus")
	assert.Contains(t, step5.CmdLine, "/tmp/db-benchmind-postschema-")
	assert.Contains(t, step5.CmdLine, "sysdba")
}

func TestSwingbenchAdapter_PostSchemaSetupContainsDBMSLockGrant(t *testing.T) {
	adapter := NewSwingbenchAdapter()
	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-priv",
			Name: "Oracle Privilege Test",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "system",
		Password:    "manager",
	}

	cmd := adapter.buildPostSchemaSetupCommand(conn, &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"sysdba_username": "sys",
			"sysdba_password": "manager",
		},
	})

	assert.Contains(t, cmd, "@/tmp/db-benchmind-postschema-")
	tempFile := strings.TrimSpace(cmd[strings.LastIndex(cmd, "@")+1:])
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "grant execute on sys.dbms_lock to soe;")
}

// TestSwingbenchAdapter_BuildRunCommand tests building run command.
func TestSwingbenchAdapter_BuildRunCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name     string
		conn     connection.Connection
		params   map[string]interface{}
		validate func(t *testing.T, cmd *Command, err error)
	}{
		{
			name: "SOE benchmark with default parameters",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-conn-1",
					Name: "Test Oracle",
				},
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Username:    "testuser",
				Password:    "testpass",
			},
			params: map[string]interface{}{
				"virtual_users": 10,
				"time":          10,
				"config_file":   "/opt/benchtools/swingbench/configs/SOE_CPU_Bound.xml",
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				assert.Contains(t, cmd.CmdLine, "charbench")
				assert.Contains(t, cmd.CmdLine, "-c /opt/benchtools/swingbench/configs/SOE_CPU_Bound.xml")
				assert.Contains(t, cmd.CmdLine, "-cs //localhost:1521/ORCL")
				assert.Contains(t, cmd.CmdLine, "-uc 10")
				assert.Contains(t, cmd.CmdLine, "-rt 00:01")
			},
		},
		{
			name: "SOE benchmark with SID instead of service name",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-conn-2",
					Name: "Test Oracle SID",
				},
				Host:     "192.168.1.100",
				Port:     1521,
				SID:      "ORCLSID",
				Username: "testuser",
				Password: "testpass",
			},
			params: map[string]interface{}{
				"virtual_users": 20,
				"time":          5,
				"config_file":   "/opt/benchtools/swingbench/configs/SOE_Disk_Bound.xml",
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				// The connection string uses service name format //host:port/SID
				assert.Contains(t, cmd.CmdLine, "//192.168.1.100:1521/ORCLSID")
			},
		},
		{
			name: "Missing config_file parameter",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-conn-3",
					Name: "Test Oracle",
				},
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Username:    "testuser",
			},
			params: map[string]interface{}{
				"virtual_users": 5,
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "config_file parameter is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Connection: tt.conn,
				Parameters: tt.params,
				WorkDir:    "/tmp/test",
			}

			cmd, err := adapter.BuildRunCommand(ctx, config)
			tt.validate(t, cmd, err)
		})
	}
}

func TestSwingbenchAdapter_BuildRunCommand_ManagedConfigUsesOfficialDefaults(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-4",
			Name: "Test Oracle Managed",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "system",
		Password:    "manager",
	}

	config := &Config{
		Connection: conn,
		Template: &template.Template{
			ID:   "tpl_swing_oe",
			Name: "Swingbench-Oracle-OE-Medium-64u-30m",
			ToolConfig: template.ToolConfig{
				Swingbench: template.SwingbenchConfig{
					Benchmark:      "orderEntry",
					Frontend:       "charbench",
					ConfigMode:     "managed",
					RunTimeSeconds: 120,
					UserCount:      8,
				},
			},
		},
		Parameters: map[string]interface{}{
			"virtual_users": 8,
			"time":          120,
		},
		WorkDir: "/tmp/db-benchmind-managed",
	}

	cmd, err := adapter.BuildRunCommand(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, cmd.CmdLine, "./charbench")
	assert.Contains(t, cmd.CmdLine, "-c ../configs/server_side_soe_v2.xml")
	assert.Contains(t, cmd.CmdLine, "-a")
	assert.Contains(t, cmd.CmdLine, "-r /tmp/db-benchmind-managed/results.xml")
	assert.Equal(t, "/tmp/db-benchmind-managed/results.xml", cmd.ResultFile)
}

// TestSwingbenchAdapter_BuildRunCommand_NonOracle tests that non-Oracle databases fail.
func TestSwingbenchAdapter_BuildRunCommand_NonOracle(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-1",
			Name: "Test MySQL",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
	}

	config := &Config{
		Connection: conn,
		WorkDir:    "/tmp/test",
	}

	_, err := adapter.BuildRunCommand(ctx, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only supports Oracle")
}

// TestSwingbenchAdapter_BuildCleanupCommand tests building cleanup command.
func TestSwingbenchAdapter_BuildCleanupCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-1",
			Name: "Test Oracle",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "testuser",
		Password:    "testpass",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"dba_username": "sys as sysdba",
			"dba_password": "testpass",
		},
		WorkDir: "/tmp/test",
	}

	cmd, err := adapter.BuildCleanupCommand(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, cmd.CmdLine, "oewizard")
	assert.Contains(t, cmd.CmdLine, "-cl")
	assert.Contains(t, cmd.CmdLine, "-drop")
	assert.Contains(t, cmd.CmdLine, "-version 2.0")
	assert.Contains(t, cmd.CmdLine, "-dt thin")
	assert.Contains(t, cmd.CmdLine, "-dba sys as sysdba")
	assert.Contains(t, cmd.CmdLine, "-dbap testpass")
	assert.Contains(t, cmd.CmdLine, "-u soe")
	assert.Contains(t, cmd.CmdLine, "-p soe")
	assert.Contains(t, cmd.CmdLine, "-ts SOE")
	assert.Contains(t, cmd.CmdLine, "-v")
	assert.Contains(t, cmd.CmdLine, "-debug")
	assert.Contains(t, cmd.CmdLine, "-debugf /tmp/oewizard_drop_debug.log")
	assert.Contains(t, cmd.CmdLine, "//localhost:1521/ORCL")
}

// TestSwingbenchAdapter_ParseRunOutput tests parsing swingbench output.
func TestSwingbenchAdapter_ParseRunOutput(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name     string
		stdout   string
		validate func(t *testing.T, result *Result)
	}{
		{
			name: "parse charbench output",
			stdout: `
Time     Users       TPM      TPS     Errors   NCR   UCD   BP    OP    PO    BO
10:58:35 [0/4]       0        0       0        0     0     0     0     0     0
10:58:37 [4/4]       0        0       0        0     0     248   414   0     0
10:58:38 [4/4]       8        8       0        0     0     32    213   0     19
`,
			validate: func(t *testing.T, result *Result) {
				assert.NotNil(t, result)
			},
		},
		{
			name:   "empty output",
			stdout: "",
			validate: func(t *testing.T, result *Result) {
				assert.NotNil(t, result)
				assert.Equal(t, 0.0, result.TPS)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.ParseRunOutput(ctx, tt.stdout, "")
			require.NoError(t, err)
			tt.validate(t, result)
		})
	}
}

func TestSwingbenchAdapter_ParseFinalResults_UsesHintedResultFile(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tempDir := t.TempDir()
	resultFile := filepath.Join(tempDir, "results.xml")
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Results>
  <TotalCompletedTransactions>120</TotalCompletedTransactions>
  <AverageTransactionsPerSecond>2.5</AverageTransactionsPerSecond>
  <TotalRunTime>0:00:48</TotalRunTime>
  <TotalFailedTransactions>3</TotalFailedTransactions>
</Results>`
	require.NoError(t, os.WriteFile(resultFile, []byte(xml), 0o644))

	result, err := adapter.ParseFinalResults(ctx, swingbenchResultFileHintPrefix+resultFile)
	require.NoError(t, err)
	assert.Equal(t, int64(120), result.TotalTransactions)
	assert.Equal(t, 2.5, result.TransactionsPerSec)
	assert.Equal(t, 150.0, result.AvgTPM)
	assert.Equal(t, 48.0, result.TotalTime)
	assert.Equal(t, int64(3), result.IgnoredErrors)
}

// TestSwingbenchAdapter_ValidateConfig tests configuration validation.
func TestSwingbenchAdapter_ValidateConfig(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid Oracle connection",
			config: &Config{
				Connection: &connection.OracleConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test-conn-1",
						Name: "Test Oracle",
					},
					Host:     "localhost",
					Port:     1521,
					SID:      "ORCL",
					Username: "testuser",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config is required",
		},
		{
			name: "nil connection",
			config: &Config{
				Connection: nil,
			},
			wantErr: true,
			errMsg:  "connection is required",
		},
		{
			name: "non-Oracle database",
			config: &Config{
				Connection: &connection.MySQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test-conn-1",
						Name: "Test MySQL",
					},
					Host:     "localhost",
					Port:     3306,
					Database: "testdb",
					Username: "root",
				},
			},
			wantErr: true,
			errMsg:  "only supports Oracle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateConfig(ctx, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSwingbenchAdapter_SupportsDatabase tests database support check.
func TestSwingbenchAdapter_SupportsDatabase(t *testing.T) {
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name   string
		dbType connection.DatabaseType
		want   bool
	}{
		{
			name:   "Oracle",
			dbType: connection.DatabaseTypeOracle,
			want:   true,
		},
		{
			name:   "MySQL",
			dbType: connection.DatabaseTypeMySQL,
			want:   false,
		},
		{
			name:   "PostgreSQL",
			dbType: connection.DatabaseTypePostgreSQL,
			want:   false,
		},
		{
			name:   "SQL Server",
			dbType: connection.DatabaseTypeSQLServer,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.SupportsDatabase(tt.dbType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestSwingbenchAdapter_buildConnectionString tests connection string building.
func TestSwingbenchAdapter_buildConnectionString(t *testing.T) {
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name          string
		conn          *connection.OracleConnection
		expectedParts []string
	}{
		{
			name: "service name",
			conn: &connection.OracleConnection{
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Username:    "testuser",
			},
			expectedParts: []string{"jdbc:oracle:thin:@//localhost:1521/ORCL"},
		},
		{
			name: "SID",
			conn: &connection.OracleConnection{
				Host:     "192.168.1.100",
				Port:     1521,
				SID:      "ORCLSID",
				Username: "testuser",
			},
			expectedParts: []string{"jdbc:oracle:thin:@192.168.1.100:1521:ORCLSID"},
		},
		{
			name: "fallback when no service name or SID",
			conn: &connection.OracleConnection{
				Host:     "localhost",
				Port:     1521,
				Username: "testuser",
			},
			expectedParts: []string{"jdbc:oracle:thin:@//localhost:1521/ORCL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.buildConnectionString(tt.conn)
			for _, part := range tt.expectedParts {
				assert.Contains(t, result, part)
			}
		})
	}
}
