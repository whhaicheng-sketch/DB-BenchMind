// Package connection provides PostgreSQL connection implementation.
// Implements: REQ-CONN-002
package connection

import (
	"context"
	"fmt"
	"time"
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
// Format: host=host port=port database=database user=username
func (c *PostgreSQLConnection) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d database=%s user=%s", c.Host, c.Port, c.Database, c.Username)
}

// GetDSNWithPassword generates a complete connection string with password.
// Format: host=host port=port database=database user=username password=password sslmode=ssl_mode
func (c *PostgreSQLConnection) GetDSNWithPassword() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}
	return fmt.Sprintf("host=%s port=%d database=%s user=%s password=%s sslmode=%s",
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

	// Validate SSL mode
	validSSLMode := map[string]bool{
		"disable":      true,
		"allow":        true,
		"prefer":       true,
		"require":      true,
		"verify-ca":    true,
		"verify-full":  true,
		"":             true, // empty is allowed (will use default)
	}
	if c.SSLMode != "" && !validSSLMode[c.SSLMode] {
		errs = append(errs, &ValidationError{
			Field:   "ssl_mode",
			Message: "ssl_mode must be one of: disable, allow, prefer, require, verify-ca, verify-full",
			Value:   c.SSLMode,
		})
	}

	if len(errs) > 0 {
		return &MultiValidationError{Errors: errs}
	}

	return nil
}

// Test tests the PostgreSQL connection availability (REQ-CONN-003, REQ-CONN-004, REQ-CONN-005).
//
// Note: PostgreSQL driver is not imported yet.
// When ready, need to import: _ "github.com/lib/pq"
func (c *PostgreSQLConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	// Placeholder - PostgreSQL driver not yet imported
	// When ready, uncomment and implement:
	// dsn := c.GetDSNWithPassword()
	// db, err := sql.Open("postgres", dsn)
	// ...

	latency := time.Since(start).Milliseconds()

	return &TestResult{
		Success:   false,
		LatencyMs: latency,
		Error:     "PostgreSQL driver not available - requires github.com/lib/pq",
	}, nil
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
