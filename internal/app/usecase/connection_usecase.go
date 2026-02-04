// Package usecase provides connection management business logic.
// Implements: REQ-CONN-001 ~ REQ-CONN-010
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

// ConnectionUseCase provides connection management business operations.
// Implements: REQ-CONN-001 ~ REQ-CONN-010
type ConnectionUseCase struct {
	repo    ConnectionRepository
	keyring keyring.Provider
}

// NewConnectionUseCase creates a new connection use case.
func NewConnectionUseCase(repo ConnectionRepository, keyring keyring.Provider) *ConnectionUseCase {
	return &ConnectionUseCase{
		repo:    repo,
		keyring: keyring,
	}
}

// GetKeyring returns the keyring provider (used by UI to load SSH passwords).
func (uc *ConnectionUseCase) GetKeyring() keyring.Provider {
	return uc.keyring
}

// =============================================================================
// Connection Operations
// Implements: REQ-CONN-001, REQ-CONN-008
// =============================================================================

// CreateConnection creates a new connection (REQ-CONN-001).
// Returns an error if:
// - Validation fails
// - Name already exists
// - Repository save fails
// - Keyring save fails
func (uc *ConnectionUseCase) CreateConnection(ctx context.Context, conn connection.Connection) error {
	// Validate connection
	if err := conn.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if name already exists
	exists, err := uc.repo.ExistsByName(ctx, conn.GetName(), "")
	if err != nil {
		return fmt.Errorf("check name exists: %w", err)
	}
	if exists {
		return &DuplicateNameError{Name: conn.GetName()}
	}

	// Generate ID if not set
	if conn.GetID() == "" {
		// Note: In a real implementation, you'd set the ID on the connection
		// For now, we'll return an error requiring ID to be set
		return fmt.Errorf("connection ID must be set before creating")
	}

	// Save password to keyring if provided
	if pwd := getPassword(conn); pwd != "" {
		if err := uc.keyring.Set(ctx, conn.GetID(), pwd); err != nil {
			return fmt.Errorf("save password to keyring: %w", err)
		}
	}

	// Save SSH password to keyring if provided
	if sshPwd := getSSHPassword(conn); sshPwd != "" {
		sshKey := conn.GetID() + ":ssh"
		if err := uc.keyring.Set(ctx, sshKey, sshPwd); err != nil {
			// Rollback: remove database password from keyring
			_ = uc.keyring.Delete(ctx, conn.GetID())
			return fmt.Errorf("save SSH password to keyring: %w", err)
		}
	}

	// Save WinRM password to keyring if provided
	if winrmPwd := getWinRMPassword(conn); winrmPwd != "" {
		winrmKey := conn.GetID() + ":winrm"
		if err := uc.keyring.Set(ctx, winrmKey, winrmPwd); err != nil {
			// Rollback: remove database and SSH passwords from keyring
			_ = uc.keyring.Delete(ctx, conn.GetID())
			_ = uc.keyring.Delete(ctx, conn.GetID()+":ssh")
			return fmt.Errorf("save WinRM password to keyring: %w", err)
		}
	}

	// Save connection to repository
	if err := uc.repo.Save(ctx, conn); err != nil {
		// Rollback: remove passwords from keyring
		_ = uc.keyring.Delete(ctx, conn.GetID())
		_ = uc.keyring.Delete(ctx, conn.GetID()+":ssh")
		_ = uc.keyring.Delete(ctx, conn.GetID()+":winrm")
		return fmt.Errorf("save connection: %w", err)
	}

	return nil
}

