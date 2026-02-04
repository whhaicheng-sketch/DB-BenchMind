// Package connection provides WinRM functionality tests.
package connection

import (
	"testing"
)

func TestWinRMConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *WinRMConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid WinRM config (HTTP)",
			config: &WinRMConfig{
				Enabled:  true,
				Host:     "192.168.1.100",
				Port:     5985,
				Username: "",
				Password: "",
				UseHTTPS: false,
			},
			wantErr: false,
		},
		{
			name: "valid WinRM config (HTTPS)",
			config: &WinRMConfig{
				Enabled:  true,
				Host:     "192.168.1.100",
				Port:     5986,
				Username: "administrator",
				Password: "password",
				UseHTTPS: true,
			},
			wantErr: false,
		},
		{
			name: "disabled WinRM config",
			config: &WinRMConfig{
				Enabled:  false,
				Host:     "",
				Port:     0,
				UseHTTPS: false,
			},
			wantErr: false,
		},
		{
			name: "invalid port - zero",
			config: &WinRMConfig{
				Enabled:  true,
				Host:    "192.168.1.100",
				Port:    0,
				UseHTTPS: false,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			config: &WinRMConfig{
				Enabled:  true,
				Host:    "192.168.1.100",
				Port:    99999,
				UseHTTPS: false,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "invalid port - HTTP with wrong port",
			config: &WinRMConfig{
				Enabled:  true,
				Host:    "192.168.1.100",
				Port:    5986,
				UseHTTPS: false,
			},
			wantErr: true,
			errMsg:  "HTTP requires port 5985",
		},
		{
			name: "invalid port - HTTPS with wrong port",
			config: &WinRMConfig{
				Enabled:  true,
				Host:    "192.168.1.100",
				Port:    5985,
				UseHTTPS: true,
			},
			wantErr: true,
			errMsg:  "HTTPS requires port 5986",
		},
		{
			name: "missing host",
			config: &WinRMConfig{
				Enabled:  true,
				Host:     "",
				Port:     5985,
				UseHTTPS: false,
			},
			wantErr: true,
			errMsg:  "host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
