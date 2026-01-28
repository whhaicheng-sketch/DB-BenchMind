// Implements: ConnectionUseCase tests
// Uses table-driven tests following constitution.md requirements
package usecase

import (
	"context"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

// MockConnectionRepository is a mock repository for testing.
type MockConnectionRepository struct {
	connections map[string]connection.Connection
	existingNames map[string]string // name -> id
}

func NewMockConnectionRepository() *MockConnectionRepository {
	return &MockConnectionRepository{
		connections: make(map[string]connection.Connection),
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
	// Remove from existing names
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
// Test Cases
// =============================================================================

// TestConnectionUseCase_CreateConnection tests connection creation.
func TestConnectionUseCase_CreateConnection(t *testing.T) {
	ctx := context.Background()
	repo := NewMockConnectionRepository()
	keyring := NewMockKeyring()
	uc := NewConnectionUseCase(repo, keyring)

	tests := []struct {
		name    string
		conn    connection.Connection
		wantErr bool
		errType string
	}{
		{
			name: "valid MySQL connection",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-1",
					Name: "Test MySQL",
				},
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "root",
				Password: "secret",
			},
			wantErr: false,
		},
		{
			name: "invalid connection - missing host",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-2",
					Name: "Invalid",
				},
				Port:     3306,
				Database: "testdb",
				Username: "root",
			},
			wantErr: true,
			errType:  "validation",
		},
		{
			name: "duplicate name",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-3",
					Name: "Test MySQL", // Same as first test
				},
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "root",
			},
			wantErr: true,
			errType:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If this is the duplicate name test, add the first connection first
			if tt.name == "duplicate name" {
				firstConn := &connection.MySQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "first-conn",
						Name: "Test MySQL",
					},
					Host:     "localhost",
					Port:     3306,
					Database: "testdb",
					Username: "root",
					Password: "secret",
				}
				_ = uc.CreateConnection(ctx, firstConn)
			}

			err := uc.CreateConnection(ctx, tt.conn)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				// Check error type
				errStr := err.Error()
				found := false
				if len(errStr) >= len(tt.errType) {
					for i := 0; i <= len(errStr)-len(tt.errType); i++ {
						if errStr[i:i+len(tt.errType)] == tt.errType {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("CreateConnection() error type = %v, want containing %q", err, tt.errType)
				}
			}
		})
	}
}

// TestConnectionUseCase_ListConnections tests listing connections.
func TestConnectionUseCase_ListConnections(t *testing.T) {
	ctx := context.Background()
	repo := NewMockConnectionRepository()
	keyring := NewMockKeyring()
	uc := NewConnectionUseCase(repo, keyring)

	// Create test connections
	conns := []connection.Connection{
		&connection.MySQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:   "conn-1",
				Name: "Alpha",
			},
			Host:     "host1",
			Port:     3306,
			Database: "db1",
			Username: "user1",
			Password: "pass1",
		},
		&connection.PostgreSQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:   "conn-2",
				Name: "Zeta",
			},
			Host:     "host2",
			Port:     5432,
			Database: "db2",
			Username: "user2",
			Password: "pass2",
		},
	}

	// Add connections (bypassing use case validation for testing)
	for _, conn := range conns {
		_ = repo.Save(ctx, conn)
		_ = keyring.Set(ctx, conn.GetID(), getPassword(conn))
	}

	// List connections
	list, err := uc.ListConnections(ctx)
	if err != nil {
		t.Fatalf("ListConnections() error = %v", err)
	}

	if len(list) != 2 {
		t.Errorf("ListConnections() count = %d, want 2", len(list))
	}
}

// TestConnectionUseCase_GetConnectionByID tests getting a connection by ID.
func TestConnectionUseCase_GetConnectionByID(t *testing.T) {
	ctx := context.Background()
	repo := NewMockConnectionRepository()
	keyring := NewMockKeyring()
	uc := NewConnectionUseCase(repo, keyring)

	// Create a connection
	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn",
			Name: "Test",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "secret",
	}

	_ = repo.Save(ctx, conn)
	_ = keyring.Set(ctx, "test-conn", "secret")

	// Get connection
	found, err := uc.GetConnectionByID(ctx, "test-conn")
	if err != nil {
		t.Fatalf("GetConnectionByID() error = %v", err)
	}

	if found.GetID() != "test-conn" {
		t.Errorf("GetConnectionByID() ID = %v, want test-conn", found.GetID())
	}

	// Verify password was loaded
	mysqlConn, ok := found.(*connection.MySQLConnection)
	if !ok {
		t.Fatal("GetConnectionByID() type is not MySQLConnection")
	}
	if mysqlConn.Password != "secret" {
		t.Errorf("GetConnectionByID() Password = %q, want secret", mysqlConn.Password)
	}
}

// TestConnectionUseCase_DeleteConnection tests deleting a connection.
func TestConnectionUseCase_DeleteConnection(t *testing.T) {
	ctx := context.Background()
	repo := NewMockConnectionRepository()
	keyring := NewMockKeyring()
	uc := NewConnectionUseCase(repo, keyring)

	// Create a connection
	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "delete-me",
			Name: "To Delete",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
		Password: "secret",
	}

	_ = repo.Save(ctx, conn)
	_ = keyring.Set(ctx, "delete-me", "secret")

	// Delete connection
	err := uc.DeleteConnection(ctx, "delete-me")
	if err != nil {
		t.Fatalf("DeleteConnection() error = %v", err)
	}

	// Verify connection is deleted
	_, err = repo.FindByID(ctx, "delete-me")
	if err == nil {
		t.Error("Connection should be deleted")
	}

	// Verify password is deleted from keyring
	_, err = keyring.Get(ctx, "delete-me")
	if err == nil {
		t.Error("Password should be deleted from keyring")
	}
}

// TestConnectionUseCase_DeleteConnection_NotFound tests deleting non-existent connection.
func TestConnectionUseCase_DeleteConnection_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewMockConnectionRepository()
	keyring := NewMockKeyring()
	uc := NewConnectionUseCase(repo, keyring)

	err := uc.DeleteConnection(ctx, "non-existent")
	if err == nil {
		t.Error("DeleteConnection() should return error for non-existent ID")
	}
}

// TestNewMySQLConnection tests factory function.
func TestNewMySQLConnection(t *testing.T) {
	conn := NewMySQLConnection("Test", "localhost", "testdb", "root", 3307)

	if conn.Name != "Test" {
		t.Errorf("Name = %q, want Test", conn.Name)
	}
	if conn.Host != "localhost" {
		t.Errorf("Host = %q, want localhost", conn.Host)
	}
	if conn.Port != 3307 {
		t.Errorf("Port = %d, want 3307", conn.Port)
	}
	if conn.Database != "testdb" {
		t.Errorf("Database = %q, want testdb", conn.Database)
	}
	if conn.Username != "root" {
		t.Errorf("Username = %q, want root", conn.Username)
	}
	if conn.SSLMode != "preferred" {
		t.Errorf("SSLMode = %q, want preferred", conn.SSLMode)
	}
	if conn.ID == "" {
		t.Error("ID should be generated")
	}
}
