// Package adapter provides HammerDB benchmark tool adapter.
// Implements: Phase 3 - HammerDB Tool Adapter
package adapter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// HammerDBAdapter implements BenchmarkAdapter for HammerDB tool.
// Implements: REQ-EXEC-001, REQ-EXEC-002, REQ-EXEC-004
type HammerDBAdapter struct {
	// Path to hammerdb executable (optional, if empty uses PATH)
	HammerDBPath string
}

// NewHammerDBAdapter creates a new hammerdb adapter.
func NewHammerDBAdapter() *HammerDBAdapter {
	return &HammerDBAdapter{
		HammerDBPath: "hammerdbcli", // Default to CLI
	}
}

// Type returns the adapter type.
func (a *HammerDBAdapter) Type() AdapterType {
	return AdapterTypeHammerDB
}

// BuildPrepareCommand builds the command for data preparation phase.
// For SQL Server, this is a two-phase operation: cleanup (TRUNCATE) + buildschema.
func (a *HammerDBAdapter) BuildPrepareCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Check if this is a SQL Server connection
	if conn.GetType() != connection.DatabaseTypeSQLServer {
		// For non-SQL Server, use the existing TCL script approach
		script := a.buildScript(ctx, conn, config, "prepare")
		cmdLine := fmt.Sprintf("echo '%s' | %s", script, a.HammerDBPath)
		return &Command{
			CmdLine: cmdLine,
			WorkDir: config.WorkDir,
			Env:     []string{"TMP=/tmp", "TMPDIR=/tmp", "TEMP=/tmp"},
		}, nil
	}

	// SQL Server: Use command sequence (cleanup + buildschema)

	// Step 1: Build cleanup command (TRUNCATE)
	cleanupCmd, err := a.buildCleanupStepCommand(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("build cleanup step: %w", err)
	}

	// Step 2: Build buildschema command
	buildschemaCmd, err := a.buildBuildschemaCommand(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("build buildschema step: %w", err)
	}

	// Return a command with the sequence
	return &Command{
		Commands: []*Command{
			cleanupCmd,
			buildschemaCmd,
		},
		Description: "Prepare: TRUNCATE tables + buildschema",
		WorkDir:     config.WorkDir,
		Env:         []string{"TMP=/tmp", "TMPDIR=/tmp", "TEMP=/tmp"},
	}, nil
}

// buildCleanupStepCommand builds the cleanup step for the prepare sequence.
func (a *HammerDBAdapter) buildCleanupStepCommand(ctx context.Context, config *Config) (*Command, error) {
	// Reuse BuildCleanupCommand logic but don't add retry (prepare has its own retry)
	cleanupCmd, err := a.BuildCleanupCommand(ctx, config)
	if err != nil {
		return nil, err
	}

	// Override step name for clarity
	cleanupCmd.StepName = "Step 1: Cleanup existing data"
	cleanupCmd.Description = "TRUNCATE all TPC-C tables before loading fresh data"

	return cleanupCmd, nil
}

// buildBuildschemaCommand builds the HammerDB buildschema command.
func (a *HammerDBAdapter) buildBuildschemaCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Get HammerDB parameters
	databaseName := a.getStringParam(config.Parameters, "database_name", "tpcc")
	warehouses := a.getIntParam(config.Parameters, "warehouses", 1)
	buildUsers := a.getIntParam(config.Parameters, "build_users", 1)

	// Build TCL script for buildschema
	var script strings.Builder

	// Database type and connection
	dbType := a.getDBType(conn)
	connectionStr := a.buildConnectionString(conn)

	script.WriteString(fmt.Sprintf("dbtype %s\n", dbType))
	script.WriteString(fmt.Sprintf("disconn %s\n", connectionStr))

	// Create database first
	a.buildDatabaseCreationScript(&script, databaseName)

	// Set buildschema parameters
	script.WriteString(fmt.Sprintf("diset tpcc mssqls_count_ware %d\n", warehouses))
	script.WriteString(fmt.Sprintf("diset tpcc mssqls_num_vu %d\n", buildUsers))
	script.WriteString("dbset bm TPC-C\n")

	// Add "yes" confirmation to automate buildschema
	script.WriteString("proc buildschema_auto {args} {\n")
	script.WriteString("    buildschema\n")
	script.WriteString("    after 1000\n")
	script.WriteString("    # Send 'yes' to confirm\n")
	script.WriteString("    # HammerDB will ask for confirmation\n")
	script.WriteString("    # We use 'yes' command to automate\n")
	script.WriteString("}\n")
	script.WriteString("buildschema_auto\n")

	// Build command line with auto-confirmation
	cmdLine := fmt.Sprintf("yes | %s << 'EOF'\n%s\nEOF\n", a.HammerDBPath, script.String())

	// Configure retry for buildschema (it can fail due to transient issues)
	retryConfig := &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  5 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}

	return &Command{
		CmdLine:     cmdLine,
		WorkDir:     config.WorkDir,
		Env:         []string{"TMP=/tmp", "TMPDIR=/tmp", "TEMP=/tmp"},
		Retry:       retryConfig,
		StepName:    "Step 2: Build schema and load data",
		Description: fmt.Sprintf("Create TPC-C schema and load %d warehouse(s)", warehouses),
	}, nil
}

