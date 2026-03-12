// Package bindings provides Wails bindings for frontend communication.
package bindings

import (
	"fmt"
	"context"
	"log/slog"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// ConnectionBinding provides Wails bindings for connection management.
// All methods are exported (uppercase) for Wails to expose to frontend.
type ConnectionBinding struct {
	uc *usecase.ConnectionUseCase
}

// NewConnectionBinding creates a new ConnectionBinding.
func NewConnectionBinding(uc *usecase.ConnectionUseCase) *ConnectionBinding {
	return &ConnectionBinding{uc: uc}
}

// ConnectionDTO represents a connection for JSON serialization to frontend.
// This is a simplified view without sensitive data like passwords.
type ConnectionDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database,omitempty"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"` // Loaded from keyring for display
	SSLMode     string `json:"ssl_mode,omitempty"`
	SID         string `json:"sid,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	ConnectType string `json:"connect_type,omitempty"`
	// SSH configuration
	SSHEnabled  bool   `json:"ssh_enabled"`
	SSHPort     int    `json:"ssh_port,omitempty"`
	SSHUsername string `json:"ssh_username,omitempty"`
	SSHPassword string `json:"ssh_password,omitempty"` // Loaded from keyring for display
	// WinRM configuration
	WinRMEnabled  bool   `json:"winrm_enabled"`
	WinRMPort     int    `json:"winrm_port,omitempty"`
	WinRMUseHTTPS bool   `json:"winrm_use_https"`
	WinRMUsername string `json:"winrm_username,omitempty"`
	WinRMPassword string `json:"winrm_password,omitempty"` // Loaded from keyring for display
	// SQL Server configuration
	TrustServerCertificate bool `json:"trust_server_certificate"`
}

// ConnectionListResult represents the result of ListConnections.
type ConnectionListResult struct {
	Connections []ConnectionDTO `json:"connections"`
	Error       string          `json:"error,omitempty"`
}

// ConnectionTestResult represents the result of TestConnection.
type ConnectionTestResult struct {
	Success         bool   `json:"success"`
	LatencyMs       int64  `json:"latency_ms"`
	DatabaseVersion string `json:"database_version,omitempty"`
	Error           string `json:"error,omitempty"`
}

// ConnectionCreateRequest represents a request to create a connection.
type ConnectionCreateRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`     // mysql, postgresql, oracle, sqlserver
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
	// Oracle specific fields
	SID         string `json:"sid,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	ConnectType string `json:"connect_type,omitempty"` // "sid" or "service_name"
	ConnectAs   string `json:"connect_as,omitempty"`   // "normal", "sysdba", "sysoper"
	// SSH Configuration
	SSHEnabled  bool   `json:"ssh_enabled"`
	SSHPort     int    `json:"ssh_port,omitempty"`
	SSHUsername string `json:"ssh_username,omitempty"`
	SSHPassword string `json:"ssh_password,omitempty"`
	// WinRM Configuration
	WinRMEnabled  bool   `json:"winrm_enabled"`
	WinRMPort     int    `json:"winrm_port,omitempty"`
	WinRMUseHTTPS bool   `json:"winrm_use_https"`
	WinRMUsername string `json:"winrm_username,omitempty"`
	WinRMPassword string `json:"winrm_password,omitempty"`
}

