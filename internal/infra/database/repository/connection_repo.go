// Package repository provides SQLite implementation of repository interfaces.
// Implements: ConnectionRepository interface from usecase layer
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// SQLiteConnectionRepository implements ConnectionRepository interface using SQLite.
// Implements: usecase.ConnectionRepository
type SQLiteConnectionRepository struct {
	db *sql.DB
}

// NewSQLiteConnectionRepository creates a new SQLite connection repository.
func NewSQLiteConnectionRepository(db *sql.DB) usecase.ConnectionRepository {
	return &SQLiteConnectionRepository{db: db}
}

// Save saves a connection to the database.
// Implements: usecase.ConnectionRepository.Save
func (r *SQLiteConnectionRepository) Save(ctx context.Context, conn connection.Connection) error {
	// Generate ID if not exists
	if conn.GetID() == "" {
		// Note: This is a simplified approach. In production, you'd set the ID
		// before calling Save, or use a different approach.
		return fmt.Errorf("connection ID must be set before saving")
	}

	// Serialize connection config to JSON (without password)
	configJSON, err := r.serializeConnection(conn)
	if err != nil {
		return fmt.Errorf("marshal connection: %w", err)
	}

	now := time.Now().Format(time.RFC3339)

	// Use INSERT OR REPLACE to handle both create and update
	query := `
		INSERT INTO connections (id, name, db_type, config_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			db_type = excluded.db_type,
			config_json = excluded.config_json,
			updated_at = excluded.updated_at
	`

	_, err = r.db.ExecContext(ctx, query,
		conn.GetID(),
		conn.GetName(),
		string(conn.GetType()),
		configJSON,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("save connection: %w", err)
	}

	return nil
}

// FindByID finds a connection by its ID.
// Implements: usecase.ConnectionRepository.FindByID
func (r *SQLiteConnectionRepository) FindByID(ctx context.Context, id string) (connection.Connection, error) {
	var name, connType, configJSON string

	err := r.db.QueryRowContext(ctx,
		"SELECT name, db_type, config_json FROM connections WHERE id = ?", id).
		Scan(&name, &connType, &configJSON)

	if err == sql.ErrNoRows {
		return nil, &ConnectionNotFoundError{ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("query connection: %w", err)
	}

	// Deserialize based on database type
	return r.deserializeConnection(id, name, connection.DatabaseType(connType), configJSON)
}

// FindAll finds all connections in the database.
// Implements: usecase.ConnectionRepository.FindAll
func (r *SQLiteConnectionRepository) FindAll(ctx context.Context) ([]connection.Connection, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, db_type, config_json FROM connections ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("query connections: %w", err)
	}
	defer rows.Close()

	var conns []connection.Connection
	for rows.Next() {
		var id, name, connType, configJSON string
		if err := rows.Scan(&id, &name, &connType, &configJSON); err != nil {
			return nil, fmt.Errorf("scan connection: %w", err)
		}

		conn, err := r.deserializeConnection(id, name, connection.DatabaseType(connType), configJSON)
		if err != nil {
			return nil, fmt.Errorf("deserialize connection %s: %w", id, err)
		}

		conns = append(conns, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate connections: %w", err)
	}

	return conns, nil
}

// Delete deletes a connection by its ID.
// Implements: usecase.ConnectionRepository.Delete
func (r *SQLiteConnectionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM connections WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return &ConnectionNotFoundError{ID: id}
	}

	return nil
}

