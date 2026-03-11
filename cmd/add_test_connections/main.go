package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

func main() {
	ctx := context.Background()

	// Initialize database
	dbPath := "./data/db-benchmind.db"
	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize keyring
	dataDir := "./data"
	keyringProvider, err := keyring.NewFileFallback(dataDir, "")
	if err != nil {
		log.Fatalf("Failed to initialize keyring: %v", err)
	}

	// Initialize repository and usecase
	connRepo := repository.NewSQLiteConnectionRepository(db)
	connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

	// Define test connections
	testConnections := []struct {
		name     string
		connType string
		createFn func() connection.Connection
	}{
		{
			name:     "Oracle Test",
			connType: "oracle",
			createFn: func() connection.Connection {
				conn := usecase.NewOracleConnection(
					"Oracle Test",
					"192.168.134.129",
					"",     // service_name (empty, use SID)
					"orcl", // sid
					"system",
					1521,
				)
				conn.SetPassword("Qwer1234")
				return conn
			},
		},
	}

	// Create each connection
	for _, tc := range testConnections {
		fmt.Printf("Creating connection: %s (%s)\n", tc.name, tc.connType)

		// Check if connection already exists
		exists, err := connRepo.ExistsByName(ctx, tc.name, "")
		if err != nil {
			log.Printf("Failed to check existence for %s: %v", tc.name, err)
			continue
		}
		if exists {
			fmt.Printf("  ⚠ Connection '%s' already exists, skipping\n", tc.name)
			continue
		}

		// Create connection object (ID is auto-generated in New*Connection functions)
		conn := tc.createFn()

		// Create connection
		err = connUC.CreateConnection(ctx, conn)
		if err != nil {
			log.Printf("  ✗ Failed to save connection %s: %v", tc.name, err)
			continue
		}

		fmt.Printf("  ✓ Connection '%s' created successfully (ID: %s)\n", tc.name, conn.GetID())
	}

	fmt.Println("\nAll test connections have been processed.")
	os.Exit(0)
}
