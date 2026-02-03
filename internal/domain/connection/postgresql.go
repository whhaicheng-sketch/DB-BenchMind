// Package connection provides PostgreSQL connection implementation.
// Implements: REQ-CONN-002
package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq" // Register PostgreSQL driver
)

// PostgreSQLConnection represents a PostgreSQL database connection configuration.
// Implements spec.md 3.2.2
type PostgreSQLConnection struct {
	// Base fields
	BaseConnection

	// Connection parameters
	Host     string `json:"host"`     // Host address
	Port     int    `json:"port"`     // Port (default 5432)
	Database string `json:"database"` // Database name
	Username string `json:"username"` // Username
	Password string `json:"-"`        // Password (stored in keyring)
	SSLMode  string `json:"ssl_mode"` // SSL mode: disable/allow/prefer/require/verify-ca/verify-full
}

// GetType returns DatabaseTypePostgreSQL.
func (c *PostgreSQLConnection) GetType() DatabaseType {
	return DatabaseTypePostgreSQL
}

// GetDSN generates a connection string without password (for logging).
// Format: host=host port=port dbname=database user=username
func (c *PostgreSQLConnection) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s", c.Host, c.Port, c.Database, c.Username)
}

// GetDSNWithPassword generates a complete connection string with password.
// Format: host=host port=port dbname=database user=username password=password sslmode=ssl_mode
func (c *PostgreSQLConnection) GetDSNWithPassword() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.Username, c.Password, sslMode)
}

// Redact returns a redacted connection string for display (REQ-CONN-008).
func (c *PostgreSQLConnection) Redact() string {
	return fmt.Sprintf("%s (***@%s:%d/%s)", c.Name, c.Host, c.Port, c.Database)
}

// ToJSON serializes the connection to JSON (without password).
func (c *PostgreSQLConnection) ToJSON() ([]byte, error) {
	return nil, fmt.Errorf("not implemented yet - will use json.Marshal")
}

// Validate validates the connection parameters (REQ-CONN-010).
func (c *PostgreSQLConnection) Validate() error {
	var errs []error

	// Validate required fields
	if err := ValidateRequired("name", c.Name); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateRequired("host", c.Host); err != nil {
		errs = append(errs, err)
	}
	// Database is required for PostgreSQL
	if err := ValidateRequired("database", c.Database); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateRequired("username", c.Username); err != nil {
		errs = append(errs, err)
	}

	// Validate port
	if err := ValidatePort(c.Port); err != nil {
		errs = append(errs, err)
	}

	// Validate SSL mode (only modes supported by most PostgreSQL servers)
	validSSLMode := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
		"":            true, // empty is allowed (will use default)
	}
	if c.SSLMode != "" && !validSSLMode[c.SSLMode] {
		errs = append(errs, &ValidationError{
			Field:   "ssl_mode",
			Message: "ssl_mode must be one of: disable, require, verify-ca, verify-full",
			Value:   c.SSLMode,
		})
	}

	if len(errs) > 0 {
		return &MultiValidationError{Errors: errs}
	}

	return nil
}

// Test tests the PostgreSQL connection availability with intelligent SSL detection.
//
// It attempts multiple SSL configurations in order:
// 1. disable (no SSL - fastest)
// 2. require (SSL without verification)
// 3. verify-ca (SSL with CA verification)
//
// Returns: TestResult with success/failure, latency, version, error.
func (c *PostgreSQLConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	// SSL modes to try in order (most common first)
	sslModes := []struct {
		mode   string
		desc   string
	}{
		{"disable", "no SSL"},
		{"require", "SSL without verification"},
		{"verify-ca", "SSL with CA verification"},
	}

	var lastErr error
	for _, sslConfig := range sslModes {
		dsn := c.buildDSNWithSSL(sslConfig.mode)

		slog.Info("PostgreSQL: Testing connection",
			"host", c.Host,
			"port", c.Port,
			"sslmode", sslConfig.mode,
			"username", c.Username)

		result, err := c.testConnection(ctx, dsn, start)
		if err != nil {
			// Context cancelled or timeout
			return nil, fmt.Errorf("test cancelled: %w", err)
		}

		if result.Success {
			slog.Info("PostgreSQL: Connection successful",
				"sslmode", sslConfig.mode,
				"latency_ms", result.LatencyMs,
				"version", result.DatabaseVersion)
			return result, nil
		}

		// Save last error for reporting
		lastErr = fmt.Errorf("sslmode=%s: %s", sslConfig.mode, result.Error)
		slog.Debug("PostgreSQL: Connection attempt failed",
			"sslmode", sslConfig.mode,
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
func (c *PostgreSQLConnection) testConnection(ctx context.Context, dsn string, start time.Time) (*TestResult, error) {
	db, err := sql.Open("postgres", dsn)
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
	err = db.QueryRowContext(testCtx, "SELECT version()").Scan(&version)
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
// Format: postgres://username:password@host:port/database?sslmode=xxx
func (c *PostgreSQLConnection) buildDSNWithSSL(sslMode string) string {
	// Build connection URL with SSL mode parameter
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, sslMode)
	return dsn
}

// SetPassword sets the password (used by keyring provider).
func (c *PostgreSQLConnection) SetPassword(password string) {
	c.Password = password
	c.UpdatedAt = time.Now()
}

// GetPassword returns the password (used by keyring provider).
func (c *PostgreSQLConnection) GetPassword() string {
	return c.Password
}
