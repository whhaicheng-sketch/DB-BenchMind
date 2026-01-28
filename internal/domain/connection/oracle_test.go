// Implements: REQ-CONN-002
package connection

import (
	"testing"
)

// TestOracleConnection_Validate tests Oracle connection validation.
func TestOracleConnection_Validate(t *testing.T) {
	tests := []struct {
		name    string
		conn    *OracleConnection
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid connection with service_name",
			conn: &OracleConnection{
				BaseConnection: BaseConnection{Name: "test-conn"},
				Host:           "localhost",
				Port:           1521,
				ServiceName:    "ORCL",
				Username:       "system",
			},
			wantErr: false,
		},
		{
			name: "valid connection with SID",
			conn: &OracleConnection{
				BaseConnection: BaseConnection{Name: "test-conn"},
				Host:           "localhost",
				Port:           1521,
				SID:            "XE",
				Username:       "system",
			},
			wantErr: false,
		},
		{
			name: "missing service_name and SID",
			conn: &OracleConnection{
				BaseConnection: BaseConnection{Name: "test-conn"},
				Host:           "localhost",
				Port:           1521,
				Username:       "system",
			},
			wantErr: true,
			errMsg:  "either service_name or sid must be specified",
		},
		{
			name: "both service_name and SID provided",
			conn: &OracleConnection{
				BaseConnection: BaseConnection{Name: "test-conn"},
				Host:           "localhost",
				Port:           1521,
				ServiceName:    "ORCL",
				SID:            "XE",
				Username:       "system",
			},
			wantErr: true,
			errMsg:  "service_name and sid are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conn.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errMsg)
					return
				}
				errStr := err.Error()
				// Simple substring check
				found := false
				if len(errStr) >= len(tt.errMsg) {
					for i := 0; i <= len(errStr)-len(tt.errMsg); i++ {
						if errStr[i:i+len(tt.errMsg)] == tt.errMsg {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
