// Implements: REQ-CONN-002
package connection

import (
	"testing"
)

// TestPostgreSQLConnection_Validate tests PostgreSQL connection validation.
func TestPostgreSQLConnection_Validate(t *testing.T) {
	tests := []struct {
		name    string
		conn    *PostgreSQLConnection
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid connection",
			conn: &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "test-conn"},
				Host:           "localhost",
				Port:           5432,
				Database:       "testdb",
				Username:       "postgres",
				SSLMode:        "require",
			},
			wantErr: false,
		},
		{
			name: "valid connection with empty SSL mode",
			conn: &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "test-conn"},
				Host:           "localhost",
				Port:           5432,
				Database:       "testdb",
				Username:       "postgres",
				SSLMode:        "",
			},
			wantErr: false,
		},
		{
			name: "invalid ssl_mode",
			conn: &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "test-conn"},
				Host:           "localhost",
				Port:           5432,
				Database:       "testdb",
				Username:       "postgres",
				SSLMode:        "invalid",
			},
			wantErr: true,
			errMsg:  "ssl_mode must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conn.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				errStr := err.Error()
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
