// Package bindings provides tests for Wails connection bindings.
// Implements: TASK-017 Unit Tests for ConnectionBinding
package bindings

import (
	"context"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

// =============================================================================
// Mock Implementations (reuse from connection_usecase_test.go)
// =============================================================================

// MockConnectionRepository is a mock repository for testing.
type MockConnectionRepository struct {
	connections   map[string]connection.Connection
	existingNames map[string]string // name -> id
}

func NewMockConnectionRepository() *MockConnectionRepository {
	return &MockConnectionRepository{
		connections:   make(map[string]connection.Connection),
		existingNames: make(map[string]string),
	}
}

func (m *MockConnectionRepository) Save(ctx context.Context, conn connection.Connection) error {
	m.connections[conn.GetID()] = conn
	m.existingNames[conn.GetName()] = conn.GetID()
	return nil
}

func (m *MockConnectionRepository) FindByID(ctx context.Context, id string) (connection.Connection, error) {
	conn, ok := m.connections[id]
	if !ok {
		return nil, &MockNotFoundError{ID: id}
	}
	return conn, nil
}

func (m *MockConnectionRepository) FindAll(ctx context.Context) ([]connection.Connection, error) {
	var result []connection.Connection
	for _, conn := range m.connections {
		result = append(result, conn)
	}
	return result, nil
}

func (m *MockConnectionRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.connections[id]; !ok {
		return &MockNotFoundError{ID: id}
	}
	delete(m.connections, id)
	for name, connID := range m.existingNames {
		if connID == id {
			delete(m.existingNames, name)
			break
		}
	}
	return nil
}

func (m *MockConnectionRepository) ExistsByName(ctx context.Context, name string, excludeID string) (bool, error) {
	id, exists := m.existingNames[name]
	if !exists {
		return false, nil
	}
	if excludeID != "" && id == excludeID {
		return false, nil
	}
	return true, nil
}

// MockNotFoundError is a mock not found error.
type MockNotFoundError struct {
	ID string
}

func (e *MockNotFoundError) Error() string {
	return "not found: " + e.ID
}

// MockKeyring is a mock keyring for testing.
type MockKeyring struct {
	passwords map[string]string
}

func NewMockKeyring() *MockKeyring {
	return &MockKeyring{
		passwords: make(map[string]string),
	}
}

func (m *MockKeyring) Set(ctx context.Context, key, password string) error {
	m.passwords[key] = password
	return nil
}

func (m *MockKeyring) Get(ctx context.Context, key string) (string, error) {
	pw, ok := m.passwords[key]
	if !ok {
		return "", &keyring.ErrNotFound{Key: key}
	}
	return pw, nil
}

func (m *MockKeyring) Delete(ctx context.Context, key string) error {
	delete(m.passwords, key)
	return nil
}

func (m *MockKeyring) Available(ctx context.Context) bool {
	return true
}

// =============================================================================
// Test TestSSHConnection
// =============================================================================

func TestConnectionBinding_TestSSHConnection_MissingHost(t *testing.T) {
	tests := []struct {
		name    string
		req     SSHTestRequest
		wantErr string
	}{
		{
			name: "empty host",
			req: SSHTestRequest{
				Host:     "",
				Port:     22,
				Username: "root",
				Password: "secret",
			},
			wantErr: "SSH host is required",
		},
		{
			name: "missing host only",
			req: SSHTestRequest{
				Username: "root",
				Password: "secret",
			},
			wantErr: "SSH host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create binding with nil usecase (we don't need it for validation tests)
			binding := &ConnectionBinding{uc: nil}

			result := binding.TestSSHConnection(tt.req)

			if result.Success {
				t.Error("TestSSHConnection() should fail when host is missing")
			}
			if result.Error != tt.wantErr {
				t.Errorf("TestSSHConnection() error = %q, want %q", result.Error, tt.wantErr)
			}
		})
	}
}

func TestConnectionBinding_TestSSHConnection_MissingUsername(t *testing.T) {
	tests := []struct {
		name    string
		req     SSHTestRequest
		wantErr string
	}{
		{
			name: "empty username",
			req: SSHTestRequest{
				Host:     "192.168.1.100",
				Port:     22,
				Username: "",
				Password: "secret",
			},
			wantErr: "SSH username is required",
		},
		{
			name: "missing username only",
			req: SSHTestRequest{
				Host:     "192.168.1.100",
				Password: "secret",
			},
			wantErr: "SSH username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding := &ConnectionBinding{uc: nil}

			result := binding.TestSSHConnection(tt.req)

			if result.Success {
				t.Error("TestSSHConnection() should fail when username is missing")
			}
			if result.Error != tt.wantErr {
				t.Errorf("TestSSHConnection() error = %q, want %q", result.Error, tt.wantErr)
			}
		})
	}
}