// BuildRunCommand builds the command for the main benchmark run.
func (a *HammerDBAdapter) BuildRunCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Build run script
	script := a.buildScript(ctx, conn, config, "run")

	cmdLine := fmt.Sprintf("echo '%s' | %s", script, a.HammerDBPath)

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
		Env:     []string{"TMP=/tmp", "TMPDIR=/tmp", "TEMP=/tmp"},
	}, nil
}

// BuildCleanupCommand builds the command for cleanup phase.
// For SQL Server, uses sqlcmd to TRUNCATE all TPC-C tables.
func (a *HammerDBAdapter) BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Check if this is a SQL Server connection
	if conn.GetType() != connection.DatabaseTypeSQLServer {
		// For non-SQL Server, use the existing TCL script approach
		script := a.buildScript(ctx, conn, config, "cleanup")
		cmdLine := fmt.Sprintf("echo '%s' | %s", script, a.HammerDBPath)
		return &Command{
			CmdLine: cmdLine,
			WorkDir: config.WorkDir,
			Env:     []string{"TMP=/tmp", "TMPDIR=/tmp", "TEMP=/tmp"},
		}, nil
	}

	// SQL Server: use sqlcmd to TRUNCATE tables
	sqlServerConn, ok := conn.(*connection.SQLServerConnection)
	if !ok {
		return nil, fmt.Errorf("invalid connection type for SQL Server")
	}

	// Build the TRUNCATE script
	truncateScript := a.buildTruncateScript()

	// Build sqlcmd command line
	cmdLine, err := a.buildSQLCmdLine(sqlServerConn, truncateScript)
	if err != nil {
		return nil, fmt.Errorf("build sqlcmd command: %w", err)
	}

	// Configure retry for cleanup operations
	// TRUNCATE can fail due to locks, so we retry
	retryConfig := &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
	}

	return &Command{
		CmdLine:     cmdLine,
		WorkDir:     config.WorkDir,
		Env:         []string{"TMP=/tmp", "TMPDIR=/tmp", "TEMP=/tmp"},
		Retry:       retryConfig,
		Description: "Cleanup: TRUNCATE all TPC-C tables",
		StepName:    "TRUNCATE Tables",
	}, nil
}

// buildTruncateScript builds SQL script to TRUNCATE all TPC-C tables.
// This is more efficient than DROP TABLE for SQL Server.
func (a *HammerDBAdapter) buildTruncateScript() string {
	// List of TPC-C tables in correct dependency order (child tables first)
	// We need to disable foreign key constraints before TRUNCATE
	var script strings.Builder

	// Disable foreign key constraints
	script.WriteString("SET XACT_ABORT ON;\n")
	script.WriteString("BEGIN TRANSACTION;\n")
	script.WriteString("-- Disable foreign key constraints\n")
	script.WriteString("EXEC sp_msforeachtable \"ALTER TABLE ? NOCHECK CONSTRAINT all\";\n")
	script.WriteString("\n")

	// TRUNCATE all TPC-C tables
	// Order: child tables first, then parent tables
	tables := []string{
		"order_line", // Child of orders
		"new_order",  // Child of orders
		"history",    // Child of customer
		"orders",     // Child of district, customer, stock, item, warehouse
		"stock",      // Child of warehouse
		"district",   // Child of warehouse
		"customer",   // Child of district
		"item",       // Independent table
		"warehouse",  // Root table
	}

	for _, table := range tables {
		script.WriteString(fmt.Sprintf("IF OBJECT_ID('%s', 'U') IS NOT NULL\n", table))
		script.WriteString(fmt.Sprintf("    TRUNCATE TABLE [%s];\n", table))
		script.WriteString(fmt.Sprintf("    PRINT 'Truncated table: %s';\n", table))
		script.WriteString("ELSE\n")
		script.WriteString(fmt.Sprintf("    PRINT 'Table does not exist: %s';\n", table))
		script.WriteString("\n")
	}

	// Re-enable foreign key constraints
	script.WriteString("-- Re-enable foreign key constraints\n")
	script.WriteString("EXEC sp_msforeachtable \"ALTER TABLE ? CHECK CONSTRAINT all\";\n")
	script.WriteString("COMMIT TRANSACTION;\n")
	script.WriteString("PRINT 'Cleanup completed successfully';\n")

	return script.String()
}

