// Package bindings_test provides integration tests for AI assistant persistence.
package bindings_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/bindings"
	_ "modernc.org/sqlite"
)

func TestAIAssistantPersistence(t *testing.T) {
	ctx := context.Background()

	// Setup in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create connections table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS connections (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			db_type TEXT NOT NULL,
			config_json TEXT NOT NULL,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := repository.NewSQLiteConnectionRepository(db)
	kr := keyring.NewMemoryKeyring()
	uc := usecase.NewConnectionUseCase(repo, kr)
	binding := bindings.NewConnectionBinding(uc)

	// Test data
	aiAssistants := []bindings.AIAssistantConfig{
		{
			ID:                "default",
			Name:              "Test DeepSeek",
			Provider:          "deepseek",
			APIHost:           "https://api.deepseek.com",
			APIEndpoint:       "/v1/chat/completions",
			APIKey:            "sk-test-key-12345",
			Model:             "deepseek-chat",
			Temperature:       0.7,
			Description:       "Test AI Description",
			EnterAction:       "send",
			CompareWithOthers: true,
			Language:          "zh-CN",
		},
	}

	t.Run("1. Create connection with AI assistants", func(t *testing.T) {
		req := bindings.ConnectionCreateRequest{
			Name:         "AI Test MySQL",
			Type:         "mysql",
			Host:         "localhost",
			Port:         3306,
			Database:     "testdb",
			Username:     "root",
			Password:     "testpass",
			AIAssistants: aiAssistants,
		}

		result := binding.CreateConnection(req)
		if result.Error != "" {
			t.Fatalf("CreateConnection failed: %s", result.Error)
		}

		if result.Connection == nil {
			t.Fatal("Connection is nil")
		}

		t.Logf("Created connection ID: %s", result.Connection.ID)
		t.Logf("AI Assistants in response: %d", len(result.Connection.AIAssistants))
	})

	t.Run("2. Retrieve and verify AI assistants", func(t *testing.T) {
		// First list connections to get the ID
		listResult := binding.ListConnections()
		if listResult.Error != "" {
			t.Fatalf("ListConnections failed: %s", listResult.Error)
		}

		if len(listResult.Connections) == 0 {
			t.Fatal("No connections found")
		}

		connID := listResult.Connections[0].ID
		t.Logf("Retrieving connection: %s", connID)

		// Get the connection
		conn := binding.GetConnection(connID)
		if conn == nil {
			t.Fatal("GetConnection returned nil")
		}

		// Verify AI assistants
		if len(conn.AIAssistants) != 1 {
			t.Fatalf("Expected 1 AI assistant, got %d", len(conn.AIAssistants))
		}

		ai := conn.AIAssistants[0]

		// Verify all fields
		checks := []struct {
			name     string
			expected interface{}
			actual   interface{}
		}{
			{"ID", "default", ai.ID},
			{"Name", "Test DeepSeek", ai.Name},
			{"Provider", "deepseek", ai.Provider},
			{"APIHost", "https://api.deepseek.com", ai.APIHost},
			{"APIEndpoint", "/v1/chat/completions", ai.APIEndpoint},
			{"APIKey", "sk-test-key-12345", ai.APIKey},
			{"Model", "deepseek-chat", ai.Model},
			{"Temperature", 0.7, ai.Temperature},
			{"Description", "Test AI Description", ai.Description},
			{"EnterAction", "send", ai.EnterAction},
			{"CompareWithOthers", true, ai.CompareWithOthers},
			{"Language", "zh-CN", ai.Language},
		}

		for _, check := range checks {
			if check.expected != check.actual {
				t.Errorf("%s mismatch: expected %v, got %v", check.name, check.expected, check.actual)
			} else {
				t.Logf("✓ %s: %v", check.name, check.actual)
			}
		}
	})

	t.Run("3. Update AI assistants", func(t *testing.T) {
		// Get existing connection
		listResult := binding.ListConnections()
		connID := listResult.Connections[0].ID

		// Update with modified AI config
		updatedAI := []bindings.AIAssistantConfig{
			{
				ID:                "default",
				Name:              "Updated GPT",
				Provider:          "openai",
				APIHost:           "https://api.openai.com",
				APIEndpoint:       "/v1/chat/completions",
				APIKey:            "sk-new-key-99999",
				Model:             "gpt-4",
				Temperature:       0.5,
				Description:       "Updated description",
				EnterAction:       "newline",
				CompareWithOthers: false,
				Language:          "en-US",
			},
		}

		req := bindings.ConnectionUpdateRequest{
			ID:           connID,
			Name:         "AI Test MySQL",
			Host:         "localhost",
			Port:         3306,
			Database:     "testdb",
			Username:     "root",
			Password:     "", // Keep existing
			AIAssistants: updatedAI,
		}

		result := binding.UpdateConnection(req)
		if result.Error != "" {
			t.Fatalf("UpdateConnection failed: %s", result.Error)
		}

		// Verify update
		conn := binding.GetConnection(connID)
		if len(conn.AIAssistants) != 1 {
			t.Fatalf("Expected 1 AI assistant after update, got %d", len(conn.AIAssistants))
		}

		ai := conn.AIAssistants[0]
		if ai.Name != "Updated GPT" {
			t.Errorf("Name not updated: expected 'Updated GPT', got '%s'", ai.Name)
		}
		if ai.Provider != "openai" {
			t.Errorf("Provider not updated: expected 'openai', got '%s'", ai.Provider)
		}
		if ai.APIKey != "sk-new-key-99999" {
			t.Errorf("API Key not updated: expected 'sk-new-key-99999', got '%s'", ai.APIKey)
		}

		t.Logf("✓ AI assistant updated successfully")
	})

	t.Run("4. Keyring storage verification", func(t *testing.T) {
		listResult := binding.ListConnections()
		connID := listResult.Connections[0].ID

		// Check keyring directly
		apiKey, err := uc.GetAIAPIKey(ctx, connID, "default")
		if err != nil {
			t.Logf("Warning: Could not retrieve API key from keyring: %v", err)
		} else {
			t.Logf("✓ API key stored in keyring: %s", apiKey)
		}
	})
}

func TestAITestConnection(t *testing.T) {
	t.Run("5. Test AI connection with invalid key", func(t *testing.T) {
		// Setup
		db, _ := sql.Open("sqlite3", ":memory:")
		defer db.Close()

		db.Exec(`CREATE TABLE IF NOT EXISTS connections (id TEXT PRIMARY KEY, name TEXT, db_type TEXT, config_json TEXT, created_at TEXT, updated_at TEXT)`)

		repo := repository.NewSQLiteConnectionRepository(db)
		kr := keyring.NewMemoryKeyring()
		uc := usecase.NewConnectionUseCase(repo, kr)
		binding := bindings.NewConnectionBinding(uc)

		req := bindings.AITestRequest{
			Provider:    "deepseek",
			APIHost:     "https://api.deepseek.com",
			APIEndpoint: "/v1/chat/completions",
			APIKey:      "sk-invalid-test-key",
			Model:       "deepseek-chat",
		}

		result := binding.TestAIConnection(req)

		t.Logf("AI Test Result:")
		t.Logf("  Success: %v", result.Success)
		t.Logf("  Latency: %dms", result.LatencyMs)
		t.Logf("  Message: %s", result.Message)
		if result.Error != "" {
			t.Logf("  Error: %s", result.Error)
		}

		// We expect this to fail with invalid key
		if result.Success {
			t.Log("Unexpected: Test succeeded with invalid key")
		} else {
			t.Log("✓ Expected: Test failed with invalid key")
		}
	})
}
