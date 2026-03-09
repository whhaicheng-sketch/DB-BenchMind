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
	SSLMode  string `json:"ssl_mode,omitempty"`
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
		}
	case "postgresql":
		conn = usecase.NewPostgreSQLConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if pgConn, ok := conn.(*connection.PostgreSQLConnection); ok {
			pgConn.SSLMode = req.SSLMode
			pgConn.SetPassword(req.Password)
		}
	case "oracle":
		conn = usecase.NewOracleConnection(req.Name, req.Host, "", "", req.Username, req.Port)
		if oraConn, ok := conn.(*connection.OracleConnection); ok {
			oraConn.SetPassword(req.Password)
		}
	case "sqlserver":
		conn = usecase.NewSQLServerConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if sqlConn, ok := conn.(*connection.SQLServerConnection); ok {
			sqlConn.SetPassword(req.Password)
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
	case *connection.OracleConnection:
		conn.SetName(req.Name)
		conn.Host = req.Host
		conn.Port = req.Port
		conn.Username = req.Username
		if req.Password != "" {
			conn.SetPassword(req.Password)
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
		}
	case "postgresql":
		conn = usecase.NewPostgreSQLConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if pgConn, ok := conn.(*connection.PostgreSQLConnection); ok {
			pgConn.SSLMode = req.SSLMode
			pgConn.SetPassword(req.Password)
		}
	case "oracle":
		conn = usecase.NewOracleConnection(req.Name, req.Host, "", "", req.Username, req.Port)
		if oraConn, ok := conn.(*connection.OracleConnection); ok {
			oraConn.SetPassword(req.Password)
		}
	case "sqlserver":
		conn = usecase.NewSQLServerConnection(req.Name, req.Host, req.Database, req.Username, req.Port)
		if sqlConn, ok := conn.(*connection.SQLServerConnection); ok {
			sqlConn.SetPassword(req.Password)
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
	case *connection.PostgreSQLConnection:
		dto.Host = c.Host
		dto.Port = c.Port
		dto.Database = c.Database
		dto.Username = c.Username
		dto.SSLMode = c.SSLMode
	case *connection.OracleConnection:
		dto.Host = c.Host
		dto.Port = c.Port
		dto.Database = c.ServiceName
		if dto.Database == "" {
			dto.Database = c.SID
		}
		dto.Username = c.Username
	case *connection.SQLServerConnection:
		dto.Host = c.Host
		dto.Port = c.Port
		dto.Database = c.Database
		dto.Username = c.Username
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
