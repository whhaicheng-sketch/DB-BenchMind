//go:build nopkgs

// Package keyring provides system keyring stub when go-keyring is unavailable.
// Implements: REQ-CONN-007 (fallback only)
package keyring

import (
	"context"
	"fmt"
)

// GoKeyring is a stub when go-keyring is not available.
type GoKeyring struct {
	fallback *FileFallback
}

// NewGoKeyring creates a new stub keyring that only uses fallback.
func NewGoKeyring(fallbackDir string) *GoKeyring {
	k := &GoKeyring{}

	if fallbackDir != "" {
		fallback, err := NewFileFallback(fallbackDir, "")
		if err == nil {
			k.fallback = fallback
		}
	}

	return k
}

// Set stores a password using fallback only.
func (k *GoKeyring) Set(ctx context.Context, key, password string) error {
	if k.fallback == nil {
		return fmt.Errorf("keyring not available and no fallback configured")
	}
	return k.fallback.Set(ctx, key, password)
}

// Get retrieves a password using fallback only.
func (k *GoKeyring) Get(ctx context.Context, key string) (string, error) {
	if k.fallback == nil {
		return "", &ErrNotFound{Key: key}
	}
	return k.fallback.Get(ctx, key)
}

// Delete removes a password using fallback only.
func (k *GoKeyring) Delete(ctx context.Context, key string) error {
	if k.fallback == nil {
		return &ErrNotFound{Key: key}
	}
	return k.fallback.Delete(ctx, key)
}

// Available always returns true (fallback is used).
func (k *GoKeyring) Available(ctx context.Context) bool {
	return k.fallback != nil && k.fallback.Available(ctx)
}

// GetFallback returns the fallback provider.
func (k *GoKeyring) GetFallback() *FileFallback {
	return k.fallback
}
