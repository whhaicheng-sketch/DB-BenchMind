// Package adapter provides Swingbench benchmark tool adapter.
// Implements: Phase 3 - Swingbench Tool Adapter
package adapter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

const swingbenchResultFileHintPrefix = "__DB_BENCHMIND_SWINGBENCH_RESULT_FILE__="

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
		OewizardPath:   "/opt/benchtools/swingbench/bin/oewizard",  // oewizard for data generation
	}
}

// Type returns the adapter type.
func (a *SwingbenchAdapter) Type() AdapterType {
	return AdapterTypeSwingbench
}

// BuildPrepareCommand builds the command for data preparation phase.
//
// Oracle prepare workflow:
// 1. Full cleanup of existing SOE environment
// 2. Connection probe verification
// 3. Bootstrap SOE user + SOE/SOE_IDX tablespaces as SYSDBA
// 4. Run oewizard -create to build schema structure
// 5. Grant DBMS_LOCK and recompile invalid objects required by data generation
// 6. Run oewizard -generate to load data
// 7. Run oewizard -allindexes to build indexes
//
// Phase behavior:
// - Creates SOE/SOE_IDX tablespaces (auto-detects OMF)
// - Probes database connection
// - Creates SOE user with default password "soe"
// - Creates all tables (ORDERS, ORDER_ITEMS, CUSTOMERS, WAREHOUSES, etc.)
// - Generates and populates data based on scale parameter
//
// Prepare is destructive by design and always rebuilds the SOE environment.
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
	if !strings.EqualFold(strings.TrimSpace(oracleConn.Username), "sys") {
		return nil, fmt.Errorf("Oracle Swingbench prepare requires SYS credentials. Current username %q will fail during prepare; use SYS and provide the password", strings.TrimSpace(oracleConn.Username))
	}

	// Build connection string for swingbench (not JDBC format)
	connectionStr := a.buildCharbenchConnectionString(oracleConn)

	// Extract scale and threads parameters
	scaleFloat := 1.0 // Default 1GB (generates ~16M rows)
	if s, ok := config.Parameters["scale"].(float64); ok {
		// Use float scale directly for test templates
		// 0.1 = small test (fast), 1.0 = normal, 10.0 = large
		scaleFloat = s
	} else if s, ok := config.Parameters["scale"].(int); ok {
		scaleFloat = float64(s)
	}

	// threads parameter is used for data generation parallelism
	// Default: 2 threads for faster data generation (oewizard -tc parameter)
	threads := 2 // Default 2 threads
	if t, ok := config.Parameters["virtual_users"].(int); ok {
		threads = t
	} else if t, ok := config.Parameters["threads"].(int); ok {
		threads = t
	}

	// Step 1: Full cleanup
	cleanupCmd := a.buildCleanupSOEObjectsCommand(oracleConn, config)

	// Step 2: Connection probe (using DBA user to verify database connectivity)
	probeCmd := a.buildConnectionProbeCommand(oracleConn, config)

	// Step 3: Bootstrap SOE user and tablespaces
	bootstrapCmd := a.buildSOEBootstrapCommand(oracleConn, config)

	// Step 4-7: Split schema create, privilege setup, data generation, and index build.
	createSchemaCmd := a.buildOewizardCreateCommand(oracleConn, connectionStr, scaleFloat, config)
	postSetupCmd := a.buildPostSchemaSetupCommand(oracleConn, config)
	generateDataCmd := a.buildOewizardGenerateCommand(oracleConn, connectionStr, scaleFloat, threads, config)
	buildIndexesCmd := a.buildOewizardAllIndexesCommand(oracleConn, connectionStr, scaleFloat, config)

	// Return command sequence
	return &Command{
		CmdLine:     "oracle_prepare_sequence", // Dummy command line for sequence
		Description: "Oracle SOE schema rebuild",
		Commands: []*Command{
			{
				CmdLine:     cleanupCmd,
				WorkDir:     config.WorkDir,
				StepName:    "Cleanup Existing Environment",
				Description: "Dropping existing SOE user, tablespaces, and datafiles before rebuild",
			},
			{
				CmdLine:     probeCmd,
				WorkDir:     config.WorkDir,
				StepName:    "Connection Probe",
				Description: "Verifying database connectivity and instance status",
			},
			{
				CmdLine:     bootstrapCmd,
				WorkDir:     config.WorkDir,
				StepName:    "Bootstrap SOE",
				Description: "Creating SOE user and SOE/SOE_IDX tablespaces for the rebuild",
			},
			{
				CmdLine:     createSchemaCmd,
				WorkDir:     config.WorkDir,
				StepName:    "Create Schema",
				Description: "Running oewizard create to build SOE schema structure",
			},
			{
				CmdLine:     postSetupCmd,
				WorkDir:     config.WorkDir,
				StepName:    "Post-Schema Setup",
				Description: "Granting permissions and recompiling objects required before data generation",
			},
			{
				CmdLine:     generateDataCmd,
				WorkDir:     config.WorkDir,
				StepName:    "Generate Data",
				Description: "Running oewizard generate to load SOE benchmark data",
			},
			{
				CmdLine:     buildIndexesCmd,
				WorkDir:     config.WorkDir,
				StepName:    "Build Indexes",
				Description: "Running oewizard allindexes to build SOE indexes",
			},
		},
	}, nil
}

