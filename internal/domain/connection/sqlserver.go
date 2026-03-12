// Package connection provides SQL Server connection implementation.
// Implements: REQ-CONN-002
package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
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

	// SSH tunnel configuration (for secure remote access)
	SSH *SSHTunnelConfig `json:"ssh,omitempty"` // SSH tunnel configuration (optional)

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
// Uses encrypt=disable to avoid TLS certificate issues with SQL Server's default certificate.
// Format: sqlserver://username:password@host:port?database=dbname&encrypt=disable&trustservercertificate=true/false
func (c *SQLServerConnection) GetDSNWithPassword() string {
	trustParam := "false"
	if c.TrustServerCertificate {
		trustParam = "true"
	}
	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable&trustservercertificate=%s",
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

// Test tests the SQL Server connection availability.
//
// This is a DIRECT database connection test (easy connection).
// It does NOT use SSH tunnel - SSH testing is handled separately by TestSSHConnection.
//
// Strategy:
// 1. Try with user's configured TrustServerCertificate setting
// 2. Use encrypt=disable to avoid TLS issues with invalid certificates
//
// Note on encrypt parameter values (go-mssqldb driver):
// - "disable" (EncryptionDisabled=3): Completely disables TLS, no handshake
// - "false" (EncryptionOff=0): TLS optional, driver may still attempt TLS
// - "true"/"mandatory" (EncryptionRequired=1): TLS required
// - "strict" (EncryptionStrict=4): Strict TLS with full validation
//
// IMPORTANT: Go 1.23+ rejects certificates with negative serial numbers at PARSE time,
// not validation time. This means TrustServerCertificate cannot help when the server
// has an invalid certificate. The only solution is to use encrypt=disable or fix the
// server certificate.
//
// Returns: TestResult with success/failure, latency, version, error.
func (c *SQLServerConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	// Direct database connection - always use original host/port
	targetHost := c.Host
	targetPort := c.Port

	// Build DSN with target host/port (direct connection)
	dsn := c.buildDSNWithConfig(targetHost, targetPort, c.TrustServerCertificate)

	slog.Info("SQL Server: Testing direct connection",
		"host", targetHost,
		"port", targetPort,
		"encrypt", "disable",
		"trust_server_certificate", c.TrustServerCertificate,
		"username", c.Username)

	result, err := c.testConnection(ctx, dsn, start)
	if err != nil {
		// Context cancelled or timeout
		return nil, fmt.Errorf("test cancelled: %w", err)
	}

	if result.Success {
		slog.Info("SQL Server: Direct connection successful",
			"latency_ms", result.LatencyMs,
			"version", result.DatabaseVersion)
		return result, nil
	}

	// Check if it's a certificate-related error
	if isCertificateError(result.Error) {
		// Provide helpful error message for negative serial number issue
		enhancedError := fmt.Sprintf(
			"TLS certificate error: %s. "+
				"This is likely caused by SQL Server's default self-signed certificate which has an invalid serial number (rejected by Go 1.23+). "+
				"Solutions: (1) Install a proper TLS certificate on SQL Server, or (2) ensure connection is tested without TLS encryption.",
			result.Error)
		latency := time.Since(start).Milliseconds()
		return &TestResult{
			Success:   false,
			LatencyMs: latency,
			Error:     enhancedError,
		}, nil
	}

	// Return the original error
	latency := time.Since(start).Milliseconds()
	return &TestResult{
		Success:   false,
		LatencyMs: latency,
		Error:     result.Error,
	}, nil
}

// isCertificateError checks if the error is related to TLS certificate issues.
func isCertificateError(errMsg string) bool {
	// Check for common certificate-related error patterns
	certErrorPatterns := []string{
		"x509:",
		"certificate",
		"TLS Handshake failed",
		"tls:",
	}
	for _, pattern := range certErrorPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}
	return false
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

// buildDSNWithConfig builds a DSN with the specified host, port, and trust settings.
// Format: sqlserver://username:password@host:port?database=xxx&encrypt=xxx&trustservercertificate=xxx
//
// IMPORTANT: We use encrypt=disable (not encrypt=false) to completely disable TLS.
// This is necessary because:
// 1. encrypt=false (EncryptionOff) still allows TLS negotiation
// 2. SQL Server's default self-signed certificate has a negative serial number
// 3. Go 1.23+ rejects such certificates at PARSE time, before TrustServerCertificate can help
// 4. Using encrypt=disable prevents any TLS handshake, avoiding the certificate parsing issue
func (c *SQLServerConnection) buildDSNWithConfig(host string, port int, trustServerCert bool) string {
	// Always use encrypt=disable to avoid TLS certificate parsing issues
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable&trustservercertificate=%t",
		c.Username, c.Password, host, port, c.Database, trustServerCert)
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

// GetSSHConfig returns the SSH tunnel configuration.
func (c *SQLServerConnection) GetSSHConfig() *SSHTunnelConfig {
	return c.SSH
}

// SetSSHConfig sets the SSH tunnel configuration.
func (c *SQLServerConnection) SetSSHConfig(config *SSHTunnelConfig) {
	c.SSH = config
	c.UpdatedAt = time.Now()
}
