// Implements: ConnectionRepository tests
// Uses table-driven tests following constitution.md requirements
package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// TestSQLiteConnectionRepository_SaveAndFind tests Save and FindByID operations.
func TestSQLiteConnectionRepository_SaveAndFind(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	repo := NewSQLiteConnectionRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		conn    connection.Connection
		wantErr bool
	}{
		{
			name: "save and find MySQL connection",
			conn: &connection.MySQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:        "test-mysql-1",
					Name:      "Test MySQL",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
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
			name: "save and find PostgreSQL connection",
			conn: &connection.PostgreSQLConnection{
				BaseConnection: connection.BaseConnection{
					ID:        "test-pg-1",
					Name:      "Test PostgreSQL",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "postgres",
				SSLMode:  "require",
			},
			wantErr: false,
		},
		{
			name: "save and find Oracle connection",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{
					ID:        "test-oracle-1",
					Name:      "Test Oracle",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Username:    "system",
			},
			wantErr: false,
		},
		{
			name: "save and find SQL Server connection",
			conn: &connection.SQLServerConnection{
				BaseConnection: connection.BaseConnection{
					ID:        "test-mssql-1",
					Name:      "Test SQL Server",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Host:                   "localhost",
				Port:                   1433,
				Database:               "testdb",
				Username:               "sa",
				TrustServerCertificate: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save
			err := repo.Save(ctx, tt.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Find
			found, err := repo.FindByID(ctx, tt.conn.GetID())
			if err != nil {
				t.Fatalf("FindByID() unexpected error = %v", err)
			}

			// Verify
			if found.GetID() != tt.conn.GetID() {
				t.Errorf("FindByID() ID = %v, want %v", found.GetID(), tt.conn.GetID())
			}
			if found.GetName() != tt.conn.GetName() {
				t.Errorf("FindByID() Name = %v, want %v", found.GetName(), tt.conn.GetName())
			}
			if found.GetType() != tt.conn.GetType() {
				t.Errorf("FindByID() Type = %v, want %v", found.GetType(), tt.conn.GetType())
			}
		})
	}
}

// TestSQLiteConnectionRepository_FindAll tests FindAll operation.
func TestSQLiteConnectionRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteConnectionRepository(db)
	ctx := context.Background()

	// Save multiple connections
	conns := []connection.Connection{
		&connection.MySQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:   "conn-1",
				Name: "Alpha",
			},
			Host:     "localhost",
			Port:     3306,
			Database: "db1",
			Username: "user1",
		},
		&connection.PostgreSQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:   "conn-2",
				Name: "Zeta",
			},
			Host:     "localhost",
			Port:     5432,
			Database: "db2",
			Username: "user2",
		},
		&connection.MySQLConnection{
			BaseConnection: connection.BaseConnection{
				ID:   "conn-3",
				Name: "Beta",
			},
			Host:     "localhost",
			Port:     3306,
			Database: "db3",
			Username: "user3",
		},
	}

	for _, conn := range conns {
		if err := repo.Save(ctx, conn); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}
	}

	// Find all
	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}

	// Should return 3 connections, ordered by name
	if len(all) != 3 {
		t.Errorf("FindAll() count = %d, want 3", len(all))
	}

	// Check ordering (should be alphabetical by name)
	if all[0].GetName() != "Alpha" {
		t.Errorf("FindAll()[0] Name = %v, want Alpha", all[0].GetName())
	}
	if all[1].GetName() != "Beta" {
		t.Errorf("FindAll()[1] Name = %v, want Beta", all[1].GetName())
	}
	if all[2].GetName() != "Zeta" {
		t.Errorf("FindAll()[2] Name = %v, want Zeta", all[2].GetName())
	}
}

// TestSQLiteConnectionRepository_Delete tests Delete operation.
func TestSQLiteConnectionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteConnectionRepository(db)
	ctx := context.Background()

	// Save a connection
	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "delete-test",
			Name: "To Be Deleted",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
	}

	if err := repo.Save(ctx, conn); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify it exists
	found, err := repo.FindByID(ctx, "delete-test")
	if err != nil {
		t.Fatalf("FindByID() before delete failed: %v", err)
	}
	if found == nil {
		t.Fatal("Connection should exist before delete")
	}

	// Delete
	err = repo.Delete(ctx, "delete-test")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = repo.FindByID(ctx, "delete-test")
	if err == nil {
		t.Error("FindByID() after delete should return error")
	}
}

// TestSQLiteConnectionRepository_Delete_NotFound tests Delete with non-existent ID.
func TestSQLiteConnectionRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteConnectionRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "non-existent")
	if err == nil {
		t.Error("Delete() should return error for non-existent ID")
	}

	if err != nil && !isConnectionNotFound(err) {
		t.Errorf("Delete() error type = %T, want ConnectionNotFoundError", err)
	}
}

// TestSQLiteConnectionRepository_ExistsByName tests ExistsByName operation.
func TestSQLiteConnectionRepository_ExistsByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteConnectionRepository(db)
	ctx := context.Background()

	// Save a connection
	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "exists-test",
			Name: "Duplicate Test",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
	}

	if err := repo.Save(ctx, conn); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	tests := []struct {
		name      string
		connName  string
		excludeID string
		want      bool
	}{
		{
			name:      "existing connection",
			connName:  "Duplicate Test",
			excludeID: "",
			want:      true,
		},
		{
			name:      "non-existing connection",
			connName:  "Not Exists",
			excludeID: "",
			want:      false,
		},
		{
			name:      "existing but excluded",
			connName:  "Duplicate Test",
			excludeID: "exists-test",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.ExistsByName(ctx, tt.connName, tt.excludeID)
			if err != nil {
				t.Fatalf("ExistsByName() error = %v", err)
			}
			if exists != tt.want {
				t.Errorf("ExistsByName() = %v, want %v", exists, tt.want)
			}
		})
	}
}

// TestSQLiteConnectionRepository_Update tests updating an existing connection.
func TestSQLiteConnectionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteConnectionRepository(db)
	ctx := context.Background()

	// Save initial connection
	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "update-test",
			Name: "Original Name",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "olddb",
		Username: "root",
	}

	if err := repo.Save(ctx, conn); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Update connection
	conn.SetName("Updated Name")
	conn.Database = "newdb"

	if err := repo.Save(ctx, conn); err != nil {
		t.Fatalf("Save() update failed: %v", err)
	}

	// Verify update
	found, err := repo.FindByID(ctx, "update-test")
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}

	if found.GetName() != "Updated Name" {
		t.Errorf("Name = %v, want Updated Name", found.GetName())
	}

	// Note: Database field requires type assertion to verify
	if mysqlConn, ok := found.(*connection.MySQLConnection); ok {
		if mysqlConn.Database != "newdb" {
			t.Errorf("Database = %v, want newdb", mysqlConn.Database)
		}
	} else {
		t.Error("Found connection is not MySQLConnection type")
	}
}

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS connections (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			db_type TEXT NOT NULL,
			config_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_connections_db_type ON connections(db_type);
		CREATE INDEX IF NOT EXISTS idx_connections_created_at ON connections(created_at);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("create tables: %v", err)
	}

	return db
}

// isConnectionNotFound checks if error is ConnectionNotFoundError.
func isConnectionNotFound(err error) bool {
	_, ok := err.(*ConnectionNotFoundError)
	return ok
}
