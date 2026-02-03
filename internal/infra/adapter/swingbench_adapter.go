// Package adapter provides Swingbench benchmark tool adapter.
// Implements: Phase 3 - Swingbench Tool Adapter
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

// SwingbenchAdapter implements BenchmarkAdapter for Swingbench tool.
// Implements: REQ-EXEC-001, REQ-EXEC-002, REQ-EXEC-004
type SwingbenchAdapter struct {
	// Path to charbench executable (CLI for running workload)
	SwingbenchPath string
	// Path to oewizard executable (for data generation and cleanup)
	OewizardPath string
}

// NewSwingbenchAdapter creates a new swingbench adapter.
func NewSwingbenchAdapter() *SwingbenchAdapter {
	return &SwingbenchAdapter{
		SwingbenchPath: "/opt/benchtools/swingbench/bin/charbench", // Default to charbench
		OewizardPath:   "/opt/benchtools/swingbench/bin/oewizard",   // oewizard for data generation
	}
}

// Type returns the adapter type.
func (a *SwingbenchAdapter) Type() AdapterType {
	return AdapterTypeSwingbench
}

// BuildPrepareCommand builds the command for data preparation phase.
// Uses oewizard to create schema and generate data.
func (a *SwingbenchAdapter) BuildPrepareCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Only Oracle is supported by Swingbench
	if conn.GetType() != connection.DatabaseTypeOracle {
		return nil, fmt.Errorf("swingbench only supports Oracle database, got %s", conn.GetType())
	}

	oracleConn, ok := conn.(*connection.OracleConnection)
	if !ok {
		return nil, fmt.Errorf("invalid connection type for swingbench: %T", conn)
	}

	// Build connection string for swingbench (not JDBC format)
	connectionStr := a.buildCharbenchConnectionString(oracleConn)

	// Build oewizard command
	cmdArgs := []string{
		"cd", "/opt/benchtools/swingbench/bin", "&&",
		a.OewizardPath,
		"-cl", // Character mode (non-interactive)
		"-create",
		"-generate",
		"-cs", connectionStr,
		"-u", oracleConn.Username,
	}

	// Add password if provided
	if oracleConn.Password != "" {
		cmdArgs = append(cmdArgs, "-p", oracleConn.Password)
	}

	// Add scale parameter (data size)
	if scale, ok := config.Parameters["scale"].(int); ok {
		cmdArgs = append(cmdArgs, "-scale", strconv.Itoa(scale))
	} else {
		cmdArgs = append(cmdArgs, "-scale", "1") // Default 1GB
	}

	// Add threads for data generation
	if threads, ok := config.Parameters["threads"].(int); ok {
		cmdArgs = append(cmdArgs, "-tc", strconv.Itoa(threads))
	} else {
		cmdArgs = append(cmdArgs, "-tc", "32") // Default 32 threads
	}

	// Add DBA credentials for schema creation
	if dbaUser, ok := config.Parameters["dba_username"].(string); ok && dbaUser != "" {
		cmdArgs = append(cmdArgs, "-dba", dbaUser)
		if dbaPass, ok := config.Parameters["dba_password"].(string); ok && dbaPass != "" {
			cmdArgs = append(cmdArgs, "-dbap", dbaPass)
		}
	}

	cmdLine := strings.Join(cmdArgs, " ")

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
	}, nil
}

// BuildRunCommand builds the command for the main benchmark run.
// Uses charbench to run the workload with specified configuration.
func (a *SwingbenchAdapter) BuildRunCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Only Oracle is supported by Swingbench
	if conn.GetType() != connection.DatabaseTypeOracle {
		return nil, fmt.Errorf("swingbench only supports Oracle database, got %s", conn.GetType())
	}

	oracleConn, ok := conn.(*connection.OracleConnection)
	if !ok {
		return nil, fmt.Errorf("invalid connection type for swingbench: %T", conn)
	}

	// Build connection string for charbench
	connectionStr := a.buildCharbenchConnectionString(oracleConn)

	// Build charbench command
	cmdArgs := []string{
		"cd", "/opt/benchtools/swingbench/bin", "&&",
		a.SwingbenchPath,
	}

	// Add config file (required for charbench)
	if configFile, ok := config.Parameters["config_file"].(string); ok && configFile != "" {
		cmdArgs = append(cmdArgs, "-c", configFile)
	} else {
		return nil, fmt.Errorf("config_file parameter is required for charbench")
	}

	// Add connection string
	cmdArgs = append(cmdArgs, "-cs", connectionStr)

	// Add username
	if oracleConn.Username != "" {
		cmdArgs = append(cmdArgs, "-u", oracleConn.Username)
	}

	// Add password if provided
	if oracleConn.Password != "" {
		cmdArgs = append(cmdArgs, "-p", oracleConn.Password)
	}

	// Add user count (concurrent users)
	if users, ok := config.Parameters["users"].(int); ok {
		cmdArgs = append(cmdArgs, "-uc", strconv.Itoa(users))
	}

	// Add runtime (in minutes)
	if runtime, ok := config.Parameters["time"].(int); ok {
		cmdArgs = append(cmdArgs, "-rt", fmt.Sprintf("%d:00", runtime))
	}

	// Add verbose output for metrics (tps, tpm, response time, errors, users)
	cmdArgs = append(cmdArgs, "-v", "tps,tpm,resp,errs,users")

	cmdLine := strings.Join(cmdArgs, " ")

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
	}, nil
}

