// Implements: REQ-CONN-010
package connection

import (
	"context"
	"testing"
	"time"
)

// TestMySQLConnection_Validate tests MySQL connection validation with table-driven approach.
func TestMySQLConnection_Validate(t *testing.T) {
	tests := []struct {
		name    string
		conn    *MySQLConnection
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid connection",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "test-conn",
				},
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "root",
				SSLMode:  "preferred",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "",
				},
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "root",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing host",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "test-conn",
				},
				Port:     3306,
				Database: "testdb",
				Username: "root",
			},
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name: "port too low",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "test-conn",
				},
				Host:     "localhost",
				Port:     0,
				Database: "testdb",
				Username: "root",
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "port too high",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "test-conn",
				},
				Host:     "localhost",
				Port:     99999,
				Database: "testdb",
				Username: "root",
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "missing database",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "test-conn",
				},
				Host:     "localhost",
				Port:     3306,
				Username: "root",
			},
			wantErr: true,
			errMsg:  "database is required",
		},
		{
			name: "missing username",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "test-conn",
				},
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "username is required",
		},
		{
			name: "invalid ssl_mode",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "test-conn",
				},
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "root",
				SSLMode:  "invalid",
			},
			wantErr: true,
			errMsg:  "ssl_mode must be one of",
		},
		{
			name: "multiple errors",
			conn: &MySQLConnection{
				BaseConnection: BaseConnection{
					Name: "",
				},
				Port: -1,
			},
			wantErr: true,
			errMsg:  "name is required", // Should contain at least this error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conn.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errMsg)
					return
				}
				// Check if error message contains expected substring
				// Note: For MultiValidationError, we check if any error matches
				errStr := err.Error()
				found := false
				if len(errStr) >= len(tt.errMsg) {
					// Simple substring check
					for i := 0; i <= len(errStr)-len(tt.errMsg); i++ {
						if errStr[i:i+len(tt.errMsg)] == tt.errMsg {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// TestMySQLConnection_GetType tests GetType method.
func TestMySQLConnection_GetType(t *testing.T) {
	conn := &MySQLConnection{
		BaseConnection: BaseConnection{
			Name: "test",
		},
	}

	if got := conn.GetType(); got != DatabaseTypeMySQL {
		t.Errorf("GetType() = %v, want %v", got, DatabaseTypeMySQL)
	}
}

// TestMySQLConnection_GetDSN tests GetDSN method (password excluded).
func TestMySQLConnection_GetDSN(t *testing.T) {
	conn := &MySQLConnection{
		BaseConnection: BaseConnection{
			Name: "test-conn",
		},
		Username: "testuser",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
	}

	want := "testuser@tcp(localhost:3306)/testdb"
	if got := conn.GetDSN(); got != want {
		t.Errorf("GetDSN() = %q, want %q", got, want)
	}
}

// TestMySQLConnection_GetDSNWithPassword tests GetDSNWithPassword method.
func TestMySQLConnection_GetDSNWithPassword(t *testing.T) {
	conn := &MySQLConnection{
		BaseConnection: BaseConnection{
			Name: "test-conn",
		},
		Username: "testuser",
		Password: "secret",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
	}

	want := "testuser:secret@tcp(localhost:3306)/testdb"
	if got := conn.GetDSNWithPassword(); got != want {
		t.Errorf("GetDSNWithPassword() = %q, want %q", got, want)
	}
}

// TestMySQLConnection_Redact tests Redact method (REQ-CONN-008).
func TestMySQLConnection_Redact(t *testing.T) {
	conn := &MySQLConnection{
		BaseConnection: BaseConnection{
			Name: "Production DB",
		},
		Host:     "prod.example.com",
		Port:     3306,
		Database: "production",
	}

	want := "Production DB (***@prod.example.com:3306/production)"
	if got := conn.Redact(); got != want {
		t.Errorf("Redact() = %q, want %q", got, want)
	}
}

// TestMySQLConnection_Test tests the Test method (REQ-CONN-003).
// Note: This test doesn't require actual MySQL connection.
func TestMySQLConnection_Test_NoDriver(t *testing.T) {
	conn := &MySQLConnection{
		BaseConnection: BaseConnection{
			Name: "test-conn",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "password",
	}

	result, err := conn.Test(context.Background())
	if err != nil {
		t.Fatalf("Test() unexpected error = %v", err)
	}

	if result.Success {
		t.Error("Test() result.Success = true, want false (no driver)")
	}

	if result.Error == "" {
		t.Error("Test() result.Error = empty, want error message")
	}
}

// TestValidatePort tests the ValidatePort helper function.
func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid port 1", 1, false},
		{"valid port 80", 80, false},
		{"valid port 3306", 3306, false},
		{"valid port 65535", 65535, false},
		{"invalid port 0", 0, true},
		{"invalid port -1", -1, true},
		{"invalid port 65536", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

// TestBaseConnection_SetName tests SetName method updates UpdatedAt.
func TestBaseConnection_SetName(t *testing.T) {
	base := BaseConnection{
		Name:      "old-name",
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Sleep to ensure time difference
	time.Sleep(10 * time.Millisecond)

	base.SetName("new-name")

	if base.Name != "new-name" {
		t.Errorf("SetName() Name = %q, want %q", base.Name, "new-name")
	}

	if !base.UpdatedAt.After(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("SetName() did not update UpdatedAt")
	}
}
