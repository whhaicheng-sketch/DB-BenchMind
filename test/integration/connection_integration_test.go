// Package integration provides end-to-end tests for DB-BenchMind.
// These tests verify the complete integration of domain, use case, and infrastructure layers.
package integration

import (
	"context"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

// TestIntegration_ConnectionWorkflow tests the complete connection management workflow.
// Scenario:
// 1. Initialize database and keyring
// 2. Create a MySQL connection
// 3. Verify connection is saved to database
// 4. Retrieve connection by ID
// 5. Verify password is retrieved from keyring
// 6. Update connection
// 7. List all connections
// 8. Delete connection
// 9. Verify connection is removed from database and keyring
func TestIntegration_ConnectionWorkflow(t *testing.T) {
	ctx := context.Background()

	// Step 1: Initialize database with temporary file
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Step 2: Initialize keyring with temporary directory
	keyringProvider, err := keyring.NewFileFallback(tmpDir, "test-master-password")
	if err != nil {
		t.Fatalf("Failed to initialize keyring: %v", err)
	}

	// Step 3: Initialize repository and use case
	connRepo := repository.NewSQLiteConnectionRepository(db)
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

	// Step 4: Create a MySQL connection
	testConn := usecase.NewMySQLConnection(
		"Integration Test MySQL",
		"localhost",
		"testdb",
		"testuser",
		3306,
	)
	testConn.Password = "test-password-123"

	err = connUC.CreateConnection(ctx, testConn)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}

	// Step 5: Verify connection is saved to database
	saved, err := connUC.GetConnectionByID(ctx, testConn.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve connection: %v", err)
	}

	if saved.GetID() != testConn.ID {
		t.Errorf("Connection ID mismatch: got %s, want %s", saved.GetID(), testConn.ID)
	}
	if saved.GetName() != testConn.Name {
		t.Errorf("Connection Name mismatch: got %s, want %s", saved.GetName(), testConn.Name)
	}

	// Step 6: Verify password is retrieved from keyring
	mysqlConn, ok := saved.(*connection.MySQLConnection)
	if !ok {
		t.Fatal("Connection is not MySQLConnection type")
	}
	if mysqlConn.Password != testConn.Password {
		t.Errorf("Password mismatch: got %s, want %s", mysqlConn.Password, testConn.Password)
	}

	// Step 7: Update connection
	mysqlConn.Name = "Updated Connection Name"
	mysqlConn.Host = "updated-host"

	err = connUC.UpdateConnection(ctx, mysqlConn)
	if err != nil {
		t.Fatalf("Failed to update connection: %v", err)
	}

	// Verify update
	updated, err := connUC.GetConnectionByID(ctx, testConn.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated connection: %v", err)
	}

	if updated.GetName() != "Updated Connection Name" {
		t.Errorf("Updated Name mismatch: got %s, want 'Updated Connection Name'", updated.GetName())
	}

	// Step 8: List all connections
	all, err := connUC.ListConnections(ctx)
	if err != nil {
		t.Fatalf("Failed to list connections: %v", err)
	}

	if len(all) != 1 {
		t.Errorf("Connection count mismatch: got %d, want 1", len(all))
	}

	// Step 9: Delete connection
	err = connUC.DeleteConnection(ctx, testConn.ID)
	if err != nil {
		t.Fatalf("Failed to delete connection: %v", err)
	}

	// Verify connection is deleted
	_, err = connUC.GetConnectionByID(ctx, testConn.ID)
	if err == nil {
		t.Error("Connection should be deleted after DeleteConnection()")
	}

	// Verify password is deleted from keyring
	_, err = keyringProvider.Get(ctx, testConn.ID)
	if err == nil {
		t.Error("Password should be deleted from keyring after connection deletion")
	}
}