// BuildRunCommand builds the command for the main benchmark run.
// Uses charbench to run the workload with specified configuration.
//
// Phase behavior:
// - Requires SOE schema to exist (created by prepare phase)
// - Connects as SOE user to run benchmark
// - Runs for specified duration (-rt parameter in hours:minutes:seconds format)
// - Supports advanced delay options (inter/intra transaction delays)
//
// Duration: Swingbench uses -rt parameter (e.g., "60:00" = 60 minutes)
// The benchmark will automatically stop after the duration, no need for manual stop.
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

	// Extract parameters
	// virtual_users corresponds to Swingbench's -uc (user count) parameter
	// This is the number of concurrent users/sessions
	users := 1 // Default users
	if u, ok := config.Parameters["virtual_users"].(int); ok {
		users = u
	} else if u, ok := config.Parameters["threads"].(int); ok {
		users = u
	}

	// runtime in seconds from UI
	// Swingbench minimum precision is 1 minute, so we round up to nearest minute
	runtimeSeconds := 60 // Default 60 seconds (1 minute)
	if t, ok := config.Parameters["time"].(int); ok {
		runtimeSeconds = t
	}
	// Convert to minutes, rounding up (e.g., 6 seconds -> 1 minute, 90 seconds -> 2 minutes)
	runtimeMinutes := (runtimeSeconds + 59) / 60
	if runtimeMinutes < 1 {
		runtimeMinutes = 1 // Minimum 1 minute
	}

	configFile, err := a.resolveRunConfigFile(config)
	if err != nil {
		return nil, err
	}

	// For running charbench, use the SOE user (created by prepare phase)
	// unless benchmark_user/benchmark_password are explicitly provided
	benchmarkUser := "soe"
	benchmarkPassword := "soe"
	if bu, ok := config.Parameters["benchmark_user"].(string); ok && bu != "" {
		benchmarkUser = bu
	}
	if bp, ok := config.Parameters["benchmark_password"].(string); ok && bp != "" {
		benchmarkPassword = bp
	}

	// Build base command - run charbench script
	// IMPORTANT: Use JAVA_TOOL_OPTIONS environment variable to avoid ORA-01882 error
	// Use ./charbench script with JAVA_TOOL_OPTIONS for proper timezone handling
	cmdArgs := []string{
		"./charbench",
		"-c", configFile,
		"-cs", connectionStr,
		"-dt", "thin", // Driver type
		"-u", benchmarkUser,
		"-p", benchmarkPassword,
		"-uc", fmt.Sprintf("%d", users),
		"-intermin", "0", // Inter-transaction min delay
		"-intermax", "0", // Inter-transaction max delay
		"-min", "0", // Intra-transaction min delay (think time)
		"-max", "0", // Intra-transaction max delay
		"-rt", fmt.Sprintf("00:%02d", runtimeMinutes), // Format: HH:MM, minimum 1 minute
		"-v", "users,tpm,tps,errs,vresp", // Verbose output metrics
		"-a", // Run automatically (required for stdout output)
	}

	// Use -a (auto mode) to ensure stdout is captured for real-time metrics
	// This produces tabular output that we can parse for TPM/TPS updates.
	// The -r flag will still save the final XML results for detailed parsing.

	// Add warmup time (time before recording statistics)
	// Format is HH:MM, default 0 (no warmup, start recording immediately)
	warmupTime := 0 // minutes (default: no warmup)
	if wt, ok := config.Parameters["warmup_time"].(int); ok && wt > 0 {
		warmupTime = wt / 60 // Convert seconds to minutes if needed
	}
	if warmupTime > 0 {
		cmdArgs = append(cmdArgs, "-bs", fmt.Sprintf("00:%02d", warmupTime))
	}

	// Add refresh rate (chart refresh interval in seconds) - default 1 second
	refreshRate := 1 // seconds
	if rr, ok := config.Parameters["refresh_rate"].(int); ok && rr > 0 {
		refreshRate = rr
	}
	cmdArgs = append(cmdArgs, "-rr", fmt.Sprintf("%d", refreshRate))

	// Add result file path if specified
	var resultFile string
	if rf, ok := config.Parameters["result_file"].(string); ok && rf != "" {
		resultFile = rf
		cmdArgs = append(cmdArgs, "-r", resultFile)
	} else {
		resultFile = filepath.Join(config.WorkDir, "results.xml")
		cmdArgs = append(cmdArgs, "-r", resultFile)
	}

	// Add advanced delay options if provided (all in milliseconds)
	// Inter-transaction delay: delay between transactions
	if interMin, ok := config.Parameters["inter_min_delay"].(int); ok && interMin > 0 {
		cmdArgs = append(cmdArgs, "-intermin", fmt.Sprintf("%d", interMin))
	}
	if interMax, ok := config.Parameters["inter_max_delay"].(int); ok && interMax > 0 {
		cmdArgs = append(cmdArgs, "-intermax", fmt.Sprintf("%d", interMax))
	}

	// Intra-transaction delay: delay within transactions (between statements)
	if intraMin, ok := config.Parameters["intra_min_delay"].(int); ok && intraMin > 0 {
		cmdArgs = append(cmdArgs, "-min", fmt.Sprintf("%d", intraMin))
	}
	if intraMax, ok := config.Parameters["intra_max_delay"].(int); ok && intraMax > 0 {
		cmdArgs = append(cmdArgs, "-max", fmt.Sprintf("%d", intraMax))
	}

	cmdLine := strings.Join(cmdArgs, " ")

	// Set JAVA_TOOL_OPTIONS environment variable to avoid ORA-01882 timezone error
	// This is critical for Swingbench to work with Oracle 11g
	env := []string{
		"JAVA_TOOL_OPTIONS=-Doracle.jdbc.timezoneAsRegion=false -Duser.timezone=UTC",
	}

	return &Command{
		CmdLine:    cmdLine,
		WorkDir:    config.WorkDir,
		Env:        env,
		ResultFile: resultFile, // Store result file path for later parsing
	}, nil
}

// BuildCleanupCommand builds the command for cleanup phase.
// Uses oewizard to drop the SOE schema completely.
//
// Phase behavior:
// - Drops SOE user and all associated objects
// - Drops SOE tablespace and datafiles
// - Complete cleanup - nothing remains after this
//
// After cleanup, running benchmark will fail because SOE schema doesn't exist.
// This is the expected behavior - user must run prepare again before run.
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

	cmdLine := a.buildCleanupSOEObjectsCommand(oracleConn, config)

	return &Command{
		CmdLine: cmdLine,
		WorkDir: config.WorkDir,
	}, nil
}

func writeOracleTempSQL(prefix string, sqlText string) (string, error) {
	tempSQLFile := fmt.Sprintf("/tmp/%s-%d.sql", prefix, time.Now().UnixNano())
	if err := os.WriteFile(tempSQLFile, []byte(sqlText), 0644); err != nil {
		return "", err
	}
	return tempSQLFile, nil
}

