package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
	"log"
)

func main() {
	db, err := sql.Open("sqlite", "./data/db-benchmind.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get current Oracle connection config
	var id, name, configJSON string
	err = db.QueryRow("SELECT id, name, config_json FROM connections WHERE db_type = 'oracle' LIMIT 1").Scan(&id, &name, &configJSON)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Current config for %s (id=%s):\n%s\n\n", name, id, configJSON)

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		log.Fatal(err)
	}

	// Update host
	oldHost := config["host"]
	config["host"] = "192.168.170.137"
	config["sid"] = "orcl" // Also fix case

	newConfigJSON, _ := json.Marshal(config)
	fmt.Printf("New config:\n%s\n\n", string(newConfigJSON))

	// Update database
	_, err = db.Exec("UPDATE connections SET config_json = ? WHERE id = ?", newConfigJSON, id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Updated Oracle connection host: %s -> %s\n", oldHost, config["host"])
	fmt.Printf("✅ Updated Oracle connection SID: ORCL -> orcl\n")
}