// TestConnectionBinding_TestSSHConnection_DefaultPort tests default port logic.
// NOTE: This test is skipped in unit tests because TestSSHConnection attempts
// actual SSH connections. Use integration tests for full coverage.
// The default port logic is tested indirectly through TestConnectionBinding_CreateConnection_WithSSH.
func TestConnectionBinding_TestSSHConnection_DefaultPort(t *testing.T) {
	t.Skip("Skipping: TestSSHConnection attempts actual SSH connection. " +
		"Default port logic is covered by TestConnectionBinding_CreateConnection_WithSSH")
}

// =============================================================================
// Test TestWinRMConnection
// =============================================================================

func TestConnectionBinding_TestWinRMConnection_MissingHost(t *testing.T) {
	tests := []struct {
		name    string
		req     WinRMTestRequest
		wantErr string
	}{
		{
			name: "empty host",
			req: WinRMTestRequest{
				Host:     "",
				Port:     5985,
				Username: "admin",
				Password: "secret",
			},
			wantErr: "WinRM host is required",
		},
		{
			name: "missing host",
			req: WinRMTestRequest{
				Port:     5985,
				Username: "admin",
				Password: "secret",
			},
			wantErr: "WinRM host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding := &ConnectionBinding{uc: nil}

			result := binding.TestWinRMConnection(tt.req)

			if result.Success {
				t.Error("TestWinRMConnection() should fail when host is missing")
			}
			if result.Error != tt.wantErr {
				t.Errorf("TestWinRMConnection() error = %q, want %q", result.Error, tt.wantErr)
			}
		})
	}
}

// TestConnectionBinding_TestWinRMConnection_DefaultPorts tests default port logic.
// NOTE: This test is skipped in unit tests because TestWinRMConnection attempts
// actual network connections. Use integration tests for full coverage.
// The default port logic is tested indirectly through TestConnectionBinding_CreateConnection_WithWinRM.
func TestConnectionBinding_TestWinRMConnection_DefaultPorts(t *testing.T) {
	t.Skip("Skipping: TestWinRMConnection attempts actual network connection. " +
		"Default port logic is covered by TestConnectionBinding_CreateConnection_WithWinRM")
}

// =============================================================================
// Test CreateConnection with SSH/WinRM
// =============================================================================

// newTestBinding creates a ConnectionBinding with mock dependencies for testing.
func newTestBinding() (*ConnectionBinding, *MockConnectionRepository, *MockKeyring) {
	repo := NewMockConnectionRepository()
	keyring := NewMockKeyring()
	uc := usecase.NewConnectionUseCase(repo, keyring)
	return NewConnectionBinding(uc), repo, keyring
}

