// Package keyring provides secure password storage using system keyring.
// Implements: REQ-CONN-006, REQ-CONN-007
package keyring

import (
	"context"
	"fmt"
)

// Provider defines the interface for password storage operations.
// This interface is defined by the infrastructure layer for keyring operations.
type Provider interface {
	// Set stores a password for the given key.
	// The key typically is the connection ID.
	Set(ctx context.Context, key, password string) error

	// Get retrieves a password for the given key.
	// Returns an error if the key is not found.
	Get(ctx context.Context, key string) (string, error)

	// Delete removes a password for the given key.
	// Returns an error if the key is not found.
	Delete(ctx context.Context, key string) error

	// Available checks if the keyring provider is available.
	// Returns false if the system keyring is not accessible.
	Available(ctx context.Context) bool
}

// ErrNotFound is returned when a key is not found in the keyring.
type ErrNotFound struct {
	Key string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("key not found: %s", e.Key)
}

// IsNotFound checks if an error is ErrNotFound.
func IsNotFound(err error) bool {
	_, ok := err.(*ErrNotFound)
	return ok
}