// BuildCleanupCommand builds the command for cleanup phase.
// Uses oewizard to drop the schema.
func (a *SwingbenchAdapter) BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Only Oracle is supported by Swingbench
	if conn.GetType() != connection.DatabaseTypeOracle {
		return nil, fmt.Errorf("swingbench only supports Oracle database, got %s", conn.GetType())
	}

	oracleConn, ok := conn.(*connection.OracleConnection)
	if !ok {
		return nil, fmt.Errorf("invalid connection type for swingbench: %T", conn)
	}

	// Build connection string for swingbench (not JDBC format)
	connectionStr := a.buildCharbenchConnectionString(oracleConn)

	// Build oewizard drop command
	cmdArgs := []string{
		"cd", "/opt/benchtools/swingbench/bin", "&&",
		a.OewizardPath,
		"-cl", // Character mode (non-interactive)
		"-drop",
		"-cs", connectionStr,
		"-u", oracleConn.Username,
	}

	// Add password if provided
	if oracleConn.Password != "" {
		cmdArgs = append(cmdArgs, "-p", oracleConn.Password)
	}

	// Add DBA credentials for schema drop
	if dbaUser, ok := config.Parameters["dba_username"].(string); ok && dbaUser != "" {
		cmdArgs = append(cmdArgs, "-dba", dbaUser)
		if dbaPass, ok := config.Parameters["dba_password"].(string); ok && dbaPass != "" {
			cmdArgs = append(cmdArgs, "-dbap", dbaPass)
		}
	}

	cmdLine := strings.Join(cmdArgs, " ")

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
	}, nil
}

// ParseRunOutput parses the output from a charbench run.
// Expected format: "Time     Users       TPM      TPS     Errors ..."
func (a *SwingbenchAdapter) ParseRunOutput(ctx context.Context, stdout string, stderr string) (*Result, error) {
	result := &Result{
		RawOutput: stdout,
	}

	lines := strings.Split(stdout, "\n")

	// Track totals for averaging
	var totalTPS float64
	var totalErrors int64
	lineCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Time") || strings.HasPrefix(line, "Author") || strings.HasPrefix(line, "Version") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Parse charbench output format:
		// Time     Users       TPM      TPS     Errors   ...
		// 10:58:37 [4/4]       0        0       0        0 ...

		// Try to parse as data line (starts with time pattern HH:MM:SS or contains [N/M])
		if len(fields) >= 5 {
			// Find TPS field (4th numeric field usually)
			for i, field := range fields {
				// Parse TPS
				if i > 2 && i < len(fields)-1 {
					if val, err := strconv.ParseFloat(field, 64); err == nil {
						// Check if this looks like TPS (reasonable range)
						if val >= 0 && val <= 100000 {
							// Check previous field for TPM
							if i > 0 {
								if prevVal, err := strconv.ParseFloat(fields[i-1], 64); err == nil {
									if prevVal >= 0 && prevVal <= 6000000 {
										// Found TPM and TPS pair
										totalTPS = val
										lineCount++
									}
								}
							}
						}
					}
				}

				// Parse Errors (field after TPS)
				if i > 3 {
					if val, err := strconv.ParseInt(field, 10, 64); err == nil {
						if val >= 0 && val <= 1000000 {
							// Check if previous field was TPS
							if i > 0 {
								if _, err := strconv.ParseFloat(fields[i-1], 64); err == nil {
									totalErrors += val
								}
							}
						}
					}
				}
			}
		}

		// Also try to find "Averages:" line which contains final averages
		if strings.Contains(line, "Averages:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				// Format: Averages: TPS TPM
				for i, part := range parts {
					if val, err := strconv.ParseFloat(part, 64); err == nil {
						if val > 0 {
							if i < len(parts)-1 {
								if nextVal, err := strconv.ParseFloat(parts[i+1], 64); err == nil && nextVal > 0 {
									// Found TPS and TPM
									result.TPS = val
									if result.TPS == 0 || result.TPS < val {
										result.TPS = val
									}
								}
							}
						}
					}
				}
			}
		}

		// Parse "Total Transactions:" line
		if strings.Contains(line, "Total") && strings.Contains(line, "Transactions") {
			re := regexp.MustCompile(`Total\s+Transactions[:\s]+(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.TotalTransactions = val
				}
			}
		}

		// Parse "Average:" response time
		if strings.Contains(line, "Average") && strings.Contains(line, ":") {
			re := regexp.MustCompile(`Average\s*:\s*(\d+\.?\d*)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyAvg = val
				}
			}
		}
	}

	// Use averages if found, otherwise use calculated values
	if result.TPS == 0 && lineCount > 0 {
		result.TPS = totalTPS / float64(lineCount)
	}

	if result.TotalErrors == 0 {
		result.TotalErrors = totalErrors
	}

	// Calculate error rate
	if result.TotalTransactions > 0 {
		result.ErrorRate = (float64(result.TotalErrors) / float64(result.TotalTransactions)) * 100
	} else if totalErrors > 0 && lineCount > 0 {
		// Fallback: estimate from parsed data
		result.ErrorRate = (float64(totalErrors) / float64(lineCount))
	}

	// Set default duration if not parsed
	if result.Duration == 0 {
		result.Duration = 60 * time.Second
	}

	return result, nil
}