// UpdateConnection updates an existing connection (REQ-CONN-008).
// Returns an error if:
// - Connection not found
// - Validation fails
// - New name already exists (excluding current connection)
// - Repository update fails
func (uc *ConnectionUseCase) UpdateConnection(ctx context.Context, conn connection.Connection) error {
	// Validate connection
	if err := conn.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if connection exists
	existing, err := uc.repo.FindByID(ctx, conn.GetID())
	if err != nil {
		return fmt.Errorf("connection not found: %w", err)
	}

	// Check if new name already exists (excluding current connection)
	if conn.GetName() != existing.GetName() {
		exists, err := uc.repo.ExistsByName(ctx, conn.GetName(), conn.GetID())
		if err != nil {
			return fmt.Errorf("check name exists: %w", err)
		}
		if exists {
			return &DuplicateNameError{Name: conn.GetName()}
		}
	}

	// Update password in keyring if changed
	if pwd := getPassword(conn); pwd != "" {
		if err := uc.keyring.Set(ctx, conn.GetID(), pwd); err != nil {
			return fmt.Errorf("update password in keyring: %w", err)
		}
	}

	// Update SSH password in keyring if changed
	if sshPwd := getSSHPassword(conn); sshPwd != "" {
		sshKey := conn.GetID() + ":ssh"
		if err := uc.keyring.Set(ctx, sshKey, sshPwd); err != nil {
			return fmt.Errorf("update SSH password in keyring: %w", err)
		}
	}

	// Update WinRM password in keyring if changed
	if winrmPwd := getWinRMPassword(conn); winrmPwd != "" {
		winrmKey := conn.GetID() + ":winrm"
		if err := uc.keyring.Set(ctx, winrmKey, winrmPwd); err != nil {
			return fmt.Errorf("update WinRM password in keyring: %w", err)
		}
	}

	// Save updated connection
	if err := uc.repo.Save(ctx, conn); err != nil {
		return fmt.Errorf("update connection: %w", err)
	}

	return nil
}

// DeleteConnection deletes a connection (REQ-CONN-009).
// Returns an error if connection not found.
// Also removes password from keyring.
func (uc *ConnectionUseCase) DeleteConnection(ctx context.Context, id string) error {
	// Check if connection exists
	_, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from repository
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}

	// Remove password from keyring (best effort, ignore if not found)
	_ = uc.keyring.Delete(ctx, id)
	_ = uc.keyring.Delete(ctx, id+":ssh")
	_ = uc.keyring.Delete(ctx, id+":winrm")

	return nil
}

// ListConnections returns all connections (REQ-CONN-001).
func (uc *ConnectionUseCase) ListConnections(ctx context.Context) ([]connection.Connection, error) {
	return uc.repo.FindAll(ctx)
}

// GetConnectionByID returns a connection by ID.
func (uc *ConnectionUseCase) GetConnectionByID(ctx context.Context, id string) (connection.Connection, error) {
	conn, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load password from keyring and set on connection
	if uc.keyring != nil {
		password, err := uc.keyring.Get(ctx, id)
		if err != nil {
			if !keyring.IsNotFound(err) {
				return nil, fmt.Errorf("get password from keyring: %w", err)
			}
			// Password not in keyring, continue without it
		} else {
			setPassword(conn, password)
		}

		// Load SSH password from keyring and set on connection
		sshKey := id + ":ssh"
		sshPassword, err := uc.keyring.Get(ctx, sshKey)
		if err != nil {
			if !keyring.IsNotFound(err) {
				return nil, fmt.Errorf("get SSH password from keyring: %w", err)
			}
			// SSH password not in keyring, continue without it
		} else {
			setSSHPassword(conn, sshPassword)
		}

		// Load WinRM password from keyring and set on connection
		winrmKey := id + ":winrm"
		winrmPassword, err := uc.keyring.Get(ctx, winrmKey)
		if err != nil {
			if !keyring.IsNotFound(err) {
				return nil, fmt.Errorf("get WinRM password from keyring: %w", err)
			}
			// WinRM password not in keyring, continue without it
		} else {
			setWinRMPassword(conn, winrmPassword)
		}
	}

	return conn, nil
}

// =============================================================================
// Connection Testing
// Implements: REQ-CONN-003, REQ-CONN-004, REQ-CONN-005
// =============================================================================

// TestConnection tests a database connection (REQ-CONN-003).
// Returns TestResult containing success/failure, latency, version, error.
// Implements: REQ-CONN-004 (success shows latency and version)
// Implements: REQ-CONN-005 (failure shows specific error)
func (uc *ConnectionUseCase) TestConnection(ctx context.Context, id string) (*connection.TestResult, error) {
	// Get connection with password
	conn, err := uc.GetConnectionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}

	// Test the connection
	result, err := conn.Test(ctx)
	if err != nil {
		return nil, fmt.Errorf("test connection: %w", err)
	}

	return result, nil
}

// =============================================================================
// Password Management
// Implements: REQ-CONN-006
// =============================================================================

// SavePassword saves a password to keyring (REQ-CONN-006).
func (uc *ConnectionUseCase) SavePassword(ctx context.Context, connID, password string) error {
	return uc.keyring.Set(ctx, connID, password)
}

