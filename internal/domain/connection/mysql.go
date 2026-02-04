// Package connection provides MySQL connection implementation.
// Implements: REQ-CONN-002, REQ-CONN-003, REQ-CONN-010
package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// MySQLConnection represents a MySQL database connection configuration.
// Implements spec.md 3.2.2
type MySQLConnection struct {
	// Base fields
	BaseConnection

	// Connection parameters
	Host     string `json:"host"`     // Host address
	Port     int    `json:"port"`     // Port (default 3306)
	Database string `json:"database"` // Database name
	Username string `json:"username"` // Username
	Password string `json:"-"`        // Password (not serialized, stored in keyring)

	// SSL configuration
	SSLMode string `json:"ssl_mode"` // SSL mode: disabled/preferred/required

	// SSH tunnel configuration
	SSH *SSHTunnelConfig `json:"ssh,omitempty"` // SSH tunnel configuration
}

// GetType returns DatabaseTypeMySQL.
func (c *MySQLConnection) GetType() DatabaseType {
	return DatabaseTypeMySQL
}

// GetDSN generates a connection string without password (for logging).
// Format: username@tcp(host:port)/database
// If database is empty, returns: username@tcp(host:port)/
func (c *MySQLConnection) GetDSN() string {
	if c.Database == "" {
		return fmt.Sprintf("%s@tcp(%s:%d)/", c.Username, c.Host, c.Port)
	}
	return fmt.Sprintf("%s@tcp(%s:%d)/%s", c.Username, c.Host, c.Port, c.Database)
}

// GetDSNWithPassword generates a complete connection string with password.
// Format: username:password@tcp(host:port)/database
// If database is empty, returns: username:password@tcp(host:port)/
func (c *MySQLConnection) GetDSNWithPassword() string {
	if c.Database == "" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.Username, c.Password, c.Host, c.Port)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
}

// Redact returns a redacted connection string for display (REQ-CONN-008).
// Format: "name (***@host:port/database)" or "name (***@host:port)" if no database
func (c *MySQLConnection) Redact() string {
	if c.Database == "" {
		return fmt.Sprintf("%s (***@%s:%d)", c.Name, c.Host, c.Port)
	}
	return fmt.Sprintf("%s (***@%s:%d/%s)", c.Name, c.Host, c.Port, c.Database)
}

// ToJSON serializes the connection to JSON (without password).
func (c *MySQLConnection) ToJSON() ([]byte, error) {
	// BaseConnection's JSON tag will handle serialization
	// Password field has json:"-" tag to exclude it
	return nil, fmt.Errorf("not implemented yet - will use json.Marshal")
}

// Validate validates the connection parameters (REQ-CONN-010).
// Returns an error if any required field is missing or invalid.
// Note: database field is optional for MySQL (can connect without specifying database).
func (c *MySQLConnection) Validate() error {
	var errs []error

	// Validate required fields
	if err := ValidateRequired("name", c.Name); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateRequired("host", c.Host); err != nil {
		errs = append(errs, err)
	}
	// Database is optional for MySQL - can connect without specifying a database
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

	// Note: SSL mode validation removed - we auto-detect the best mode
	// c.SSLMode field is kept for backward compatibility but not validated

	if len(errs) > 0 {
		return &MultiValidationError{Errors: errs}
	}

	return nil
}

// Test tests the MySQL connection availability with intelligent SSL detection.
//
// If SSH tunnel is enabled, it establishes the tunnel first.
//
// It attempts multiple SSL configurations in order:
// 1. disabled (no SSL - fastest)
// 2. preferred (auto-detect, fallback to no SSL)
// 3. required (force SSL)
//
// Returns: TestResult with success/failure, latency, version, error.
func (c *MySQLConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	// Variables to track connection target
	targetHost := c.Host
	targetPort := c.Port

	// Create SSH tunnel if enabled
	var tunnel *SSHTunnel
	if c.SSH != nil && c.SSH.Enabled {
		var err error
		tunnel, err = NewSSHTunnel(ctx, c.SSH, c.Host, c.Port)
		if err != nil {
			slog.Error("MySQL: Failed to create SSH tunnel", "error", err)
			return &TestResult{
				Success:   false,
				LatencyMs: time.Since(start).Milliseconds(),
				Error:     fmt.Sprintf("SSH tunnel failed: %v", err),
			}, nil
		}
		defer tunnel.Close()

		// Use tunnel's local port
		targetHost = "127.0.0.1"
		targetPort = tunnel.GetLocalPort()
		slog.Info("MySQL: Using SSH tunnel", "local_port", targetPort)
	}

	// SSL modes to try in order (most common first)
	sslModes := []string{"disabled", "preferred", "required"}

	var lastErr error
	for _, sslMode := range sslModes {
		dsn := c.buildDSNWithSSL(sslMode, targetHost, targetPort)

		slog.Info("MySQL: Testing connection",
			"host", targetHost,
			"port", targetPort,
			"ssh_tunnel", tunnel != nil,
			"ssl_mode", sslMode,
			"username", c.Username)

		result, err := c.testConnection(ctx, dsn, start)
		if err != nil {
			// Context cancelled or timeout
			return nil, fmt.Errorf("test cancelled: %w", err)
		}

		if result.Success {
			slog.Info("MySQL: Connection successful",
				"ssl_mode", sslMode,
				"ssh_tunnel", tunnel != nil,
				"latency_ms", result.LatencyMs,
				"version", result.DatabaseVersion)
			return result, nil
		}

		// Save last error for reporting
		lastErr = fmt.Errorf("ssl_mode=%s: %s", sslMode, result.Error)
		slog.Debug("MySQL: Connection attempt failed",
			"ssl_mode", sslMode,
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
func (c *MySQLConnection) testConnection(ctx context.Context, dsn string, start time.Time) (*TestResult, error) {
	db, err := sql.Open("mysql", dsn)
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
	err = db.QueryRowContext(testCtx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		version = "unknown"
	}

	return &TestResult{
		Success:         true,
		LatencyMs:       latency,
		DatabaseVersion: version,
	}, nil
}

// buildDSNWithSSL builds a DSN with the specified SSL mode.
// Format: username:password@tcp(host:port)/database?tls=xxx
// If database is empty: username:password@tcp(host:port)/?tls=xxx
func (c *MySQLConnection) buildDSNWithSSL(sslMode string, host string, port int) string {
	var dsn string
	if c.Database == "" {
		// No database specified
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/?tls=%s",
			c.Username, c.Password, host, port, sslMode)
	} else {
		// With database
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s",
			c.Username, c.Password, host, port, c.Database, sslMode)
	}
	return dsn
}

// MultiValidationError represents multiple validation errors.
type MultiValidationError struct {
	Errors []error
}

func (e *MultiValidationError) Error() string {
	var msg string
	for _, err := range e.Errors {
		if msg != "" {
			msg += "; "
		}
		msg += err.Error()
	}
	return msg
}

// SetPassword sets the password (used by keyring provider).
func (c *MySQLConnection) SetPassword(password string) {
	c.Password = password
	c.UpdatedAt = time.Now()
}

// GetPassword returns the password (used by keyring provider).
func (c *MySQLConnection) GetPassword() string {
	return c.Password
}
