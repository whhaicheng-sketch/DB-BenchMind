package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

func main() {
	ctx := context.Background()

	// Initialize database
	db, err := database.InitializeSQLite(ctx, "./data/db-benchmind.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize connection repository
	connRepo := repository.NewSQLiteConnectionRepository(db)

	// Initialize keyring
	keyringProvider, err := keyring.NewFileFallback("./data", "")
	if err != nil {
		log.Fatalf("Failed to initialize keyring: %v", err)
	}

	// Get all connections
	conns, err := connRepo.FindAll(ctx)
	if err != nil {
		log.Fatalf("Failed to get connections: %v", err)
	}

	// Find PostgreSQL connection
	var pgConn *connection.PostgreSQLConnection
	var connID string
	for _, c := range conns {
		if pc, ok := c.(*connection.PostgreSQLConnection); ok && pc.Name == "PostgreSQL13.14" {
			pgConn = pc
			connID = pc.ID
			break
		}
	}

	if pgConn == nil {
		log.Fatal("❌ PostgreSQL13.14 connection not found")
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║        PostgreSQL Connection Debug Info                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Connection Object Fields:\n")
	fmt.Printf("  Name:     '%s'\n", pgConn.Name)
	fmt.Printf("  Host:     '%s'\n", pgConn.Host)
	fmt.Printf("  Port:     %d\n", pgConn.Port)
	fmt.Printf("  Database: '%s'\n", pgConn.Database)
	fmt.Printf("  Username: '%s'\n", pgConn.Username)
	fmt.Printf("  SSL Mode: '%s'\n", pgConn.SSLMode)
	fmt.Printf("  ID:       '%s'\n", connID)
	fmt.Println()

	// Get password from keyring
	password, err := keyringProvider.Get(ctx, "connection-"+connID)
	if err != nil {
		log.Printf("⚠️  Failed to get password from keyring: %v", err)
		password = "Qwer1234"
	}

	// Set password and get DSN
	pgConn.Password = password

	fmt.Println("Generated Connection String (DSN):")
	fmt.Println("─────────────────────────────────────────────────────────────")
	// Mask password in display
	maskedDSN := password
	for i := range password {
		if i > 0 && i < len(password)-1 {
			maskedDSN = maskedDSN[:i] + "*" + maskedDSN[i+1:]
		}
	}
	fmt.Printf("  host=%s port=%d database=%s user=%s password=%s sslmode=%s\n\n",
		pgConn.Host, pgConn.Port, pgConn.Database, pgConn.Username, maskedDSN, pgConn.SSLMode)
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("🔍 Testing connection...")
	fmt.Println()

	// Test connection
	result, err := pgConn.Test(ctx)
	if err != nil {
		log.Fatalf("❌ Test failed with error: %v", err)
	}

	if result.Success {
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║                   ✅ CONNECTION SUCCESSFUL!                 ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Printf("  Latency:     %d ms\n", result.LatencyMs)
		fmt.Printf("  Version:     %s\n", result.DatabaseVersion)
		fmt.Println()
		os.Exit(0)
	} else {
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║                    ❌ CONNECTION FAILED                     ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Printf("  Error:       %s\n", result.Error)
		fmt.Println()
		os.Exit(1)
	}
}