// buildSQLCmdLine builds a sqlcmd command line to execute SQL script.
func (a *HammerDBAdapter) buildSQLCmdLine(conn *connection.SQLServerConnection, script string) (string, error) {
	// Build host:port string
	hostPort := fmt.Sprintf("%s,%d", conn.Host, conn.Port)

	// Build sqlcmd command with proper options
	// -S: server
	// -U: username
	// -P: password
	// -d: database
	// -C: trust server certificate
	// -Q: execute query and exit
	// Note: We use -Q with the script, but need to escape properly
	// Instead, we'll use echo with pipe to handle multi-line scripts

	cmdLine := fmt.Sprintf("echo '%s' | sqlcmd -S %s -U %s -P '%s' -d %s -C -b -r1",
		script,
		hostPort,
		conn.Username,
		conn.Password,
		conn.Database)

	return cmdLine, nil
}

// buildScript builds a TCL script for HammerDB.
func (a *HammerDBAdapter) buildScript(ctx context.Context, conn connection.Connection, config *Config, phase string) string {
	var script strings.Builder

	// Database type and connection
	dbType := a.getDBType(conn)
	connectionStr := a.buildConnectionString(conn)

	script.WriteString(fmt.Sprintf("dbtype %s\n", dbType))
	script.WriteString(fmt.Sprintf("disconn %s\n", connectionStr))

	// SQL Server specific parameters
	if conn.GetType() == connection.DatabaseTypeSQLServer {
		a.buildSQLServerScript(&script, config, phase)
	} else {
		a.buildGenericScript(&script, config)
	}

	return script.String()
}

// buildSQLServerScript builds SQL Server specific TCL script.
func (a *HammerDBAdapter) buildSQLServerScript(script *strings.Builder, config *Config, phase string) {
	// Get HammerDB parameters with defaults
	warehouses := a.getIntParam(config.Parameters, "warehouses", 1)
	users := a.getIntParam(config.Parameters, "virtual_users", 1) // Get from Tasks page
	buildUsers := a.getIntParam(config.Parameters, "build_users", 1)
	rampUp := a.getIntParam(config.Parameters, "rampup", 1)
	duration := a.getIntParam(config.Parameters, "duration", 1)
	iterations := a.getIntParam(config.Parameters, "iterations", 1000000)
	driver := a.getStringParam(config.Parameters, "driver", "timed")
	allWarehouse := a.getBoolParam(config.Parameters, "all_warehouse", "false")
	databaseName := a.getStringParam(config.Parameters, "database_name", "tpcc")

	// Set benchmark to TPC-C
	script.WriteString("dbset bm TPC-C\n")

	// Phase-specific configuration
	switch phase {
	case "prepare":
		// Create database first (before buildschema)
		a.buildDatabaseCreationScript(script, databaseName)

		// Build schema phase
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_count_ware %d\n", warehouses))
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_num_vu %d\n", buildUsers))
		script.WriteString("buildschema\n")
		script.WriteString("waittocomplete\n")

	case "run":
		// Run benchmark phase with tcstart for real-time monitoring
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_driver %s\n", driver))

		// Ramp up is only used in timed mode
		if driver == "timed" {
			script.WriteString(fmt.Sprintf("diset tpcc mssqls_rampup %d\n", rampUp))
			script.WriteString(fmt.Sprintf("diset tpcc mssqls_duration %d\n", duration))
		} else {
			// Iterations mode
			script.WriteString(fmt.Sprintf("diset tpcc mssqls_iterations %d\n", iterations))
		}

		script.WriteString(fmt.Sprintf("diset tpcc mssqls_allwarehouse %s\n", allWarehouse))

		script.WriteString("loadscript\n")
		script.WriteString(fmt.Sprintf("vuset vu %d\n", users))
		script.WriteString("vucreate\n")

		// Start transaction counter for real-time TPM/NOPM monitoring
		// Set refresh rate to 1 second for real-time updates
		script.WriteString("tcset refreshrate 1\n")
		script.WriteString("tcstart\n")

		// Run the benchmark
		script.WriteString("vurun\n")

		// Stop transaction counter after run completes
		script.WriteString("tcstop\n")
		script.WriteString("vudestroy\n")

	case "cleanup":
		// Cleanup phase - no specific cleanup for SQL Server
		script.WriteString("delete virtualmachine\n")
	}
}

