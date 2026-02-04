// Package connection provides SQL Server connection implementation.
// Implements: REQ-CONN-002
package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/microsoft/go-mssqldb" // SQL Server driver
)

// SQLServerConnection represents a SQL Server database connection configuration.
// Implements spec.md 3.2.2
type SQLServerConnection struct {
	// Base fields
	BaseConnection

	// Connection parameters
	Host                   string `json:"host"`                     // Host address
	Port                   int    `json:"port"`                     // Port (default 1433)
	Database               string `json:"database"`                 // Database name
	Username               string `json:"username"`                 // Username
	Password               string `json:"-"`                        // Password (stored in keyring)
	TrustServerCertificate bool   `json:"trust_server_certificate"` // Trust server certificate

	// WinRM configuration (for Windows Server monitoring)
	WinRM *WinRMConfig `json:"winrm,omitempty"` // WinRM configuration (optional)
}

// GetType returns DatabaseTypeSQLServer.
func (c *SQLServerConnection) GetType() DatabaseType {
	return DatabaseTypeSQLServer
}

// GetDSN generates a connection string without password (for logging).
// Format: sqlserver://username@host:port?database=dbname
func (c *SQLServerConnection) GetDSN() string {
	return fmt.Sprintf("sqlserver://%s@%s:%d?database=%s", c.Username, c.Host, c.Port, c.Database)
}

// GetDSNWithPassword generates a complete connection string with password.
// Format: sqlserver://username:password@host:port?database=dbname&trustservercertificate=true/false
func (c *SQLServerConnection) GetDSNWithPassword() string {
	trustParam := "false"
	if c.TrustServerCertificate {
		trustParam = "true"
	}
	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&trustservercertificate=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, trustParam)
}

// Redact returns a redacted connection string for display (REQ-CONN-008).
func (c *SQLServerConnection) Redact() string {
	return fmt.Sprintf("%s (***@%s:%d/%s)", c.Name, c.Host, c.Port, c.Database)
}

// ToJSON serializes the connection to JSON (without password).
func (c *SQLServerConnection) ToJSON() ([]byte, error) {
	return nil, fmt.Errorf("not implemented yet - will use json.Marshal")
}

// Validate validates the connection parameters (REQ-CONN-010).
func (c *SQLServerConnection) Validate() error {
	var errs []error

	// Validate required fields
	if err := ValidateRequired("name", c.Name); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateRequired("host", c.Host); err != nil {
		errs = append(errs, err)
	}
	// Database is optional for SQL Server - can connect without specifying a database
	// if err := ValidateRequired("database", c.Database); err != nil {
	// 	errs = append(errs, err)
	// }
	if err := ValidateRequired("username", c.Username); err != nil {
		errs = append(errs, err)
	}

	// Validate port
	if err := ValidatePort(c.Port); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return &MultiValidationError{Errors: errs}
	}

	return nil
}

// Test tests the SQL Server connection availability with intelligent encryption detection.
//
// It attempts multiple encryption configurations in order:
// 1. No encryption, trust certificate (most common)
// 2. Encryption enabled, trust certificate
// 3. No encryption, no trust
// 4. Encryption enabled, no trust
//
// Returns: TestResult with success/failure, latency, version, error.
func (c *SQLServerConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	// Connection configurations to try in order
	configs := []struct {
		encrypt                bool
		trustServerCertificate bool
		desc                   string
	}{
		{false, true, "no encryption, trust certificate"},
		{true, true, "encryption enabled, trust certificate"},
		{false, false, "no encryption, no trust"},
		{true, false, "encryption enabled, no trust"},
	}

	var lastErr error
	for _, config := range configs {
		dsn := c.buildDSNWithConfig(config.encrypt, config.trustServerCertificate)

		slog.Info("SQL Server: Testing connection",
			"host", c.Host,
			"port", c.Port,
			"encrypt", config.encrypt,
			"trust_server_certificate", config.trustServerCertificate,
			"username", c.Username)

		result, err := c.testConnection(ctx, dsn, start)
		if err != nil {
			// Context cancelled or timeout
			return nil, fmt.Errorf("test cancelled: %w", err)
		}

		if result.Success {
			slog.Info("SQL Server: Connection successful",
				"config", config.desc,
				"latency_ms", result.LatencyMs,
				"version", result.DatabaseVersion)
			return result, nil
		}

		// Save last error for reporting
		lastErr = fmt.Errorf("%s: %s", config.desc, result.Error)
		slog.Debug("SQL Server: Connection attempt failed",
			"config", config.desc,
			"error", result.Error)
	}

	// All attempts failed
	latency := time.Since(start).Milliseconds()
	return &TestResult{
		Success:   false,
		LatencyMs: latency,
		Error:     fmt.Sprintf("all connection attempts failed. Last error: %v", lastErr),
	}, nil
}

// testConnection performs a single connection attempt with the given DSN.
func (c *SQLServerConnection) testConnection(ctx context.Context, dsn string, start time.Time) (*TestResult, error) {
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return &TestResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to open connection: %v", err),
			LatencyMs: time.Since(start).Milliseconds(),
		}, nil
	}
	defer db.Close()

	// Set timeout for this connection attempt (5 seconds per attempt)
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Attempt to ping the database
	err = db.PingContext(testCtx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &TestResult{
			Success:   false,
			LatencyMs: latency,
			Error:     fmt.Sprintf("%v", err),
		}, nil
	}

	// Get database version
	var version string
	err = db.QueryRowContext(testCtx, "SELECT @@VERSION").Scan(&version)
	if err != nil {
		version = "unknown"
	}

	return &TestResult{
		Success:         true,
		LatencyMs:       latency,
		DatabaseVersion: version,
	}, nil
}

// buildDSNWithConfig builds a DSN with the specified encryption and trust settings.
// Format: sqlserver://username:password@host:port?database=xxx&encrypt=xxx&trustservercertificate=xxx
func (c *SQLServerConnection) buildDSNWithConfig(encrypt, trustServerCert bool) string {
	// Build connection URL with encryption parameters
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=%t&trustservercertificate=%t",
		c.Username, c.Password, c.Host, c.Port, c.Database, encrypt, trustServerCert)
	return dsn
}

// SetPassword sets the password (used by keyring provider).
func (c *SQLServerConnection) SetPassword(password string) {
	c.Password = password
	c.UpdatedAt = time.Now()
}

// GetPassword returns the password (used by keyring provider).
func (c *SQLServerConnection) GetPassword() string {
	return c.Password
}

// GetWinRMConfig returns the WinRM configuration.
func (c *SQLServerConnection) GetWinRMConfig() *WinRMConfig {
	return c.WinRM
}

// SetWinRMConfig sets the WinRM configuration.
func (c *SQLServerConnection) SetWinRMConfig(config *WinRMConfig) {
	c.WinRM = config
	c.UpdatedAt = time.Now()
}
