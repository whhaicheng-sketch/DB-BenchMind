package bindings

import (
	"context"
	"database/sql"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
	_ "modernc.org/sqlite"
)

func TestOracleBindingCreateConnection_BasicServiceName(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS connections (id TEXT PRIMARY KEY, name TEXT NOT NULL, db_type TEXT NOT NULL, config_json TEXT NOT NULL, created_at TEXT, updated_at TEXT)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	binding := NewConnectionBinding(usecase.NewConnectionUseCase(repository.NewSQLiteConnectionRepository(db), keyring.NewMemoryKeyring()))

	result := binding.CreateConnection(ConnectionCreateRequest{
		Name:              "oracle-basic-service",
		Type:              "oracle",
		Host:              "db.local",
		Port:              1521,
		Username:          "system",
		ConnectType:       "basic",
		OracleConnectMode: "basic",
		IdentifierType:    "service_name",
		ServiceName:       "ORCL",
	})
	if result.Error != "" {
		t.Fatalf("CreateConnection() error = %s", result.Error)
	}
	if result.Connection == nil {
		t.Fatal("CreateConnection() returned nil connection")
	}
	if result.Connection.ConnectType != "basic" {
		t.Fatalf("ConnectType = %q, want basic", result.Connection.ConnectType)
	}
	if result.Connection.ServiceName != "ORCL" {
		t.Fatalf("ServiceName = %q, want ORCL", result.Connection.ServiceName)
	}
	if result.Connection.SID != "" {
		t.Fatalf("SID = %q, want empty", result.Connection.SID)
	}
}

func TestOracleBindingCreateConnection_TNS(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS connections (id TEXT PRIMARY KEY, name TEXT NOT NULL, db_type TEXT NOT NULL, config_json TEXT NOT NULL, created_at TEXT, updated_at TEXT)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	uc := usecase.NewConnectionUseCase(repository.NewSQLiteConnectionRepository(db), keyring.NewMemoryKeyring())
	binding := NewConnectionBinding(uc)

	result := binding.CreateConnection(ConnectionCreateRequest{
		Name:              "oracle-tns",
		Type:              "oracle",
		Username:          "system",
		ConnectType:       "tns",
		OracleConnectMode: "tns",
		TNSName:           "ORCLCDB_HIGH",
	})
	if result.Error != "" {
		t.Fatalf("CreateConnection() error = %s", result.Error)
	}

	dto := binding.GetConnection(result.Connection.ID)
	if dto == nil {
		t.Fatal("GetConnection() returned nil")
	}
	if dto.ConnectType != "tns" {
		t.Fatalf("ConnectType = %q, want tns", dto.ConnectType)
	}
	if dto.TNSName != "ORCLCDB_HIGH" {
		t.Fatalf("TNSName = %q, want ORCLCDB_HIGH", dto.TNSName)
	}
	if dto.Host != "" || dto.Port != 0 {
		t.Fatalf("expected TNS mode to avoid basic host/port requirements, got host=%q port=%d", dto.Host, dto.Port)
	}

	if _, err := uc.GetConnectionByID(context.Background(), result.Connection.ID); err != nil {
		t.Fatalf("GetConnectionByID() error = %v", err)
	}
}