// buildGenericScript builds generic TCL script for non-SQL Server databases.
func (a *HammerDBAdapter) buildGenericScript(script *strings.Builder, config *Config) {
	script.WriteString(fmt.Sprintf("vu %d\n", a.getIntParam(config.Parameters, "virtual_users", 1)))
	script.WriteString(fmt.Sprintf("vucount %d\n", a.getIntParam(config.Parameters, "vu_count", 1)))
	script.WriteString(fmt.Sprintf("vuverbose %s\n", a.getBoolParam(config.Parameters, "vu_verbose", "false")))
	script.WriteString(fmt.Sprintf("iterations %d\n", a.getIntParam(config.Parameters, "iterations", 1)))
	script.WriteString(fmt.Sprintf("tcname %s\n", a.getStringParam(config.Parameters, "testcase", "TPC-C")))
	script.WriteString(fmt.Sprintf("tcstatus %s\n", a.getStringParam(config.Parameters, "tcstatus", "")))
	script.WriteString(fmt.Sprintf("rampup %d\n", a.getIntParam(config.Parameters, "rampup", 0)))
	script.WriteString(fmt.Sprintf("duration %d\n", a.getIntParam(config.Parameters, "duration", 1)))
	script.WriteString(fmt.Sprintf("alliterations %s\n", a.getBoolParam(config.Parameters, "all_iterations", "true")))
	script.WriteString(fmt.Sprintf("times %s\n", a.getBoolParam(config.Parameters, "times", "true")))
	script.WriteString(fmt.Sprintf("background %s\n", a.getBoolParam(config.Parameters, "background", "false")))
	script.WriteString(fmt.Sprintf("nozip %s\n", a.getBoolParam(config.Parameters, "no_zip", "false")))
	script.WriteString(fmt.Sprintf("suppress_output %s\n", a.getBoolParam(config.Parameters, "suppress_output", "false")))
	script.WriteString(fmt.Sprintf("hwscale %s\n", a.getBoolParam(config.Parameters, "hwscale", "false")))
	script.WriteString(fmt.Sprintf("hwmem %s\n", a.getBoolParam(config.Parameters, "hwmem", "false")))
	script.WriteString(fmt.Sprintf("clearlog %s\n", a.getBoolParam(config.Parameters, "clear_log", "true")))
	script.WriteString(fmt.Sprintf("logtotemp %s\n", a.getBoolParam(config.Parameters, "log_to_temp", "false")))
}

