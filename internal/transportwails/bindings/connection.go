// Package bindings provides Wails bindings for frontend communication.
package bindings

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/bindings/aiprovider"
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
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database,omitempty"`
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"` // Loaded from keyring for display
	SSLMode     string `json:"ssl_mode,omitempty"`
	SID         string `json:"sid,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	ConnectType string `json:"connect_type,omitempty"`
	ConnectAs   string `json:"connect_as,omitempty"`
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
	// AI Assistant configuration
	AIAssistants []AIAssistantConfig `json:"ai_assistants,omitempty"`
}

// AIAssistantConfig represents AI assistant configuration for a connection.
type AIAssistantConfig struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Provider          string  `json:"provider"`
	APIHost           string  `json:"api_host"`
	APIEndpoint       string  `json:"api_endpoint"`
	APIKey            string  `json:"api_key,omitempty"`
	Model             string  `json:"model"`
	Temperature       float64 `json:"temperature"`
	Description       string  `json:"description"`
	EnterAction       string  `json:"enter_action"`
	CompareWithOthers bool    `json:"compare_with_others"`
	Language          string  `json:"language"`
}

// toDomainAIAssistants converts binding AIAssistantConfig slice to domain AIAssistantConfig slice.
func toDomainAIAssistants(assistants []AIAssistantConfig) []connection.AIAssistantConfig {
	if len(assistants) == 0 {
		return nil
	}
	result := make([]connection.AIAssistantConfig, len(assistants))
	for i, a := range assistants {
		result[i] = connection.AIAssistantConfig{
			ID:                a.ID,
			Name:              a.Name,
			Provider:          a.Provider,
			APIHost:           a.APIHost,
			APIEndpoint:       a.APIEndpoint,
			APIKey:            a.APIKey,
			Model:             a.Model,
			Temperature:       a.Temperature,
			Description:       a.Description,
			EnterAction:       a.EnterAction,
			CompareWithOthers: a.CompareWithOthers,
			Language:          a.Language,
		}
	}
	return result
}

// toBindingAIAssistants converts domain AIAssistantConfig slice to binding AIAssistantConfig slice.
func toBindingAIAssistants(assistants []connection.AIAssistantConfig) []AIAssistantConfig {
	if len(assistants) == 0 {
		return nil
	}
	result := make([]AIAssistantConfig, len(assistants))
	for i, a := range assistants {
		result[i] = AIAssistantConfig{
			ID:                a.ID,
			Name:              a.Name,
			Provider:          a.Provider,
			APIHost:           a.APIHost,
			APIEndpoint:       a.APIEndpoint,
			APIKey:            a.APIKey,
			Model:             a.Model,
			Temperature:       a.Temperature,
			Description:       a.Description,
			EnterAction:       a.EnterAction,
			CompareWithOthers: a.CompareWithOthers,
			Language:          a.Language,
		}
	}
	return result
}

// ConnectionListResult represents the result of ListConnections.
type ConnectionListResult struct {
	Connections []ConnectionDTO `json:"connections"`
	Error       string          `json:"error,omitempty"`
}

// ConnectionCreateResult represents the result of CreateConnection.
type ConnectionCreateResult struct {
	Connection *ConnectionDTO `json:"connection,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// ConnectionUpdateResult represents the result of UpdateConnection.
