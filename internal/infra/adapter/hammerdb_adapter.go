// Package adapter provides HammerDB benchmark tool adapter.
// Implements: Phase 3 - HammerDB Tool Adapter
package adapter

import (
	"bufio"
	"context"
	"encoding/base64"
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
// For SQL Server, this is a two-phase operation: cleanup + rebuild schema/data.
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
		Description: "Prepare: full cleanup + rebuild schema/data",
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
// Uses HammerDB 4.10+ format with diset connection mssqls_* for SQL Server.
// For SQL Server, this includes: DROP DATABASE -> CREATE DATABASE -> buildschema
func (a *HammerDBAdapter) buildBuildschemaCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Get HammerDB parameters
	databaseName := a.getStringParam(config.Parameters, "database_name", "tpcc")
	warehouses := a.getIntParam(config.Parameters, "warehouses", 1)
	buildUsers := a.getIntParam(config.Parameters, "build_users", 1)

	// Build TCL script for buildschema
	var script strings.Builder

	// Database type
	dbType := a.getDBType(conn)
	script.WriteString(fmt.Sprintf("dbset db %s\n", dbType))

	// Set benchmark (TPROC-C is the correct HammerDB benchmark name)
	script.WriteString("dbset bm TPROC-C\n")

	// SQL Server specific: build complete prepare script
	if conn.GetType() == connection.DatabaseTypeSQLServer {
		a.buildSQLServerConnection(&script, conn)
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_dbase %s\n", databaseName))
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_count_ware %d\n", warehouses))
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_num_vu %d\n", buildUsers))
		script.WriteString("vudestroy\n")
		script.WriteString("buildschema\n")
	} else {
		// Non-SQL Server: use legacy format
		connectionStr := a.buildConnectionString(conn)
		script.WriteString(fmt.Sprintf("disconn %s\n", connectionStr))
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_dbase %s\n", databaseName))
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_count_ware %d\n", warehouses))
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_num_vu %d\n", buildUsers))
		script.WriteString("buildschema\n")
	}

	// Build command line with auto-confirmation using yes pipe
	// Use base64 encoding to avoid quoting issues with heredoc and nested quotes
	encodedScript := base64.StdEncoding.EncodeToString([]byte(script.String()))
	cmdLine := fmt.Sprintf("bash -c 'echo %s | base64 -d > /tmp/hdb_$$.tcl && yes | %s auto /tmp/hdb_$$.tcl; rm -f /tmp/hdb_$$.tcl'", encodedScript, a.HammerDBPath)

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

	// Use base64 encoding to avoid quoting issues with heredoc and nested quotes
	// Command flow:
	// 1. Decode base64 script and write to temp file (using $$ for unique filename in same shell context)
	// 2. Execute hammerdbcli auto with temp file - "auto" mode waits for script completion
	// 3. Cleanup temp file after execution
	// All in single shell context to ensure $$ is consistent
	encodedScript := base64.StdEncoding.EncodeToString([]byte(script))
	cmdLine := fmt.Sprintf("bash -c 'echo %s | base64 -d > /tmp/hdb_run_$$.tcl && %s auto /tmp/hdb_run_$$.tcl; rm -f /tmp/hdb_run_$$.tcl'", encodedScript, a.HammerDBPath)

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
		Env:     []string{"TMP=/tmp", "TMPDIR=/tmp", "TEMP=/tmp"},
	}, nil
}

// BuildCleanupCommand builds the command for cleanup phase.
// Uses HammerDB's deleteschema command to drop the entire database.
// This follows the standard HammerDB workflow: deleteschema -> buildschema -> loadscript -> vurun
// IMPORTANT: deleteschema requires confirmation, we use 'yes' pipe to auto-confirm
func (a *HammerDBAdapter) BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// All databases use the TCL script approach with deleteschema
	script := a.buildScript(ctx, conn, config, "cleanup")

	// Use base64 encoding to avoid quoting issues with heredoc and nested quotes
	// Command flow:
	// 1. Decode base64 script and write to temp file (using $$ for unique filename in same shell context)
	// 2. Execute hammerdbcli auto with temp file
	// 3. yes pipe auto-confirms the prompt
	// 4. Cleanup temp file after execution
	// All in single shell context to ensure $$ is consistent
	encodedScript := base64.StdEncoding.EncodeToString([]byte(script))
	cmdLine := fmt.Sprintf("bash -c 'echo %s | base64 -d > /tmp/hdb_$$.tcl && yes | %s auto /tmp/hdb_$$.tcl; rm -f /tmp/hdb_$$.tcl'", encodedScript, a.HammerDBPath)

	return &Command{
		CmdLine:     cmdLine,
		WorkDir:     config.WorkDir,
		Env:         []string{"TMP=/tmp", "TMPDIR=/tmp", "TEMP=/tmp"},
		Description: "Cleanup: Delete TPC-C schema (DROP DATABASE)",
		StepName:    "Delete Schema",
	}, nil
}

