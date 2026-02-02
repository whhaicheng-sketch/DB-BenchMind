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
func (a *HammerDBAdapter) BuildPrepareCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Build prepare script
	script := a.buildScript(ctx, conn, config, "prepare")

	cmdLine := fmt.Sprintf("echo '%s' | %s", script, a.HammerDBPath)

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
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
	}, nil
}

// BuildCleanupCommand builds the command for cleanup phase.
func (a *HammerDBAdapter) BuildCleanupCommand(ctx context.Context, config *Config) (*Command, error) {
	conn := config.Connection

	// Build cleanup script
	script := a.buildScript(ctx, conn, config, "cleanup")

	cmdLine := fmt.Sprintf("echo '%s' | %s", script, a.HammerDBPath)

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
	}, nil
}

// buildScript builds a TCL script for HammerDB.
func (a *HammerDBAdapter) buildScript(ctx context.Context, conn connection.Connection, config *Config, phase string) string {
	var script strings.Builder

	// Database type and connection
	dbType := a.getDBType(conn)
	connectionStr := a.buildConnectionString(conn)

	script.WriteString(fmt.Sprintf("dbtype %s\n", dbType))
	script.WriteString(fmt.Sprintf("disconn %s\n", connectionStr))
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

	// Phase-specific commands
	switch phase {
	case "prepare":
		script.WriteString("loadscript\n")
		script.WriteString("create virtualmachine\n")
		script.WriteString("vucreate\n")
		script.WriteString("vurun\n")
	case "run":
		script.WriteString("loadscript\n")
		script.WriteString("create virtualmachine\n")
		script.WriteString("vucreate\n")
		script.WriteString("vurun\n")
		script.WriteString("vudestroy\n")
	case "cleanup":
		script.WriteString("delete virtualmachine\n")
	}

	return script.String()
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
