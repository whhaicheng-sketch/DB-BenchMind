// Package adapter provides Sysbench benchmark tool adapter.
// Implements: Phase 3 - Sysbench Tool Adapter
package adapter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// SysbenchAdapter implements BenchmarkAdapter for sysbench tool.
// Implements: REQ-EXEC-001, REQ-EXEC-002, REQ-EXEC-004
type SysbenchAdapter struct {
	// Path to sysbench executable (optional, if empty uses PATH)
	SysbenchPath string
}

// NewSysbenchAdapter creates a new sysbench adapter.
func NewSysbenchAdapter() *SysbenchAdapter {
	return &SysbenchAdapter{
		SysbenchPath: "sysbench", // Default to PATH
	}
}

// BuildCreateDatabaseCommand builds a command to create the database if it doesn't exist.
// This should be called before BuildPrepareCommand to ensure the database exists.
func (a *SysbenchAdapter) BuildCreateDatabaseCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Get database name from connection or parameters
	var dbName string
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		dbName = c.Database
	case *connection.PostgreSQLConnection:
		dbName = c.Database
	default:
		return nil, fmt.Errorf("unsupported connection type for database creation")
	}

	// If connection database is empty, try to get from parameters
	if dbName == "" {
		if db, ok := config.Parameters["db_name"].(string); ok && db != "" {
			dbName = db
		}
	}

	// If still empty, use default
	if dbName == "" {
		dbName = "sbtest"
	}

	// Build SQL command to create database if not exists
	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", dbName)

	// Build command based on database type
	var cmdLine string
	var env []string

	switch c := conn.(type) {
	case *connection.MySQLConnection:
		// MySQL: mysql -h host -P port -u user -e "CREATE DATABASE IF NOT EXISTS `db`;"
		// Use MYSQL_PWD environment variable for password
		if c.Password != "" {
			env = append(env, fmt.Sprintf("MYSQL_PWD=%s", c.Password))
		}
		slog.Info("SysbenchAdapter: Building create database command",
			"host", c.Host, "port", c.Port, "user", c.Username,
			"has_password", c.Password != "", "db", dbName)
		cmdLine = fmt.Sprintf("mysql -h %s -P %d -u %s -e \"%s\"",
			c.Host, c.Port, c.Username, createSQL)

	case *connection.PostgreSQLConnection:
		// PostgreSQL: psql -h host -p port -U user -c "CREATE DATABASE \"db\";"
		cmdLine = fmt.Sprintf("psql -h %s -p %d -U %s -c \"%s\"",
			c.Host, c.Port, c.Username, createSQL)
		// Password is set via PGPASSWORD environment variable
		if c.Password != "" {
			env = append(env, fmt.Sprintf("PGPASSWORD=%s", c.Password))
		}
	}

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
		Env:     env,
	}, nil
}

// Type returns the adapter type.
func (a *SysbenchAdapter) Type() AdapterType {
	return AdapterTypeSysbench
}

// BuildPrepareCommand builds the command for data preparation phase.
func (a *SysbenchAdapter) BuildPrepareCommand(ctx context.Context, config *Config) (*Command, error) {
	// Get connection details
	conn := config.Connection

	// Get database type for db-driver parameter
	dbDriver := a.getDBType(conn)

	// Determine sysbench script name from template ID or default
	scriptName := a.getScriptName(config.Template)

	// Build prepare command
	cmdArgs := []string{
		a.SysbenchPath,
		scriptName,
		fmt.Sprintf("--db-driver=%s", dbDriver),
	}

	// Add connection-specific arguments
	cmdArgs = append(cmdArgs, a.buildConnectionArgs(conn, config)...)

	// Add template parameters
	if tables, ok := config.Parameters["tables"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--tables=%d", tables))
	}
	if tableSize, ok := config.Parameters["table_size"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--table-size=%d", tableSize))
	}

	cmdArgs = append(cmdArgs, "prepare")

	cmdLine := strings.Join(cmdArgs, " ")

	slog.Info("SysbenchAdapter: Built prepare command",
		"cmd", cmdLine)

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
		Env:     a.buildEnvVars(conn),
	}, nil
}