func TestConnectionBinding_CreateConnection_WithSSH(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		req         ConnectionCreateRequest
		wantSSH     bool
		sshPort     int
		sshUsername string
	}{
		{
			name: "MySQL with SSH enabled",
			req: ConnectionCreateRequest{
				Name:        "Test MySQL SSH",
				Type:        "mysql",
				Host:        "192.168.1.100",
				Port:        3306,
				Database:    "testdb",
				Username:    "root",
				Password:    "secret",
				SSLMode:     "preferred",
				SSHEnabled:  true,
				SSHPort:     22,
				SSHUsername: "sshuser",
				SSHPassword: "sshpass",
			},
			wantSSH:     true,
			sshPort:     22,
			sshUsername: "sshuser",
		},
		{
			name: "MySQL with SSH but default port",
			req: ConnectionCreateRequest{
				Name:        "Test MySQL SSH Default",
				Type:        "mysql",
				Host:        "192.168.1.100",
				Port:        3306,
				Database:    "testdb",
				Username:    "root",
				Password:    "secret",
				SSLMode:     "preferred",
				SSHEnabled:  true,
				SSHPort:     0, // Should default to 22
				SSHUsername: "sshuser",
				SSHPassword: "sshpass",
			},
			wantSSH:     true,
			sshPort:     22,
			sshUsername: "sshuser",
		},
		{
			name: "PostgreSQL with SSH enabled",
			req: ConnectionCreateRequest{
				Name:        "Test PostgreSQL SSH",
				Type:        "postgresql",
				Host:        "192.168.1.100",
				Port:        5432,
				Database:    "testdb",
				Username:    "postgres",
				Password:    "secret",
				SSLMode:     "disable", // PostgreSQL valid ssl_mode: disable, require, verify-ca, verify-full
				SSHEnabled:  true,
				SSHPort:     2222,
				SSHUsername: "sshuser",
				SSHPassword: "sshpass",
			},
			wantSSH:     true,
			sshPort:     2222,
			sshUsername: "sshuser",
		},
		{
			name: "Oracle with SSH enabled",
			req: ConnectionCreateRequest{
				Name:        "Test Oracle SSH",
				Type:        "oracle",
				Host:        "192.168.1.100",
				Port:        1521,
				Database:    "ORCL",
				Username:    "system",
				Password:    "secret",
				SSHEnabled:  true,
				SSHPort:     22,
				SSHUsername: "oracle",
				SSHPassword: "sshpass",
			},
			wantSSH:     true,
			sshPort:     22,
			sshUsername: "oracle",
		},
		{
			name: "MySQL without SSH",
			req: ConnectionCreateRequest{
				Name:       "Test MySQL No SSH",
				Type:       "mysql",
				Host:       "192.168.1.100",
				Port:       3306,
				Database:   "testdb",
				Username:   "root",
				Password:   "secret",
				SSLMode:    "preferred",
				SSHEnabled: false,
			},
			wantSSH: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding, _, _ := newTestBinding()
			result := binding.CreateConnection(tt.req)
			dto := result.Connection

			if dto == nil {
				t.Fatalf("CreateConnection() returned nil connection, error=%q", result.Error)
			}

			// Check SSH status in DTO
			if dto.SSHEnabled != tt.wantSSH {
				t.Errorf("CreateConnection() SSHEnabled = %v, want %v", dto.SSHEnabled, tt.wantSSH)
			}

			// Verify SSH config was set correctly on the connection
			conn, err := binding.uc.GetConnectionByID(ctx, dto.ID)
			if err != nil {
				t.Fatalf("GetConnectionByID() error = %v", err)
			}

			sshConfig := getSSHConfig(conn)
			if tt.wantSSH {
				if sshConfig == nil {
					t.Error("getSSHConfig() returned nil, expected SSH config")
					return
				}
				if !sshConfig.Enabled {
					t.Error("SSH config not enabled")
				}
				if sshConfig.Port != tt.sshPort {
					t.Errorf("SSH port = %d, want %d", sshConfig.Port, tt.sshPort)
				}
				if sshConfig.Username != tt.sshUsername {
					t.Errorf("SSH username = %q, want %q", sshConfig.Username, tt.sshUsername)
				}
			} else {
				if sshConfig != nil {
					t.Errorf("getSSHConfig() should return nil for non-SSH connection, got %+v", sshConfig)
				}
			}
		})
	}
}

