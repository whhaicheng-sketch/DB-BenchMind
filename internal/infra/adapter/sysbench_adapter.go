// Package adapter provides Sysbench benchmark tool adapter.
// Implements: Phase 3 - Sysbench Tool Adapter
package adapter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
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

// Type returns the adapter type.
func (a *SysbenchAdapter) Type() AdapterType {
	return AdapterTypeSysbench
}

// BuildPrepareCommand builds the command for data preparation phase.
func (a *SysbenchAdapter) BuildPrepareCommand(ctx context.Context, config *Config) (*Command, error) {
	// Get connection details
	conn := config.Connection
	dbType := a.getDBType(conn)
	connectionStr := a.buildConnectionString(conn, dbType)

	// Build prepare command
	cmdArgs := []string{
		a.SysbenchPath,
		dbType,
	}

	// Add connection-specific arguments
	cmdArgs = append(cmdArgs, a.buildConnectionArgs(conn, dbType, connectionStr)...)

	// Add template parameters
	if tables, ok := config.Parameters["tables"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--tables=%d", tables))
	}
	if tableSize, ok := config.Parameters["table_size"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--table-size=%d", tableSize))
	}

	cmdArgs = append(cmdArgs, "prepare")

	cmdLine := strings.Join(cmdArgs, " ")

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
	}, nil
}

// BuildRunCommand builds the command for the main benchmark run.
func (a *SysbenchAdapter) BuildRunCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection
	dbType := a.getDBType(conn)
	connectionStr := a.buildConnectionString(conn, dbType)

	// Build run command
	cmdArgs := []string{
		a.SysbenchPath,
		dbType,
	}

	// Add connection-specific arguments
	cmdArgs = append(cmdArgs, a.buildConnectionArgs(conn, dbType, connectionStr)...)

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

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
		Env:     a.buildEnvVars(conn),
	}, nil
}

// BuildCleanupCommand builds the command for cleanup phase.
func (a *SysbenchAdapter) BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection
	dbType := a.getDBType(conn)
	connectionStr := a.buildConnectionString(conn, dbType)

	cmdArgs := []string{
		a.SysbenchPath,
		dbType,
	}

	cmdArgs = append(cmdArgs, a.buildConnectionArgs(conn, dbType, connectionStr)...)

	if tables, ok := config.Parameters["tables"].(int); ok {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--tables=%d", tables))
	}

	cmdArgs = append(cmdArgs, "cleanup")

	cmdLine := strings.Join(cmdArgs, " ")

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
func (a *SysbenchAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader, stderr io.Reader) (<-chan Sample, <-chan error) {
	sampleCh := make(chan Sample, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(sampleCh)
		defer close(errCh)

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			// Parse intermediate results
			// Format: "[ 10s ] threads: 8 tps: 1234.56 qps: 5678.90 (rt: 12.34ms) 95%: 23.45ms"
			if matches := regexp.MustCompile(`\[\s*\d+s\s*\].*?tps:\s*(\d+\.?\d*)`).FindStringSubmatch(line); len(matches) > 1 {
				tps, err := strconv.ParseFloat(matches[1], 64)
				if err != nil {
					continue
				}

				// Try to extract latency if available
				var latencyAvg float64
				if latencyMatches := regexp.MustCompile(`rt:\s*(\d+\.?\d*)ms`).FindStringSubmatch(line); len(latencyMatches) > 1 {
					latencyAvg, _ = strconv.ParseFloat(latencyMatches[1], 64)
				}

				sample := Sample{
					Timestamp:   time.Now(),
					TPS:         tps,
					LatencyAvg:  latencyAvg,
					ErrorRate:   0, // No error rate in intermediate output
				}

				select {
				case sampleCh <- sample:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case errCh <- fmt.Errorf("scan stdout: %w", err):
			case <-ctx.Done():
			}
		}
	}()

	// Also monitor stderr for errors
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(strings.ToLower(line), "error") ||
				strings.Contains(strings.ToLower(line), "failed") {
				select {
				case errCh <- fmt.Errorf("sysbench error: %s", line):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return sampleCh, errCh
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

	// Validate required parameters
	requiredParams := []string{"threads", "time"}
	for _, param := range requiredParams {
		if _, ok := config.Parameters[param]; !ok {
			return fmt.Errorf("required parameter '%s' is missing", param)
		}
	}

	// Validate threads value
	if threads, ok := config.Parameters["threads"].(int); ok {
		if threads < 1 || threads > 1024 {
			return fmt.Errorf("threads must be between 1 and 1024, got %d", threads)
		}
	}

	// Validate time value
	if runTime, ok := config.Parameters["time"].(int); ok {
		if runTime < 10 || runTime > 86400 {
			return fmt.Errorf("time must be between 10 and 86400 seconds, got %d", runTime)
		}
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
func (a *SysbenchAdapter) buildConnectionArgs(conn connection.Connection, dbType string, connectionStr string) []string {
	var args []string

	switch c := conn.(type) {
	case *connection.MySQLConnection:
		args = append(args,
			fmt.Sprintf("--mysql-host=%s", c.Host),
			fmt.Sprintf("--mysql-port=%d", c.Port),
			fmt.Sprintf("--mysql-user=%s", c.Username),
			// Password is set via environment variable for security
			fmt.Sprintf("--mysql-db=%s", c.Database),
		)
		if c.SSLMode != "" && c.SSLMode != "disabled" {
			args = append(args, "--mysql-ssl=ON")
		}

	case *connection.PostgreSQLConnection:
		args = append(args,
			fmt.Sprintf("--pgsql-host=%s", c.Host),
			fmt.Sprintf("--pgsql-port=%d", c.Port),
			fmt.Sprintf("--pgsql-user=%s", c.Username),
			// Password is set via environment variable for security
			fmt.Sprintf("--pgsql-db=%s", c.Database),
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
	tpsPattern    *regexp.Regexp
	latencyPattern *regexp.Regexp
	once          sync.Once
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
