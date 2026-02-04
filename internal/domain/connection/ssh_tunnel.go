// Package connection provides SSH tunnel functionality for database connections.
package connection

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHTunnelConfig represents SSH tunnel configuration.
type SSHTunnelConfig struct {
	Enabled  bool   `json:"enabled"`   // Whether SSH tunnel is enabled
	Host     string `json:"host"`      // SSH server host
	Port     int    `json:"port"`      // SSH server port (default 22)
	Username string `json:"username"`  // SSH username
	Password string `json:"-"`         // SSH password (stored in keyring)
	KeyPath  string `json:"key_path"`  // SSH private key path (optional)
	LocalPort int    `json:"local_port"` // Local binding port (0 = auto-assign)
}

// SSHTunnel manages an SSH tunnel connection.
type SSHTunnel struct {
	config    *SSHTunnelConfig
	client    *ssh.Client
	listener  net.Listener
	localPort int
	cancel    context.CancelFunc
	mu        sync.Mutex
	closed    bool
}

// NewSSHTunnel creates a new SSH tunnel.
// Returns an error if the tunnel cannot be established.
func NewSSHTunnel(ctx context.Context, config *SSHTunnelConfig, remoteHost string, remotePort int) (*SSHTunnel, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("SSH tunnel is not enabled")
	}

	slog.Info("SSH: Creating tunnel",
		"op", "ssh_tunnel_create",
		"ssh_host", config.Host,
		"ssh_port", config.Port,
		"remote_host", remoteHost,
		"remote_port", remotePort,
		"username", config.Username)

	// Validate required fields
	if config.Host == "" {
		return nil, fmt.Errorf("SSH host is required")
	}
	if config.Username == "" {
		return nil, fmt.Errorf("SSH username is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return nil, fmt.Errorf("SSH port must be between 1 and 65535")
	}

	// Create SSH client config
	sshConfig, err := config.buildSSHConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH config: %w", err)
	}

	// Connect to SSH server
	sshAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(30 * time.Second)
	}

	client, err := net.DialTimeout("tcp", sshAddr, time.Until(deadline))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server %s: %w", sshAddr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(client, sshAddr, sshConfig)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("SSH handshake failed: %w", err)
	}

	sshClient := ssh.NewClient(sshConn, chans, reqs)

	// Create local listener
	localPort := config.LocalPort
	if localPort == 0 {
		// Auto-assign port
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			sshClient.Close()
			return nil, fmt.Errorf("failed to create local listener: %w", err)
		}
		localPort = listener.Addr().(*net.TCPAddr).Port
		slog.Info("SSH: Auto-assigned local port", "port", localPort)

		tunnel := &SSHTunnel{
			config:    config,
			client:    sshClient,
			listener:  listener,
			localPort: localPort,
		}

		// Start forwarding in background
		if err := tunnel.startForwarding(remoteHost, remotePort); err != nil {
			listener.Close()
			sshClient.Close()
			return nil, err
		}

		slog.Info("SSH: Tunnel created successfully",
			"op", "ssh_tunnel_created",
			"local_port", localPort,
			"remote_target", fmt.Sprintf("%s:%d", remoteHost, remotePort))

		return tunnel, nil
	}

	// Use specified port
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to listen on port %d: %w", localPort, err)
	}

	tunnel := &SSHTunnel{
		config:    config,
		client:    sshClient,
		listener:  listener,
		localPort: localPort,
	}

	if err := tunnel.startForwarding(remoteHost, remotePort); err != nil {
		listener.Close()
		sshClient.Close()
		return nil, err
	}

	slog.Info("SSH: Tunnel created successfully",
		"op", "ssh_tunnel_created",
		"local_port", localPort,
		"remote_target", fmt.Sprintf("%s:%d", remoteHost, remotePort))

	return tunnel, nil
}

// buildSSHConfig creates SSH client config from SSHTunnelConfig.
func (c *SSHTunnelConfig) buildSSHConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User: c.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 30 * time.Second,
	}

	// Use password auth if password is provided
	if c.Password != "" {
		slog.Info("SSH: Using password authentication",
			"username", c.Username,
			"password_length", len(c.Password))
		config.Auth = append(config.Auth, ssh.Password(c.Password))
	}

	// Use key auth if key path is provided
	if c.KeyPath != "" {
		key, err := ssh.ParsePrivateKey([]byte(c.KeyPath))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(key))
	}

	// At least one auth method is required
	if len(config.Auth) == 0 {
		slog.Error("SSH: No authentication method configured",
			"has_password", c.Password != "",
			"has_key", c.KeyPath != "")
		return nil, fmt.Errorf("SSH requires either password or private key")
	}

	slog.Info("SSH: Auth methods configured", "count", len(config.Auth))
	return config, nil
}

// startForwarding starts forwarding connections through the tunnel.
func (t *SSHTunnel) startForwarding(remoteHost string, remotePort int) error {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := t.listener.Accept()
				if err != nil {
					if t.closed {
						return
					}
					slog.Error("SSH: Failed to accept connection", "error", err)
					continue
				}

				go t.forwardConnection(conn, remoteHost, remotePort)
			}
		}
	}()

	return nil
}

// forwardConnection forwards a single connection through the tunnel.
func (t *SSHTunnel) forwardConnection(localConn net.Conn, remoteHost string, remotePort int) {
	defer localConn.Close()

	t.mu.Lock()
	client := t.client
	t.mu.Unlock()

	if client == nil {
		slog.Warn("SSH: Tunnel client is nil, cannot forward")
		return
	}

	remoteAddr := fmt.Sprintf("%s:%d", remoteHost, remotePort)
	remoteConn, err := client.Dial("tcp", remoteAddr)
	if err != nil {
		slog.Error("SSH: Failed to dial remote", "error", err, "remote", remoteAddr)
		return
	}
	defer remoteConn.Close()

	// Bidirectional copy
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	// Wait for both directions to finish
	<-done
	<-done
}

// GetLocalPort returns the local port number of the tunnel.
func (t *SSHTunnel) GetLocalPort() int {
	return t.localPort
}

// Close closes the SSH tunnel and releases resources.
func (t *SSHTunnel) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	slog.Info("SSH: Closing tunnel",
		"op", "ssh_tunnel_close",
		"local_port", t.localPort)

	if t.cancel != nil {
		t.cancel()
	}

	var errs []error

	if t.listener != nil {
		if err := t.listener.Close(); err != nil {
			errs = append(errs, fmt.Errorf("listener close: %w", err))
		}
	}

	if t.client != nil {
		if err := t.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("client close: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing tunnel: %v", errs)
	}

	return nil
}

// IsClosed returns whether the tunnel is closed.
func (t *SSHTunnel) IsClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}
