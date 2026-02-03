package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()

	db, err := sql.Open("sqlite", "./data/db-benchmind.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	var id, configJSON string
	query := `SELECT id, config_json FROM connections WHERE name = 'PostgreSQL13.14'`
	err = db.QueryRowContext(ctx, query).Scan(&id, &configJSON)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Printf("Connection ID: %s\n", id)
	fmt.Printf("\nRaw config_json:\n%s\n\n", configJSON)

	// Parse to verify
	var config map[string]interface{}
	err = json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	fmt.Println("Parsed config:")
	for k, v := range config {
		fmt.Printf("  %s: %v\n", k, v)
	}
}