func TestConnectionBinding_CreateConnection_WithWinRM(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		req           ConnectionCreateRequest
		wantWinRM     bool
		winrmPort     int
		winrmUseHTTPS bool
		winrmUsername string
	}{
		{
			name: "SQL Server with WinRM HTTP",
			req: ConnectionCreateRequest{
				Name:          "Test SQL Server WinRM HTTP",
				Type:          "sqlserver",
				Host:          "192.168.1.100",
				Port:          1433,
				Database:      "testdb",
				Username:      "sa",
				Password:      "secret",
				WinRMEnabled:  true,
				WinRMPort:     5985,
				WinRMUseHTTPS: false,
				WinRMUsername: "admin",
				WinRMPassword: "winrmpass",
			},
			wantWinRM:     true,
			winrmPort:     5985,
			winrmUseHTTPS: false,
			winrmUsername: "admin",
		},
		{
			name: "SQL Server with WinRM HTTPS",
			req: ConnectionCreateRequest{
				Name:          "Test SQL Server WinRM HTTPS",
				Type:          "sqlserver",
				Host:          "192.168.1.100",
				Port:          1433,
				Database:      "testdb",
				Username:      "sa",
				Password:      "secret",
				WinRMEnabled:  true,
				WinRMPort:     0, // Should default to 5986 for HTTPS
				WinRMUseHTTPS: true,
				WinRMUsername: "admin",
				WinRMPassword: "winrmpass",
			},
			wantWinRM:     true,
			winrmPort:     5986,
			winrmUseHTTPS: true,
			winrmUsername: "admin",
		},
		{
			name: "SQL Server with WinRM HTTP default port",
			req: ConnectionCreateRequest{
				Name:          "Test SQL Server WinRM Default",
				Type:          "sqlserver",
				Host:          "192.168.1.100",
				Port:          1433,
				Database:      "testdb",
				Username:      "sa",
				Password:      "secret",
				WinRMEnabled:  true,
				WinRMPort:     0, // Should default to 5985 for HTTP
				WinRMUseHTTPS: false,
				WinRMUsername: "admin",
				WinRMPassword: "winrmpass",
			},
			wantWinRM:     true,
			winrmPort:     5985,
			winrmUseHTTPS: false,
			winrmUsername: "admin",
		},
		{
			name: "SQL Server without WinRM",
			req: ConnectionCreateRequest{
				Name:         "Test SQL Server No WinRM",
				Type:         "sqlserver",
				Host:         "192.168.1.100",
				Port:         1433,
				Database:     "testdb",
				Username:     "sa",
				Password:     "secret",
				WinRMEnabled: false,
			},
			wantWinRM: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding, _, _ := newTestBinding()
			result := binding.CreateConnection(tt.req)
			dto := result.Connection

			if dto == nil {
				t.Fatalf("CreateConnection() returned nil connection, error=%q", result.Error)
			}

			// Check WinRM status in DTO
			if dto.WinRMEnabled != tt.wantWinRM {
				t.Errorf("CreateConnection() WinRMEnabled = %v, want %v", dto.WinRMEnabled, tt.wantWinRM)
			}

			// Verify WinRM config was set correctly on the connection
			conn, err := binding.uc.GetConnectionByID(ctx, dto.ID)
			if err != nil {
				t.Fatalf("GetConnectionByID() error = %v", err)
			}

			winrmConfig := getWinRMConfig(conn)
			if tt.wantWinRM {
				if winrmConfig == nil {
					t.Error("getWinRMConfig() returned nil, expected WinRM config")
					return
				}
				if !winrmConfig.Enabled {
					t.Error("WinRM config not enabled")
				}
				if winrmConfig.Port != tt.winrmPort {
					t.Errorf("WinRM port = %d, want %d", winrmConfig.Port, tt.winrmPort)
				}
				if winrmConfig.UseHTTPS != tt.winrmUseHTTPS {
					t.Errorf("WinRM UseHTTPS = %v, want %v", winrmConfig.UseHTTPS, tt.winrmUseHTTPS)
				}
				if winrmConfig.Username != tt.winrmUsername {
					t.Errorf("WinRM username = %q, want %q", winrmConfig.Username, tt.winrmUsername)
				}
			} else {
				if winrmConfig != nil {
					t.Errorf("getWinRMConfig() should return nil for non-WinRM connection, got %+v", winrmConfig)
				}
			}
		})
	}
}

func TestConnectionBinding_CreateConnection_UnknownType(t *testing.T) {
	binding, _, _ := newTestBinding()

	req := ConnectionCreateRequest{
		Name:     "Unknown Type",
		Type:     "unknown",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "secret",
	}

	result := binding.CreateConnection(req)

	if result.Connection != nil {
		t.Error("CreateConnection() should return nil connection for unknown type")
	}
	if result.Error == "" {
		t.Error("CreateConnection() should return error for unknown type")
	}
}

// =============================================================================
// Test UpdateConnection with SSH/WinRM
// =============================================================================

