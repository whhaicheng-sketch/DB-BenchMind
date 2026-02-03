// Package connection provides Oracle connection implementation.
// Implements: REQ-CONN-002
package connection

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/sijms/go-ora/v2" // Oracle driver
	"database/sql"
)

// OracleConnection represents an Oracle database connection configuration.
// Implements spec.md 3.2.2
type OracleConnection struct {
	// Base fields
	BaseConnection

	// Connection parameters
	Host        string `json:"host"`         // Host address
	Port        int    `json:"port"`         // Port (default 1521)
	ServiceName string `json:"service_name"` // Service name
	SID         string `json:"sid"`          // SID (alternative to ServiceName)
	Username    string `json:"username"`     // Username
	Password    string `json:"-"`            // Password (stored in keyring)
}

// GetType returns DatabaseTypeOracle.
func (c *OracleConnection) GetType() DatabaseType {
	return DatabaseTypeOracle
}

// GetDSN generates a connection string without password (for logging).
// Format: oracle://username@host:port/service_name or oracle://username@host:port/sid
func (c *OracleConnection) GetDSN() string {
	identifier := c.SID
	if c.ServiceName != "" {
		identifier = c.ServiceName
	}
	return fmt.Sprintf("oracle://%s@%s:%d/%s", c.Username, c.Host, c.Port, identifier)
}

// GetDSNWithPassword generates a complete connection string with password.
// Format: oracle://username:password@host:port/service_name or oracle://username:password@host:port/sid
func (c *OracleConnection) GetDSNWithPassword() string {
	identifier := c.SID
	if c.ServiceName != "" {
		identifier = c.ServiceName
	}
	return fmt.Sprintf("oracle://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, identifier)
}

// Redact returns a redacted connection string for display (REQ-CONN-008).
func (c *OracleConnection) Redact() string {
	identifier := c.ServiceName
	if identifier == "" {
		identifier = c.SID
	}
	return fmt.Sprintf("%s (***@%s:%d/%s)", c.Name, c.Host, c.Port, identifier)
}

// ToJSON serializes the connection to JSON (without password).
func (c *OracleConnection) ToJSON() ([]byte, error) {
	return nil, fmt.Errorf("not implemented yet - will use json.Marshal")
}

// Validate validates the connection parameters (REQ-CONN-010).
func (c *OracleConnection) Validate() error {
	var errs []error

	// Validate required fields
	if err := ValidateRequired("name", c.Name); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateRequired("host", c.Host); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateRequired("username", c.Username); err != nil {
		errs = append(errs, err)
	}

	// Validate port
	if err := ValidatePort(c.Port); err != nil {
		errs = append(errs, err)
	}

	// SID is required (ServiceName is not used in our UI)
	if c.SID == "" {
		errs = append(errs, &ValidationError{
			Field:   "sid",
			Message: "SID is required",
			Value:   c.SID,
		})
	}

	// Validate that ServiceName and SID are not both specified (mutually exclusive)
	if c.ServiceName != "" && c.SID != "" {
		errs = append(errs, &ValidationError{
			Field:   "service_name/sid",
			Message: "service_name and sid are mutually exclusive (specify only one)",
		})
	}

	if len(errs) > 0 {
		return &MultiValidationError{Errors: errs}
	}

	return nil
}

// Test tests the Oracle connection availability (REQ-CONN-003, REQ-CONN-004, REQ-CONN-005).
//
// It attempts to connect to the Oracle database and returns:
// - Success: Whether the connection succeeded
// - LatencyMs: Time taken to establish connection
// - Version: Oracle version string (on success)
// - Error: Error message (on failure)
//
// The context supports cancellation and timeout.
func (c *OracleConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	// Log connection parameters for debugging
	identifier := c.SID
	if identifier == "" {
		identifier = c.ServiceName
	}
	slog.Info("Oracle: Testing connection",
		"host", c.Host,
		"port", c.Port,
		"sid", c.SID,
		"service_name", c.ServiceName,
		"identifier", identifier,
		"username", c.Username,
		"password_set", c.Password != "")

	// Build DSN with password
	dsn := c.GetDSNWithPassword()
	slog.Info("Oracle: Generated DSN", "dsn", dsn)

	// Try to open connection
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		slog.Error("Oracle: Failed to open connection", "error", err)
		return &TestResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to open connection: %v", err),
			LatencyMs: time.Since(start).Milliseconds(),
		}, nil
	}
	defer db.Close()

	// Set timeout for connection test (10 seconds)
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

	// Query Oracle version
	var version string
	err = db.QueryRowContext(testCtx, "SELECT * FROM v$version WHERE rownum = 1").Scan(&version)
	if err != nil {
		version = "unknown"
	}

	return &TestResult{
		Success:        true,
		LatencyMs:      latency,
		DatabaseVersion: version,
	}, nil
}

// SetPassword sets the password (used by keyring provider).
func (c *OracleConnection) SetPassword(password string) {
	c.Password = password
	c.UpdatedAt = time.Now()
}

// GetPassword returns the password (used by keyring provider).
func (c *OracleConnection) GetPassword() string {
	return c.Password
}