// buildDatabaseCreationScript builds TCL script to create SQL Server database with appropriate settings.
func (a *HammerDBAdapter) buildDatabaseCreationScript(script *strings.Builder, databaseName string) {
	// Check if database already exists, if not create it
	// Also detect AlwaysOn and set recovery mode accordingly
	script.WriteString("puts \"Checking database existence...\"\n")
	script.WriteString(fmt.Sprintf("set db_exists [tcldb \"IF EXISTS (SELECT name FROM sys.databases WHERE name = '%s') SELECT 1 ELSE SELECT 0\"]\n", databaseName))
	script.WriteString("if {$db_exists == 0} {\n")
	script.WriteString("    puts \"Database does not exist, creating...\"\n")

	// Check for AlwaysOn availability groups
	script.WriteString("    puts \"Checking for AlwaysOn availability groups...\"\n")
	script.WriteString("    set is_alwayson [tcldb \"IF EXISTS (SELECT * FROM sys.dm_hadr_ag_replicas WHERE is_local = 1) SELECT 1 ELSE SELECT 0\"]\n")
	script.WriteString("    if {$is_alwayson == 1} {\n")
	script.WriteString("        puts \"AlwaysOn detected: setting recovery mode to FULL\"\n")
	script.WriteString("        set recovery_mode \"FULL\"\n")
	script.WriteString("    } else {\n")
	script.WriteString("        puts \"Standalone instance: setting recovery mode to SIMPLE\"\n")
	script.WriteString("        set recovery_mode \"SIMPLE\"\n")
	script.WriteString("    }\n")

	// Create database with specified file settings
	script.WriteString(fmt.Sprintf("    puts \"Creating database %s with recovery mode $recovery_mode...\"\n", databaseName))
	// Build the SQL statement directly using Go string formatting
	createSQL := fmt.Sprintf("CREATE DATABASE [%s] ON PRIMARY (NAME = N'%s_data', FILENAME = N'/var/opt/mssql/data/%s_data.mdf', SIZE = 20480MB, MAXSIZE = UNLIMITED, FILEGROWTH = 500MB) LOG ON (NAME = N'%s_log', FILENAME = N'/var/opt/mssql/data/%s_log.ldf', SIZE = 10240MB, MAXSIZE = UNLIMITED, FILEGROWTH = 500MB)",
		databaseName, databaseName, databaseName, databaseName, databaseName)
	script.WriteString(fmt.Sprintf("    tcldb {%s}\n", createSQL))
	script.WriteString(fmt.Sprintf("    puts \"Database %s created successfully\"\n", databaseName))

	// Set recovery mode
	script.WriteString(fmt.Sprintf("    puts \"Setting recovery mode to $recovery_mode for database %s...\"\n", databaseName))
	script.WriteString(fmt.Sprintf("    tcldb \"ALTER DATABASE [%s] SET RECOVERY $recovery_mode WITH NO_WAIT\"\n", databaseName))
	script.WriteString("    puts \"Recovery mode set successfully\"\n")
	script.WriteString("} else {\n")
	script.WriteString(fmt.Sprintf("    puts \"Database %s already exists, skipping creation\"\n", databaseName))
	script.WriteString("}\n")
	script.WriteString("\n")
}

// ParseRunOutput parses the output from a hammerdb run.
func (a *HammerDBAdapter) ParseRunOutput(ctx context.Context, stdout string, stderr string) (*Result, error) {
	result := &Result{
		RawOutput: stdout,
	}

	lines := strings.Split(stdout, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse TPM (Transactions Per Minute) or NOPM (New Orders Per Minute)
		// Format: "TEST RESULT : System achieved 12345 NOPM from 1 Virtual Users"
		if strings.Contains(line, "NOPM") || strings.Contains(line, "TPM") {
			re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:NOPM|TPM)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.TPS = val / 60 // Convert to TPS
				}
			}
		}

		// Parse response time
		// Format: "Response Time: 250ms" or "Average response time: 250.00ms"
		if strings.Contains(line, "Response") || strings.Contains(line, "response") {
			if strings.Contains(line, "Average") {
				re := regexp.MustCompile(`(?:Average|average)\s+response\s+time:\s*(\d+(?:\.\d+)?)\s*ms`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
						result.LatencyAvg = val
					}
				}
			}
		}

		// Parse 95th percentile
		if strings.Contains(line, "95th") || strings.Contains(line, "95th") {
			re := regexp.MustCompile(`95(?:th)?\s+percentile:\s*(\d+(?:\.\d+)?)\s*ms`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyP95 = val
				}
			}
		}

		// Parse errors
		if strings.Contains(strings.ToLower(line), "error") || strings.Contains(strings.ToLower(line), "failed") {
			re := regexp.MustCompile(`error[s]?:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.TotalErrors = val
				}
			}
		}

		// Parse transactions count
		if strings.Contains(strings.ToLower(line), "transaction") {
			re := regexp.MustCompile(`transaction[s]?:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.TotalTransactions = val
				}
			}
		}
	}

	// Calculate error rate
	if result.TotalTransactions > 0 {
		result.ErrorRate = (float64(result.TotalErrors) / float64(result.TotalTransactions)) * 100
	}

	// Set default duration if not parsed
	if result.Duration == 0 {
		result.Duration = 60 * time.Second
	}

	return result, nil
}