// ConnectionUpdateRequest represents a request to update a connection.
type ConnectionUpdateRequest struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"` // Optional, empty means keep existing
	SSLMode  string `json:"ssl_mode"`
	// Oracle specific fields
	SID         string `json:"sid,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	ConnectType string `json:"connect_type,omitempty"` // "sid" or "service_name"
	ConnectAs   string `json:"connect_as,omitempty"`   // "normal", "sysdba", "sysoper"
	// SSH Configuration
	SSHEnabled  bool   `json:"ssh_enabled"`
	SSHPort     int    `json:"ssh_port,omitempty"`
	SSHUsername string `json:"ssh_username,omitempty"`
	SSHPassword string `json:"ssh_password,omitempty"`
	// WinRM Configuration
	WinRMEnabled  bool   `json:"winrm_enabled"`
	WinRMPort     int    `json:"winrm_port,omitempty"`
	WinRMUseHTTPS bool   `json:"winrm_use_https"`
	WinRMUsername string `json:"winrm_username,omitempty"`
	WinRMPassword string `json:"winrm_password,omitempty"`
}

// WinRMTestRequest represents a request to test WinRM connection.
type WinRMTestRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	UseHTTPS bool   `json:"use_https"`
}

// ListConnections returns all connections (Wails binding).
func (b *ConnectionBinding) ListConnections() ConnectionListResult {
	ctx := context.Background()
	conns, err := b.uc.ListConnections(ctx)
	if err != nil {
		slog.Error("ListConnections failed", "error", err)
		return ConnectionListResult{
			Error: err.Error(),
		}
	}

	dtos := make([]ConnectionDTO, 0, len(conns))
	for _, conn := range conns {
		dto := b.toDTO(conn)
		dtos = append(dtos, dto)
	}

	return ConnectionListResult{
		Connections: dtos,
	}
}

// GetConnection returns a single connection by ID (Wails binding).
func (b *ConnectionBinding) GetConnection(id string) *ConnectionDTO {
	ctx := context.Background()
	conn, err := b.uc.GetConnectionByID(ctx, id)
	if err != nil {
		slog.Error("GetConnection failed", "id", id, "error", err)
		return nil
	}
	if conn == nil {
		return nil
	}

	dto := b.toDTO(conn)
	return &dto
}

// CreateConnection creates a new connection (Wails binding).
func (b *ConnectionBinding) CreateConnection(req ConnectionCreateRequest) *ConnectionDTO {
	ctx := context.Background()

	var conn connection.Connection
	switch req.Type {
	case "mysql":
		conn = usecase.NewMySQLConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if mysqlConn, ok := conn.(*connection.MySQLConnection); ok {
			mysqlConn.SSLMode = req.SSLMode
			mysqlConn.SetPassword(req.Password)
			// SSH configuration
			if req.SSHEnabled {
				mysqlConn.SSH = &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.SSHPort,
					Username: req.SSHUsername,
					Password: req.SSHPassword,
				}
			}
		}
	case "postgresql":
		conn = usecase.NewPostgreSQLConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if pgConn, ok := conn.(*connection.PostgreSQLConnection); ok {
			pgConn.SSLMode = req.SSLMode
			pgConn.SetPassword(req.Password)
			// SSH configuration
			if req.SSHEnabled {
				pgConn.SSH = &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.SSHPort,
					Username: req.SSHUsername,
					Password: req.SSHPassword,
				}
			}
		}
	case "oracle":
		// Determine SID/ServiceName based on connect_type
		sid := ""
		serviceName := ""
		if req.ConnectType == "sid" {
			sid = req.Database
		} else if req.ConnectType == "service_name" {
			serviceName = req.Database
		} else if req.SID != "" {
			sid = req.SID
		} else if req.ServiceName != "" {
			serviceName = req.ServiceName
		} else {
			// Default: treat database field as SID for backward compatibility
			sid = req.Database
		}
		conn = usecase.NewOracleConnection(req.Name, req.Host, serviceName, sid, req.Username, req.Port)
		if oraConn, ok := conn.(*connection.OracleConnection); ok {
			oraConn.SetPassword(req.Password)
			// SSH configuration
			if req.SSHEnabled {
				oraConn.SSH = &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.SSHPort,
					Username: req.SSHUsername,
					Password: req.SSHPassword,
				}
			}
		}
	case "sqlserver":
		conn = usecase.NewSQLServerConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if sqlConn, ok := conn.(*connection.SQLServerConnection); ok {
			sqlConn.SetPassword(req.Password)
			// SSH configuration
			if req.SSHEnabled {
				sqlConn.SSH = &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.SSHPort,
					Username: req.SSHUsername,
					Password: req.SSHPassword,
				}
			}
			// WinRM configuration
			if req.WinRMEnabled {
				sqlConn.WinRM = &connection.WinRMConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.WinRMPort,
					Username: req.WinRMUsername,
					Password: req.WinRMPassword,
					UseHTTPS: req.WinRMUseHTTPS,
				}
			}
		}
	default:
		slog.Error("Unknown connection type", "type", req.Type)
		return nil
	}

	if err := b.uc.CreateConnection(ctx, conn); err != nil {
		slog.Error("CreateConnection failed", "error", err)
		return nil
	}

	dto := b.toDTO(conn)
	return &dto
}

// UpdateConnection updates an existing connection (Wails binding).
func (b *ConnectionBinding) UpdateConnection(req ConnectionUpdateRequest) *ConnectionDTO {
	ctx := context.Background()

	// Get existing connection first
	existing, err := b.uc.GetConnectionByID(ctx, req.ID)
	if err != nil {
		slog.Error("UpdateConnection: failed to get existing connection", "id", req.ID, "error", err)
		return nil
	}

	// Update fields based on type
	switch conn := existing.(type) {
	case *connection.MySQLConnection:
		conn.SetName(req.Name)
		conn.Host = req.Host
		conn.Port = req.Port
		conn.Database = req.Database
		conn.Username = req.Username
		if req.SSLMode != "" {
			conn.SSLMode = req.SSLMode
		}
		if req.Password != "" {
			conn.SetPassword(req.Password)
		}
		// SSH configuration
		if req.SSHEnabled {
			conn.SSH = &connection.SSHTunnelConfig{
				Enabled:  true,
				Host:     req.Host,
				Port:     req.SSHPort,
				Username: req.SSHUsername,
				Password: req.SSHPassword,
			}
		} else {
			conn.SSH = nil
		}
	case *connection.PostgreSQLConnection:
		conn.SetName(req.Name)
		conn.Host = req.Host
		conn.Port = req.Port
		conn.Database = req.Database
		conn.Username = req.Username
		if req.SSLMode != "" {
			conn.SSLMode = req.SSLMode
		}
		if req.Password != "" {
			conn.SetPassword(req.Password)
		}
		// SSH configuration
		if req.SSHEnabled {
			conn.SSH = &connection.SSHTunnelConfig{
				Enabled:  true,
				Host:     req.Host,
				Port:     req.SSHPort,
				Username: req.SSHUsername,
				Password: req.SSHPassword,
			}
		} else {
			conn.SSH = nil
		}
	case *connection.OracleConnection:
		conn.SetName(req.Name)
		conn.Host = req.Host
		conn.Port = req.Port
		conn.Username = req.Username
		// Update SID/ServiceName based on connect_type
		if req.ConnectType == "sid" {
			conn.SID = req.Database
			conn.ServiceName = ""
		} else if req.ConnectType == "service_name" {
			conn.ServiceName = req.Database
			conn.SID = ""
		} else if req.SID != "" {
			conn.SID = req.SID
		} else if req.ServiceName != "" {
			conn.ServiceName = req.ServiceName
		}
		if req.Password != "" {
			conn.SetPassword(req.Password)
		}
		// SSH configuration
		if req.SSHEnabled {
			conn.SSH = &connection.SSHTunnelConfig{
				Enabled:  true,
				Host:     req.Host,
				Port:     req.SSHPort,
				Username: req.SSHUsername,
				Password: req.SSHPassword,
			}
		} else {
			conn.SSH = nil
		}
	case *connection.SQLServerConnection:
		conn.SetName(req.Name)
		conn.Host = req.Host
		conn.Port = req.Port
		conn.Database = req.Database
		conn.Username = req.Username
		if req.Password != "" {
			conn.SetPassword(req.Password)
		}
		// SSH configuration
		if req.SSHEnabled {
			conn.SSH = &connection.SSHTunnelConfig{
				Enabled:  true,
				Host:     req.Host,
				Port:     req.SSHPort,
				Username: req.SSHUsername,
				Password: req.SSHPassword,
			}
		} else {
			conn.SSH = nil
		}
		// WinRM configuration
		if req.WinRMEnabled {
			conn.WinRM = &connection.WinRMConfig{
				Enabled:  true,
				Host:     req.Host,
				Port:     req.WinRMPort,
				Username: req.WinRMUsername,
				Password: req.WinRMPassword,
				UseHTTPS: req.WinRMUseHTTPS,
			}
		} else {
			conn.WinRM = nil
		}
	}

	if err := b.uc.UpdateConnection(ctx, existing); err != nil {
		slog.Error("UpdateConnection failed", "error", err)
		return nil
	}

	dto := b.toDTO(existing)
	return &dto
}

// DeleteConnection deletes a connection by ID (Wails binding).
func (b *ConnectionBinding) DeleteConnection(id string) bool {
	ctx := context.Background()
	if err := b.uc.DeleteConnection(ctx, id); err != nil {
		slog.Error("DeleteConnection failed", "id", id, "error", err)
		return false
	}
	return true
}

// TestConnection tests a connection by ID (Wails binding).
func (b *ConnectionBinding) TestConnection(id string) ConnectionTestResult {
	ctx := context.Background()
	result, err := b.uc.TestConnection(ctx, id)
	if err != nil {
		slog.Error("TestConnection failed", "id", id, "error", err)
		return ConnectionTestResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return ConnectionTestResult{
		Success:         result.Success,
		LatencyMs:       result.LatencyMs,
		DatabaseVersion: result.DatabaseVersion,
		Error:           result.Error,
	}
}

// TestConnectionDirect tests a connection directly from request data (Wails binding).
// This allows testing without saving to database first.
func (b *ConnectionBinding) TestConnectionDirect(req ConnectionCreateRequest) ConnectionTestResult {
	ctx := context.Background()

	var conn connection.Connection
	switch req.Type {
	case "mysql":
		conn = usecase.NewMySQLConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if mysqlConn, ok := conn.(*connection.MySQLConnection); ok {
			mysqlConn.SSLMode = req.SSLMode
			mysqlConn.SetPassword(req.Password)
			// SSH configuration
			if req.SSHEnabled {
				mysqlConn.SSH = &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.SSHPort,
					Username: req.SSHUsername,
					Password: req.SSHPassword,
				}
			}
		}
	case "postgresql":
		conn = usecase.NewPostgreSQLConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if pgConn, ok := conn.(*connection.PostgreSQLConnection); ok {
			pgConn.SSLMode = req.SSLMode
			pgConn.SetPassword(req.Password)
			// SSH configuration
			if req.SSHEnabled {
				pgConn.SSH = &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.SSHPort,
					Username: req.SSHUsername,
					Password: req.SSHPassword,
				}
			}
		}
	case "oracle":
		// Determine SID/ServiceName based on connect_type
		sid := ""
		serviceName := ""
		if req.ConnectType == "sid" {
			sid = req.Database
		} else if req.ConnectType == "service_name" {
			serviceName = req.Database
		} else if req.SID != "" {
			sid = req.SID
		} else if req.ServiceName != "" {
			serviceName = req.ServiceName
		} else {
			// Default: treat database field as SID for backward compatibility
			sid = req.Database
		}
		conn = usecase.NewOracleConnection(req.Name, req.Host, serviceName, sid, req.Username, req.Port)
		if oraConn, ok := conn.(*connection.OracleConnection); ok {
			oraConn.SetPassword(req.Password)
			// SSH configuration
			if req.SSHEnabled {
				oraConn.SSH = &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.SSHPort,
					Username: req.SSHUsername,
					Password: req.SSHPassword,
				}
			}
		}
	case "sqlserver":
		conn = usecase.NewSQLServerConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if sqlConn, ok := conn.(*connection.SQLServerConnection); ok {
			sqlConn.SetPassword(req.Password)
			// SSH configuration
			if req.SSHEnabled {
				sqlConn.SSH = &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.SSHPort,
					Username: req.SSHUsername,
					Password: req.SSHPassword,
				}
			}
			// WinRM configuration
			if req.WinRMEnabled {
				sqlConn.WinRM = &connection.WinRMConfig{
					Enabled:  true,
					Host:     req.Host,
					Port:     req.WinRMPort,
					Username: req.WinRMUsername,
					Password: req.WinRMPassword,
					UseHTTPS: req.WinRMUseHTTPS,
				}
			}
		}
	default:
		return ConnectionTestResult{
			Success: false,
			Error:   "Unknown connection type: " + req.Type,
		}
	}

	// Test the connection
	result, err := conn.Test(ctx)
	if err != nil {
		slog.Error("TestConnectionDirect failed", "type", req.Type, "host", req.Host, "error", err)
		return ConnectionTestResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return ConnectionTestResult{
		Success:         result.Success,
		LatencyMs:       result.LatencyMs,
		DatabaseVersion: result.DatabaseVersion,
		Error:           result.Error,
	}
}

// toDTO converts a Connection to ConnectionDTO.
func (b *ConnectionBinding) toDTO(conn connection.Connection) ConnectionDTO {
	dto := ConnectionDTO{
		ID:   conn.GetID(),
		Name: conn.GetName(),
		Type: string(conn.GetType()),
	}

	// Load password from keyring for display
	if pwd, err := b.uc.GetPassword(context.Background(), conn.GetID()); err == nil {
		dto.Password = pwd
	}

	// Type-specific fields
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		dto.Host = c.Host
		dto.Port = c.Port
		dto.Database = c.Database
		dto.Username = c.Username
		dto.SSLMode = c.SSLMode
		// SSH configuration
		if c.SSH != nil {
			dto.SSHEnabled = c.SSH.Enabled
			dto.SSHPort = c.SSH.Port
			dto.SSHUsername = c.SSH.Username
			// Load SSH password from keyring
			if sshPwd, err := b.uc.GetSSHPassword(context.Background(), conn.GetID()); err == nil {
				dto.SSHPassword = sshPwd
			}
		}
	case *connection.PostgreSQLConnection:
		dto.Host = c.Host
		dto.Port = c.Port
		dto.Database = c.Database
		dto.Username = c.Username
		dto.SSLMode = c.SSLMode
		// SSH configuration
		if c.SSH != nil {
			dto.SSHEnabled = c.SSH.Enabled
			dto.SSHPort = c.SSH.Port
			dto.SSHUsername = c.SSH.Username
			// Load SSH password from keyring
			if sshPwd, err := b.uc.GetSSHPassword(context.Background(), conn.GetID()); err == nil {
				dto.SSHPassword = sshPwd
			}
		}
	case *connection.OracleConnection:
		dto.Host = c.Host
		dto.Port = c.Port
		dto.Database = c.ServiceName
		if dto.Database == "" {
			dto.Database = c.SID
		}
		dto.SID = c.SID
		dto.ServiceName = c.ServiceName
		// Determine connect_type based on which field is populated
		if c.SID != "" && c.ServiceName == "" {
			dto.ConnectType = "sid"
		} else {
			dto.ConnectType = "service_name"
		}
		dto.Username = c.Username
		// SSH configuration
		if c.SSH != nil {
			dto.SSHEnabled = c.SSH.Enabled
			dto.SSHPort = c.SSH.Port
			dto.SSHUsername = c.SSH.Username
			// Load SSH password from keyring
			if sshPwd, err := b.uc.GetSSHPassword(context.Background(), conn.GetID()); err == nil {
				dto.SSHPassword = sshPwd
			}
		}
	case *connection.SQLServerConnection:
		dto.Host = c.Host
		dto.Port = c.Port
		dto.Database = c.Database
		dto.Username = c.Username
		dto.TrustServerCertificate = c.TrustServerCertificate
		// SSH configuration
		if c.SSH != nil {
			dto.SSHEnabled = c.SSH.Enabled
			dto.SSHPort = c.SSH.Port
			dto.SSHUsername = c.SSH.Username
			// Load SSH password from keyring
			if sshPwd, err := b.uc.GetSSHPassword(context.Background(), conn.GetID()); err == nil {
				dto.SSHPassword = sshPwd
			}
		}
		// WinRM configuration
		if c.WinRM != nil {
			dto.WinRMEnabled = c.WinRM.Enabled
			dto.WinRMPort = c.WinRM.Port
			dto.WinRMUseHTTPS = c.WinRM.UseHTTPS
			dto.WinRMUsername = c.WinRM.Username
			// Load WinRM password from keyring
			if winrmPwd, err := b.uc.GetWinRMPassword(context.Background(), conn.GetID()); err == nil {
				dto.WinRMPassword = winrmPwd
			}
		}
	}

	return dto
}

// TestWinRMConnection tests a WinRM connection (Wails binding).
func (b *ConnectionBinding) TestWinRMConnection(req WinRMTestRequest) ConnectionTestResult {
	ctx := context.Background()

	// Create WinRM config
	winrmConfig := &connection.WinRMConfig{
		Enabled:  true,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		UseHTTPS: req.UseHTTPS,
	}

	// Create WinRM client and test
	client, err := connection.NewWinRMClient(ctx, winrmConfig)
	if err != nil {
		slog.Error("WinRM: Failed to create client", "error", err)
		return ConnectionTestResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create WinRM client: %v", err),
		}
	}
	defer client.Close()

	// Test the connection
	result, err := client.Test(ctx)
	if err != nil {
		slog.Error("WinRM: Test failed", "error", err)
		return ConnectionTestResult{
			Success: false,
			Error:   fmt.Sprintf("WinRM test failed: %v", err),
		}
	}

	return ConnectionTestResult{
		Success:         result.Success,
		LatencyMs:       result.LatencyMs,
		DatabaseVersion: result.DatabaseVersion,
		Error:           result.Error,
	}
}

// SSHTestRequest represents a request to test SSH connection.
type SSHTestRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// TestSSHConnection tests an SSH connection (Wails binding).
func (b *ConnectionBinding) TestSSHConnection(req SSHTestRequest) ConnectionTestResult {
	ctx := context.Background()

	slog.Info("TestSSHConnection called", "host", req.Host, "port", req.Port, "username", req.Username)

	// Create SSH config
	sshConfig := &connection.SSHTunnelConfig{
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
	}
	sshConfig.Password = req.Password

	success, latencyMs, err := connection.TestSSHConnection(ctx, sshConfig)
	if err != nil {
		slog.Error("TestSSHConnection failed", "error", err)
		return ConnectionTestResult{
			Success: false,
			Error:   fmt.Sprintf("SSH test failed: %v", err),
		}
	}

	slog.Info("TestSSHConnection success", "host", req.Host, "latencyMs", latencyMs)
	return ConnectionTestResult{
		Success:   success,
		LatencyMs: latencyMs,
	}
}
