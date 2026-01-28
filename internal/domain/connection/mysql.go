// Package connection provides MySQL connection implementation.
// Implements: REQ-CONN-002, REQ-CONN-003, REQ-CONN-010
package connection

import (
	"context"
	"database/sql"
	"fmt"
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
}

// GetType returns DatabaseTypeMySQL.
func (c *MySQLConnection) GetType() DatabaseType {
	return DatabaseTypeMySQL
}

// GetDSN generates a connection string without password (for logging).
// Format: username@tcp(host:port)/database
// If database is empty, returns: username@tcp(host:port)
func (c *MySQLConnection) GetDSN() string {
	if c.Database == "" {
		return fmt.Sprintf("%s@tcp(%s:%d)", c.Username, c.Host, c.Port)
	}
	return fmt.Sprintf("%s@tcp(%s:%d)/%s", c.Username, c.Host, c.Port, c.Database)
}

// GetDSNWithPassword generates a complete connection string with password.
// Format: username:password@tcp(host:port)/database
// If database is empty, returns: username:password@tcp(host:port)
func (c *MySQLConnection) GetDSNWithPassword() string {
	if c.Database == "" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)", c.Username, c.Password, c.Host, c.Port)
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

	// Validate SSL mode
	if c.SSLMode != "" && c.SSLMode != "disabled" && c.SSLMode != "preferred" && c.SSLMode != "required" {
		errs = append(errs, &ValidationError{
			Field:   "ssl_mode",
			Message: "ssl_mode must be one of: disabled, preferred, required",
			Value:   c.SSLMode,
		})
	}

	if len(errs) > 0 {
		return &MultiValidationError{Errors: errs}
	}

	return nil
}

// Test tests the MySQL connection availability (REQ-CONN-003, REQ-CONN-004, REQ-CONN-005).
//
// It attempts to connect to the MySQL database and returns:
// - Success: Whether the connection succeeded
// - LatencyMs: Time taken to establish connection
// - DatabaseVersion: MySQL version string (on success)
// - Error: Error message (on failure)
//
// The context supports cancellation and timeout.
// Note: This method requires the MySQL driver to be imported.
// Currently returns an error stating driver not available.
func (c *MySQLConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	dsn := c.GetDSNWithPassword()

	// Try to open connection
	// Note: MySQL driver is not imported yet to avoid dependency
	// When ready, uncomment: _ "github.com/go-sql-driver/mysql"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return &TestResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to open connection: %v", err),
			LatencyMs: 0,
		}, nil
	}
	defer db.Close()

	// Set timeout for connection test
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Attempt to ping the database
	err = db.PingContext(testCtx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &TestResult{
			Success:   false,
			LatencyMs: latency,
			Error:     fmt.Sprintf("connection failed: %v", err),
		}, nil
	}

	// Get database version (REQ-CONN-004)
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