type ConnectionUpdateResult struct {
	Connection *ConnectionDTO `json:"connection,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// ConnectionDeleteResult represents the result of DeleteConnection.
type ConnectionDeleteResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// SSHWinRMTestResult represents the result of SSH or WinRM connection test.
type SSHWinRMTestResult struct {
	Success   bool   `json:"success"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

// ConnectionTestResult represents the result of TestConnection.
// Includes DB test result and optional SSH/WinRM test results if configured.
type ConnectionTestResult struct {
	Success         bool   `json:"success"`
	LatencyMs       int64  `json:"latency_ms"`
	DatabaseVersion string `json:"database_version,omitempty"`
	Error           string `json:"error,omitempty"`
	// SSH/WinRM test results (only populated when configured and tested)
	SSHResult   *SSHWinRMTestResult `json:"ssh_result,omitempty"`
	WinRMResult *SSHWinRMTestResult `json:"winrm_result,omitempty"`
}

// ConnectionCreateRequest represents a request to create a connection.
type ConnectionCreateRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // mysql, postgresql, oracle, sqlserver
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
	// AI Assistant Configuration
	AIAssistants []AIAssistantConfig `json:"ai_assistants,omitempty"`
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
	// AI Assistant Configuration
	AIAssistants []AIAssistantConfig `json:"ai_assistants,omitempty"`
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
func (b *ConnectionBinding) CreateConnection(req ConnectionCreateRequest) ConnectionCreateResult {
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
			if req.ConnectAs != "" {
				oraConn.ConnectAs = req.ConnectAs
			}
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
		return ConnectionCreateResult{Error: "Unknown connection type: " + req.Type}
	}

	// Set AI assistants (common to all connection types)
	if len(req.AIAssistants) > 0 {
		switch c := conn.(type) {
		case *connection.MySQLConnection:
			c.SetAIAssistants(toDomainAIAssistants(req.AIAssistants))
		case *connection.PostgreSQLConnection:
			c.SetAIAssistants(toDomainAIAssistants(req.AIAssistants))
		case *connection.OracleConnection:
			c.SetAIAssistants(toDomainAIAssistants(req.AIAssistants))
		case *connection.SQLServerConnection:
			c.SetAIAssistants(toDomainAIAssistants(req.AIAssistants))
		}
	}

	if err := b.uc.CreateConnection(ctx, conn); err != nil {
		slog.Error("CreateConnection failed", "error", err)
		return ConnectionCreateResult{Error: err.Error()}
	}

	// Store AI API keys in keyring (after connection is created with ID)
	if len(req.AIAssistants) > 0 {
		connID := conn.GetID()
		for _, a := range req.AIAssistants {
			if a.APIKey != "" {
				if err := b.uc.SetAIAPIKey(ctx, connID, a.ID, a.APIKey); err != nil {
					slog.Warn("Failed to store AI API key in keyring", "assistant_id", a.ID, "error", err)
				}
			}
		}
	}

	dto := b.toDTO(conn)
	return ConnectionCreateResult{Connection: &dto}
}