func TestConnectionBinding_UpdateConnection_WithSSH(t *testing.T) {
	ctx := context.Background()
	binding, _, _ := newTestBinding()

	// First create a MySQL connection
	createReq := ConnectionCreateRequest{
		Name:       "Original MySQL",
		Type:       "mysql",
		Host:       "192.168.1.100",
		Port:       3306,
		Database:   "testdb",
		Username:   "root",
		Password:   "secret",
		SSLMode:    "preferred",
		SSHEnabled: false,
	}
	createResult := binding.CreateConnection(createReq)
	createdDTO := createResult.Connection
	if createdDTO == nil {
		t.Fatalf("Failed to create connection: %s", createResult.Error)
	}

	// Now update with SSH enabled
	updateReq := ConnectionUpdateRequest{
		ID:          createdDTO.ID,
		Name:        "Updated MySQL with SSH",
		Host:        "192.168.1.200",
		Port:        3307,
		Database:    "newdb",
		Username:    "newuser",
		Password:    "newsecret",
		SSLMode:     "required",
		SSHEnabled:  true,
		SSHPort:     2222,
		SSHUsername: "sshuser",
		SSHPassword: "sshpass",
	}

	updateResult := binding.UpdateConnection(updateReq)
	updatedDTO := updateResult.Connection
	if updatedDTO == nil {
		t.Fatalf("UpdateConnection() returned nil connection, error=%q", updateResult.Error)
	}

	if updatedDTO.Name != "Updated MySQL with SSH" {
		t.Errorf("UpdateConnection() Name = %q, want %q", updatedDTO.Name, "Updated MySQL with SSH")
	}

	if !updatedDTO.SSHEnabled {
		t.Error("UpdateConnection() SSHEnabled should be true")
	}

	// Verify SSH config on connection
	conn, err := binding.uc.GetConnectionByID(ctx, updatedDTO.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID() error = %v", err)
	}

	sshConfig := getSSHConfig(conn)
	if sshConfig == nil {
		t.Fatal("getSSHConfig() returned nil after update with SSH")
	}
	if !sshConfig.Enabled {
		t.Error("SSH should be enabled after update")
	}
	if sshConfig.Port != 2222 {
		t.Errorf("SSH port = %d, want 2222", sshConfig.Port)
	}
	if sshConfig.Username != "sshuser" {
		t.Errorf("SSH username = %q, want sshuser", sshConfig.Username)
	}
}

func TestConnectionBinding_UpdateConnection_DisableSSH(t *testing.T) {
	ctx := context.Background()
	binding, _, _ := newTestBinding()

	// Create MySQL connection with SSH
	createReq := ConnectionCreateRequest{
		Name:        "MySQL with SSH",
		Type:        "mysql",
		Host:        "192.168.1.100",
		Port:        3306,
		Database:    "testdb",
		Username:    "root",
		Password:    "secret",
		SSLMode:     "preferred",
		SSHEnabled:  true,
		SSHPort:     22,
		SSHUsername: "sshuser",
		SSHPassword: "sshpass",
	}
	createResult := binding.CreateConnection(createReq)
	createdDTO := createResult.Connection
	if createdDTO == nil {
		t.Fatalf("Failed to create connection: %s", createResult.Error)
	}

	// Update to disable SSH
	updateReq := ConnectionUpdateRequest{
		ID:          createdDTO.ID,
		Name:        "MySQL without SSH",
		Host:        "192.168.1.100",
		Port:        3306,
		Database:    "testdb",
		Username:    "root",
		Password:    "", // Keep existing
		SSLMode:     "preferred",
		SSHEnabled:  false,
		SSHPort:     0,
		SSHUsername: "",
		SSHPassword: "",
	}

	updateResult := binding.UpdateConnection(updateReq)
	updatedDTO := updateResult.Connection
	if updatedDTO == nil {
		t.Fatalf("UpdateConnection() returned nil connection, error=%q", updateResult.Error)
	}

	if updatedDTO.SSHEnabled {
		t.Error("UpdateConnection() SSHEnabled should be false")
	}

	// Verify SSH config is nil after disabling
	conn, err := binding.uc.GetConnectionByID(ctx, updatedDTO.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID() error = %v", err)
	}

	sshConfig := getSSHConfig(conn)
	if sshConfig != nil {
		t.Errorf("getSSHConfig() should return nil after disabling SSH, got %+v", sshConfig)
	}
}

func TestConnectionBinding_UpdateConnection_WithWinRM(t *testing.T) {
	ctx := context.Background()
	binding, _, _ := newTestBinding()

	// Create SQL Server connection without WinRM
	createReq := ConnectionCreateRequest{
		Name:         "Original SQL Server",
		Type:         "sqlserver",
		Host:         "192.168.1.100",
		Port:         1433,
		Database:     "testdb",
		Username:     "sa",
		Password:     "secret",
		WinRMEnabled: false,
	}
	createResult := binding.CreateConnection(createReq)
	createdDTO := createResult.Connection
	if createdDTO == nil {
		t.Fatalf("Failed to create connection: %s", createResult.Error)
	}

	// Update with WinRM enabled
	updateReq := ConnectionUpdateRequest{
		ID:            createdDTO.ID,
		Name:          "Updated SQL Server with WinRM",
		Host:          "192.168.1.200",
		Port:          1433,
		Database:      "newdb",
		Username:      "sa",
		Password:      "newsecret",
		WinRMEnabled:  true,
		WinRMPort:     5986,
		WinRMUseHTTPS: true,
		WinRMUsername: "admin",
		WinRMPassword: "winrmpass",
	}

	updateResult := binding.UpdateConnection(updateReq)
	updatedDTO := updateResult.Connection
	if updatedDTO == nil {
		t.Fatalf("UpdateConnection() returned nil connection, error=%q", updateResult.Error)
	}

	if !updatedDTO.WinRMEnabled {
		t.Error("UpdateConnection() WinRMEnabled should be true")
	}

	// Verify WinRM config on connection
	conn, err := binding.uc.GetConnectionByID(ctx, updatedDTO.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID() error = %v", err)
	}

	winrmConfig := getWinRMConfig(conn)
	if winrmConfig == nil {
		t.Fatal("getWinRMConfig() returned nil after update with WinRM")
	}
	if !winrmConfig.Enabled {
		t.Error("WinRM should be enabled after update")
	}
	if winrmConfig.Port != 5986 {
		t.Errorf("WinRM port = %d, want 5986", winrmConfig.Port)
	}
	if !winrmConfig.UseHTTPS {
		t.Error("WinRM UseHTTPS should be true")
	}
}

