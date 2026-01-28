// Package adapter provides unit tests for sysbench adapter.
package adapter

import (
	"context"
	"strings"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// TestSysbenchAdapter_Type tests adapter type.
func TestSysbenchAdapter_Type(t *testing.T) {
	adapter := NewSysbenchAdapter()
	if adapter.Type() != AdapterTypeSysbench {
		t.Errorf("Type() = %v, want %v", adapter.Type(), AdapterTypeSysbench)
	}
}

// TestSysbenchAdapter_SupportsDatabase tests database support.
func TestSysbenchAdapter_SupportsDatabase(t *testing.T) {
	adapter := NewSysbenchAdapter()

	tests := []struct {
		name    string
		dbType  connection.DatabaseType
		want    bool
	}{
		{"MySQL supported", connection.DatabaseTypeMySQL, true},
		{"PostgreSQL supported", connection.DatabaseTypePostgreSQL, true},
		{"Oracle not supported", connection.DatabaseTypeOracle, false},
		{"SQL Server not supported", connection.DatabaseTypeSQLServer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := adapter.SupportsDatabase(tt.dbType); got != tt.want {
				t.Errorf("SupportsDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSysbenchAdapter_BuildPrepareCommand tests prepare command building.
func TestSysbenchAdapter_BuildPrepareCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSysbenchAdapter()

	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn",
			Name: "Test MySQL",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"tables":     10,
			"table_size": 10000,
		},
		WorkDir: "/tmp/work",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, config)
	if err != nil {
		t.Fatalf("BuildPrepareCommand() failed: %v", err)
	}

	if !strings.Contains(cmd.CmdLine, "sysbench") {
		t.Errorf("CmdLine should contain 'sysbench', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "mysql") {
		t.Errorf("CmdLine should contain 'mysql', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "--tables=10") {
		t.Errorf("CmdLine should contain '--tables=10', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "--table-size=10000") {
		t.Errorf("CmdLine should contain '--table-size=10000', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "prepare") {
		t.Errorf("CmdLine should contain 'prepare', got: %s", cmd.CmdLine)
	}
}

// TestSysbenchAdapter_BuildRunCommand tests run command building.
func TestSysbenchAdapter_BuildRunCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSysbenchAdapter()

	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn",
			Name: "Test MySQL",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"threads": 8,
			"time":    60,
			"tables":  10,
			"rate":    0,
		},
		WorkDir: "/tmp/work",
	}

	cmd, err := adapter.BuildRunCommand(ctx, config)
	if err != nil {
		t.Fatalf("BuildRunCommand() failed: %v", err)
	}

	if !strings.Contains(cmd.CmdLine, "sysbench") {
		t.Errorf("CmdLine should contain 'sysbench', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "--threads=8") {
		t.Errorf("CmdLine should contain '--threads=8', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "--time=60") {
		t.Errorf("CmdLine should contain '--time=60', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "--report-interval=1") {
		t.Errorf("CmdLine should contain '--report-interval=1', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "run") {
		t.Errorf("CmdLine should contain 'run', got: %s", cmd.CmdLine)
	}
}

// TestSysbenchAdapter_BuildCleanupCommand tests cleanup command building.
func TestSysbenchAdapter_BuildCleanupCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSysbenchAdapter()

	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn",
			Name: "Test MySQL",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"tables": 10,
		},
		WorkDir: "/tmp/work",
	}

	cmd, err := adapter.BuildCleanupCommand(ctx, config)
	if err != nil {
		t.Fatalf("BuildCleanupCommand() failed: %v", err)
	}

	if !strings.Contains(cmd.CmdLine, "sysbench") {
		t.Errorf("CmdLine should contain 'sysbench', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "--tables=10") {
		t.Errorf("CmdLine should contain '--tables=10', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "cleanup") {
		t.Errorf("CmdLine should contain 'cleanup', got: %s", cmd.CmdLine)
	}
}

// TestSysbenchAdapter_ParseRunOutput tests output parsing.
func TestSysbenchAdapter_ParseRunOutput(t *testing.T) {
	ctx := context.Background()
	adapter := NewSysbenchAdapter()

	stdout := `
SQL statistics:
    queries performed:
        read:                            140000
        write:                           40000
        other:                           20000
        total:                           200000
    transactions:                        20000  (1234.56 per sec.)
    queries:                             200000 (12345.67 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                          0      (0.00 per sec.)

General statistics:
    total time:                          16.2045s
    total number of events:              20000

Latency (ms):
         min:                                    3.23
         avg:                                    6.45
         max:                                   45.67
         95th percentile:                       12.34
         sum:                                129000.00

Threads fairness:
    events (avg/stddev):           2500.00/0.00
    execution time (avg/stddev):  16.1995/0.00

`

	result, err := adapter.ParseRunOutput(ctx, stdout, "")
	if err != nil {
		t.Fatalf("ParseRunOutput() failed: %v", err)
	}

	// Verify parsed metrics
	if result.TPS != 1234.56 {
		t.Errorf("TPS = %v, want 1234.56", result.TPS)
	}
	if result.LatencyAvg != 6.45 {
		t.Errorf("LatencyAvg = %v, want 6.45", result.LatencyAvg)
	}
	if result.LatencyMin != 3.23 {
		t.Errorf("LatencyMin = %v, want 3.23", result.LatencyMin)
	}
	if result.LatencyMax != 45.67 {
		t.Errorf("LatencyMax = %v, want 45.67", result.LatencyMax)
	}
	if result.LatencyP95 != 12.34 {
		t.Errorf("LatencyP95 = %v, want 12.34", result.LatencyP95)
	}
	if result.TotalTransactions != 20000 {
		t.Errorf("TotalTransactions = %v, want 20000", result.TotalTransactions)
	}
}

// TestSysbenchAdapter_ValidateConfig tests configuration validation.
func TestSysbenchAdapter_ValidateConfig(t *testing.T) {
	ctx := context.Background()
	adapter := NewSysbenchAdapter()

	tmpl := &template.Template{
		ID:            "sysbench-oltp-read-write",
		Name:          "Sysbench OLTP",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql", "postgresql"},
		CommandTemplate: template.CommandTemplate{
			Run: "run",
		},
		OutputParser: template.OutputParser{
			Type: template.ParserTypeRegex,
		},
	}

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Connection: &connection.MySQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test",
						Name: "Test",
					},
					Host:     "localhost",
					Port:     3306,
					Database: "testdb",
					Username: "root",
				},
				Template: tmpl,
				Parameters: map[string]interface{}{
					"threads": 8,
					"time":    60,
				},
			},
			wantErr: false,
		},
		{
			name: "missing connection",
			config: &Config{
				Template: tmpl,
				Parameters: map[string]interface{}{
					"threads": 8,
					"time":    60,
				},
			},
			wantErr: true,
		},
		{
			name: "missing template",
			config: &Config{
				Connection: &connection.MySQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test",
						Name: "Test",
					},
					Host:     "localhost",
					Port:     3306,
					Database: "testdb",
					Username: "root",
				},
				Parameters: map[string]interface{}{
					"threads": 8,
					"time":    60,
				},
			},
			wantErr: true,
		},
		{
			name: "missing threads parameter",
			config: &Config{
				Connection: &connection.MySQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test",
						Name: "Test",
					},
					Host:     "localhost",
					Port:     3306,
					Database: "testdb",
					Username: "root",
				},
				Template: tmpl,
				Parameters: map[string]interface{}{
					"time": 60,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid threads value",
			config: &Config{
				Connection: &connection.MySQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test",
						Name: "Test",
					},
					Host:     "localhost",
					Port:     3306,
					Database: "testdb",
					Username: "root",
				},
				Template: tmpl,
				Parameters: map[string]interface{}{
					"threads": 0,
					"time":    60,
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported database",
			config: &Config{
				Connection: &connection.OracleConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test",
						Name: "Test",
					},
					Host:     "localhost",
					Port:     1521,
					Username: "sys",
				},
				Template: tmpl,
				Parameters: map[string]interface{}{
					"threads": 8,
					"time":    60,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateConfig(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSysbenchAdapter_PostgreSQL tests PostgreSQL support.
func TestSysbenchAdapter_PostgreSQL(t *testing.T) {
	ctx := context.Background()
	adapter := NewSysbenchAdapter()

	conn := &connection.PostgreSQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn",
			Name: "Test PG",
		},
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "password",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"threads": 4,
			"time":    30,
		},
		WorkDir: "/tmp/work",
	}

	cmd, err := adapter.BuildRunCommand(ctx, config)
	if err != nil {
		t.Fatalf("BuildRunCommand() failed: %v", err)
	}

	if !strings.Contains(cmd.CmdLine, "sysbench") {
		t.Errorf("CmdLine should contain 'sysbench', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "pgsql") {
		t.Errorf("CmdLine should contain 'pgsql', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "--pgsql-host=localhost") {
		t.Errorf("CmdLine should contain '--pgsql-host=localhost', got: %s", cmd.CmdLine)
	}
	if !strings.Contains(cmd.CmdLine, "--pgsql-port=5432") {
		t.Errorf("CmdLine should contain '--pgsql-port=5432', got: %s", cmd.CmdLine)
	}
}

// TestSysbenchAdapter_ParseIntermediateOutput tests intermediate output parsing.
func TestSysbenchAdapter_ParseIntermediateOutput(t *testing.T) {
	adapter := NewSysbenchAdapter()

	line := "[ 5s ] threads: 8 tps: 1234.56 qps: 5678.90 (rt: 6.45ms) 95%: 12.34ms"

	sample := adapter.ParseIntermediateOutput(line)

	if sample.TPS != 1234.56 {
		t.Errorf("TPS = %v, want 1234.56", sample.TPS)
	}
	if sample.LatencyAvg != 6.45 {
		t.Errorf("LatencyAvg = %v, want 6.45", sample.LatencyAvg)
	}
}