// buildScript builds a TCL script for HammerDB.
func (a *HammerDBAdapter) buildScript(ctx context.Context, conn connection.Connection, config *Config, phase string) string {
	var script strings.Builder

	// Database type
	dbType := a.getDBType(conn)
	script.WriteString(fmt.Sprintf("dbset db %s\n", dbType))

	// SQL Server uses new HammerDB 4.10+ format with diset connection mssqls_*
	if conn.GetType() == connection.DatabaseTypeSQLServer {
		// Build connection settings using HammerDB 4.10+ format
		a.buildSQLServerConnection(&script, conn)
		a.buildSQLServerScript(&script, config, phase)
	} else {
		// Legacy format for other databases
		connectionStr := a.buildConnectionString(conn)
		script.WriteString(fmt.Sprintf("disconn %s\n", connectionStr))
		a.buildGenericScript(&script, config)
	}

	return script.String()
}

// buildSQLServerConnection builds SQL Server connection settings using HammerDB 4.10+ format.
// Uses Linux-specific parameters (mssqls_linux_server, mssqls_tcp) for HammerDB on Linux.
func (a *HammerDBAdapter) buildSQLServerConnection(script *strings.Builder, conn connection.Connection) {
	sqlServerConn, ok := conn.(*connection.SQLServerConnection)
	if !ok {
		return
	}

	// HammerDB 4.10+ connection format for Linux
	// Use mssqls_linux_server instead of mssqls_server for Linux
	script.WriteString(fmt.Sprintf("diset connection mssqls_linux_server %s\n", sqlServerConn.Host))
	// Enable TCP connection (required for Linux)
	script.WriteString("diset connection mssqls_tcp true\n")
	script.WriteString(fmt.Sprintf("diset connection mssqls_port %d\n", sqlServerConn.Port))
	script.WriteString(fmt.Sprintf("diset connection mssqls_uid %s\n", sqlServerConn.Username))
	script.WriteString(fmt.Sprintf("diset connection mssqls_pass %s\n", sqlServerConn.Password))
}

