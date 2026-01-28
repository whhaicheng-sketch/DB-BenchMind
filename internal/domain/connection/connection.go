// Package connection provides database connection domain models and interfaces.
// Implements: REQ-CONN-001 ~ REQ-CONN-010
package connection

import (
	"context"
	"time"
)

// DatabaseType represents the type of database.
type DatabaseType string

const (
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeOracle     DatabaseType = "oracle"
	DatabaseTypeSQLServer  DatabaseType = "sqlserver"
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
)

// Connection interface defines the contract for all database connections.
// Implements: REQ-CONN-002, REQ-CONN-003
type Connection interface {
	// GetID returns the connection ID.
	GetID() string

	// GetName returns the connection name.
	GetName() string

	// SetName sets the connection name.
	SetName(name string)

	// GetType returns the database type.
	GetType() DatabaseType

	// Validate validates the connection parameters (REQ-CONN-010).
	// Returns an error if any required field is missing or invalid.
	Validate() error

	// Test tests the connection availability (REQ-CONN-003).
	// Returns TestResult containing success/failure, latency, version, and error message.
	// Context supports cancellation and timeout.
	Test(ctx context.Context) (*TestResult, error)

	// GetDSN generates a connection string without password (for logging).
	GetDSN() string

	// GetDSNWithPassword generates a complete connection string with password (for actual connection).
	GetDSNWithPassword() string

	// Redact returns a redacted connection string for display (REQ-CONN-008).
	// Format: "name (***@host:port/db)" or similar.
	Redact() string

	// ToJSON serializes the connection to JSON (without password).
	ToJSON() ([]byte, error)
}

// TestResult represents the result of a connection test.
// Implements: REQ-CONN-004, REQ-CONN-005
type TestResult struct {
	Success         bool   `json:"success"`           // Whether the test succeeded
	LatencyMs       int64  `json:"latency_ms"`         // Connection latency in milliseconds
	DatabaseVersion string `json:"database_version"`   // Database version information
	Error           string `json:"error,omitempty"`    // Error message if failed
}

// ValidatePort validates that a port number is in valid range (1-65535).
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return &ValidationError{
			Field:   "port",
			Message: "port must be between 1 and 65535",
			Value:   port,
		}
	}
	return nil
}

// ValidateRequired validates that a required string field is not empty.
func ValidateRequired(fieldName, value string) error {
	if value == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " is required",
		}
	}
	return nil
}

// ValidationError represents a validation error for a specific field.
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Value != nil {
		return e.Message + ": " + toString(e.Value)
	}
	return e.Message
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return "<?>"
}

// BaseConnection contains common fields for all connection types.
type BaseConnection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetID returns the connection ID.
func (b *BaseConnection) GetID() string {
	return b.ID
}

// GetName returns the connection name.
func (b *BaseConnection) GetName() string {
	return b.Name
}

// SetName sets the connection name.
func (b *BaseConnection) SetName(name string) {
	b.Name = name
	b.UpdatedAt = time.Now()
}