func TestConnectionBinding_UpdateConnection_DisableWinRM(t *testing.T) {
	ctx := context.Background()
	binding, _, _ := newTestBinding()

	// Create SQL Server connection with WinRM
	createReq := ConnectionCreateRequest{
		Name:          "SQL Server with WinRM",
		Type:          "sqlserver",
		Host:          "192.168.1.100",
		Port:          1433,
		Database:      "testdb",
		Username:      "sa",
		Password:      "secret",
		WinRMEnabled:  true,
		WinRMPort:     5985,
		WinRMUseHTTPS: false,
		WinRMUsername: "admin",
		WinRMPassword: "winrmpass",
	}
	createResult := binding.CreateConnection(createReq)
	createdDTO := createResult.Connection
	if createdDTO == nil {
		t.Fatalf("Failed to create connection: %s", createResult.Error)
	}

	// Update to disable WinRM
	updateReq := ConnectionUpdateRequest{
		ID:            createdDTO.ID,
		Name:          "SQL Server without WinRM",
		Host:          "192.168.1.100",
		Port:          1433,
		Database:      "testdb",
		Username:      "sa",
		Password:      "",
		WinRMEnabled:  false,
		WinRMPort:     0,
		WinRMUseHTTPS: false,
		WinRMUsername: "",
		WinRMPassword: "",
	}

	updateResult := binding.UpdateConnection(updateReq)
	updatedDTO := updateResult.Connection
	if updatedDTO == nil {
		t.Fatalf("UpdateConnection() returned nil connection, error=%q", updateResult.Error)
	}

	if updatedDTO.WinRMEnabled {
		t.Error("UpdateConnection() WinRMEnabled should be false")
	}

	// Verify WinRM config is nil after disabling
	conn, err := binding.uc.GetConnectionByID(ctx, updatedDTO.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID() error = %v", err)
	}

	winrmConfig := getWinRMConfig(conn)
	if winrmConfig != nil {
		t.Errorf("getWinRMConfig() should return nil after disabling WinRM, got %+v", winrmConfig)
	}
}

// =============================================================================
// Test Helper Functions
// =============================================================================

func TestGetSSHConfig(t *testing.T) {
	tests := []struct {
		name     string
		conn     connection.Connection
		wantNil  bool
		wantHost string
		wantPort int
	}{
		{
			name: "MySQL with SSH",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{ID: "test-1"},
				Host:     "192.168.1.100",
				Port:     3306,
				Database: "testdb",
				Username: "root",
				SSH: &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     "192.168.1.100",
					Port:     22,
					Username: "sshuser",
				},
			},
			wantNil:  false,
			wantHost: "192.168.1.100",
			wantPort: 22,
		},
		{
			name: "MySQL without SSH",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{ID: "test-2"},
				Host:     "192.168.1.100",
				Port:     3306,
				Database: "testdb",
				Username: "root",
				SSH:      nil,
			},
			wantNil: true,
		},
		{
			name: "PostgreSQL with SSH",
			conn: &connection.PostgreSQLConnection{
				BaseConnection: connection.BaseConnection{ID: "test-3"},
				Host:     "192.168.1.100",
				Port:     5432,
				Database: "testdb",
				Username: "postgres",
				SSH: &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     "192.168.1.100",
					Port:     2222,
					Username: "sshuser",
				},
			},
			wantNil:  false,
			wantHost: "192.168.1.100",
			wantPort: 2222,
		},
		{
			name: "Oracle with SSH",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{ID: "test-4"},
				Host:     "192.168.1.100",
				Port:     1521,
				SID:      "ORCL",
				Username: "system",
				SSH: &connection.SSHTunnelConfig{
					Enabled:  true,
					Host:     "192.168.1.100",
					Port:     22,
					Username: "oracle",
				},
			},
			wantNil:  false,
			wantHost: "192.168.1.100",
			wantPort: 22,
		},
		{
			name: "SQL Server returns nil (no SSH support)",
			conn: &connection.SQLServerConnection{
				BaseConnection: connection.BaseConnection{ID: "test-5"},
				Host:     "192.168.1.100",
				Port:     1433,
				Database: "testdb",
				Username: "sa",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSSHConfig(tt.conn)

			if tt.wantNil {
				if result != nil {
					t.Errorf("getSSHConfig() = %+v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("getSSHConfig() returned nil, want non-nil")
			}
			if result.Host != tt.wantHost {
				t.Errorf("getSSHConfig() Host = %q, want %q", result.Host, tt.wantHost)
			}
			if result.Port != tt.wantPort {
				t.Errorf("getSSHConfig() Port = %d, want %d", result.Port, tt.wantPort)
			}
		})
	}
}

