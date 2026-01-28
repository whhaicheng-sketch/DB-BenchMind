// Package keyring provides encrypted file-based fallback for password storage.
// Implements: REQ-CONN-007 (fallback when keyring is unavailable)
package keyring

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileFallback provides encrypted file-based password storage.
// This is used when system keyring is not available (REQ-CONN-007).
type FileFallback struct {
	keyFile string // Path to the encryption key file
	dataDir string // Directory for encrypted password files
	secret  []byte // Derived encryption key
}

// NewFileFallback creates a new file-based keyring fallback.
// The masterPassword is used to derive the encryption key.
// If masterPassword is empty, a default password is used (less secure).
func NewFileFallback(dataDir, masterPassword string) (*FileFallback, error) {
	if dataDir == "" {
		return nil, errors.New("data directory is required")
	}

	// Create directory if not exists
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	// Use default password if not provided (less secure but functional)
	if masterPassword == "" {
		masterPassword = "db-benchmind-default-key" // Less secure
	}

	// Derive encryption key from master password
	// In production, you might want to use a more secure approach
	secret := deriveKey(masterPassword, "db-benchmind-salt")

	return &FileFallback{
		keyFile: filepath.Join(dataDir, ".key"),
		dataDir: dataDir,
		secret:  secret,
	}, nil
}

// deriveKey derives a 32-byte encryption key from a password using repeated SHA256.
// This is a simplified derivation (not as secure as PBKDF2, but functional).
func deriveKey(password, salt string) []byte {
	// Simple key derivation using SHA256
	// In production, use PBKDF2, scrypt, or Argon2
	key := []byte(password + salt)
	for i := 0; i < 10000; i++ {
		hash := sha256.Sum256(key)
		key = hash[:]
	}
	return key
}

// Set stores an encrypted password for the given key.
func (f *FileFallback) Set(ctx context.Context, key, password string) error {
	// Encrypt password
	encrypted, err := f.encrypt(password)
	if err != nil {
		return fmt.Errorf("encrypt password: %w", err)
	}

	// Save to file
	filePath := f.getPasswordPath(key)
	if err := os.WriteFile(filePath, encrypted, 0600); err != nil {
		return fmt.Errorf("write password file: %w", err)
	}

	return nil
}

// Get retrieves and decrypts a password for the given key.
func (f *FileFallback) Get(ctx context.Context, key string) (string, error) {
	filePath := f.getPasswordPath(key)

	// Read encrypted file
	encrypted, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", &ErrNotFound{Key: key}
		}
		return "", fmt.Errorf("read password file: %w", err)
	}

	// Decrypt
	password, err := f.decrypt(encrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt password: %w", err)
	}

	return password, nil
}

// Delete removes the password file for the given key.
func (f *FileFallback) Delete(ctx context.Context, key string) error {
	filePath := f.getPasswordPath(key)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return &ErrNotFound{Key: key}
		}
		return fmt.Errorf("delete password file: %w", err)
	}

	return nil
}

// Available checks if the file-based fallback is available.
// It's always available if we can write to the data directory.
func (f *FileFallback) Available(ctx context.Context) bool {
	// Try to create a test file
	testFile := filepath.Join(f.dataDir, ".available-test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return false
	}
	os.Remove(testFile)
	return true
}

// getPasswordPath returns the file path for a password key.
func (f *FileFallback) getPasswordPath(key string) string {
	// Use hex encoding to safely use the key as a filename
	safeKey := hex.EncodeToString([]byte(key))
	return filepath.Join(f.dataDir, safeKey+".enc")
}

// encrypt encrypts plaintext using AES-GCM.
func (f *FileFallback) encrypt(plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(f.secret)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and append nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// decrypt decrypts ciphertext using AES-GCM.
func (f *FileFallback) decrypt(ciphertext []byte) (string, error) {
	block, err := aes.NewCipher(f.secret)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce
	nonce, cipher := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// passwordFile represents the structure of a password file.
// This is for potential future use if we want to store metadata.
type passwordFile struct {
	CreatedAt int64  `json:"created_at"`
	Password  string `json:"password"` // Encrypted
}

// savePasswordFile saves a password file with metadata.
func (f *FileFallback) savePasswordFile(key string, pf *passwordFile) error {
	filePath := f.getPasswordPath(key)
	data, err := json.Marshal(pf)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0600)
}

// loadPasswordFile loads a password file.
func (f *FileFallback) loadPasswordFile(key string) (*passwordFile, error) {
	filePath := f.getPasswordPath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var pf passwordFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, err
	}

	return &pf, nil
}