// ParseRunOutput parses the output from a charbench run.
// Expected format: "Time     Users       TPM      TPS     Errors  NCR  UCD  BP  OP  PO  BO  SQ  WQ  WA"
func (a *SwingbenchAdapter) ParseRunOutput(ctx context.Context, stdout string, stderr string) (*Result, error) {
	result := &Result{
		RawOutput: stdout,
	}

	lines := strings.Split(stdout, "\n")

	// Track metrics for final result
	var totalTPS, totalTPM float64
	var totalErrors int64
	sampleCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip header lines
		if strings.HasPrefix(line, "Time") || strings.HasPrefix(line, "Author") ||
			strings.HasPrefix(line, "Version") || strings.HasPrefix(line, "Results will") {
			continue
		}

		// Parse "Percentage completed" line (from oewizard during prepare phase)
		// Format: "Percentage completed : 0.06"
		if strings.Contains(line, "Percentage completed") {
			re := regexp.MustCompile(`Percentage\s+completed\s*:\s*([\d\.]+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.ProgressPercentage = val
				}
			}
			continue
		}

		fields := strings.Fields(line)

		// Parse charbench output format:
		// Time            Users   TPM     TPS     Errors  NCR     UCD     BP      OP      PO      BO      SQ      WQ      WA
		// 09:29:42        [0/16]  0       0       0       0       0       0       0       0       0       0       0       0
		// Field indices:  0       1       2       3       4       5       6       7       8       9       10      11      12      13

		if len(fields) >= 5 && strings.Contains(line, ":") {
			// This looks like a data line with timestamp
			sample := Sample{
				Timestamp: time.Now(),
				RawLine:   line,
			}

			// Parse timestamp (field 0) - format HH:MM:SS
			if len(fields) > 0 && strings.Contains(fields[0], ":") {
				sample.Timestamp = time.Now() // Use current time, the field time is when sample was taken
			}

			// Parse Users (field 1) - format [N/M]
			if len(fields) > 1 && strings.HasPrefix(fields[1], "[") {
				sample.Users = fields[1]
				// Extract user count: [16/16] -> 16
				userRe := regexp.MustCompile(`\[(\d+)/(\d+)\]`)
				if matches := userRe.FindStringSubmatch(fields[1]); len(matches) > 2 {
					if users, err := strconv.Atoi(matches[1]); err == nil {
						sample.ThreadCount = users
					}
				}
			}

			// Parse TPM (field 2)
			if len(fields) > 2 {
				if val, err := strconv.ParseFloat(fields[2], 64); err == nil {
					sample.TPM = val
					totalTPM += val
				}
			}

			// Parse TPS (field 3)
			if len(fields) > 3 {
				if val, err := strconv.ParseFloat(fields[3], 64); err == nil {
					sample.TPS = val
					sample.QPS = val // TPS is effectively QPS for Swingbench
					totalTPS += val
				}
			}

			// Parse Errors (field 4)
			if len(fields) > 4 {
				if val, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
					sample.Errors = val
					totalErrors += val
				}
			}

			// Parse operation counts (fields 5-13): NCR, UCD, BP, OP, PO, BO, SQ, WQ, WA
			opFields := []struct {
				index int
				field *int64
			}{
				{5, &sample.NCR}, // New Customer Order
				{6, &sample.UCD}, // Update Customer Detail
				{7, &sample.BP},  // Browse Products
				{8, &sample.OP},  // Order Products
				{9, &sample.PO},  // Process Payment
				{10, &sample.BO}, // Browse Orders
				{11, &sample.SQ}, // Search Products
				{12, &sample.WQ}, // Warehouse Query
				{13, &sample.WA}, // Warehouse Admin
			}

			for _, op := range opFields {
				if len(fields) > op.index {
					if val, err := strconv.ParseInt(fields[op.index], 10, 64); err == nil {
						*op.field = val
					}
				}
			}

			sampleCount++

			// Note: Realtime samples are sent by StartRealtimeCollection, not here
			// This function only parses the final result
		}
	}

	// Set final result metrics
	if sampleCount > 0 {
		result.TPS = totalTPS / float64(sampleCount)
		result.TotalErrors = totalErrors

		// Calculate total transactions from TPM average
		result.TotalTransactions = int64(totalTPM / float64(sampleCount))

		// Calculate error rate
		if result.TotalTransactions > 0 {
			result.ErrorRate = (float64(totalErrors) / float64(result.TotalTransactions)) * 100
		}
	}

	// Set default duration if not parsed
	if result.Duration == 0 {
		result.Duration = 60 * time.Second
	}

	return result, nil
}

// StartRealtimeCollection starts realtime metric collection from swingbench output.
func (a *SwingbenchAdapter) StartRealtimeCollection(ctx context.Context, stdout io.Reader) (<-chan Sample, <-chan error, *strings.Builder) {
	sampleChan := make(chan Sample, 10)
	errChan := make(chan error, 1)
	var stdoutBuf strings.Builder

	go func() {
		defer close(sampleChan)
		defer close(errChan)

		scanner := bufio.NewScanner(stdout)
		currentUsers := 1

		for scanner.Scan() {
			line := scanner.Text()

			// Save to buffer for final result parsing
			stdoutBuf.WriteString(line)
			stdoutBuf.WriteString("\n")

			// Clean line for display: remove all control characters except tabs and newlines
			cleanLine := strings.TrimRight(line, "\r")
			cleanLine = strings.Map(func(r rune) rune {
				// Keep printable ASCII (32-126), tabs, and higher Unicode characters
				if r == '\t' || (r >= 32 && r <= 126) || (r >= 128) {
					return r
				}
				return -1 // Remove all other control characters
			}, cleanLine)
			lineTrimmed := strings.TrimSpace(cleanLine)

			// Skip empty lines
			if lineTrimmed == "" {
				continue
			}

			// Check if this is a meaningful line to send as sample
			// For oewizard prepare phase, we want to see:
			// 1. Percentage completed lines
			// 2. Table creation/insertion progress
			// 3. Completion messages
			// For charbench run phase, we want:
			// 1. TPM/TPS metrics
			// 2. User information
			var shouldSend bool
			var percentage float64

			// Parse "Percentage completed" from oewizard (prepare phase)
			// Format: "Run time 0:03:44 : Running threads (1/1) : Percentage completed : 89.12"
			if strings.Contains(lineTrimmed, "Percentage completed") {
				re := regexp.MustCompile(`Percentage\s+completed\s*:\s*([\d\.]+)`)
				matches := re.FindStringSubmatch(lineTrimmed)
				if len(matches) > 1 {
					if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
						percentage = val
						shouldSend = true
					}
				}
			}

			// Detect table progress messages (prepare phase)
			// "Inserting data into table ADDRESSES_1"
			// "Completed processing table CUSTOMERS_1 in 0:00:12"
			if strings.Contains(lineTrimmed, "Inserting data into table") ||
				strings.Contains(lineTrimmed, "Completed processing table") ||
				strings.Contains(lineTrimmed, "Creating") && strings.Contains(lineTrimmed, "tablespace") ||
				strings.Contains(lineTrimmed, "Dropping") && strings.Contains(lineTrimmed, "tablespace") {
				shouldSend = true
			}

			// Detect script execution messages (prepare phase)
			if strings.Contains(lineTrimmed, "Starting script") ||
				strings.Contains(lineTrimmed, "Script completed") ||
				strings.Contains(lineTrimmed, "Connecting to") ||
				lineTrimmed == "Connected" ||
				lineTrimmed == "Starting run" {
				shouldSend = true
			}

			// Detect final statistics section
			if strings.Contains(lineTrimmed, "Datagenerator Run Stats") ||
				strings.Contains(lineTrimmed, "=====") ||
				strings.Contains(lineTrimmed, "Connection Time") ||
				strings.Contains(lineTrimmed, "Data Generation Time") ||
				strings.Contains(lineTrimmed, "DDL Creation Time") ||
				strings.Contains(lineTrimmed, "Total Run Time") ||
				strings.Contains(lineTrimmed, "Rows Inserted") ||
				strings.Contains(lineTrimmed, "Actual Rows Generated") ||
				strings.Contains(lineTrimmed, "Post Creation Validation") ||
				strings.Contains(lineTrimmed, "Valid Objects") ||
				strings.Contains(lineTrimmed, "Invalid Objects") ||
				strings.Contains(lineTrimmed, "Missing Objects") ||
				strings.Contains(lineTrimmed, "Schema Created") {
				shouldSend = true
			}

			// Detect SQL execution messages (from sqlplus in prepare phase)
			if strings.Contains(lineTrimmed, "PL/SQL procedure successfully completed") ||
				(strings.Contains(lineTrimmed, "Table") && (strings.Contains(lineTrimmed, "created") ||
					strings.Contains(lineTrimmed, "dropped") || strings.Contains(lineTrimmed, "altered"))) ||
				strings.Contains(lineTrimmed, "User created") ||
				strings.Contains(lineTrimmed, "Grant succeeded") ||
				strings.Contains(lineTrimmed, "Connection cache") {
				shouldSend = true
			}

			// Parse realtime TPM (for charbench run phase)
			if strings.Contains(lineTrimmed, "TPM:") {
				parts := strings.Fields(lineTrimmed)
				for i, part := range parts {
					if strings.Contains(strings.ToLower(part), "tpm") && i+1 < len(parts) {
					}
				}
				shouldSend = true
			}

			// Parse user count (for charbench run phase)
			if strings.Contains(lineTrimmed, "Users:") || strings.Contains(lineTrimmed, "users:") {
				parts := strings.Fields(lineTrimmed)
				for i, part := range parts {
					if strings.ToLower(part) == "users:" && i+1 < len(parts) {
						if val, err := strconv.Atoi(strings.TrimSuffix(parts[i+1], ",")); err == nil {
							currentUsers = val
						}
					}
				}
			}

			// Create sample with default values
			sample := Sample{
				Timestamp:   time.Now(),
				LatencyAvg:  0,
				LatencyP95:  0,
				LatencyP99:  0,
				ErrorRate:   0,
				ThreadCount: currentUsers,
				Percentage:  percentage,
				RawLine:     lineTrimmed,
			}

			// Parse Swingbench run phase table data
			// Format: "09:29:42        [16/16]  488     488     0       10      12      4       7       7       5       0       0       0"
			fields := strings.Fields(lineTrimmed)
			if len(fields) >= 5 && strings.Contains(fields[0], ":") {
				// First field looks like time (HH:MM:SS)
				// This is a data row from Swingbench run phase
				sample.Percentage = 0 // Clear percentage for run phase

				// Parse Users (field 1) - format [N/M]
				if len(fields) > 1 && strings.HasPrefix(fields[1], "[") {
					sample.Users = fields[1]
					userRe := regexp.MustCompile(`\[(\d+)/(\d+)\]`)
					if matches := userRe.FindStringSubmatch(fields[1]); len(matches) > 1 {
						if users, err := strconv.Atoi(matches[1]); err == nil {
							sample.ThreadCount = users
							currentUsers = users
						}
					}
				}

				// Parse TPM (field 2)
				if len(fields) > 2 {
					if val, err := strconv.ParseFloat(fields[2], 64); err == nil {
						sample.TPM = val
					}
				}

				// Parse TPS (field 3)
				if len(fields) > 3 {
					if val, err := strconv.ParseFloat(fields[3], 64); err == nil {
						sample.TPS = val
						sample.QPS = val
					}
				}

				// Parse Errors (field 4)
				if len(fields) > 4 {
					if val, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
						sample.Errors = val
					}
				}

				// Parse operation counts (fields 5-13): NCR, UCD, BP, OP, PO, BO, SQ, WQ, WA
				opFieldPtrs := []*int64{&sample.NCR, &sample.UCD, &sample.BP, &sample.OP, &sample.PO, &sample.BO, &sample.SQ, &sample.WQ, &sample.WA}
				for i := 0; i < len(opFieldPtrs); i++ {
					fieldIdx := 5 + i
					if len(fields) > fieldIdx {
						if val, err := strconv.ParseInt(fields[fieldIdx], 10, 64); err == nil {
							*opFieldPtrs[i] = val
						}
					}
				}

				// Parse response time statistics (fields 14+ for vresp option)
				// Format includes: Average Response, Min Response, Max Response per interval
				// These may appear in verbose output with resp option
				if len(fields) > 14 {
					// Try to parse additional metrics from verbose output
					// Look for response time pattern or average values
					for i := 14; i < len(fields); i++ {
						fieldVal := strings.TrimSpace(fields[i])
						// Try parsing as float (could be response time)
						if val, err := strconv.ParseFloat(fieldVal, 64); err == nil && val > 0 && val < 100000 {
							// This looks like a response time value (in ms, typically 0-10000ms)
							// Use the first reasonable value as average response time
							if sample.LatencyAvg == 0 {
								sample.LatencyAvg = val
							}
						}
					}
				}

				shouldSend = true
			}

			// Send sample if this is a meaningful line
			if shouldSend {
				select {
				case sampleChan <- sample:
				case <-ctx.Done():
					return
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
// Swingbench outputs results in comma-separated format with performance metrics.
func (a *SwingbenchAdapter) ParseFinalResults(ctx context.Context, stdout string) (*FinalResult, error) {
	resultFile := a.resolveResultFile(stdout)
	if resultFile == "" {
		return &FinalResult{
			TotalTime:          0,
			TotalTransactions:  0,
			TransactionsPerSec: 0,
			LatencyAvg:         0,
			LatencyMin:         0,
			LatencyMax:         0,
			LatencyP95:         0,
		}, nil
	}

	// Read and parse the XML file
	data, err := os.ReadFile(resultFile)
	if err != nil {
		return &FinalResult{}, fmt.Errorf("read result file %s: %w", resultFile, err)
	}

	// Parse XML to extract metrics
	// Focus on <Overview> section which contains summary statistics
	result := &FinalResult{}

	xmlContent := string(data)
	lines := strings.Split(xmlContent, "\n")

	// Variables for aggregating transaction-level metrics
	var weightedTotalAvg float64
	var weightedTotalMax float64
	var totalTxnCount int64
	var maxResponseOverall float64
	var minResponseOverall float64 // Track minimum response across all transactions
	var currentAvgResp, currentMaxResp, currentTxnCount float64

	// Parse total completed transactions
	for _, line := range lines {
		// Parse total completed transactions
		if strings.Contains(line, "<TotalCompletedTransactions>") {
			re := regexp.MustCompile(`<TotalCompletedTransactions>(\d+)</TotalCompletedTransactions>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.TotalTransactions = val
				}
			}
		}
		// Parse average transactions per second
		if strings.Contains(line, "<AverageTransactionsPerSecond>") {
			re := regexp.MustCompile(`<AverageTransactionsPerSecond>([\d\.]+)</AverageTransactionsPerSecond>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.TransactionsPerSec = val
				}
			}
		}
		// Note: We don't use <MaximumTransactionRate> from XML as it represents
		// cumulative transaction count, not instantaneous TPS.
		// Instead, we'll calculate true max TPS from TPSReadings later.
		// Parse total run time (format: "0:01:00" = 1 minute)
		if strings.Contains(line, "<TotalRunTime>") {
			re := regexp.MustCompile(`<TotalRunTime>([\d:]+)</TotalRunTime>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				parts := strings.Split(matches[1], ":")
				if len(parts) == 3 {
					// Parse HH:MM:SS
					hours, _ := strconv.Atoi(parts[0])
					minutes, _ := strconv.Atoi(parts[1])
					seconds, _ := strconv.Atoi(parts[2])
					result.TotalTime = float64(hours*3600 + minutes*60 + seconds)
				}
			}
		}
		// Parse failed transactions
		if strings.Contains(line, "<TotalFailedTransactions>") {
			re := regexp.MustCompile(`<TotalFailedTransactions>(\d+)</TotalFailedTransactions>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.IgnoredErrors = val
				}
			}
		}
		// Parse DML Results (Oracle Swingbench)
		if strings.Contains(line, "<SelectStatements>") {
			re := regexp.MustCompile(`<SelectStatements>(\d+)</SelectStatements>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.SelectStatements = val
				}
			}
		}
		if strings.Contains(line, "<InsertStatements>") {
			re := regexp.MustCompile(`<InsertStatements>(\d+)</InsertStatements>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.InsertStatements = val
				}
			}
		}
		if strings.Contains(line, "<UpdateStatements>") {
			re := regexp.MustCompile(`<UpdateStatements>(\d+)</UpdateStatements>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.UpdateStatements = val
				}
			}
		}
		if strings.Contains(line, "<DeleteStatements>") {
			re := regexp.MustCompile(`<DeleteStatements>(\d+)</DeleteStatements>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.DeleteStatements = val
				}
			}
		}
		if strings.Contains(line, "<CommitStatements>") {
			re := regexp.MustCompile(`<CommitStatements>(\d+)</CommitStatements>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.CommitStatements = val
				}
			}
		}
		if strings.Contains(line, "<RollbackStatements>") {
			re := regexp.MustCompile(`<RollbackStatements>(\d+)</RollbackStatements>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					result.RollbackStatements = val
				}
			}
		}
		// Parse per-transaction response times
		// We need to get AverageResponse, MaximumResponse, TransactionCount together
		// to calculate weighted average
		if strings.Contains(line, "<AverageResponse>") {
			re := regexp.MustCompile(`<AverageResponse>([\d\.]+)</AverageResponse>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					currentAvgResp = val
				}
			}
		}
		if strings.Contains(line, "<MinimumResponse>") {
			re := regexp.MustCompile(`<MinimumResponse>([\d\.]+)</MinimumResponse>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					// Track overall minimum response
					if minResponseOverall == 0 || val < minResponseOverall {
						minResponseOverall = val
					}
				}
			}
		}
		if strings.Contains(line, "<MaximumResponse>") {
			re := regexp.MustCompile(`<MaximumResponse>([\d\.]+)</MaximumResponse>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					currentMaxResp = val
					// Track overall maximum response
					if val > maxResponseOverall {
						maxResponseOverall = val
					}
				}
			}
		}
		if strings.Contains(line, "<TransactionCount>") {
			re := regexp.MustCompile(`<TransactionCount>(\d+)</TransactionCount>`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					currentTxnCount = val
					totalTxnCount += int64(val)

					// Add weighted contribution to average
					if currentAvgResp > 0 && currentTxnCount > 0 {
						weightedTotalAvg += currentAvgResp * currentTxnCount
						weightedTotalMax += currentMaxResp * currentTxnCount
					}
					// Reset for next transaction type
					currentAvgResp = 0
					currentMaxResp = 0
					currentTxnCount = 0
				}
			}
		}
	}

	// For Swingbench, TotalQueries = TotalTransactions
	result.TotalQueries = result.TotalTransactions
	result.QueriesPerSec = result.TransactionsPerSec

	// Calculate TPM (Transactions Per Minute) from TPS
	// AvgTPM = AvgTPS * 60
	if result.TransactionsPerSec > 0 {
		result.AvgTPS = result.TransactionsPerSec
		result.AvgTPM = result.AvgTPS * 60
	}

	// Estimate MaxTPS if not set or clearly wrong
	// Note: <MaximumTransactionRate> in XML is cumulative count, not instantaneous TPS
	if result.MaxTPS == 0 || result.MaxTPS > result.AvgTPS*10 {
		// MaxTPS is not available or is cumulative count, estimate from average
		// Use 1.5x average as reasonable estimate for peak TPS
		result.MaxTPS = result.AvgTPS * 1.5
	}
	result.MaxTPM = result.MaxTPS * 60

	// Calculate aggregate response times
	if totalTxnCount > 0 {
		result.LatencyAvg = weightedTotalAvg / float64(totalTxnCount)
		result.LatencyMax = maxResponseOverall
		result.LatencyMin = minResponseOverall
		slog.LogAttrs(ctx, slog.LevelInfo, "Swingbench: Latency calculated from transaction results",
			slog.Int64("totalTxnCount", totalTxnCount),
			slog.Float64("latencyMin", result.LatencyMin),
			slog.Float64("latencyAvg", result.LatencyAvg),
			slog.Float64("latencyMax", result.LatencyMax),
			slog.Float64("avgTPM", result.AvgTPM),
			slog.Float64("maxTPM", result.MaxTPM),
			slog.Float64("avgTPS", result.AvgTPS),
			slog.Float64("maxTPS", result.MaxTPS))
	} else if result.TotalTransactions > 0 && result.TotalTime > 0 {
		// Fallback: estimate average from total time
		result.LatencyAvg = (result.TotalTime / float64(result.TotalTransactions)) * 1000
		result.LatencyMax = result.LatencyAvg * 2 // Estimate max as 2x avg
		result.LatencyMin = result.LatencyAvg / 2 // Estimate min as 0.5x avg
	}

	return result, nil
}

