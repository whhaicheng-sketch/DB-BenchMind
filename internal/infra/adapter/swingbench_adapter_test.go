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
		WorkDir:    "/tmp/test",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, cmd.CmdLine, "no action required")
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
			},
			params: map[string]interface{}{
				"users":       10,
				"cycles":      100,
				"think_time":  1000,
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				assert.Contains(t, cmd.CmdLine, "oowbench")
				assert.Contains(t, cmd.CmdLine, "-cs")
				assert.Contains(t, cmd.CmdLine, "-bt")
				assert.Contains(t, cmd.CmdLine, "-u 10")
				assert.Contains(t, cmd.CmdLine, "-c 100")
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
			},
			params: map[string]interface{}{
				"users": 20,
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				assert.Contains(t, cmd.CmdLine, "192.168.1.100:1521")
			},
		},
		{
			name: "CALLING benchmark type",
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
				"benchmark_type": "CALLING",
				"users":          5,
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				assert.Contains(t, cmd.CmdLine, "-bt CALLING")
				assert.Contains(t, cmd.CmdLine, "-u 5")
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
	}

	config := &Config{
		Connection: conn,
		WorkDir:    "/tmp/test",
	}

	cmd, err := adapter.BuildCleanupCommand(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, cmd.CmdLine, "no action required")
}

// TestSwingbenchAdapter_ParseRunOutput tests parsing swingbench output.
func TestSwingbenchAdapter_ParseRunOutput(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name      string
		stdout    string
		validate  func(t *testing.T, result *Result)
	}{
		{
			name: "parse standard output",
			stdout: `
Averaged Results:
Benchmark: SOE
Users: 10
Transactions: 1000
TPM: 5000
Average response time: 250ms
Minimum response time: 10ms
Maximum response time: 1000ms
Errors: 5
`,
			validate: func(t *testing.T, result *Result) {
				assert.Equal(t, 5000/60, int(result.TPS)) // TPM converted to TPS
				assert.Equal(t, 250.0, result.LatencyAvg)
				assert.Equal(t, 10.0, result.LatencyMin)
				assert.Equal(t, 1000.0, result.LatencyMax)
				assert.Equal(t, int64(5), result.TotalErrors)
			},
		},
		{
			name: "parse output with percentiles",
			stdout: `
Transaction Results:
TPM: 6000
Average response: 200ms
95th percentile: 400ms
99th percentile: 800ms
Errors: 0
`,
			validate: func(t *testing.T, result *Result) {
				assert.Equal(t, 6000/60, int(result.TPS))
				assert.Equal(t, 200.0, result.LatencyAvg)
			},
		},
		{
			name: "empty output",
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
					Host:        "localhost",
					Port:        1521,
					ServiceName: "ORCL",
					Username:    "testuser",
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
		name    string
		dbType  connection.DatabaseType
		want    bool
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