// UpdateConnection updates an existing connection (Wails binding).
func (b *ConnectionBinding) UpdateConnection(req ConnectionUpdateRequest) ConnectionUpdateResult {
	ctx := context.Background()

	// Get existing connection first
	existing, err := b.uc.GetConnectionByID(ctx, req.ID)
	if err != nil {
		slog.Error("UpdateConnection: failed to get existing connection", "id", req.ID, "error", err)
		return ConnectionUpdateResult{Error: err.Error()}
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
		if req.ConnectAs != "" {
			conn.ConnectAs = req.ConnectAs
		} else {
			conn.ConnectAs = "normal"
		}
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

	// Update AI assistants (common to all connection types)
	// First, collect old assistant IDs for cleanup
	oldAssistantIDs := make(map[string]bool)
	for _, a := range existing.GetAIAssistants() {
		oldAssistantIDs[a.ID] = true
	}

	// Set new assistants
	switch c := existing.(type) {
	case *connection.MySQLConnection:
		c.SetAIAssistants(toDomainAIAssistants(req.AIAssistants))
	case *connection.PostgreSQLConnection:
		c.SetAIAssistants(toDomainAIAssistants(req.AIAssistants))
	case *connection.OracleConnection:
		c.SetAIAssistants(toDomainAIAssistants(req.AIAssistants))
	case *connection.SQLServerConnection:
		c.SetAIAssistants(toDomainAIAssistants(req.AIAssistants))
	}

	// Clean up API keys for removed assistants
	newAssistantIDs := make(map[string]bool)
	for _, a := range req.AIAssistants {
		newAssistantIDs[a.ID] = true
	}
	for oldID := range oldAssistantIDs {
		if !newAssistantIDs[oldID] {
			if err := b.uc.DeleteAIAPIKey(ctx, req.ID, oldID); err != nil {
				slog.Warn("Failed to delete old AI API key", "assistant_id", oldID, "error", err)
			} else {
				slog.Info("Cleaned up old AI API key", "connection_id", req.ID, "assistant_id", oldID)
			}
		}
	}

	// Store AI API keys in keyring
	for _, a := range req.AIAssistants {
		if a.APIKey != "" {
			if err := b.uc.SetAIAPIKey(ctx, req.ID, a.ID, a.APIKey); err != nil {
				slog.Warn("Failed to store AI API key in keyring", "assistant_id", a.ID, "error", err)
			}
		}
	}

	if err := b.uc.UpdateConnection(ctx, existing); err != nil {
		slog.Error("UpdateConnection failed", "error", err)
		return ConnectionUpdateResult{Error: err.Error()}
	}

	dto := b.toDTO(existing)
	return ConnectionUpdateResult{Connection: &dto}
}

// DeleteConnection deletes a connection by ID (Wails binding).
func (b *ConnectionBinding) DeleteConnection(id string) ConnectionDeleteResult {
	ctx := context.Background()
	if err := b.uc.DeleteConnection(ctx, id); err != nil {
		slog.Error("DeleteConnection failed", "id", id, "error", err)
		return ConnectionDeleteResult{Success: false, Error: err.Error()}
	}
	return ConnectionDeleteResult{Success: true}
}

// TestConnection tests a connection by ID (Wails binding).
// Tests DB connection and Also tests SSH tunnel if configured.
// Returns combined results for both tests.
func (b *ConnectionBinding) TestConnection(id string) ConnectionTestResult {
	ctx := context.Background()

	// Get connection to check SSH configuration
	conn, err := b.uc.GetConnectionByID(ctx, id)
	if err != nil {
		slog.Error("TestConnection failed to get connection", "id", id, "error", err)
		return ConnectionTestResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to get connection: %v", err),
		}
	}

	// Test DB connection
	dbResult, err := b.uc.TestConnection(ctx, id)
	if err != nil {
		slog.Error("TestConnection DB test failed", "id", id, "error", err)
		return ConnectionTestResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	result := ConnectionTestResult{
		Success:         dbResult.Success,
		LatencyMs:       dbResult.LatencyMs,
		DatabaseVersion: dbResult.DatabaseVersion,
		Error:           dbResult.Error,
	}

	// If SSH is configured, test SSH connection too
	// Use type assertion to get SSH config from concrete types
	var sshConfig *connection.SSHTunnelConfig
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		sshConfig = c.SSH
	case *connection.PostgreSQLConnection:
		sshConfig = c.SSH
	case *connection.OracleConnection:
		sshConfig = c.SSH
	case *connection.SQLServerConnection:
		sshConfig = c.SSH
	}

	// Test SSH if configured
	if sshConfig != nil && sshConfig.Host != "" {
		slog.Info("TestConnection: Testing SSH tunnel",
			"id", id,
			"ssh_host", sshConfig.Host,
			"ssh_port", sshConfig.Port)

		sshSuccess, sshLatencyMs, sshErr := connection.TestSSHConnection(ctx, sshConfig)
		if sshErr != nil {
			slog.Error("TestConnection SSH test failed", "id", id, "error", sshErr)
			result.SSHResult = &SSHWinRMTestResult{
				Success: false,
				Error:   fmt.Sprintf("SSH test failed: %v", sshErr),
			}
		} else {
			result.SSHResult = &SSHWinRMTestResult{
				Success:   sshSuccess,
				LatencyMs: sshLatencyMs,
			}
			if !sshSuccess {
				result.SSHResult.Error = "SSH connection failed"
			}
			slog.Info("TestConnection SSH test completed",
				"id", id,
				"ssh_success", sshSuccess,
				"ssh_latency_ms", sshLatencyMs)
		}
	}

	return result
}