// TestIntegration_MultipleConnectionTypes tests creating connections of all supported types.
func TestIntegration_MultipleConnectionTypes(t *testing.T) {
	ctx := context.Background()

	// Setup
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	keyringProvider, err := keyring.NewFileFallback(tmpDir, "test-password")
	if err != nil {
		t.Fatalf("Failed to initialize keyring: %v", err)
	}

	connRepo := repository.NewSQLiteConnectionRepository(db)
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

	// Create connections of all types
	connections := []struct {
		name string
		conn connection.Connection
	}{
		{
			name: "MySQL",
			conn: usecase.NewMySQLConnection("Test MySQL", "localhost", "db1", "user1", 3306),
		},
		{
			name: "PostgreSQL",
			conn: usecase.NewPostgreSQLConnection("Test PostgreSQL", "localhost", "db2", "user2", 5432),
		},
		{
			name: "Oracle",
			conn: usecase.NewOracleConnection("Test Oracle", "localhost", "ORCL", "", "sys", 1521),
		},
		{
			name: "SQL Server",
			conn: usecase.NewSQLServerConnection("Test SQL Server", "localhost", "db3", "sa", 1433),
		},
	}

	// Set passwords and create connections
	for _, tc := range connections {
		switch c := tc.conn.(type) {
		case *connection.MySQLConnection:
			c.Password = "secret123"
		case *connection.PostgreSQLConnection:
			c.Password = "secret456"
		case *connection.OracleConnection:
			c.Password = "secret789"
		case *connection.SQLServerConnection:
			c.Password = "secret000"
		}

		err := connUC.CreateConnection(ctx, tc.conn)
		if err != nil {
			t.Fatalf("Failed to create %s connection: %v", tc.name, err)
		}
	}

	// List all connections
	all, err := connUC.ListConnections(ctx)
	if err != nil {
		t.Fatalf("Failed to list connections: %v", err)
	}

	// Verify all connections were created
	if len(all) != 4 {
		t.Errorf("Connection count mismatch: got %d, want 4", len(all))
	}

	// Verify each connection type
	typeCount := map[connection.DatabaseType]int{
		connection.DatabaseTypeMySQL:      0,
		connection.DatabaseTypePostgreSQL: 0,
		connection.DatabaseTypeOracle:     0,
		connection.DatabaseTypeSQLServer:  0,
	}

	for _, conn := range all {
		typeCount[conn.GetType()]++
	}

	if typeCount[connection.DatabaseTypeMySQL] != 1 {
		t.Errorf("MySQL connection count: got %d, want 1", typeCount[connection.DatabaseTypeMySQL])
	}
	if typeCount[connection.DatabaseTypePostgreSQL] != 1 {
		t.Errorf("PostgreSQL connection count: got %d, want 1", typeCount[connection.DatabaseTypePostgreSQL])
	}
	if typeCount[connection.DatabaseTypeOracle] != 1 {
		t.Errorf("Oracle connection count: got %d, want 1", typeCount[connection.DatabaseTypeOracle])
	}
	if typeCount[connection.DatabaseTypeSQLServer] != 1 {
		t.Errorf("SQL Server connection count: got %d, want 1", typeCount[connection.DatabaseTypeSQLServer])
	}
}

// TestIntegration_DuplicateNameError tests that duplicate connection names are rejected.
func TestIntegration_DuplicateNameError(t *testing.T) {
	ctx := context.Background()

	// Setup
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	keyringProvider, err := keyring.NewFileFallback(tmpDir, "test-password")
	if err != nil {
		t.Fatalf("Failed to initialize keyring: %v", err)
	}

	connRepo := repository.NewSQLiteConnectionRepository(db)
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

	// Create first connection
	conn1 := usecase.NewMySQLConnection("Duplicate", "host1", "db1", "user1", 3306)
	conn1.Password = "pass1"

	err = connUC.CreateConnection(ctx, conn1)
	if err != nil {
		t.Fatalf("Failed to create first connection: %v", err)
	}

	// Try to create second connection with same name
	conn2 := usecase.NewMySQLConnection("Duplicate", "host2", "db2", "user2", 3306)
	conn2.Password = "pass2"
	conn2.ID = "different-id"

	err = connUC.CreateConnection(ctx, conn2)
	if err == nil {
		t.Error("CreateConnection should return error for duplicate name")
	}

	// Verify error type
	if err != nil {
		_, isDuplicate := err.(*usecase.DuplicateNameError)
		if !isDuplicate {
			t.Errorf("Error type should be DuplicateNameError, got %T", err)
		}
	}
}

