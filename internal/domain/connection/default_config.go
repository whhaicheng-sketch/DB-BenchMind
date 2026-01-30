// Package connection provides default connection configuration.
package connection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// DefaultConnectionConfig stores default connection parameters for each database type.
type DefaultConnectionConfig struct {
	// Version is the configuration version.
	Version int `json:"version"`

	// MySQL stores default MySQL connection parameters.
	MySQL *MySQLDefaults `json:"mysql,omitempty"`

	// PostgreSQL stores default PostgreSQL connection parameters.
	PostgreSQL *PostgreSQLDefaults `json:"postgresql,omitempty"`

	// Oracle stores default Oracle connection parameters.
	Oracle *OracleDefaults `json:"oracle,omitempty"`

	// SQLServer stores default SQL Server connection parameters.
	SQLServer *SQLServerDefaults `json:"sqlserver,omitempty"`
}

// MySQLDefaults stores default MySQL connection parameters.
type MySQLDefaults struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	SSLMode  string `json:"ssl_mode"`
}

// PostgreSQLDefaults stores default PostgreSQL connection parameters.
type PostgreSQLDefaults struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	SSLMode  string `json:"ssl_mode"`
}

// OracleDefaults stores default Oracle connection parameters.
type OracleDefaults struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	ServiceName string `json:"service_name,omitempty"`
	SID         string `json:"sid,omitempty"`
	Username    string `json:"username"`
}

// SQLServerDefaults stores default SQL Server connection parameters.
type SQLServerDefaults struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	// TrustServerCertificate is stored as bool
	TrustServerCertificate bool `json:"trust_server_certificate"`
}

var (
	defaultConfig     *DefaultConnectionConfig
	defaultConfigOnce sync.Once
	defaultConfigMu   sync.RWMutex
)

// GetDefaultConfigPath returns the path to the default connection config file.
func GetDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".db-benchmind", "default_connection.json"), nil
}

// LoadDefaultConnectionConfig loads the default connection configuration from file.
func LoadDefaultConnectionConfig() (*DefaultConnectionConfig, error) {
	defaultConfigOnce.Do(func() {
		defaultConfig = &DefaultConnectionConfig{Version: 1}
	})

	configPath, err := GetDefaultConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read default config: %w", err)
	}

	defaultConfigMu.Lock()
	defer defaultConfigMu.Unlock()

	if err := json.Unmarshal(data, defaultConfig); err != nil {
		return nil, fmt.Errorf("parse default config: %w", err)
	}

	return defaultConfig, nil
}

// SaveDefaultConnectionConfig saves the default connection configuration to file.
func SaveDefaultConnectionConfig(config *DefaultConnectionConfig) error {
	configPath, err := GetDefaultConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	defaultConfigMu.Lock()
	defer defaultConfigMu.Unlock()

	defaultConfig = config

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// SaveConnectionAsDefault saves a connection as the default for its type.
func SaveConnectionAsDefault(conn Connection) error {
	config, err := LoadDefaultConnectionConfig()
	if err != nil {
		return err
	}

	config.Version = 1

	switch c := conn.(type) {
	case *MySQLConnection:
		config.MySQL = &MySQLDefaults{
			Host:     c.Host,
			Port:     c.Port,
			Database: c.Database,
			Username: c.Username,
			SSLMode:  c.SSLMode,
		}
	case *PostgreSQLConnection:
		config.PostgreSQL = &PostgreSQLDefaults{
			Host:     c.Host,
			Port:     c.Port,
			Database: c.Database,
			Username: c.Username,
			SSLMode:  c.SSLMode,
		}
	case *OracleConnection:
		config.Oracle = &OracleDefaults{
			Host:        c.Host,
			Port:        c.Port,
			ServiceName: c.ServiceName,
			SID:         c.SID,
			Username:    c.Username,
		}
	case *SQLServerConnection:
		config.SQLServer = &SQLServerDefaults{
			Host:                   c.Host,
			Port:                   c.Port,
			Database:               c.Database,
			Username:               c.Username,
			TrustServerCertificate: c.TrustServerCertificate,
		}
	default:
		return fmt.Errorf("unsupported connection type: %T", conn)
	}

	return SaveDefaultConnectionConfig(config)
}