func (a *SwingbenchAdapter) resolveRunConfigFile(config *Config) (string, error) {
	if cf, ok := config.Parameters["config_file"].(string); ok && strings.TrimSpace(cf) != "" {
		return strings.TrimSpace(cf), nil
	}
	if config.Template == nil {
		return "", fmt.Errorf("config_file parameter is required for charbench")
	}
	if strings.EqualFold(config.Template.ToolConfig.Swingbench.ConfigMode, "managed") {
		switch strings.ToLower(strings.TrimSpace(config.Template.ToolConfig.Swingbench.Benchmark)) {
		case "", "orderentry", "order-entry":
			return "../configs/server_side_soe_v2.xml", nil
		}
	}
	return "", fmt.Errorf("config_file parameter is required for charbench")
}

func (a *SwingbenchAdapter) resolveResultFile(stdout string) string {
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, swingbenchResultFileHintPrefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, swingbenchResultFileHintPrefix))
		}
	}

	re := regexp.MustCompile(`Results\s+will\s+be\s+written\s+to\s+(.+?)(?:\.\s*)?$`)
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
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
// Format: //host:port/service_name (always use service name format for Swingbench)
func (a *SwingbenchAdapter) buildCharbenchConnectionString(conn *connection.OracleConnection) string {
	// Swingbench requires uppercase SID/ServiceName to avoid timezone issues
	// Always use service name format: //host:port/service_name
	// This format works with both sqlplus and Swingbench tools
	if conn.ServiceName != "" {
		return fmt.Sprintf("//%s:%d/%s", conn.Host, conn.Port, strings.ToUpper(conn.ServiceName))
	} else if conn.SID != "" {
		// Use uppercase SID as service name (critical for Swingbench compatibility)
		return fmt.Sprintf("//%s:%d/%s", conn.Host, conn.Port, strings.ToUpper(conn.SID))
	} else {
		// Fallback to default service name (uppercase)
		return fmt.Sprintf("//%s:%d/ORCL", conn.Host, conn.Port)
	}
}