// StartRealtimeCollection starts realtime metric collection from swingbench output.
func (a *SwingbenchAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader, stderr io.Reader) (<-chan Sample, <-chan error, *strings.Builder) {
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

			// Parse realtime TPM
			if strings.Contains(line, "TPM:") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if strings.Contains(strings.ToLower(part), "tpm") && i+1 < len(parts) {
						if val, err := strconv.ParseFloat(strings.TrimSuffix(parts[i+1], ","), 64); err == nil {
							currentTPM = val
						}
					}
				}

				sample := Sample{
					Timestamp:   time.Now(),
					TPS:         currentTPM / 60,
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

			// Parse user count
			if strings.Contains(line, "Users:") || strings.Contains(line, "users:") {
				parts := strings.Fields(line)
				for i, part := range parts {
					if strings.ToLower(part) == "users:" && i+1 < len(parts) {
						if val, err := strconv.Atoi(strings.TrimSuffix(parts[i+1], ",")); err == nil {
							currentUsers = val
						}
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

// ParseFinalResults parses final results from swingbench output.
// TODO: Implement swingbench-specific parsing
func (a *SwingbenchAdapter) ParseFinalResults(ctx context.Context, stdout string) (*FinalResult, error) {
	// Stub implementation for now
	return &FinalResult{}, fmt.Errorf("parse final results not implemented for swingbench")
}

// ValidateConfig validates the configuration for swingbench.
func (a *SwingbenchAdapter) ValidateConfig(ctx context.Context, config *Config) error {
	if config == nil {
		return fmt.Errorf("config is required")
	}

	if config.Connection == nil {
		return fmt.Errorf("connection is required")
	}

	// Swingbench only supports Oracle
	if config.Connection.GetType() != connection.DatabaseTypeOracle {
		return fmt.Errorf("swingbench only supports Oracle database, got %s", config.Connection.GetType())
	}

	// Validate connection
	if err := config.Connection.Validate(); err != nil {
		return fmt.Errorf("invalid connection: %w", err)
	}

	return nil
}

// SupportsDatabase checks if swingbench supports the given database type.
func (a *SwingbenchAdapter) SupportsDatabase(dbType connection.DatabaseType) bool {
	return dbType == connection.DatabaseTypeOracle
}

// buildConnectionString builds a Swingbench connection string for Oracle.
func (a *SwingbenchAdapter) buildConnectionString(conn *connection.OracleConnection) string {
	// Swingbench format: jdbc:oracle:thin:@//host:port/service_name or jdbc:oracle:thin:@host:port:sid
	var connectionStr string

	if conn.ServiceName != "" {
		connectionStr = fmt.Sprintf("jdbc:oracle:thin:@//%s:%d/%s",
			conn.Host, conn.Port, conn.ServiceName)
	} else if conn.SID != "" {
		connectionStr = fmt.Sprintf("jdbc:oracle:thin:@%s:%d:%s",
			conn.Host, conn.Port, conn.SID)
	} else {
		// Fallback to localhost
		connectionStr = fmt.Sprintf("jdbc:oracle:thin:@//%s:%d/ORCL",
			conn.Host, conn.Port)
	}

	// Add username/password if available
	if conn.Username != "" {
		connectionStr = fmt.Sprintf("%s/%s@%s",
			conn.Username,
			"*****", // Password is redacted
			connectionStr)
	}

	return connectionStr
}

// buildCharbenchConnectionString builds a charbench/oewizard connection string for Oracle.
// Format: //host:port/service_name or host:port:sid (Easy Connect format)
func (a *SwingbenchAdapter) buildCharbenchConnectionString(conn *connection.OracleConnection) string {
	if conn.ServiceName != "" {
		return fmt.Sprintf("//%s:%d/%s", conn.Host, conn.Port, conn.ServiceName)
	} else if conn.SID != "" {
		return fmt.Sprintf("%s:%d:%s", conn.Host, conn.Port, conn.SID)
	} else {
		// Fallback to default service name
		return fmt.Sprintf("//%s:%d/ORCL", conn.Host, conn.Port)
	}
}
