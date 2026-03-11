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
		log.Fatal(err)
	}
	defer db.Close()

	var name, configJSON string
	err = db.QueryRowContext(ctx, "SELECT name, config_json FROM connections WHERE name = ?", "Oracle Test").Scan(&name, &configJSON)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Oracle Test Connection ===")
	fmt.Println("Name:", name)

	var config map[string]interface{}
	json.Unmarshal([]byte(configJSON), &config)

	fmt.Println("\nConfig Details:")
	fmt.Printf("  Host: %v\n", config["host"])
	fmt.Printf("  Port: %v\n", config["port"])
	fmt.Printf("  Service Name: %v\n", config["service_name"])
	fmt.Printf("  SID: %v\n", config["sid"])
	fmt.Printf("  Username: %v\n", config["username"])
}