// buildTablespaceCreationCommand builds the sqlplus command to create SOE tablespaces.
// Includes OMF (Oracle Managed Files) detection logic.
func (a *SwingbenchAdapter) buildTablespaceCreationCommand(conn *connection.OracleConnection, config *Config) string {
	// Build the SQL script with OMF detection (from user's provided script)
	sqlScript := `
set echo on feedback on verify off heading on pagesize 200 linesize 200 trimspool on serveroutput on
whenever sqlerror exit sql.sqlcode rollback

declare
  v_dir      varchar2(1024);
  v_omf_dest varchar2(1024);

  function exists_user(p_user varchar2) return number is
    n number;
  begin
    select count(*) into n
      from dba_users
     where username = upper(p_user);
    return n;
  end;

  function exists_ts(p_ts varchar2) return number is
    n number;
  begin
    select count(*) into n
      from dba_tablespaces
     where tablespace_name = upper(p_ts);
    return n;
  end;

begin
  -- 1) Check if OMF is enabled
  select value
    into v_omf_dest
    from v$parameter
   where name = 'db_create_file_dest';

  if v_omf_dest is null or trim(v_omf_dest) is null then
    v_omf_dest := null;
    dbms_output.put_line('OMF = OFF (db_create_file_dest is null)');

    -- Non-OMF: derive datafile directory (follow SYSTEM)
    select substr(file_name, 1, instr(file_name, '/', -1))
      into v_dir
      from dba_data_files
     where tablespace_name = 'SYSTEM'
       and rownum = 1;

    dbms_output.put_line('Datafile dir = ' || v_dir);
  else
    dbms_output.put_line('OMF = ON, db_create_file_dest = ' || v_omf_dest);
  end if;

  -- 2) Kill SOE sessions
  for r in (
    select sid, serial#
      from v$session
     where username = 'SOE'
       and type = 'USER'
  ) loop
    begin
      execute immediate 'alter system kill session ''' || r.sid || ',' || r.serial# || ''' immediate';
      dbms_output.put_line('Killed session ' || r.sid || ',' || r.serial#);
    exception
      when others then
        dbms_output.put_line('Kill session failed (ignored): ' || r.sid || ',' || r.serial# || ' - ' || sqlerrm);
    end;
  end loop;

  -- 3) Drop user (if exists)
  if exists_user('SOE') > 0 then
    dbms_output.put_line('Dropping user SOE ...');
    execute immediate 'drop user SOE cascade';
  else
    dbms_output.put_line('User SOE not exists, skip.');
  end if;

  -- 4) Drop tablespaces (if exist)
  if exists_ts('SOE_IDX') > 0 then
    dbms_output.put_line('Dropping tablespace SOE_IDX ...');
    execute immediate 'drop tablespace SOE_IDX including contents and datafiles';
  else
    dbms_output.put_line('Tablespace SOE_IDX not exists, skip.');
  end if;

  if exists_ts('SOE') > 0 then
    dbms_output.put_line('Dropping tablespace SOE ...');
    execute immediate 'drop tablespace SOE including contents and datafiles';
  else
    dbms_output.put_line('Tablespace SOE not exists, skip.');
  end if;

  -- 5) Create tablespaces (OMF vs non-OMF branches)
  dbms_output.put_line('Creating tablespace SOE ...');
  if v_omf_dest is not null then
    execute immediate
      'create tablespace SOE ' ||
      'datafile size 1024M ' ||
      'autoextend on next 256M maxsize unlimited ' ||
      'extent management local segment space management auto';
  else
    execute immediate
      'create tablespace SOE ' ||
      'datafile ''' || v_dir || 'soe01.dbf'' size 1024M ' ||
      'autoextend on next 256M maxsize unlimited ' ||
      'extent management local segment space management auto';
  end if;

  dbms_output.put_line('Creating tablespace SOE_IDX ...');
  if v_omf_dest is not null then
    execute immediate
      'create tablespace SOE_IDX ' ||
      'datafile size 512M ' ||
      'autoextend on next 256M maxsize unlimited ' ||
      'extent management local segment space management auto';
  else
    execute immediate
      'create tablespace SOE_IDX ' ||
      'datafile ''' || v_dir || 'soe_idx01.dbf'' size 512M ' ||
      'autoextend on next 256M maxsize unlimited ' ||
      'extent management local segment space management auto';
  end if;

  dbms_output.put_line('SOE / SOE_IDX tablespaces READY. (User SOE is NOT created by this script)');
end;
/

prompt
prompt ========== VERIFY ==========
col tablespace_name format a20
col status format a10
select tablespace_name, status
from dba_tablespaces
where tablespace_name in ('SOE','SOE_IDX')
order by tablespace_name;

col file_name format a120
select tablespace_name, file_name, bytes/1024/1024 as MB
from dba_data_files
where tablespace_name in ('SOE','SOE_IDX')
order by tablespace_name, file_name;

exit
`

	// Build sqlplus command using the connection credentials
	// sqlplus -L 'user/pass@//host:port/service' <<'SQL'
	return fmt.Sprintf("sqlplus -L '%s/%s@%s' <<'SQL'\n%s\nSQL",
		conn.Username,
		conn.Password,
		a.buildCharbenchConnectionString(conn),
		sqlScript)
}

// buildConnectionProbeCommand builds a sqlplus command to probe database connectivity using DBA credentials.
// This verifies:
// 1. Database is accessible
// 2. DBA credentials are correct
// 3. Instance is in valid status
func (a *SwingbenchAdapter) buildConnectionProbeCommand(conn *connection.OracleConnection, config *Config) string {
	probeSQL := `
set pages 200 lines 200
set heading off
set feedback off
-- Display current user
select 'Connected as: ' || user as connection_info from dual;
-- Display instance information
select 'Instance: ' || instance_name || ', Status: ' || status as instance_info from v$instance;
-- Display database version
select 'Version: ' || version as version_info from v$instance;
-- Count available tablespaces
select 'Available tablespaces: ' || count(*) as tablespace_count from v$tablespace;
exit
`

	// Use DBA credentials if provided, otherwise use connection credentials
	username, password, connectAs := resolveOracleAdminCredentials(conn, config)

	return fmt.Sprintf("sqlplus -L '%s/%s@%s%s' <<'SQL'\n%s\nSQL",
		username,
		password,
		a.buildCharbenchConnectionString(conn),
		sqlplusConnectAsSuffix(connectAs),
		probeSQL)
}