// BuildRunCommand builds the command for the main benchmark run.
func (a *SysbenchAdapter) BuildRunCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Get database type for db-driver parameter
	dbDriver := a.getDBType(conn)

	// Determine sysbench script name from template ID or default
	scriptName := a.getScriptName(config.Template)

	// Build run command
	cmdArgs := []string{
		a.SysbenchPath,
		scriptName,
		fmt.Sprintf("--db-driver=%s", dbDriver),
	}

	// Add connection-specific arguments
	cmdArgs = append(cmdArgs, a.buildConnectionArgs(conn, config)...)

	// Add template parameters
	if tables, ok := config.Parameters["tables"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--tables=%d", tables))
	}
	if threads, ok := config.Parameters["threads"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--threads=%d", threads))
	}
	if runTime, ok := config.Parameters["time"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--time=%d", runTime))
	}
	if rate, ok := config.Parameters["rate"].(int); ok && rate > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--rate=%d", rate))
	}

	// Add report interval for realtime monitoring
	cmdArgs = append(cmdArgs, "--report-interval=1")

	cmdArgs = append(cmdArgs, "run")

	cmdLine := strings.Join(cmdArgs, " ")

	slog.Info("SysbenchAdapter: Built run command",
		"cmd", cmdLine)

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
		Env:     a.buildEnvVars(conn),
	}, nil
}

// BuildCleanupCommand builds the command for cleanup phase.
func (a *SysbenchAdapter) BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Get database type for db-driver parameter
	dbDriver := a.getDBType(conn)

	// Build script path or name
	scriptName := a.getScriptName(config.Template)

	cmdArgs := []string{
		a.SysbenchPath,
		scriptName,
		fmt.Sprintf("--db-driver=%s", dbDriver),
	}

	cmdArgs = append(cmdArgs, a.buildConnectionArgs(conn, config)...)

	if tables, ok := config.Parameters["tables"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--tables=%d", tables))
	}

	cmdArgs = append(cmdArgs, "cleanup")

	cmdLine := strings.Join(cmdArgs, " ")

	slog.Info("SysbenchAdapter: Built cleanup command",
		"cmd", cmdLine)

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
		Env:     a.buildEnvVars(conn),
	}, nil
}

