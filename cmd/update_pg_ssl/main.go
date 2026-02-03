package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/lib/pq"
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

	fmt.Printf("PostgreSQL Server: %s:%d\n", config.Host, config.Port)
	fmt.Printf("Database: %s\n", config.Database)
	fmt.Printf("User: %s\n", config.Username)
	fmt.Printf("Current SSL mode: '%s'\n", config.SSLMode)
	fmt.Println()

	// The server has SSL disabled, so we must use sslmode=disable
	if config.SSLMode != "disable" {
		fmt.Printf("Updating SSL mode to 'disable' (server has SSL disabled)...\n")
		config.SSLMode = "disable"

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
		fmt.Println("✅ SSL mode updated!")
	} else {
		fmt.Println("SSL mode is already 'disable'")
	}
}