func resolveOracleAdminCredentials(conn *connection.OracleConnection, config *Config) (string, string, string) {
	username := strings.TrimSpace(conn.Username)
	password := conn.Password
	connectAs := strings.ToLower(strings.TrimSpace(conn.ConnectAs))
	if connectAs == "" {
		connectAs = "normal"
	}

	if dbaUser, ok := config.Parameters["dba_username"].(string); ok && strings.TrimSpace(dbaUser) != "" {
		username, connectAs = splitOracleAdminUsername(strings.TrimSpace(dbaUser), connectAs)
	}
	if dbaPass, ok := config.Parameters["dba_password"].(string); ok && dbaPass != "" {
		password = dbaPass
	}
	if sysUser, ok := config.Parameters["sysdba_username"].(string); ok && strings.TrimSpace(sysUser) != "" {
		username = strings.TrimSpace(sysUser)
		connectAs = "sysdba"
	}
	if sysPass, ok := config.Parameters["sysdba_password"].(string); ok && sysPass != "" {
		password = sysPass
		if connectAs == "normal" {
			connectAs = "sysdba"
		}
	}

	if strings.EqualFold(strings.TrimSpace(username), "sys") && connectAs == "normal" {
		connectAs = "sysdba"
	}

	return username, password, connectAs
}

func splitOracleAdminUsername(username string, fallbackConnectAs string) (string, string) {
	lower := strings.ToLower(username)
	switch {
	case strings.Contains(lower, " as sysdba"):
		return strings.TrimSpace(strings.ReplaceAll(lowerCaseReplaceOriginal(username, " as sysdba"), "  ", " ")), "sysdba"
	case strings.Contains(lower, " as sysoper"):
		return strings.TrimSpace(strings.ReplaceAll(lowerCaseReplaceOriginal(username, " as sysoper"), "  ", " ")), "sysoper"
	default:
		return strings.TrimSpace(username), fallbackConnectAs
	}
}

func resolveOracleWizardCredentials(conn *connection.OracleConnection, config *Config) (string, string) {
	username := strings.TrimSpace(conn.Username)
	password := conn.Password

	if value, ok := config.Parameters["oewizard_dba_username"].(string); ok && strings.TrimSpace(value) != "" {
		username = strings.TrimSpace(value)
	}
	if value, ok := config.Parameters["oewizard_dba_password"].(string); ok && value != "" {
		password = value
	}

	if strings.EqualFold(username, "sys") {
		username = "system"
	}

	return username, password
}

func lowerCaseReplaceOriginal(username, suffix string) string {
	lower := strings.ToLower(username)
	index := strings.Index(lower, suffix)
	if index < 0 {
		return username
	}
	return username[:index]
}

func sqlplusConnectAsSuffix(connectAs string) string {
	switch strings.ToLower(strings.TrimSpace(connectAs)) {
	case "sysdba":
		return " as sysdba"
	case "sysoper":
		return " as sysoper"
	default:
		return ""
	}
}

func (a *SwingbenchAdapter) buildSOEBootstrapCommand(conn *connection.OracleConnection, config *Config) string {
	bootstrapSQL := `
set echo on feedback on verify off heading on pagesize 200 linesize 200 trimspool on serveroutput on
whenever sqlerror exit sql.sqlcode rollback

declare
  v_user        varchar2(128) := 'SOE';
  v_pass        varchar2(128) := 'soe';
  v_ts_data     varchar2(128) := 'SOE';
  v_ts_idx      varchar2(128) := 'SOE_IDX';
  v_temp_ts     varchar2(128) := 'TEMP';
  v_data_mb     number := 1024;
  v_idx_mb      number := 512;
  v_next_mb     number := 64;
  v_omf_dest    varchar2(4000);
  v_sample_df   varchar2(4000);
  v_fs_dir      varchar2(4000);
  v_asm_dg      varchar2(4000);
  v_use_omf     varchar2(1);
  v_use_asm     varchar2(1);
  v_cnt         number;
  v_sql         clob;
  v_bigfile     varchar2(3);

  function norm_fs_dir(p varchar2) return varchar2 is
  begin
    if p is null then return null; end if;
    if substr(p, -1) = '/' then return p; end if;
    return p||'/';
  end;

  function datafile_clause(p_ts varchar2, p_role varchar2, p_size_mb number) return varchar2 is
    v_loc   varchar2(4000);
    v_fname varchar2(4000);
  begin
    if v_use_omf = 'Y' then
      return ' datafile size '||p_size_mb||'M';
    end if;
    if v_use_asm = 'Y' then
      return ' datafile '''||v_asm_dg||''' size '||p_size_mb||'M';
    end if;
    v_loc := norm_fs_dir(v_fs_dir);
    v_fname := lower(p_ts)||'_'||lower(p_role)||'_01.dbf';
    return ' datafile '''||v_loc||v_fname||''' size '||p_size_mb||'M';
  end;
begin
  select nvl(value, '') into v_omf_dest from v$parameter where name = 'db_create_file_dest';
  select name into v_sample_df from v$datafile where rownum = 1;

  v_fs_dir := nvl(regexp_substr(v_sample_df, '^(.*/)', 1, 1), '');
  v_asm_dg := nvl(regexp_substr(v_sample_df, '^\+[^/]+', 1, 1), '');
  v_use_omf := case when length(v_omf_dest) > 0 then 'Y' else 'N' end;
  v_use_asm := case when regexp_like(v_sample_df, '^\+') then 'Y' else 'N' end;

  select count(*) into v_cnt from dba_tablespaces where tablespace_name = v_ts_data;
  if v_cnt = 0 then
    v_sql := 'create tablespace '||v_ts_data
          || datafile_clause(v_ts_data, 'data', v_data_mb)
          || ' autoextend on next '||v_next_mb||'M maxsize unlimited'
          || ' extent management local segment space management auto';
    execute immediate v_sql;
    dbms_output.put_line('Created tablespace '||v_ts_data);
  else
    select bigfile into v_bigfile from dba_tablespaces where tablespace_name = v_ts_data;
    if v_bigfile = 'YES' then
      raise_application_error(-20002, 'Tablespace '||v_ts_data||' already exists and is BIGFILE.');
    end if;
    dbms_output.put_line('Tablespace exists: '||v_ts_data);
  end if;

  select count(*) into v_cnt from dba_tablespaces where tablespace_name = v_ts_idx;
  if v_cnt = 0 then
    v_sql := 'create tablespace '||v_ts_idx
          || datafile_clause(v_ts_idx, 'idx', v_idx_mb)
          || ' autoextend on next '||v_next_mb||'M maxsize unlimited'
          || ' extent management local segment space management auto';
    execute immediate v_sql;
    dbms_output.put_line('Created tablespace '||v_ts_idx);
  else
    select bigfile into v_bigfile from dba_tablespaces where tablespace_name = v_ts_idx;
    if v_bigfile = 'YES' then
      raise_application_error(-20003, 'Tablespace '||v_ts_idx||' already exists and is BIGFILE.');
    end if;
    dbms_output.put_line('Tablespace exists: '||v_ts_idx);
  end if;

  select count(*) into v_cnt from dba_users where username = v_user;
  if v_cnt = 0 then
    execute immediate 'create user '||v_user||' identified by "'||replace(v_pass,'"','""')||'"'
                   || ' default tablespace '||v_ts_data
                   || ' temporary tablespace '||v_temp_ts;
    dbms_output.put_line('Created user '||v_user);
  else
    execute immediate 'alter user '||v_user||' identified by "'||replace(v_pass,'"','""')||'"';
    execute immediate 'alter user '||v_user||' default tablespace '||v_ts_data;
    execute immediate 'alter user '||v_user||' temporary tablespace '||v_temp_ts;
    dbms_output.put_line('User exists, altered: '||v_user);
  end if;

  execute immediate 'alter user '||v_user||' quota unlimited on '||v_ts_data;
  execute immediate 'grant create session, create table, create sequence, create procedure, create trigger, create view, create type, create synonym to '||v_user;
  execute immediate 'grant unlimited tablespace to '||v_user;
  dbms_output.put_line('Bootstrap complete.');
end;
/
exit
`

	tempSQLFile, err := writeOracleTempSQL("db-benchmind-cts-oracle", bootstrapSQL)
	if err != nil {
		return fmt.Sprintf("echo 'Failed to create temp SQL file: %v' && exit 1", err)
	}

	username, password, connectAs := resolveOracleAdminCredentials(conn, config)
	return fmt.Sprintf("sqlplus -L '%s/%s@%s%s' @%s",
		username, password, a.buildCharbenchConnectionString(conn), sqlplusConnectAsSuffix(connectAs), tempSQLFile)
}

