// Package keyring provides secure password storage using system keyring.
package keyring

import (
	"context"
	"sync"
)

// MemoryKeyring is an in-memory implementation of Provider for testing.
type MemoryKeyring struct {
	mu    sync.RWMutex
	store map[string]string
}

// NewMemoryKeyring creates a new in-memory keyring for testing.
func NewMemoryKeyring() *MemoryKeyring {
	return &MemoryKeyring{
		store: make(map[string]string),
	}
}

// Set stores a password for the given key.
func (m *MemoryKeyring) Set(ctx context.Context, key, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = password
	return nil
}

// Get retrieves a password for the given key.
func (m *MemoryKeyring) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.store[key]; ok {
		return val, nil
	}
	return "", &ErrNotFound{Key: key}
}

// Delete removes a password for the given key.
func (m *MemoryKeyring) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.store[key]; !ok {
		return &ErrNotFound{Key: key}
	}
	delete(m.store, key)
	return nil
}

// Available always returns true for memory keyring.
func (m *MemoryKeyring) Available(ctx context.Context) bool {
	return true
}
