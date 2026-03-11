package main

import (
	"context"
	"database/sql"
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

	rows, err := db.QueryContext(ctx, "SELECT id, name, db_type FROM connections ORDER BY name")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== Database Connections ===")
	for rows.Next() {
		var id, name, dbType string
		if err := rows.Scan(&id, &name, &dbType); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %s | Name: %-20s | Type: %s\n", id[:8], name, dbType)
	}
}