// ParseRunOutput parses the output from a benchmark run.
// Implements: REQ-EXEC-004, REQ-EXEC-008
func (a *SysbenchAdapter) ParseRunOutput(ctx context.Context, stdout string, stderr string) (*Result, error) {
	result := &Result{
		RawOutput: stdout,
	}

	// Parse using regex patterns
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		// TPS: "transactions:                        20000  (1234.56 per sec.)"
		if matches := regexp.MustCompile(`transactions:\s*\d+\s*\(\s*(\d+\.?\d*)\s*per sec\.`).FindStringSubmatch(line); len(matches) > 1 {
			tps, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result.TPS = tps
			}
		}

		// Latency avg: "         avg:                                    6.45"
		if matches := regexp.MustCompile(`avg:\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
			avg, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result.LatencyAvg = avg
			}
		}

		// Latency min: "         min:                                    3.23"
		if matches := regexp.MustCompile(`min:\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
			min, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result.LatencyMin = min
			}
		}

		// Latency max: "         max:                                   45.67"
		if matches := regexp.MustCompile(`max:\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
			max, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result.LatencyMax = max
			}
		}

		// 95th percentile: "         95th percentile:                       12.34"
		if matches := regexp.MustCompile(`95th percentile:\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
			p95, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result.LatencyP95 = p95
			}
		}

		// Queries: "queries:                             200000 (12345.67 per sec.)"
		if matches := regexp.MustCompile(`queries:\s*\d+\s*\(\s*(\d+\.?\d*)\s*per sec\.`).FindStringSubmatch(line); len(matches) > 1 {
			qps, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result.TotalQueries = int64(qps * 60) // Approximate for 1 minute
			}
		}

		// Errors: "    ignored errors:                      0      (0.00 per sec.)"
		if matches := regexp.MustCompile(`ignored errors:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			errors, err := strconv.ParseInt(matches[1], 10, 64)
			if err == nil {
				result.TotalErrors = errors
			}
		}

		// Reconnects: "    reconnects:                        0      (0.00 per sec.)"
		if matches := regexp.MustCompile(`reconnects:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			// Track reconnects as part of errors
		}

		// Total transactions: "    total number of events:              20000"
		if matches := regexp.MustCompile(`total number of events:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			total, err := strconv.ParseInt(matches[1], 10, 64)
			if err == nil {
				result.TotalTransactions = total
			}
		}
	}

	// Calculate error rate
	if result.TotalTransactions > 0 && result.TotalErrors > 0 {
		result.ErrorRate = (float64(result.TotalErrors) / float64(result.TotalTransactions)) * 100
	}

	return result, nil
}

// StartRealtimeCollection starts realtime metric collection from the running process.
// Implements: REQ-EXEC-004 (realtime monitoring)
//
// Parses sysbench intermediate output format:
// [ 10s ] thds: 4 tps: 342.03 qps: 6846.39 (r/w/o: 4792.91/1369.02/684.46) lat (ms,95%): 13.46 err/s: 0.00 reconn/s: 0.00
//
// Also returns a buffer containing the complete stdout for final result parsing.
func (a *SysbenchAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader) (<-chan Sample, <-chan error, *strings.Builder) {
	sampleCh := make(chan Sample, 10)
	errCh := make(chan error, 1)
	var stdoutBuf strings.Builder

	go func() {
		defer close(sampleCh)
		defer close(errCh)

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			// Save to buffer for final result parsing
			stdoutBuf.WriteString(line)
			stdoutBuf.WriteString("\n")

			// Parse intermediate results - check for time marker first
			if !regexp.MustCompile(`\[\s*\d+s\s*\]`).MatchString(line) {
				continue
			}

			// Extract TPS
			var tps float64
			if matches := regexp.MustCompile(`tps:\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
				tps, _ = strconv.ParseFloat(matches[1], 64)
			} else {
				continue // Not a valid metrics line
			}

			// Extract QPS
			var qps float64
			if matches := regexp.MustCompile(`qps:\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
				qps, _ = strconv.ParseFloat(matches[1], 64)
			}

			// Extract thread count
			var threadCount int
			if matches := regexp.MustCompile(`thds:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				threadCount, _ = strconv.Atoi(matches[1])
			}

			// Extract 95th percentile latency
			var latencyP95 float64
			if matches := regexp.MustCompile(`lat\s*\(ms,95%\):\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
				latencyP95, _ = strconv.ParseFloat(matches[1], 64)
			}

			// Extract average latency (rt: response time)
			var latencyAvg float64
			if matches := regexp.MustCompile(`rt:\s*(\d+\.?\d*)ms`).FindStringSubmatch(line); len(matches) > 1 {
				latencyAvg, _ = strconv.ParseFloat(matches[1], 64)
			}

			// Extract error rate
			var errorRate float64
			if matches := regexp.MustCompile(`err/s:\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
				errorRate, _ = strconv.ParseFloat(matches[1], 64)
			}

			sample := Sample{
				Timestamp:   time.Now(),
				TPS:         tps,
				QPS:         qps,
				LatencyAvg:  latencyAvg,
				LatencyP95:  latencyP95,
				ErrorRate:   errorRate,
				ThreadCount: threadCount,
				RawLine:     line, // Save original output line
			}

			slog.Debug("SysbenchAdapter: Parsed realtime sample",
				"tps", tps, "qps", qps, "threads", threadCount, "latency_p95", latencyP95, "err_rate", errorRate)

			select {
			case sampleCh <- sample:
			case <-ctx.Done():
				return
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case errCh <- fmt.Errorf("scan stdout: %w", err):
			case <-ctx.Done():
			}
		}
	}()

	return sampleCh, errCh, &stdoutBuf
}

