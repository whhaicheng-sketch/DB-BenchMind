package connection

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestPostgreSQLConnection_Validate_ValidInput tests validation with valid input
func TestPostgreSQLConnection_Validate_ValidInput(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Test PG",
		},
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		SSLMode:  "prefer",
	}

	err := conn.Validate()
	if err != nil {
		t.Errorf("Validate() should succeed with valid input, got error: %v", err)
	}
}

// TestPostgreSQLConnection_Validate_MissingRequiredFields tests validation with missing required fields
func TestPostgreSQLConnection_Validate_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		conn    *PostgreSQLConnection
		wantErr bool
	}{
		{
			name: "Missing Name",
			conn: &PostgreSQLConnection{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
			},
			wantErr: true,
		},
		{
			name: "Missing Host",
			conn: &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "Test"},
				Port:           5432,
				Username:       "postgres",
			},
			wantErr: true,
		},
		{
			name: "Missing Username",
			conn: &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "Test"},
				Host:           "localhost",
				Port:           5432,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conn.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPostgreSQLConnection_Validate_InvalidPort tests validation with invalid port
func TestPostgreSQLConnection_Validate_InvalidPort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"Port zero", 0, true},
		{"Port negative", -1, true},
		{"Port too large", 65536, true},
		{"Valid port", 5432, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "Test"},
				Host:           "localhost",
				Port:           tt.port,
				Username:       "postgres",
			}
			err := conn.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPostgreSQLConnection_Validate_InvalidSSLMode tests validation with invalid SSL mode
func TestPostgreSQLConnection_Validate_InvalidSSLMode(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
		Host:           "localhost",
		Port:           5432,
		Username:       "postgres",
		SSLMode:        "invalid-mode",
	}

	err := conn.Validate()
	if err == nil {
		t.Error("Validate() should fail with invalid SSL mode")
	}
}

// TestPostgreSQLConnection_Validate_ValidSSLModes tests validation with all valid SSL modes
func TestPostgreSQLConnection_Validate_ValidSSLModes(t *testing.T) {
	validModes := []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full", ""}

	for _, mode := range validModes {
		t.Run("SSLMode_"+mode, func(t *testing.T) {
			conn := &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "Test"},
				Host:           "localhost",
				Port:           5432,
				Username:       "postgres",
				SSLMode:        mode,
			}
			err := conn.Validate()
			if err != nil {
				t.Errorf("Validate() should succeed with SSLMode=%q, got error: %v", mode, err)
			}
		})
	}
}

// TestPostgreSQLConnection_GetDSN tests DSN generation
func TestPostgreSQLConnection_GetDSN(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
		Host:           "localhost",
		Port:           5432,
		Database:       "testdb",
		Username:       "postgres",
	}

	expected := "host=localhost port=5432 database=testdb user=postgres"
	got := conn.GetDSN()

	if got != expected {
		t.Errorf("GetDSN() = %q, want %q", got, expected)
	}
}

// TestPostgreSQLConnection_GetDSNWithPassword tests DSN generation with password
func TestPostgreSQLConnection_GetDSNWithPassword(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
		Host:           "localhost",
		Port:           5432,
		Database:       "testdb",
		Username:       "postgres",
		Password:       "secret",
		SSLMode:        "require",
	}

	expected := "host=localhost port=5432 database=testdb user=postgres password=secret sslmode=require"
	got := conn.GetDSNWithPassword()

	if got != expected {
		t.Errorf("GetDSNWithPassword() = %q, want %q", got, expected)
	}
}

// TestPostgreSQLConnection_GetDSN_EmptyDatabase tests DSN with empty database
func TestPostgreSQLConnection_GetDSN_EmptyDatabase(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
		Host:           "localhost",
		Port:           5432,
		Database:       "", // Empty database
		Username:       "postgres",
	}

	expected := "host=localhost port=5432 database= user=postgres"
	got := conn.GetDSN()

	if got != expected {
		t.Errorf("GetDSN() = %q, want %q", got, expected)
	}
}

// TestPostgreSQLConnection_Redact tests redaction for display
func TestPostgreSQLConnection_Redact(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Production DB",
		},
		Host:     "prod.example.com",
		Port:     5432,
		Database: "production",
	}

	expected := "Production DB (***@prod.example.com:5432/production)"
	got := conn.Redact()

	if got != expected {
		t.Errorf("Redact() = %q, want %q", got, expected)
	}
}

// TestPostgreSQLConnection_SetPassword_GetPassword tests password management
func TestPostgreSQLConnection_SetPassword_GetPassword(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
	}

	// Test SetPassword
	conn.SetPassword("my-secret-password")
	if conn.Password != "my-secret-password" {
		t.Errorf("SetPassword() failed, Password = %q", conn.Password)
	}

	// Test GetPassword
	got := conn.GetPassword()
	if got != "my-secret-password" {
		t.Errorf("GetPassword() = %q, want %q", got, "my-secret-password")
	}

	// Verify UpdatedAt is set
	if conn.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set after SetPassword")
	}
}

// TestPostgreSQLConnection_GetType tests GetType returns correct database type
func TestPostgreSQLConnection_GetType(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
	}

	if conn.GetType() != DatabaseTypePostgreSQL {
		t.Errorf("GetType() = %q, want %q", conn.GetType(), DatabaseTypePostgreSQL)
	}
}

// TestPostgreSQLConnection_Test_ConnectionFailure tests failed connection test
func TestPostgreSQLConnection_Test_ConnectionFailure(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Test",
		},
		Host:     "invalid-host-that-does-not-exist.local",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "test",
		SSLMode:  "disable",
	}

	ctx := context.Background()
	result, err := conn.Test(ctx)

	if err != nil {
		t.Fatalf("Test() should not return error, got: %v", err)
	}

	if result.Success {
		t.Error("Test() Success = true, want false")
	}

	if result.Error == "" {
		t.Error("Test() Error should not be empty on failure")
	}

	if result.LatencyMs <= 0 {
		t.Errorf("Test() LatencyMs = %d, want > 0", result.LatencyMs)
	}
}

// TestPostgreSQLConnection_Test_ContextCancellation tests connection test with context cancellation
func TestPostgreSQLConnection_Test_ContextCancellation(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Test",
		},
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "test",
	}

	// Create context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := conn.Test(ctx)

	if err != nil {
		t.Fatalf("Test() should not return error, got: %v", err)
	}

	// Should fail due to context cancellation
	if result.Success {
		t.Error("Test() Success = true, want false (context cancelled)")
	}
}

// TestPostgreSQLConnection_Test_Timeout tests connection test with timeout
func TestPostgreSQLConnection_Test_Timeout(t *testing.T) {
	// Use a very short timeout to test timeout behavior
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Test",
		},
		Host:     "slow-postgres.example.com", // Unresponsive host
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "test",
	}

	// Create context with 1ms timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err := conn.Test(ctx)

	if err != nil {
		t.Fatalf("Test() should not return error, got: %v", err)
	}

	// Should fail due to timeout
	if result.Success {
		t.Error("Test() Success = true, want false (timeout)")
	}

	// Error message should indicate timeout or connection failure
	if result.Error == "" {
		t.Error("Test() Error should not be empty on timeout")
	}

	// Check if error mentions timeout or connection failure
	errorLower := strings.ToLower(result.Error)
	isTimeoutOrConnError := strings.Contains(errorLower, "timeout") ||
		strings.Contains(errorLower, "connection") ||
		strings.Contains(errorLower, "context deadline")

	if !isTimeoutOrConnError {
		t.Errorf("Test() Error should mention timeout or connection error, got: %q", result.Error)
	}
}
