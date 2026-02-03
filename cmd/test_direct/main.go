package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Direct PostgreSQL Driver Test")
	fmt.Println("================================")
	fmt.Println()

	// Test 1: Correct DSN
	fmt.Println("Test 1: Using correct connection string")
	dsn1 := "host=192.168.170.137 port=5432 database=postgres user=admin password=Qwer1234 sslmode=disable"
	fmt.Printf("DSN: %s\n\n", dsn1)

	db1, err := sql.Open("postgres", dsn1)
	if err != nil {
		log.Fatalf("Failed to open: %v", err)
	}
	defer db1.Close()

	err = db1.Ping()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n\n", err)
	} else {
		fmt.Println("✅ Success!")
		var version string
		err = db1.QueryRow("SELECT version()").Scan(&version)
		if err == nil {
			fmt.Printf("Version: %s\n\n", version)
		}
	}

	// Test 2: Wrong DSN (database=admin)
	fmt.Println("Test 2: Using WRONG connection string (database=admin)")
	dsn2 := "host=192.168.170.137 port=5432 database=admin user=admin password=Qwer1234 sslmode=disable"
	fmt.Printf("DSN: %s\n\n", dsn2)

	db2, err := sql.Open("postgres", dsn2)
	if err != nil {
		log.Fatalf("Failed to open: %v", err)
	}
	defer db2.Close()

	err = db2.Ping()
	if err != nil {
		fmt.Printf("❌ Failed (expected): %v\n\n", err)
	} else {
		fmt.Println("✅ Success (unexpected)")
	}

	os.Exit(0)
}