// ParseFinalResults parses the final benchmark results from sysbench output.
// Implements: REQ-EXEC-005 (result collection)
func (a *SysbenchAdapter) ParseFinalResults(ctx context.Context, stdout string) (*FinalResult, error) {
	result := &FinalResult{}

	lines := strings.Split(stdout, "\n")

	// Parse SQL statistics
	for i, line := range lines {
		// Total transactions: 20466  (340.98 per sec.)
		if matches := regexp.MustCompile(`transactions:\s*(\d+)\s*\(\s*(\d+\.?\d*)\s*per sec\.\)`).FindStringSubmatch(line); len(matches) > 2 {
			result.TotalTransactions, _ = strconv.ParseInt(matches[1], 10, 64)
			result.TransactionsPerSec, _ = strconv.ParseFloat(matches[2], 64)
		}

		// queries: 409320 (6819.55 per sec.)
		if matches := regexp.MustCompile(`queries:\s*(\d+)\s*\(\s*(\d+\.?\d*)\s*per sec\.\)`).FindStringSubmatch(line); len(matches) > 2 {
			result.TotalQueries, _ = strconv.ParseInt(matches[1], 10, 64)
			result.QueriesPerSec, _ = strconv.ParseFloat(matches[2], 64)
		}

		// read:    286524
		if matches := regexp.MustCompile(`read:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			result.ReadQueries, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		// write:   81864
		if matches := regexp.MustCompile(`write:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			result.WriteQueries, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		// other:   40932
		if matches := regexp.MustCompile(`other:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			result.OtherQueries, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		// ignored errors:  0      (0.00 per sec.)
		if matches := regexp.MustCompile(`ignored errors:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			result.IgnoredErrors, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		// reconnects:  0      (0.00 per sec.)
		if matches := regexp.MustCompile(`reconnects:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			result.Reconnects, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		// General statistics: total time:                          60.0202s
		if strings.Contains(line, "total time:") {
			if matches := regexp.MustCompile(`total time:\s*(\d+\.?\d*)s`).FindStringSubmatch(line); len(matches) > 1 {
				result.TotalTime, _ = strconv.ParseFloat(matches[1], 64)
			}
		}

		// total number of events:              20466
		if strings.Contains(line, "total number of events:") {
			if matches := regexp.MustCompile(`total number of events:\s*(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				result.TotalEvents, _ = strconv.ParseInt(matches[1], 10, 64)
			}
		}

		// Latency (ms): section
		if strings.TrimSpace(line) == "Latency (ms):" {
			// Parse next few lines for min, avg, max, 95th percentile
			for j := i + 1; j < len(lines) && j < i+10; j++ {
				latencyLine := strings.TrimSpace(lines[j])
				if latencyLine == "" {
					continue
				}

				// min:                                    8.42
				if matches := regexp.MustCompile(`min:\s*(\d+\.?\d*)`).FindStringSubmatch(latencyLine); len(matches) > 1 {
					result.LatencyMin, _ = strconv.ParseFloat(matches[1], 64)
				}

				// avg:                                   11.73
				if matches := regexp.MustCompile(`avg:\s*(\d+\.?\d*)`).FindStringSubmatch(latencyLine); len(matches) > 1 {
					result.LatencyAvg, _ = strconv.ParseFloat(matches[1], 64)
				}

				// max:                                   31.18
				if matches := regexp.MustCompile(`max:\s*(\d+\.?\d*)`).FindStringSubmatch(latencyLine); len(matches) > 1 {
					result.LatencyMax, _ = strconv.ParseFloat(matches[1], 64)
				}

				// 95th percentile:                       13.70
				if matches := regexp.MustCompile(`95th percentile:\s*(\d+\.?\d*)`).FindStringSubmatch(latencyLine); len(matches) > 1 {
					result.LatencyP95, _ = strconv.ParseFloat(matches[1], 64)
				}

				// 99th percentile (if present)
				if matches := regexp.MustCompile(`99th percentile:\s*(\d+\.?\d*)`).FindStringSubmatch(latencyLine); len(matches) > 1 {
					result.LatencyP99, _ = strconv.ParseFloat(matches[1], 64)
				}

				// sum:                               239982.82
				if matches := regexp.MustCompile(`sum:\s*(\d+\.?\d*)`).FindStringSubmatch(latencyLine); len(matches) > 1 {
					result.LatencySum, _ = strconv.ParseFloat(matches[1], 64)
				}
			}
		}

		// Threads fairness: events (avg/stddev):           5116.5000/4.15
		if strings.Contains(line, "events (avg/stddev):") {
			if matches := regexp.MustCompile(`events\s*\(avg/stddev\):\s*(\d+\.?\d*)/(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 2 {
				result.EventsAvg, _ = strconv.ParseFloat(matches[1], 64)
				result.EventsStddev, _ = strconv.ParseFloat(matches[2], 64)
			}
		}

		// execution time (avg/stddev):   59.9957/0.00
		if strings.Contains(line, "execution time (avg/stddev):") {
			if matches := regexp.MustCompile(`execution time\s*\(avg/stddev\):\s*(\d+\.?\d*)/(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 2 {
				result.ExecTimeAvg, _ = strconv.ParseFloat(matches[1], 64)
				result.ExecTimeStddev, _ = strconv.ParseFloat(matches[2], 64)
			}
		}
	}

	slog.Info("SysbenchAdapter: Parsed final results",
		"total_transactions", result.TotalTransactions,
		"tps", result.TransactionsPerSec,
		"qps", result.QueriesPerSec,
		"latency_avg", result.LatencyAvg,
		"latency_p95", result.LatencyP95)

	return result, nil
}

// ValidateConfig validates the configuration for sysbench.
// Implements: REQ-EXEC-001 (pre-check)
func (a *SysbenchAdapter) ValidateConfig(ctx context.Context, config *Config) error {
	// Validate connection
	if config.Connection == nil {
		return fmt.Errorf("connection is required")
	}

	// Validate database type support
	if !a.SupportsDatabase(config.Connection.GetType()) {
		return fmt.Errorf("database type %s not supported by sysbench", config.Connection.GetType())
	}

	// Validate template
	if config.Template == nil {
		return fmt.Errorf("template is required")
	}

	// Detect execution phase from options
	// Prepare-only mode: SkipCleanup=true, time=0
	// Cleanup-only mode: SkipPrepare=true, time=0
	// Run mode: both skip=false, time>0
	isPreparePhase := config.Options.SkipCleanup && !config.Options.SkipPrepare
	isCleanupPhase := config.Options.SkipPrepare && !config.Options.SkipCleanup
	isRunPhase := !config.Options.SkipPrepare && !config.Options.SkipCleanup

	slog.Info("SysbenchAdapter: Validating config",
		"is_prepare_phase", isPreparePhase,
		"is_cleanup_phase", isCleanupPhase,
		"is_run_phase", isRunPhase,
		"skip_prepare", config.Options.SkipPrepare,
		"skip_cleanup", config.Options.SkipCleanup)

	// Validate required parameters based on phase
	if isRunPhase {
		// Run phase requires threads and time
		requiredParams := []string{"threads", "time"}
		for _, param := range requiredParams {
			if _, ok := config.Parameters[param]; !ok {
				return fmt.Errorf("required parameter '%s' is missing for run phase", param)
			}
		}

		// Validate threads value
		if threads, ok := config.Parameters["threads"].(int); ok {
			if threads < 1 || threads > 1024 {
				return fmt.Errorf("threads must be between 1 and 1024, got %d", threads)
			}
		}

		// Validate time value for run phase
		if runTime, ok := config.Parameters["time"].(int); ok {
			if runTime < 10 || runTime > 86400 {
				return fmt.Errorf("time must be between 10 and 86400 seconds, got %d", runTime)
			}
		}
	} else if isPreparePhase || isCleanupPhase {
		// Prepare/cleanup phases only require that time parameter exists (can be 0)
		// threads is not required for prepare/cleanup
		slog.Info("SysbenchAdapter: Prepare/cleanup phase - skipping time validation")
	}

	return nil
}

// SupportsDatabase checks if this adapter supports the given database type.
func (a *SysbenchAdapter) SupportsDatabase(dbType connection.DatabaseType) bool {
	switch dbType {
	case connection.DatabaseTypeMySQL, connection.DatabaseTypePostgreSQL:
		return true
	default:
		return false
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// getScriptName determines the sysbench script name from template.
func (a *SysbenchAdapter) getScriptName(template *domaintemplate.Template) string {
	// Sysbench Lua scripts are typically located in /usr/share/sysbench/
	// Return full path for reliability
	const sysbenchScriptPath = "/usr/share/sysbench"

	if template == nil {
		return filepath.Join(sysbenchScriptPath, "oltp_read_write.lua") // Default fallback
	}

	// Extract script name from template ID
	// Template IDs are like: "sysbench-oltp-read-write", "sysbench-oltp-read-only", etc.
	scriptName := template.ID
	if strings.HasPrefix(scriptName, "sysbench-") {
		scriptName = strings.TrimPrefix(scriptName, "sysbench-")
		// Replace hyphens with underscores for Lua script names
		scriptName = strings.ReplaceAll(scriptName, "-", "_")
		return filepath.Join(sysbenchScriptPath, scriptName+".lua")
	}

	// Fallback to default based on template name/description
	if strings.Contains(strings.ToLower(template.Name), "read only") {
		return filepath.Join(sysbenchScriptPath, "oltp_read_only.lua")
	}
	if strings.Contains(strings.ToLower(template.Name), "write only") {
		return filepath.Join(sysbenchScriptPath, "oltp_write_only.lua")
	}

	return filepath.Join(sysbenchScriptPath, "oltp_read_write.lua") // Default
}

// getDBType converts connection type to sysbench database type.
func (a *SysbenchAdapter) getDBType(conn connection.Connection) string {
	switch conn.GetType() {
	case connection.DatabaseTypeMySQL:
		return "mysql"
	case connection.DatabaseTypePostgreSQL:
		return "pgsql"
	default:
		return ""
	}
}

// buildConnectionString builds a sysbench connection string.
func (a *SysbenchAdapter) buildConnectionString(conn connection.Connection, dbType string) string {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		// MySQL: --mysql-host=localhost --mysql-port=3306 --mysql-user=root --mysql-password=pass --mysql-db=testdb
		return c.Host // Return host, individual args are built separately
	case *connection.PostgreSQLConnection:
		// PostgreSQL: --pgsql-host=localhost --pgsql-port=5432 --pgsql-user=user --pgsql-password=pass --pgsql-db=testdb
		return c.Host
	default:
		return ""
	}
}

// buildConnectionArgs builds connection-specific command line arguments.
func (a *SysbenchAdapter) buildConnectionArgs(conn connection.Connection, config *Config) []string {
	var args []string

	switch c := conn.(type) {
	case *connection.MySQLConnection:
		// Get database name from connection or parameters
		dbName := c.Database
		if dbName == "" {
			if db, ok := config.Parameters["db_name"].(string); ok && db != "" {
				dbName = db
			}
		}
		if dbName == "" {
			dbName = "sbtest"
		}

		args = append(args,
			fmt.Sprintf("--mysql-host=%s", c.Host),
			fmt.Sprintf("--mysql-port=%d", c.Port),
			fmt.Sprintf("--mysql-user=%s", c.Username),
			// Password is set via environment variable for security
			fmt.Sprintf("--mysql-db=%s", dbName),
		)
		if c.SSLMode != "" && c.SSLMode != "disabled" {
			args = append(args, "--mysql-ssl=ON")
		}

	case *connection.PostgreSQLConnection:
		// Get database name from connection or parameters
		dbName := c.Database
		if dbName == "" {
			if db, ok := config.Parameters["db_name"].(string); ok && db != "" {
				dbName = db
			}
		}
		if dbName == "" {
			dbName = "postgres"
		}

		args = append(args,
			fmt.Sprintf("--pgsql-host=%s", c.Host),
			fmt.Sprintf("--pgsql-port=%d", c.Port),
			fmt.Sprintf("--pgsql-user=%s", c.Username),
			// Password is set via environment variable for security
			fmt.Sprintf("--pgsql-db=%s", dbName),
		)
		if c.SSLMode != "" && c.SSLMode != "disable" {
			args = append(args, "--pgsql-ssl=ON")
		}
	}

	return args
}

// buildEnvVars builds environment variables for the command.
func (a *SysbenchAdapter) buildEnvVars(conn connection.Connection) []string {
	var env []string

	// Set password via environment variable for security
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		if c.Password != "" {
			env = append(env, fmt.Sprintf("MYSQL_PWD=%s", c.Password))
		}
	case *connection.PostgreSQLConnection:
		if c.Password != "" {
			env = append(env, fmt.Sprintf("PGPASSWORD=%s", c.Password))
		}
	}

	return env
}

// sysbenchOutputParser is a helper for parsing sysbench output.
type sysbenchOutputParser struct {
	tpsPattern     *regexp.Regexp
	latencyPattern *regexp.Regexp
	once           sync.Once
}

func (p *sysbenchOutputParser) init() {
	p.once.Do(func() {
		p.tpsPattern = regexp.MustCompile(`tps:\s*(\d+\.?\d*)`)
		p.latencyPattern = regexp.MustCompile(`rt:\s*(\d+\.?\d*)ms`)
	})
}

// ParseIntermediateOutput parses intermediate output from sysbench.
func (a *SysbenchAdapter) ParseIntermediateOutput(line string) *Sample {
	p := &sysbenchOutputParser{}
	p.init()

	sample := &Sample{
		Timestamp: time.Now(),
	}

	// Extract TPS
	if matches := p.tpsPattern.FindStringSubmatch(line); len(matches) > 1 {
		if tps, err := strconv.ParseFloat(matches[1], 64); err == nil {
			sample.TPS = tps
		}
	}

	// Extract latency
	if matches := p.latencyPattern.FindStringSubmatch(line); len(matches) > 1 {
		if latency, err := strconv.ParseFloat(matches[1], 64); err == nil {
			sample.LatencyAvg = latency
		}
	}

	return sample
}