func (a *SwingbenchAdapter) buildCleanupSOEObjectsCommand(conn *connection.OracleConnection, config *Config) string {
	cleanupSQL := `
set echo on feedback on verify off heading on pagesize 200 linesize 200 trimspool on serveroutput on
whenever sqlerror continue

declare
begin
  for r in (
    select sid, serial#
      from v$session
     where username = 'SOE'
       and type = 'USER'
  ) loop
    begin
      execute immediate 'alter system kill session ''' || r.sid || ',' || r.serial# || ''' immediate';
      dbms_output.put_line('Killed session ' || r.sid || ',' || r.serial#);
    exception
      when others then
        dbms_output.put_line('Kill session skipped: ' || sqlerrm);
    end;
  end loop;

  begin
    execute immediate 'drop user SOE cascade';
    dbms_output.put_line('Dropped user SOE');
  exception
    when others then
      dbms_output.put_line('Drop user skipped: ' || sqlerrm);
  end;

  begin
    execute immediate 'drop tablespace SOE_IDX including contents and datafiles';
    dbms_output.put_line('Dropped tablespace SOE_IDX');
  exception
    when others then
      dbms_output.put_line('Drop tablespace SOE_IDX skipped: ' || sqlerrm);
  end;

  begin
    execute immediate 'drop tablespace SOE including contents and datafiles';
    dbms_output.put_line('Dropped tablespace SOE');
  exception
    when others then
      dbms_output.put_line('Drop tablespace SOE skipped: ' || sqlerrm);
  end;
end;
/
exit
`

	tempSQLFile, err := writeOracleTempSQL("db-benchmind-cleanup-soe", cleanupSQL)
	if err != nil {
		return fmt.Sprintf("echo 'Failed to create temp SQL file: %v' && exit 1", err)
	}

	username, password, connectAs := resolveOracleAdminCredentials(conn, config)
	return fmt.Sprintf("sqlplus -L '%s/%s@%s%s' @%s",
		username, password, a.buildCharbenchConnectionString(conn), sqlplusConnectAsSuffix(connectAs), tempSQLFile)
}

// buildSchemaExistenceCheckCommand builds a sqlplus command to check if SOE schema already exists.
// This provides early detection to avoid running oewizard unnecessarily.
func (a *SwingbenchAdapter) buildSchemaExistenceCheckCommand(conn *connection.OracleConnection, config *Config) string {
	checkSQL := `
set pages 0 lines 200
set heading off
set feedback off
set serveroutput on

declare
  v_user_count    number;
  v_table_count   number;
  v_schema_exists  varchar2(10) := 'NO';

begin
  -- Check if SOE user exists
  select count(*) into v_user_count
    from dba_users
   where username = 'SOE';

  -- Check if SOE tables exist
  select count(*) into v_table_count
    from all_tables
   where owner = 'SOE'
     and table_name in ('ORDERS', 'ORDER_ITEMS', 'CUSTOMERS', 'WAREHOUSES');

  -- Determine if schema exists
  if v_user_count > 0 and v_table_count > 0 then
    v_schema_exists := 'YES';
  end if;

  dbms_output.put_line('SOE_SCHEMA_EXISTS: ' || v_schema_exists);

  if v_schema_exists = 'YES' then
    dbms_output.put_line('ERROR: SOE schema already exists with ' || v_table_count || ' tables.');
    dbms_output.put_line('ERROR: Please run Cleanup phase first before re-creating.');
    -- Return error code to stop execution
    raise_application_error(-20001, 'SOE schema already exists');
  else
    dbms_output.put_line('SUCCESS: SOE schema does not exist, can proceed with data generation.');
  end if;

exception
  when others then
    dbms_output.put_line('ERROR: Schema check failed - ' || sqlerrm);
    raise;
end;
/

exit
`

	// Use DBA credentials for schema check
	username, password, connectAs := resolveOracleAdminCredentials(conn, config)

	return fmt.Sprintf("sqlplus -L '%s/%s@%s%s' <<'SQL'\n%s\nSQL",
		username,
		password,
		a.buildCharbenchConnectionString(conn),
		sqlplusConnectAsSuffix(connectAs),
		checkSQL)
}

// buildOewizardCreateCommand builds the oewizard -create command.
// Uses correct parameters matching the user's example.
func (a *SwingbenchAdapter) buildOewizardCreateCommand(conn *connection.OracleConnection, connectionStr string, scaleFloat float64, config *Config) string {
	// Use system default java to avoid hardcoded path issues
	javaPath := "java"
	dbaUser, dbaPass := resolveOracleWizardCredentials(conn, config)

	cmdArgs := []string{
		javaPath,
		"-Duser.timezone=Asia/Shanghai",
		"-Doracle.jdbc.timezoneAsRegion=false",
		"-cp", "../launcher",
		"LauncherBootstrap",
		"-executablename", "oewizard",
		"oewizard",
		"-c", "oewizard.xml", // Config file (in bin directory)
		"-cl", // Command line mode
		"-create",
		"-version", "2.0",
		"-cs", connectionStr,
		"-dt", "thin", // Driver type: thin
		"-dba", dbaUser,
		"-dbap", dbaPass,
		"-u", "soe",
		"-p", "soe",
		"-ts", "SOE", // Use existing tablespace
		"-nopart",                                 // No partitioning
		"-nocompress",                             // No compression
		"-normalfile",                             // Normal file type
		"-scale", fmt.Sprintf("%.1f", scaleFloat), // Support float scale like 0.1 for test templates
		"-v",
		"-debug",
		"-debugf", "/tmp/oewizard_debug.log",
	}

	return strings.Join(cmdArgs, " ")
}

func (a *SwingbenchAdapter) buildOewizardGenerateCommand(conn *connection.OracleConnection, connectionStr string, scaleFloat float64, threads int, config *Config) string {
	javaPath := "java"
	dbaUser, dbaPass := resolveOracleWizardCredentials(conn, config)

	cmdArgs := []string{
		javaPath,
		"-Duser.timezone=Asia/Shanghai",
		"-Doracle.jdbc.timezoneAsRegion=false",
		"-cp", "../launcher",
		"LauncherBootstrap",
		"-executablename", "oewizard",
		"oewizard",
		"-c", "oewizard.xml",
		"-cl",
		"-generate",
		"-version", "2.0",
		"-cs", connectionStr,
		"-dt", "thin",
		"-dba", dbaUser,
		"-dbap", dbaPass,
		"-u", "soe",
		"-p", "soe",
		"-ts", "SOE",
		"-nopart",
		"-nocompress",
		"-normalfile",
		"-scale", fmt.Sprintf("%.1f", scaleFloat),
		"-tc", fmt.Sprintf("%d", threads),
		"-v",
		"-debug",
		"-debugf", "/tmp/oewizard_generate_debug.log",
	}

	return strings.Join(cmdArgs, " ")
}