// ExistsByName checks if a connection with the given name exists.
// Implements: usecase.ConnectionRepository.ExistsByName
func (r *SQLiteConnectionRepository) ExistsByName(ctx context.Context, name string, excludeID string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM connections WHERE name = ?"
	args := []interface{}{name}

	if excludeID != "" {
		query += " AND id != ?"
		args = append(args, excludeID)
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check connection exists: %w", err)
	}

	return count > 0, nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// serializeConnection serializes a connection to JSON.
func (r *SQLiteConnectionRepository) serializeConnection(conn connection.Connection) (string, error) {
	// Create a map that includes all connection fields except password
	data := map[string]interface{}{
		"id":         conn.GetID(),
		"name":       conn.GetName(),
		"type":       string(conn.GetType()),
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	// Add type-specific fields
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		data["host"] = c.Host
		data["port"] = c.Port
		data["database"] = c.Database
		data["username"] = c.Username
		data["ssl_mode"] = c.SSLMode
	case *connection.OracleConnection:
		data["host"] = c.Host
		data["port"] = c.Port
		data["service_name"] = c.ServiceName
		data["sid"] = c.SID
		data["username"] = c.Username
	case *connection.SQLServerConnection:
		data["host"] = c.Host
		data["port"] = c.Port
		data["database"] = c.Database
		data["username"] = c.Username
		data["trust_server_certificate"] = c.TrustServerCertificate
	case *connection.PostgreSQLConnection:
		data["host"] = c.Host
		data["port"] = c.Port
		data["database"] = c.Database
		data["username"] = c.Username
		data["ssl_mode"] = c.SSLMode
	default:
		return "", fmt.Errorf("unsupported connection type: %T", conn)
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal connection data: %w", err)
	}

	return string(bytes), nil
}

// deserializeConnection deserializes a connection from JSON.
func (r *SQLiteConnectionRepository) deserializeConnection(id, name string, connType connection.DatabaseType, configJSON string) (connection.Connection, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &data); err != nil {
		return nil, fmt.Errorf("unmarshal connection config: %w", err)
	}

	// Parse created_at and updated_at
	createdAt, _ := time.Parse(time.RFC3339, getString(data, "created_at"))
	updatedAt, _ := time.Parse(time.RFC3339, getString(data, "updated_at"))

	base := connection.BaseConnection{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	switch connType {
	case connection.DatabaseTypeMySQL:
		conn := &connection.MySQLConnection{
			BaseConnection: base,
			Host:           getString(data, "host"),
			Port:           getInt(data, "port"),
			Database:       getString(data, "database"),
			Username:       getString(data, "username"),
			SSLMode:        getString(data, "ssl_mode"),
		}
		// Set default port if not specified
		if conn.Port == 0 {
			conn.Port = 3306
		}
		return conn, nil

	case connection.DatabaseTypeOracle:
		conn := &connection.OracleConnection{
			BaseConnection: base,
			Host:           getString(data, "host"),
			Port:           getInt(data, "port"),
			ServiceName:    getString(data, "service_name"),
			SID:            getString(data, "sid"),
			Username:       getString(data, "username"),
		}
		if conn.Port == 0 {
			conn.Port = 1521
		}
		return conn, nil

	case connection.DatabaseTypeSQLServer:
		conn := &connection.SQLServerConnection{
			BaseConnection: base,
			Host:           getString(data, "host"),
			Port:           getInt(data, "port"),
			Database:       getString(data, "database"),
			Username:       getString(data, "username"),
			TrustServerCertificate: getBool(data, "trust_server_certificate"),
		}
		if conn.Port == 0 {
			conn.Port = 1433
		}
		return conn, nil

	case connection.DatabaseTypePostgreSQL:
		conn := &connection.PostgreSQLConnection{
			BaseConnection: base,
			Host:           getString(data, "host"),
			Port:           getInt(data, "port"),
			Database:       getString(data, "database"),
			Username:       getString(data, "username"),
			SSLMode:        getString(data, "ssl_mode"),
		}
		if conn.Port == 0 {
			conn.Port = 5432
		}
		return conn, nil

	default:
		return nil, fmt.Errorf("unknown connection type: %s", connType)
	}
}

// =============================================================================
// Error Types
// =============================================================================

// ConnectionNotFoundError is returned when a connection is not found.
type ConnectionNotFoundError struct {
	ID string
}

func (e *ConnectionNotFoundError) Error() string {
	return fmt.Sprintf("connection not found: %s", e.ID)
}

// =============================================================================
// Helper Functions for JSON Parsing
// =============================================================================

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
