// Package connection provides SQL Server connection implementation.
// Implements: REQ-CONN-002
package connection

import (
	"context"
	"database/sql"
	"fmt"
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

// Test tests the SQL Server connection availability (REQ-CONN-003, REQ-CONN-004, REQ-CONN-005).
//
// It attempts to connect to the SQL Server database and returns:
// - Success: Whether the connection succeeded
// - LatencyMs: Time taken to establish connection
// - DatabaseVersion: SQL Server version string (on success)
// - Error: Error message (on failure)
//
// The context supports cancellation and timeout.
func (c *SQLServerConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	dsn := c.GetDSNWithPassword()

	// Try to open connection
	db, err := sql.Open("sqlserver", dsn)
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
		errorMsg := fmt.Sprintf("connection failed: %v", err)

		// Provide helpful hints for common errors
		if !c.TrustServerCertificate {
			errorMsg += "\n\n💡 Hint: Try enabling 'Trust Server Certificate' if using self-signed certificates"
		}

		return &TestResult{
			Success:   false,
			LatencyMs: latency,
			Error:     errorMsg,
		}, nil
	}

	// Get database version (REQ-CONN-004)
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

// SetPassword sets the password (used by keyring provider).
func (c *SQLServerConnection) SetPassword(password string) {
	c.Password = password
	c.UpdatedAt = time.Now()
}

// GetPassword returns the password (used by keyring provider).
func (c *SQLServerConnection) GetPassword() string {
	return c.Password
}