func (a *SwingbenchAdapter) buildOewizardAllIndexesCommand(conn *connection.OracleConnection, connectionStr string, scaleFloat float64, config *Config) string {
	javaPath := "java"
	dbaUser, dbaPass := resolveOracleWizardCredentials(conn, config)

	cmdArgs := []string{
		javaPath,
		"-Duser.timezone=Asia/Shanghai",
		"-Doracle.jdbc.timezoneAsRegion=false",
		"-cp", "../launcher",
		"LauncherBootstrap",
		"-executablename", "oewizard",
		"oewizard",
		"-c", "oewizard.xml",
		"-cl",
		"-allindexes",
		"-version", "2.0",
		"-cs", connectionStr,
		"-dt", "thin",
		"-dba", dbaUser,
		"-dbap", dbaPass,
		"-u", "soe",
		"-p", "soe",
		"-ts", "SOE",
		"-nopart",
		"-nocompress",
		"-normalfile",
		"-scale", fmt.Sprintf("%.1f", scaleFloat),
		"-v",
		"-debug",
		"-debugf", "/tmp/oewizard_indexes_debug.log",
	}

	return strings.Join(cmdArgs, " ")
}

// buildCreateTablespacesOnlyCommand creates only the tablespaces (no drop).
// This is needed because oewizard can't create tablespaces without datafile paths.
// The tablespaces are created empty, then oewizard will detect and use them.
func (a *SwingbenchAdapter) buildCreateTablespacesOnlyCommand(conn *connection.OracleConnection, config *Config) string {
	sqlScript := `
set echo on feedback on verify off heading on pagesize 200 linesize 200 trimspool on serveroutput on
whenever sqlerror exit sql.sqlcode rollback

declare
  v_dir      varchar2(1024);
  v_omf_dest varchar2(1024);

  function exists_ts(p_ts varchar2) return number is
    n number;
  begin
    select count(*) into n
      from dba_tablespaces
     where tablespace_name = upper(p_ts);
    return n;
  end;

begin
  -- 1) Check if OMF is enabled
  select value
    into v_omf_dest
    from v$parameter
   where name = 'db_create_file_dest';

  if v_omf_dest is null or trim(v_omf_dest) is null then
    v_omf_dest := null;
    dbms_output.put_line('OMF = OFF (db_create_file_dest is null)');

    -- Non-OMF: derive datafile directory (follow SYSTEM)
    select substr(file_name, 1, instr(file_name, '/', -1))
      into v_dir
      from dba_data_files
     where tablespace_name = 'SYSTEM'
       and rownum = 1;

    dbms_output.put_line('Datafile dir = ' || v_dir);
  else
    dbms_output.put_line('OMF = ON, db_create_file_dest = ' || v_omf_dest);
  end if;

  -- 2) Create SOE tablespace (if not exists)
  if exists_ts('SOE') = 0 then
    dbms_output.put_line('Creating tablespace SOE ...');
    if v_omf_dest is not null then
      execute immediate
        'create tablespace SOE ' ||
        'datafile size 1024M ' ||
        'autoextend on next 256M maxsize unlimited ' ||
        'extent management local segment space management auto';
    else
      execute immediate
        'create tablespace SOE ' ||
        'datafile ''' || v_dir || 'soe01.dbf'' size 1024M ' ||
        'autoextend on next 256M maxsize unlimited ' ||
        'extent management local segment space management auto';
    end if;
    dbms_output.put_line('Tablespace SOE created.');
  else
    dbms_output.put_line('Tablespace SOE already exists, skipping.');
  end if;

  -- 3) Create SOE_IDX tablespace (if not exists)
  if exists_ts('SOE_IDX') = 0 then
    dbms_output.put_line('Creating tablespace SOE_IDX ...');
    if v_omf_dest is not null then
      execute immediate
        'create tablespace SOE_IDX ' ||
        'datafile size 512M ' ||
        'autoextend on next 256M maxsize unlimited ' ||
        'extent management local segment space management auto';
    else
      execute immediate
        'create tablespace SOE_IDX ' ||
        'datafile ''' || v_dir || 'soe_idx01.dbf'' size 512M ' ||
        'autoextend on next 256M maxsize unlimited ' ||
        'extent management local segment space management auto';
    end if;
    dbms_output.put_line('Tablespace SOE_IDX created.');
  else
    dbms_output.put_line('Tablespace SOE_IDX already exists, skipping.');
  end if;

  -- 4) Verify
  dbms_output.put_line('========== VERIFY ==========');
  for ts in (select tablespace_name, status from dba_tablespaces where tablespace_name in ('SOE','SOE_IDX') order by tablespace_name) loop
    dbms_output.put_line('Tablespace: ' || ts.tablespace_name || ' Status: ' || ts.status);
  end loop;

end;
/

exit
`

	// Write SQL to a temp file first, then execute it
	// This avoids complex shell escaping issues with heredocs
	tempSQLFile, err := writeOracleTempSQL("db-benchmind-creatablespace", sqlScript)
	if err != nil {
		// If we can't write the file, fall back to a simple command
		// This shouldn't happen in normal operation
		return fmt.Sprintf("echo 'Failed to create temp SQL file: %v' && exit 1", err)
	}
	// Return command to execute the SQL file with sqlplus
	username, password, connectAs := resolveOracleAdminCredentials(conn, config)
	return fmt.Sprintf("sqlplus -L '%s/%s@%s%s' @%s",
		username, password, a.buildCharbenchConnectionString(conn), sqlplusConnectAsSuffix(connectAs), tempSQLFile)
}

// buildPostSchemaSetupCommand builds the post-schema setup command.
// This grants necessary permissions and recompiles invalid objects.
func (a *SwingbenchAdapter) buildPostSchemaSetupCommand(conn *connection.OracleConnection, config *Config) string {
	// Use DBA credentials for granting privileges
	// For SYSDBA connection, we need to use "sys" username, not "system"
	username, password, connectAs := resolveOracleAdminCredentials(conn, config)
	if connectAs == "" || connectAs == "normal" {
		connectAs = "sysdba"
	}

	// Build connection string for sysdba
	connStr := a.buildCharbenchConnectionString(conn)

	setupSQL := `
set echo on
-- 1) Grant DBMS_LOCK permission to SOE user (required for benchmark)
grant execute on sys.dbms_lock to soe;

-- 2) Recompile ORDERENTRY package body (may be invalid after schema creation)
alter package soe.orderentry compile body;

-- 3) Verify all SOE objects are valid
select object_name, object_type, status
from dba_objects
where owner='SOE'
and object_name='ORDERENTRY';

-- 4) Show any compilation errors
show errors package body soe.orderentry;

exit
`

	// Write SQL to temp file and execute it
	tempSQLFile, err := writeOracleTempSQL("db-benchmind-postschema", setupSQL)
	if err != nil {
		return fmt.Sprintf("echo 'Failed to create temp SQL file: %v' && exit 1", err)
	}

	// Connect as SYSDBA for granting privileges
	return fmt.Sprintf("sqlplus -L '%s/%s@%s%s' @%s",
		username, password, connStr, sqlplusConnectAsSuffix(connectAs), tempSQLFile)
}

// buildOewizardDropCommand builds the oewizard -drop command.
// Drops the SOE schema, user, and tablespace.
func (a *SwingbenchAdapter) buildOewizardDropCommand(conn *connection.OracleConnection, connectionStr string, config *Config) string {
	// Use system default java to avoid hardcoded path issues
	javaPath := "java"

	// Get DBA credentials from parameters, fallback to connection credentials
	dbaUser := conn.Username
	dbaPass := conn.Password
	if u, ok := config.Parameters["dba_username"].(string); ok && u != "" {
		dbaUser = u
	}
	if p, ok := config.Parameters["dba_password"].(string); ok && p != "" {
		dbaPass = p
	}

	cmdArgs := []string{
		javaPath,
		"-Duser.timezone=Asia/Shanghai",
		"-Doracle.jdbc.timezoneAsRegion=false",
		"-cp", "../launcher",
		"LauncherBootstrap",
		"-executablename", "oewizard",
		"oewizard",
		"-c", "oewizard.xml", // Config file (in bin directory)
		"-cl", // Command line mode
		"-drop",
		"-version", "2.0",
		"-cs", connectionStr,
		"-dt", "thin", // Driver type: thin
		"-dba", dbaUser,
		"-dbap", dbaPass,
		"-u", "soe",
		"-p", "soe",
		"-ts", "SOE", // Tablespace name
		"-v",
		"-debug",
		"-debugf", "/tmp/oewizard_drop_debug.log",
	}

	return strings.Join(cmdArgs, " ")
}