// TestIntegration_ConnectionValidation tests that invalid connections are rejected.
func TestIntegration_ConnectionValidation(t *testing.T) {
	ctx := context.Background()

	// Setup
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	keyringProvider, _ := keyring.NewFileFallback(tmpDir, "test-password")
	connRepo := repository.NewSQLiteConnectionRepository(db)
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

	// Test cases with invalid connections
	tests := []struct {
		name    string
		conn    connection.Connection
		wantErr bool
	}{
		{
			name: "missing name",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{
					ID: "test-1",
				},
				Host: "localhost",
				Port: 3306,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-2",
					Name: "Test",
				},
				Host: "localhost",
				Port: -1,
			},
			wantErr: true,
		},
		{
			name: "missing host",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-3",
					Name: "Test",
				},
				Port: 3306,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := connUC.CreateConnection(ctx, tt.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIntegration_PersistenceAcrossRestart tests that data persists across "restarts".
func TestIntegration_PersistenceAcrossRestart(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	// First "session": create connection
	db1, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database (session 1): %v", err)
	}

	keyring1, _ := keyring.NewFileFallback(tmpDir, "test-password")
	connRepo1 := repository.NewSQLiteConnectionRepository(db1)
	connUC1 := usecase.NewConnectionUseCase(connRepo1, keyring1)

	conn := usecase.NewMySQLConnection("Persistent Test", "localhost", "db", "user", 3306)
	conn.Password = "persistent-password"

	err = connUC1.CreateConnection(ctx, conn)
	if err != nil {
		db1.Close()
		t.Fatalf("Failed to create connection (session 1): %v", err)
	}

	db1.Close()

	// Second "session": reopen database and verify connection exists
	db2, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database (session 2): %v", err)
	}
	defer db2.Close()

	keyring2, _ := keyring.NewFileFallback(tmpDir, "test-password")
	connRepo2 := repository.NewSQLiteConnectionRepository(db2)
	connUC2 := usecase.NewConnectionUseCase(connRepo2, keyring2)

	// Verify connection persists
	retrieved, err := connUC2.GetConnectionByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve connection (session 2): %v", err)
	}

	if retrieved.GetName() != "Persistent Test" {
		t.Errorf("Connection Name mismatch: got %s, want 'Persistent Test'", retrieved.GetName())
	}

	// Verify password persists in keyring
	mysqlConn, ok := retrieved.(*connection.MySQLConnection)
	if !ok {
		t.Fatal("Connection is not MySQLConnection type")
	}
	if mysqlConn.Password != "persistent-password" {
		t.Errorf("Password not persisted: got %s, want 'persistent-password'", mysqlConn.Password)
	}
}

// TestIntegration_ConnectionListOrdering tests that connections are ordered by name.
func TestIntegration_ConnectionListOrdering(t *testing.T) {
	ctx := context.Background()

	// Setup
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	keyringProvider, _ := keyring.NewFileFallback(tmpDir, "test-password")
	connRepo := repository.NewSQLiteConnectionRepository(db)
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

	// Create connections with specific names (not alphabetical order)
	connections := []struct {
		name string
		conn connection.Connection
	}{
		{"Zeta", usecase.NewMySQLConnection("Zeta", "localhost", "db1", "user1", 3306)},
		{"Alpha", usecase.NewMySQLConnection("Alpha", "localhost", "db2", "user2", 3306)},
		{"Beta", usecase.NewMySQLConnection("Beta", "localhost", "db3", "user3", 3306)},
	}

	for _, tc := range connections {
		switch c := tc.conn.(type) {
		case *connection.MySQLConnection:
			c.Password = "pass"
		}
		if err := connUC.CreateConnection(ctx, tc.conn); err != nil {
			t.Fatalf("Failed to create connection %s: %v", tc.name, err)
		}
	}

	// List connections
	all, err := connUC.ListConnections(ctx)
	if err != nil {
		t.Fatalf("Failed to list connections: %v", err)
	}

	// Verify ordering (alphabetical by name)
	if len(all) != 3 {
		t.Fatalf("Connection count mismatch: got %d, want 3", len(all))
	}

	expectedOrder := []string{"Alpha", "Beta", "Zeta"}
	for i, conn := range all {
		if conn.GetName() != expectedOrder[i] {
			t.Errorf("Connection at index %d: got %s, want %s", i, conn.GetName(), expectedOrder[i])
		}
	}
}
