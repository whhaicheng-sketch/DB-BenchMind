package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
	_ "modernc.org/sqlite"
)

type pgConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	SSLMode  string `json:"ssl_mode"`
}

func main() {
	ctx := context.Background()

	// Open database
	db, err := sql.Open("sqlite", "./data/db-benchmind.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get the current config
	var id, configJSON string
	query := `SELECT id, config_json FROM connections WHERE name = 'PostgreSQL13.14' AND db_type = 'postgresql'`
	err = db.QueryRowContext(ctx, query).Scan(&id, &configJSON)
	if err != nil {
		log.Fatalf("Failed to query connection: %v", err)
	}

	// Parse config
	var config pgConfig
	err = json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	fmt.Printf("Current config:\n")
	fmt.Printf("  Host: %s\n", config.Host)
	fmt.Printf("  Port: %d\n", config.Port)
	fmt.Printf("  Database: %s\n", config.Database)
	fmt.Printf("  Username: %s\n", config.Username)
	fmt.Printf("  SSL Mode: %s\n", config.SSLMode)
	fmt.Println()

	// Ensure database is "postgres"
	if config.Database != "postgres" {
		fmt.Printf("⚠️  Database is '%s', changing to 'postgres'...\n", config.Database)
		config.Database = "postgres"
	}

	// Marshal back to JSON
	newConfigJSON, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	// Update database
	updateQuery := `UPDATE connections SET config_json = ?, updated_at = datetime('now') WHERE id = ?`
	_, err = db.ExecContext(ctx, updateQuery, string(newConfigJSON), id)
	if err != nil {
		log.Fatalf("Failed to update connection: %v", err)
	}

	fmt.Println("✅ Config updated!")

	// Update password in keyring
	dataDir := "./data"
	kr, err := keyring.NewFileFallback(dataDir, "")
	if err != nil {
		fmt.Printf("⚠️  Failed to initialize keyring: %v\n", err)
		fmt.Println("   You may need to update the password manually in the GUI")
		return
	}

	err = kr.Set(ctx, "connection-"+id, "Qwer1234")
	if err != nil {
		fmt.Printf("⚠️  Failed to update password in keyring: %v\n", err)
		fmt.Println("   You may need to update the password manually in the GUI")
	} else {
		fmt.Println("✅ Password updated in keyring!")
	}

	fmt.Println("\nConnection configuration complete!")
	fmt.Println("Please restart the GUI and test the connection again.")
}
