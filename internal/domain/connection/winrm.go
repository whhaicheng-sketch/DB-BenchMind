// Package connection provides WinRM functionality for Windows Server connections.
package connection

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/masterzen/winrm"
)

// WinRMConfig represents WinRM configuration.
// Implements: REQ-WINRM-001 ~ REQ-WINRM-015
type WinRMConfig struct {
	Enabled  bool   `json:"enabled"`    // Whether WinRM is enabled
	Host     string `json:"host"`       // WinRM host (use Database Host)
	Port     int    `json:"port"`       // WinRM port (5985 HTTP, 5986 HTTPS)
	Username string `json:"username"`   // Username (empty = current Windows user)
	Password string `json:"-"`          // Password (stored in keyring)
	UseHTTPS bool   `json:"use_https"`  // Whether to use HTTPS
}

// Validate validates the WinRM configuration.
func (c *WinRMConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	// Validate standard ports
	if c.UseHTTPS && c.Port != 5986 {
		return fmt.Errorf("HTTPS requires port 5986, got %d", c.Port)
	}
	if !c.UseHTTPS && c.Port != 5985 {
		return fmt.Errorf("HTTP requires port 5985, got %d", c.Port)
	}

	return nil
}

// WinRMClient manages a WinRM connection.
type WinRMClient struct {
	config *WinRMConfig
	client *winrm.Client
}

// NewWinRMClient creates a new WinRM client.
// Returns an error if the client cannot be created.
func NewWinRMClient(ctx context.Context, config *WinRMConfig) (*WinRMClient, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("WinRM is not enabled")
	}

	slog.Info("WinRM: Creating client",
		"op", "winrm_create",
		"host", config.Host,
		"port", config.Port,
		"https", config.UseHTTPS,
		"username", config.Username)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid WinRM configuration: %w", err)
	}

	// Create WinRM endpoint with 60s timeout
	endpoint := winrm.NewEndpoint(
		config.Host,
		config.Port,
		config.UseHTTPS,
		false, // insecure
		nil,   // Cacert
		nil,   // cert
		nil,   // key
		60*time.Second, // timeout
	)

	// Create WinRM client
	// Note: For integrated Windows auth, username and password should be empty
	username := config.Username
	password := config.Password
	if username == "" {
		// Use current Windows user (integrated auth)
		username = ""
		password = ""
	}

	client, err := winrm.NewClientWithParameters(endpoint, username, password, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create WinRM client: %w", err)
	}

	slog.Info("WinRM: Client created successfully",
		"op", "winrm_created",
		"host", config.Host,
		"port", config.Port)

	return &WinRMClient{
		config: config,
		client: client,
	}, nil
}

// Test tests the WinRM connection.
// Returns TestResult containing success/failure, latency, error.
func (c *WinRMClient) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	// Simple WinRM test: execute "hostname" command
	shell, err := c.client.CreateShell()
	if err != nil {
		latency := time.Since(start).Milliseconds()
		return &TestResult{
			Success:   false,
			LatencyMs: latency,
			Error:     fmt.Sprintf("failed to create shell: %v", err),
		}, nil
	}
	defer shell.Close()

	// Execute hostname command with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = shell.ExecuteWithContext(ctx, "hostname")
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &TestResult{
			Success:   false,
			LatencyMs: latency,
			Error:     fmt.Sprintf("WinRM command failed: %v", err),
		}, nil
	}

	slog.Info("WinRM: Connection test successful",
		"op", "winrm_test_success",
		"latency_ms", latency)

	return &TestResult{
		Success:         true,
		LatencyMs:       latency,
		DatabaseVersion: "WinRM Connected",
	}, nil
}

// Close closes the WinRM client.
func (c *WinRMClient) Close() error {
	// WinRM client doesn't have explicit close method
	// Resources are cleaned up automatically
	slog.Info("WinRM: Client closed",
		"op", "winrm_close",
		"host", c.config.Host)
	return nil
}
