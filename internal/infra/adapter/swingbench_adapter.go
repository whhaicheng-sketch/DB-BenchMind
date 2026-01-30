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
	// Path to swingbench executable (optional, if empty uses PATH)
	SwingbenchPath string
}

// NewSwingbenchAdapter creates a new swingbench adapter.
func NewSwingbenchAdapter() *SwingbenchAdapter {
	return &SwingbenchAdapter{
		SwingbenchPath: "oowbench", // Default to oowbench
	}
}

// Type returns the adapter type.
func (a *SwingbenchAdapter) Type() AdapterType {
	return AdapterTypeSwingbench
}

// BuildPrepareCommand builds the command for data preparation phase.
// Swingbench doesn't have a separate prepare phase - data is created during run.
func (a *SwingbenchAdapter) BuildPrepareCommand(ctx context.Context, config *Config) (*Command, error) {
	// Swingbench doesn't have a separate prepare command
	// We return a minimal command that does nothing
	return &Command{
		CmdLine: "echo 'Swingbench prepare phase - no action required'",
		WorkDir: config.WorkDir,
	}, nil
}

// BuildRunCommand builds the command for the main benchmark run.
func (a *SwingbenchAdapter) BuildRunCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Only Oracle is supported by Swingbench
	if conn.GetType() != connection.DatabaseTypeOracle {
		return nil, fmt.Errorf("swingbench only supports Oracle database, got %s", conn.GetType())
	}

	// Build run command
	cmdArgs := []string{
		a.SwingbenchPath,
	}

	// Add connection parameters
	oracleConn, ok := conn.(*connection.OracleConnection)
	if !ok {
		return nil, fmt.Errorf("invalid connection type for swingbench: %T", conn)
	}

	// Build connection string
	connectionStr := a.buildConnectionString(oracleConn)
	cmdArgs = append(cmdArgs, "-cs", connectionStr)

	// Add benchmark type from template
	benchmarkType := "SOE" // Default
	if bt, ok := config.Parameters["benchmark_type"].(string); ok {
		benchmarkType = bt
	}
	cmdArgs = append(cmdArgs, "-bt", benchmarkType)

	// Add template parameters
	if users, ok := config.Parameters["users"].(int); ok {
		cmdArgs = append(cmdArgs, "-u", strconv.Itoa(users))
	}
	if cycles, ok := config.Parameters["cycles"].(int); ok {
		cmdArgs = append(cmdArgs, "-c", strconv.Itoa(cycles))
	}
	if thinkTime, ok := config.Parameters["think_time"].(int); ok {
		cmdArgs = append(cmdArgs, "-t", strconv.Itoa(thinkTime))
	}
	if minDelay, ok := config.Parameters["min_delay"].(int); ok {
		cmdArgs = append(cmdArgs, "-a", strconv.Itoa(minDelay))
	}
	if maxDelay, ok := config.Parameters["max_delay"].(int); ok {
		cmdArgs = append(cmdArgs, "-b", strconv.Itoa(maxDelay))
	}

	// Add output file
	cmdArgs = append(cmdArgs, "-o", "swingbench_output.txt")

	cmdLine := strings.Join(cmdArgs, " ")

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
	}, nil
}

// BuildCleanupCommand builds the command for cleanup phase.
func (a *SwingbenchAdapter) BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error) {
	// Swingbench doesn't have a separate cleanup command
	return &Command{
		CmdLine: "echo 'Swingbench cleanup phase - no action required'",
		WorkDir: config.WorkDir,
	}, nil
}

// ParseRunOutput parses the output from a swingbench run.
func (a *SwingbenchAdapter) ParseRunOutput(ctx context.Context, stdout string, stderr string) (*Result, error) {
	result := &Result{
		RawOutput: stdout,
	}

	lines := strings.Split(stdout, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse TPM (Transactions Per Minute)
		if strings.Contains(strings.ToLower(line), "tpm") {
			// Try format: "TPM: 5000" or "TPM:5000"
			re := regexp.MustCompile(`(?i)tpm:\s*(\d+(?:\.\d+)?)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.TPS = val / 60 // Convert TPM to TPS
				}
			}
		}

		// Parse average response time (format: "Average response time: 250ms")
		if strings.Contains(line, "Average") && (strings.Contains(line, "response") || strings.Contains(line, "Response")) {
			re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*ms`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyAvg = val
				}
			}
		}

		// Parse minimum response time
		if strings.Contains(line, "Minimum") && (strings.Contains(line, "response") || strings.Contains(line, "Response")) {
			re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*ms`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyMin = val
				}
			}
		}

		// Parse maximum response time
		if strings.Contains(line, "Maximum") && (strings.Contains(line, "response") || strings.Contains(line, "Response")) {
			re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*ms`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.LatencyMax = val
				}
			}
		}

		// Parse errors (format: "Errors: 5" or "Errors:5")
		if strings.Contains(strings.ToLower(line), "error") {
			re := regexp.MustCompile(`(?i)error[s]?:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.TotalErrors = val
				}
			}
		}

		// Parse transactions count
		if strings.Contains(strings.ToLower(line), "transaction") {
			re := regexp.MustCompile(`(?i)transaction[s]?:\s*(\d+)`)
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