// StartRealtimeCollection starts realtime metric collection from hammerdb output.
func (a *HammerDBAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader) (<-chan Sample, <-chan error, *strings.Builder) {
	sampleChan := make(chan Sample, 10)
	errChan := make(chan error, 1)
	var stdoutBuf strings.Builder

	go func() {
		defer close(sampleChan)
		defer close(errChan)

		scanner := bufio.NewScanner(stdout)
		currentTPM := 0.0
		currentUsers := 1

		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)

			// Save to stdout buffer
			stdoutBuf.WriteString(line)
			stdoutBuf.WriteString("\n")

			// Parse realtime TPM/NOPM
			if strings.Contains(line, "NOPM") || strings.Contains(line, "TPM") {
				re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:NOPM|TPM)`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
						currentTPM = val / 60
					}
				}

				sample := Sample{
					Timestamp:   time.Now(),
					TPS:         currentTPM,
					LatencyAvg:  0,
					LatencyP95:  0,
					LatencyP99:  0,
					ErrorRate:   0,
					ThreadCount: currentUsers,
				}

				select {
				case sampleChan <- sample:
				case <-ctx.Done():
					return
				}
			}

			// Parse virtual user count
			if strings.Contains(line, "Virtual") && strings.Contains(line, "Users") {
				re := regexp.MustCompile(`(\d+)\s+Virtual\s+Users`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					if val, err := strconv.Atoi(matches[1]); err == nil {
						currentUsers = val
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case errChan <- fmt.Errorf("scanner error: %w", err):
			case <-ctx.Done():
			}
		}
	}()

	return sampleChan, errChan, &stdoutBuf
}

// ParseFinalResults parses final results from hammerdb output.
// TODO: Implement hammerdb-specific parsing
func (a *HammerDBAdapter) ParseFinalResults(ctx context.Context, stdout string) (*FinalResult, error) {
	// Stub implementation for now
	return &FinalResult{}, fmt.Errorf("parse final results not implemented for hammerdb")
}

// ValidateConfig validates the configuration for hammerdb.
func (a *HammerDBAdapter) ValidateConfig(ctx context.Context, config *Config) error {
	if config == nil {
		return fmt.Errorf("config is required")
	}

	if config.Connection == nil {
		return fmt.Errorf("connection is required")
	}

	// HammerDB supports multiple database types
	if !a.SupportsDatabase(config.Connection.GetType()) {
		return fmt.Errorf("hammerdb does not support database type %s", config.Connection.GetType())
	}

	// Validate connection
	if err := config.Connection.Validate(); err != nil {
		return fmt.Errorf("invalid connection: %w", err)
	}

	return nil
}

// SupportsDatabase checks if hammerdb supports the given database type.
func (a *HammerDBAdapter) SupportsDatabase(dbType connection.DatabaseType) bool {
	switch dbType {
	case connection.DatabaseTypeMySQL,
		connection.DatabaseTypeOracle,
		connection.DatabaseTypeSQLServer,
		connection.DatabaseTypePostgreSQL:
		return true
	default:
		return false
	}
}

// getDBType returns the HammerDB database type string.
func (a *HammerDBAdapter) getDBType(conn connection.Connection) string {
	switch conn.GetType() {
	case connection.DatabaseTypeMySQL:
		return "MySQL"
	case connection.DatabaseTypeOracle:
		return "Oracle"
	case connection.DatabaseTypeSQLServer:
		return "MSSQLServer"
	case connection.DatabaseTypePostgreSQL:
		return "Postgres"
	default:
		return "Unknown"
	}
}

// buildConnectionString builds a HammerDB connection string.
func (a *HammerDBAdapter) buildConnectionString(conn connection.Connection) string {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		return fmt.Sprintf("%s@%s:%d/%s", c.Username, c.Host, c.Port, c.Database)
	case *connection.OracleConnection:
		if c.ServiceName != "" {
			return fmt.Sprintf("%s@//%s:%d/%s", c.Username, c.Host, c.Port, c.ServiceName)
		}
		return fmt.Sprintf("%s@%s:%d:%s", c.Username, c.Host, c.Port, c.SID)
	case *connection.SQLServerConnection:
		return fmt.Sprintf("%s@%s:%d/%s", c.Username, c.Host, c.Port, c.Database)
	case *connection.PostgreSQLConnection:
		return fmt.Sprintf("%s@%s:%d/%s", c.Username, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

// Helper functions for parameter extraction
func (a *HammerDBAdapter) getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

func (a *HammerDBAdapter) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}

func (a *HammerDBAdapter) getBoolParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case bool:
			if v {
				return "true"
			}
			return "false"
		case string:
			return v
		}
	}
	return defaultValue
}