func TestGetWinRMConfig(t *testing.T) {
	tests := []struct {
		name      string
		conn      connection.Connection
		wantNil   bool
		wantHost  string
		wantPort  int
		wantHTTPS bool
	}{
		{
			name: "SQL Server with WinRM HTTP",
			conn: &connection.SQLServerConnection{
				BaseConnection: connection.BaseConnection{ID: "test-1"},
				Host:     "192.168.1.100",
				Port:     1433,
				Database: "testdb",
				Username: "sa",
				WinRM: &connection.WinRMConfig{
					Enabled:  true,
					Host:     "192.168.1.100",
					Port:     5985,
					UseHTTPS: false,
					Username: "admin",
				},
			},
			wantNil:   false,
			wantHost:  "192.168.1.100",
			wantPort:  5985,
			wantHTTPS: false,
		},
		{
			name: "SQL Server with WinRM HTTPS",
			conn: &connection.SQLServerConnection{
				BaseConnection: connection.BaseConnection{ID: "test-2"},
				Host:     "192.168.1.100",
				Port:     1433,
				Database: "testdb",
				Username: "sa",
				WinRM: &connection.WinRMConfig{
					Enabled:  true,
					Host:     "192.168.1.100",
					Port:     5986,
					UseHTTPS: true,
					Username: "admin",
				},
			},
			wantNil:   false,
			wantHost:  "192.168.1.100",
			wantPort:  5986,
			wantHTTPS: true,
		},
		{
			name: "SQL Server without WinRM",
			conn: &connection.SQLServerConnection{
				BaseConnection: connection.BaseConnection{ID: "test-3"},
				Host:     "192.168.1.100",
				Port:     1433,
				Database: "testdb",
				Username: "sa",
				WinRM:    nil,
			},
			wantNil: true,
		},
		{
			name: "MySQL returns nil (no WinRM support)",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{ID: "test-4"},
				Host:     "192.168.1.100",
				Port:     3306,
				Database: "testdb",
				Username: "root",
			},
			wantNil: true,
		},
		{
			name: "PostgreSQL returns nil (no WinRM support)",
			conn: &connection.PostgreSQLConnection{
				BaseConnection: connection.BaseConnection{ID: "test-5"},
				Host:     "192.168.1.100",
				Port:     5432,
				Database: "testdb",
				Username: "postgres",
			},
			wantNil: true,
		},
		{
			name: "Oracle returns nil (no WinRM support)",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{ID: "test-6"},
				Host:     "192.168.1.100",
				Port:     1521,
				SID:      "ORCL",
				Username: "system",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getWinRMConfig(tt.conn)

			if tt.wantNil {
				if result != nil {
					t.Errorf("getWinRMConfig() = %+v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("getWinRMConfig() returned nil, want non-nil")
			}
			if result.Host != tt.wantHost {
				t.Errorf("getWinRMConfig() Host = %q, want %q", result.Host, tt.wantHost)
			}
			if result.Port != tt.wantPort {
				t.Errorf("getWinRMConfig() Port = %d, want %d", result.Port, tt.wantPort)
			}
			if result.UseHTTPS != tt.wantHTTPS {
				t.Errorf("getWinRMConfig() UseHTTPS = %v, want %v", result.UseHTTPS, tt.wantHTTPS)
			}
		})
	}
}

// =============================================================================
// Test toDTO function
// =============================================================================