// TestConnectionDirect tests a connection directly from request data (Wails binding).
// This is a DIRECT database connection test (easy connection).
// It does NOT use SSH tunnel - SSH testing is handled separately by TestSSHConnection.
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
			// NOTE: SSH is intentionally NOT set - this is a direct DB test only
		}
	case "postgresql":
		conn = usecase.NewPostgreSQLConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if pgConn, ok := conn.(*connection.PostgreSQLConnection); ok {
			pgConn.SSLMode = req.SSLMode
			pgConn.SetPassword(req.Password)
			// NOTE: SSH is intentionally NOT set - this is a direct DB test only
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
			// NOTE: SSH is intentionally NOT set - this is a direct DB test only
		}
	case "sqlserver":
		conn = usecase.NewSQLServerConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if sqlConn, ok := conn.(*connection.SQLServerConnection); ok {
			sqlConn.SetPassword(req.Password)
			// Use default TrustServerCertificate=false for direct connection test
			// NOTE: SSH and WinRM are intentionally NOT set - this is a direct DB test only
		}
	default:
		return ConnectionTestResult{
			Success: false,
			Error:   "Unknown connection type: " + req.Type,
		}
	}

	// Test the connection (direct database connection only)
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
		dto.ConnectAs = c.ConnectAs
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

	// Load AI assistants (common to all connection types)
	assistants := conn.GetAIAssistants()
	if len(assistants) > 0 {
		dto.AIAssistants = toBindingAIAssistants(assistants)
		// Load AI API keys from keyring
		for i, a := range assistants {
			if apiKey, err := b.uc.GetAIAPIKey(context.Background(), conn.GetID(), a.ID); err == nil {
				dto.AIAssistants[i].APIKey = apiKey
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

// AITestRequest represents a request to test AI API connection.
type AITestRequest struct {
	Provider    string `json:"provider"`
	APIHost     string `json:"api_host"`
	APIEndpoint string `json:"api_endpoint"`
	APIKey      string `json:"api_key"`
	Model       string `json:"model"`
}

// AITestResult represents the result of AI API connection test.
type AITestResult struct {
	Success   bool   `json:"success"`
	LatencyMs int64  `json:"latency_ms"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// AIChatTestRequest represents a request to send a real prompt to the AI model.
type AIChatTestRequest struct {
	Provider    string  `json:"provider"`
	APIHost     string  `json:"api_host"`
	APIEndpoint string  `json:"api_endpoint"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
}

// AIChatTestResult represents the result of an AI prompt test.
type AIChatTestResult struct {
	Success   bool   `json:"success"`
	LatencyMs int64  `json:"latency_ms"`
	Content   string `json:"content,omitempty"`
	Error     string `json:"error,omitempty"`
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

// TestAIConnection tests an AI API connection (Wails binding).
// This is a standalone test that only verifies AI API connectivity.
// It does NOT test database or SSH connections.
func (b *ConnectionBinding) TestAIConnection(req AITestRequest) AITestResult {
	ctx := context.Background()

	slog.Info("TestAIConnection called",
		"provider", req.Provider,
		"api_host", req.APIHost,
		"model", req.Model)

	// Delegate to aiprovider package
	result := aiprovider.TestConnection(ctx, aiprovider.TestRequest{
		Provider:    req.Provider,
		APIHost:     req.APIHost,
		APIEndpoint: req.APIEndpoint,
		APIKey:      req.APIKey,
		Model:       req.Model,
	})

	return AITestResult{
		Success:   result.Success,
		LatencyMs: result.LatencyMs,
		Message:   result.Message,
		Error:     result.Error,
	}
}

// TestAIChat sends a real prompt to the configured AI model and returns the response text.
func (b *ConnectionBinding) TestAIChat(req AIChatTestRequest) AIChatTestResult {
	ctx := context.Background()

	slog.Info("TestAIChat called",
		"provider", req.Provider,
		"api_host", req.APIHost,
		"model", req.Model)

	result := aiprovider.SendChat(ctx, aiprovider.ChatRequest{
		Provider:    req.Provider,
		APIHost:     req.APIHost,
		APIEndpoint: req.APIEndpoint,
		APIKey:      req.APIKey,
		Model:       req.Model,
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
	})

	return AIChatTestResult{
		Success:   result.Success,
		LatencyMs: result.LatencyMs,
		Content:   result.Content,
		Error:     result.Error,
	}
}

// AIModelQueryRequest represents a request to query available AI models.
type AIModelQueryRequest struct {
	Provider string `json:"provider"`
	APIHost  string `json:"api_host"`
	APIKey   string `json:"api_key"`
}

// AIModelInfo represents a single AI model's information.
type AIModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// AIModelQueryResult represents the result of AI model query.
type AIModelQueryResult struct {
	Success bool          `json:"success"`
	Models  []AIModelInfo `json:"models,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// QueryAIModels queries available AI models from the provider (Wails binding).
// For cloud providers: uses /v1/models endpoint with API key.
// For Ollama (local): uses /api/tags endpoint without API key.
func (b *ConnectionBinding) QueryAIModels(req AIModelQueryRequest) AIModelQueryResult {
	ctx := context.Background()

	slog.Info("QueryAIModels called",
		"provider", req.Provider,
		"api_host", req.APIHost)

	// Delegate to aiprovider package
	result := aiprovider.QueryModels(ctx, aiprovider.QueryModelsRequest{
		Provider: req.Provider,
		APIHost:  req.APIHost,
		APIKey:   req.APIKey,
	})

	// Convert result
	models := make([]AIModelInfo, 0, len(result.Models))
	for _, m := range result.Models {
		models = append(models, AIModelInfo{
			ID:   m.ID,
			Name: m.Name,
		})
	}

	return AIModelQueryResult{
		Success: result.Success,
		Models:  models,
		Error:   result.Error,
	}
}
