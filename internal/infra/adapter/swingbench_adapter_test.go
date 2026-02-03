// Package adapter provides unit tests for Swingbench adapter.
package adapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
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
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"scale":         1,
			"threads":       32,
			"dba_username":  "sys as sysdba",
			"dba_password":  "testpass",
		},
		WorkDir: "/tmp/test",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, cmd.CmdLine, "oewizard")
	assert.Contains(t, cmd.CmdLine, "-cl")
	assert.Contains(t, cmd.CmdLine, "-create")
	assert.Contains(t, cmd.CmdLine, "-generate")
	assert.Contains(t, cmd.CmdLine, "-scale 1")
	assert.Contains(t, cmd.CmdLine, "-tc 32")
	assert.Contains(t, cmd.CmdLine, "//localhost:1521/ORCL")
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
				"users":       10,
				"time":        10,
				"config_file": "/opt/benchtools/swingbench/configs/SOE_CPU_Bound.xml",
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				assert.Contains(t, cmd.CmdLine, "charbench")
				assert.Contains(t, cmd.CmdLine, "-c /opt/benchtools/swingbench/configs/SOE_CPU_Bound.xml")
				assert.Contains(t, cmd.CmdLine, "-cs //localhost:1521/ORCL")
				assert.Contains(t, cmd.CmdLine, "-uc 10")
				assert.Contains(t, cmd.CmdLine, "-rt 10:00")
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
				"users":       20,
				"time":        5,
				"config_file": "/opt/benchtools/swingbench/configs/SOE_Disk_Bound.xml",
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				assert.Contains(t, cmd.CmdLine, "192.168.1.100:1521:ORCLSID")
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
				"users": 5,
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