func TestConnectionBinding_toDTO(t *testing.T) {
	binding, _, _ := newTestBinding()

	tests := []struct {
		name         string
		conn         connection.Connection
		wantType     string
		wantSSH      bool
		wantWinRM    bool
		wantTrust    bool
	}{
		{
			name: "MySQL with SSH",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "mysql-1",
					Name: "MySQL SSH",
				},
				Host:     "192.168.1.100",
				Port:     3306,
				Database: "testdb",
				Username: "root",
				SSLMode:  "preferred",
				SSH: &connection.SSHTunnelConfig{
					Enabled: true,
				},
			},
			wantType:  "mysql",
			wantSSH:   true,
			wantWinRM: false,
		},
		{
			name: "PostgreSQL without SSH",
			conn: &connection.PostgreSQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "pg-1",
					Name: "PostgreSQL",
				},
				Host:     "192.168.1.100",
				Port:     5432,
				Database: "testdb",
				Username: "postgres",
				SSLMode:  "prefer",
				SSH:      nil,
			},
			wantType:  "postgresql",
			wantSSH:   false,
			wantWinRM: false,
		},
		{
			name: "Oracle with SSH",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "oracle-1",
					Name: "Oracle",
				},
				Host:     "192.168.1.100",
				Port:     1521,
				SID:      "ORCL",
				Username: "system",
				SSH: &connection.SSHTunnelConfig{
					Enabled: true,
				},
			},
			wantType:  "oracle",
			wantSSH:   true,
			wantWinRM: false,
		},
		{
			name: "SQL Server with WinRM and Trust",
			conn: &connection.SQLServerConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "sqlserver-1",
					Name: "SQL Server",
				},
				Host:                   "192.168.1.100",
				Port:                   1433,
				Database:               "testdb",
				Username:               "sa",
				TrustServerCertificate: true,
				WinRM: &connection.WinRMConfig{
					Enabled: true,
				},
			},
			wantType:  "sqlserver",
			wantSSH:   false,
			wantWinRM: true,
			wantTrust: true,
		},
		{
			name: "SQL Server without WinRM",
			conn: &connection.SQLServerConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "sqlserver-2",
					Name: "SQL Server No WinRM",
				},
				Host:                   "192.168.1.100",
				Port:                   1433,
				Database:               "testdb",
				Username:               "sa",
				TrustServerCertificate: false,
				WinRM:                  nil,
			},
			wantType:  "sqlserver",
			wantSSH:   false,
			wantWinRM: false,
			wantTrust: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := binding.toDTO(tt.conn)

			if dto.ID != tt.conn.GetID() {
				t.Errorf("toDTO() ID = %q, want %q", dto.ID, tt.conn.GetID())
			}
			if dto.Name != tt.conn.GetName() {
				t.Errorf("toDTO() Name = %q, want %q", dto.Name, tt.conn.GetName())
			}
			if dto.Type != tt.wantType {
				t.Errorf("toDTO() Type = %q, want %q", dto.Type, tt.wantType)
			}
			if dto.SSHEnabled != tt.wantSSH {
				t.Errorf("toDTO() SSHEnabled = %v, want %v", dto.SSHEnabled, tt.wantSSH)
			}
			if dto.WinRMEnabled != tt.wantWinRM {
				t.Errorf("toDTO() WinRMEnabled = %v, want %v", dto.WinRMEnabled, tt.wantWinRM)
			}
			if dto.TrustServerCertificate != tt.wantTrust {
				t.Errorf("toDTO() TrustServerCertificate = %v, want %v", dto.TrustServerCertificate, tt.wantTrust)
			}
		})
	}
}

// =============================================================================
// Test keyring provider integration
// =============================================================================

// Verify MockKeyring implements keyring.Provider interface
var _ keyring.Provider = (*MockKeyring)(nil)

// getSSHConfig extracts SSH config from a connection.
func getSSHConfig(conn connection.Connection) *connection.SSHTunnelConfig {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		return c.SSH
	case *connection.PostgreSQLConnection:
		return c.SSH
	case *connection.OracleConnection:
		return c.SSH
	case *connection.SQLServerConnection:
		// SQL Server doesn't support SSH
		return nil
	}
	return nil
}

// getWinRMConfig extracts WinRM config from a connection.
func getWinRMConfig(conn connection.Connection) *connection.WinRMConfig {
	switch c := conn.(type) {
	case *connection.SQLServerConnection:
		return c.WinRM
	default:
		return nil
	}
}
