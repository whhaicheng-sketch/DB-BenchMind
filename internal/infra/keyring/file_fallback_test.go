// Implements: Keyring tests (file-based fallback)
package keyring

import (
	"context"
	"testing"
)

// TestFileFallback_SetAndGet tests Set and Get operations.
func TestFileFallback_SetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	provider, err := NewFileFallback(tmpDir, "test-password")
	if err != nil {
		t.Fatalf("NewFileFallback() failed: %v", err)
	}
	ctx := context.Background()

	tests := []struct {
		name     string
		key      string
		password string
	}{
		{
			name:     "simple password",
			key:      "conn-1",
			password: "secret123",
		},
		{
			name:     "password with special characters",
			key:      "conn-2",
			password: "p@$$w0rd!#$%",
		},
		{
			name:     "long password",
			key:      "conn-3",
			password: string(make([]byte, 1000)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set
			err := provider.Set(ctx, tt.key, tt.password)
			if err != nil {
				t.Errorf("Set() error = %v", err)
				return
			}

			// Get
			got, err := provider.Get(ctx, tt.key)
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}

			if got != tt.password {
				t.Errorf("Get() = %q, want %q", got, tt.password)
			}
		})
	}
}

// TestFileFallback_Get_NotFound tests Get with non-existent key.
func TestFileFallback_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	provider, err := NewFileFallback(tmpDir, "test-password")
	if err != nil {
		t.Fatalf("NewFileFallback() failed: %v", err)
	}
	ctx := context.Background()

	_, err = provider.Get(ctx, "non-existent")
	if err == nil {
		t.Error("Get() should return error for non-existent key")
	}

	if !IsNotFound(err) {
		t.Errorf("Get() error type = %T, want ErrNotFound", err)
	}
}

// TestFileFallback_Delete tests Delete operation.
func TestFileFallback_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	provider, err := NewFileFallback(tmpDir, "test-password")
	if err != nil {
		t.Fatalf("NewFileFallback() failed: %v", err)
	}
	ctx := context.Background()

	// Set a password
	key := "delete-test"
	password := "to-be-deleted"
	if err := provider.Set(ctx, key, password); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Verify it exists
	got, err := provider.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() before delete failed: %v", err)
	}
	if got != password {
		t.Fatalf("Get() = %q, want %q", got, password)
	}

	// Delete
	if err := provider.Delete(ctx, key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = provider.Get(ctx, key)
	if err == nil {
		t.Error("Get() after delete should return error")
	}
	if !IsNotFound(err) {
		t.Errorf("Get() error type = %T, want ErrNotFound", err)
	}
}

// TestFileFallback_Available tests Available method.
func TestFileFallback_Available(t *testing.T) {
	tmpDir := t.TempDir()
	provider, err := NewFileFallback(tmpDir, "test-password")
	if err != nil {
		t.Fatalf("NewFileFallback() failed: %v", err)
	}
	ctx := context.Background()

	if !provider.Available(ctx) {
		t.Error("Available() = false, want true")
	}

	// Note: We skip testing non-writable directories as:
	// 1. Running as root might bypass permission checks
	// 2. Different OS behave differently with directory permissions
}

// TestFileFallback_Update tests updating an existing password.
func TestFileFallback_Update(t *testing.T) {
	tmpDir := t.TempDir()
	provider, err := NewFileFallback(tmpDir, "test-password")
	if err != nil {
		t.Fatalf("NewFileFallback() failed: %v", err)
	}
	ctx := context.Background()

	key := "update-test"
	oldPassword := "old-password"
	newPassword := "new-password"

	// Set initial password
	if err := provider.Set(ctx, key, oldPassword); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Update with new password
	if err := provider.Set(ctx, key, newPassword); err != nil {
		t.Fatalf("Set() update failed: %v", err)
	}

	// Verify new password
	got, err := provider.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if got != newPassword {
		t.Errorf("Get() = %q, want %q", got, newPassword)
	}
}

// TestFileFallback_DifferentPasswords tests that different keys have different passwords.
func TestFileFallback_DifferentPasswords(t *testing.T) {
	tmpDir := t.TempDir()
	provider, err := NewFileFallback(tmpDir, "test-password")
	if err != nil {
		t.Fatalf("NewFileFallback() failed: %v", err)
	}
	ctx := context.Background()

	passwords := map[string]string{
		"key-1": "password-1",
		"key-2": "password-2",
		"key-3": "password-3",
	}

	// Set all passwords
	for key, password := range passwords {
		if err := provider.Set(ctx, key, password); err != nil {
			t.Fatalf("Set(%s) failed: %v", key, err)
		}
	}

	// Verify all passwords
	for key, expectedPassword := range passwords {
		got, err := provider.Get(ctx, key)
		if err != nil {
			t.Errorf("Get(%s) error = %v", key, err)
			continue
		}
		if got != expectedPassword {
			t.Errorf("Get(%s) = %q, want %q", key, got, expectedPassword)
		}
	}
}