// GetPassword retrieves a password from keyring.
func (uc *ConnectionUseCase) GetPassword(ctx context.Context, connID string) (string, error) {
	return uc.keyring.Get(ctx, connID)
}

// DeletePassword removes a password from keyring.
func (uc *ConnectionUseCase) DeletePassword(ctx context.Context, connID string) error {
	return uc.keyring.Delete(ctx, connID)
}

// =============================================================================
// Helper Functions
// =============================================================================

// getPassword extracts password from a connection (type-specific).
func getPassword(conn connection.Connection) string {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		return c.Password
	case *connection.OracleConnection:
		return c.Password
	case *connection.SQLServerConnection:
		return c.Password
	case *connection.PostgreSQLConnection:
		return c.Password
	default:
		return ""
	}
}

// setPassword sets password on a connection (type-specific).
func setPassword(conn connection.Connection, password string) {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		c.SetPassword(password)
	case *connection.OracleConnection:
		c.SetPassword(password)
	case *connection.SQLServerConnection:
		c.SetPassword(password)
	case *connection.PostgreSQLConnection:
		c.SetPassword(password)
	}
}

// getSSHPassword gets SSH password from a connection (type-specific).
func getSSHPassword(conn connection.Connection) string {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		if c.SSH != nil {
			return c.SSH.Password
		}
	case *connection.PostgreSQLConnection:
		if c.SSH != nil {
			return c.SSH.Password
		}
	case *connection.OracleConnection:
		if c.SSH != nil {
			return c.SSH.Password
		}
	}
	return ""
}

// setSSHPassword sets SSH password on a connection (type-specific).
func setSSHPassword(conn connection.Connection, password string) {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		if c.SSH != nil {
			c.SSH.Password = password
		}
	case *connection.PostgreSQLConnection:
		if c.SSH != nil {
			c.SSH.Password = password
		}
	case *connection.OracleConnection:
		if c.SSH != nil {
			c.SSH.Password = password
		}
	}
}

// getWinRMPassword gets WinRM password from a connection (type-specific).
func getWinRMPassword(conn connection.Connection) string {
	switch c := conn.(type) {
	case *connection.SQLServerConnection:
		if c.WinRM != nil {
			return c.WinRM.Password
		}
	}
	return ""
}

// setWinRMPassword sets WinRM password on a connection (type-specific).
func setWinRMPassword(conn connection.Connection, password string) {
	switch c := conn.(type) {
	case *connection.SQLServerConnection:
		if c.WinRM != nil {
			c.WinRM.Password = password
		}
	}
}

// =============================================================================
// Error Types
// =============================================================================

// DuplicateNameError is returned when a connection name already exists.
type DuplicateNameError struct {
	Name string
}

func (e *DuplicateNameError) Error() string {
	return fmt.Sprintf("connection name already exists: %s", e.Name)
}

// =============================================================================
// Factory Functions
// =============================================================================

// NewMySQLConnection creates a new MySQL connection with default values.
func NewMySQLConnection(name, host, database, username string, port int) *connection.MySQLConnection {
	return &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:        uuid.New().String(),
			Name:      name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		SSLMode:  "preferred", // Default SSL mode
	}
}

// NewPostgreSQLConnection creates a new PostgreSQL connection with default values.
func NewPostgreSQLConnection(name, host, database, username string, port int) *connection.PostgreSQLConnection {
	return &connection.PostgreSQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:        uuid.New().String(),
			Name:      name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		SSLMode:  "prefer", // Default SSL mode
	}
}

// NewOracleConnection creates a new Oracle connection with default values.
func NewOracleConnection(name, host, serviceName, sid, username string, port int) *connection.OracleConnection {
	return &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:        uuid.New().String(),
			Name:      name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Host:        host,
		Port:        port,
		ServiceName: serviceName,
		SID:         sid,
		Username:    username,
	}
}

// NewSQLServerConnection creates a new SQL Server connection with default values.
func NewSQLServerConnection(name, host, database, username string, port int) *connection.SQLServerConnection {
	return &connection.SQLServerConnection{
		BaseConnection: connection.BaseConnection{
			ID:        uuid.New().String(),
			Name:      name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Host:                   host,
		Port:                   port,
		Database:               database,
		Username:               username,
		TrustServerCertificate: false, // Default: don't trust self-signed certs
	}
}