// buildSQLServerScript builds SQL Server specific TCL script.
func (a *HammerDBAdapter) buildSQLServerScript(script *strings.Builder, config *Config, phase string) {
	// Get HammerDB parameters with defaults
	warehouses := a.getIntParam(config.Parameters, "warehouses", 1)
	users := a.getIntParam(config.Parameters, "virtual_users", 1) // Get from Tasks page
	buildUsers := a.getIntParam(config.Parameters, "build_users", 1)
	rampUpSeconds := a.getIntParam(config.Parameters, "rampup", 0) // Default 0 = no rampup
	durationSeconds := a.getIntParam(config.Parameters, "duration", 60)
	iterations := a.getIntParam(config.Parameters, "iterations", 1000000)
	driver := a.getStringParam(config.Parameters, "driver", "timed")
	allWarehouse := a.getBoolParam(config.Parameters, "all_warehouse", "false")
	databaseName := a.getStringParam(config.Parameters, "database_name", "tpcc")

	// Convert seconds to minutes for HammerDB (HammerDB uses minutes for duration/rampup)
	// Round up to ensure we run at least the requested time
	// Note: rampup of 0 means no rampup (immediate start)
	var rampUpMinutes, durationMinutes int
	if rampUpSeconds > 0 {
		rampUpMinutes = (rampUpSeconds + 59) / 60
		if rampUpMinutes < 1 {
			rampUpMinutes = 1
		}
	}
	durationMinutes = (durationSeconds + 59) / 60
	if durationMinutes < 1 {
		durationMinutes = 1
	}

	// Set benchmark (TPROC-C is HammerDB's benchmark name)
	script.WriteString("dbset bm TPROC-C\n")

	// Phase-specific configuration
	switch phase {
	case "prepare":
		// Build schema phase using the same HammerDB TCL that is validated manually.
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_dbase %s\n", databaseName))
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_count_ware %d\n", warehouses))
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_num_vu %d\n", buildUsers))
		script.WriteString("vudestroy\n")
		script.WriteString("buildschema\n")

	case "run":
		// Run benchmark phase with tcstart for real-time monitoring
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_driver %s\n", driver))

		// Ramp up is only used in timed mode
		if driver == "timed" {
			script.WriteString(fmt.Sprintf("diset tpcc mssqls_rampup %d\n", rampUpMinutes))
			script.WriteString(fmt.Sprintf("diset tpcc mssqls_duration %d\n", durationMinutes))
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
		// Cleanup phase - use deleteschema to drop the entire database
		// This follows the standard HammerDB workflow
		// Set the database name so deleteschema knows which database to drop
		script.WriteString(fmt.Sprintf("diset tpcc mssqls_dbase %s\n", databaseName))
		script.WriteString("deleteschema\n")
		script.WriteString("waittocomplete\n")
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
// Streams ALL output lines to the sample channel for logging, not just metrics.
func (a *HammerDBAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader) (<-chan Sample, <-chan error, *strings.Builder) {
	sampleChan := make(chan Sample, 100) // Increased buffer for all output lines
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
			originalLine := line
			line = strings.TrimSpace(line)

			// Skip empty lines
			if line == "" {
				continue
			}

			// Save to stdout buffer
			stdoutBuf.WriteString(originalLine)
			stdoutBuf.WriteString("\n")

			// Parse realtime TPM/NOPM for metrics
			// HammerDB outputs: "9420 MSSQLServer tpm" or "38040 MSSQLServer tpm"
			if strings.Contains(line, "tpm") || strings.Contains(line, "TPM") {
				re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:MSSQLServer\s+)?tpm`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
						currentTPM = val // Store actual TPM value
					}
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

			// Send EVERY line as a sample with RawLine set for logging
			// This ensures buildschema output like "Loading Items - 50000" is captured
			sample := Sample{
				Timestamp:   time.Now(),
				TPM:         currentTPM,      // Actual TPM from HammerDB
				TPS:         currentTPM / 60, // Calculate TPS from TPM
				LatencyAvg:  0,
				LatencyP95:  0,
				LatencyP99:  0,
				ErrorRate:   0,
				ThreadCount: currentUsers,
				RawLine:     line, // Include raw line for logging
			}

			select {
			case sampleChan <- sample:
			case <-ctx.Done():
				return
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
// HammerDB outputs results in format: "TEST RESULT : System achieved 12345 NOPM from 1 Virtual Users"
// Or with TPM: "TEST RESULT : System achieved 50000 TPM from 10 Virtual Users"
func (a *HammerDBAdapter) ParseFinalResults(ctx context.Context, stdout string) (*FinalResult, error) {
	result := &FinalResult{}

	lines := strings.Split(stdout, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse TEST RESULT line - format: "TEST RESULT : System achieved 12345 NOPM from 1 Virtual Users"
		// or "TEST RESULT : System achieved 50000 TPM from 10 Virtual Users"
		if strings.Contains(line, "TEST RESULT") && (strings.Contains(line, "NOPM") || strings.Contains(line, "TPM")) {
			// Try NOPM first (more specific)
			re := regexp.MustCompile(`achieved\s+(\d+(?:\.\d+)?)\s+NOPM`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.TPM = val
					result.TransactionsPerSec = val / 60 // Convert to TPS
					result.AvgTPS = result.TransactionsPerSec
				}
			} else {
				// Try TPM
				re = regexp.MustCompile(`achieved\s+(\d+(?:\.\d+)?)\s+TPM`)
				matches = re.FindStringSubmatch(line)
				if len(matches) > 1 {
					if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
						result.TPM = val
						result.TransactionsPerSec = val / 60 // Convert to TPS
						result.AvgTPS = result.TransactionsPerSec
					}
				}
			}
		}

		// Parse Average response time: "Average response time: 150.50ms"
		if strings.Contains(line, "Average response time") {
			re := regexp.MustCompile(`Average response time:\s*(\d+(?:\.\d+)?)\s*ms`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyAvg = val
				}
			}
		}

		// Parse Min response time: "Min response time: 50.25ms"
		if strings.Contains(line, "Min response time") {
			re := regexp.MustCompile(`Min response time:\s*(\d+(?:\.\d+)?)\s*ms`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyMin = val
				}
			}
		}

		// Parse Max response time: "Max response time: 450.75ms"
		if strings.Contains(line, "Max response time") {
			re := regexp.MustCompile(`Max response time:\s*(\d+(?:\.\d+)?)\s*ms`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyMax = val
				}
			}
		}

		// Parse 95th percentile response: "95th percentile response: 300.00ms"
		if strings.Contains(line, "95th percentile") {
			re := regexp.MustCompile(`95th percentile(?:\s+response)?:\s*(\d+(?:\.\d+)?)\s*ms`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyP95 = val
				}
			}
		}

		// Parse Errors: "Errors: 5"
		if strings.Contains(line, "Errors:") {
			re := regexp.MustCompile(`Errors:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.IgnoredErrors = val
				}
			}
		}

		// Parse Total time: "Total time: 120.0s"
		if strings.Contains(line, "Total time:") {
			re := regexp.MustCompile(`Total time:\s*(\d+(?:\.\d+)?)\s*s`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.TotalTime = val
				}
			}
		}
	}

	// Calculate transaction count if we have TPM and total time
	if result.TPM > 0 && result.TotalTime > 0 {
		result.TotalTransactions = int64(result.TPM * (result.TotalTime / 60))
	}

	return result, nil
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
