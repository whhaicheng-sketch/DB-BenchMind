// Package adapter provides unit tests for Swingbench adapter.
package adapter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// TestSwingbenchAdapter_Type tests the Type method.
func TestSwingbenchAdapter_Type(t *testing.T) {
	adapter := NewSwingbenchAdapter()
	assert.Equal(t, AdapterTypeSwingbench, adapter.Type())
}

// TestSwingbenchAdapter_BuildPrepareCommand tests building prepare command.
func TestSwingbenchAdapter_BuildPrepareCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-1",
			Name: "Test Oracle",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "sys",
		Password:    "testpass",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"scale":        0.1, // Test with float scale (100MB)
			"threads":      32,
			"dba_username": "system",
			"dba_password": "testpass",
		},
		WorkDir: "/tmp/test",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, config)
	require.NoError(t, err)

	// Prepare must follow the exact hand-validated lifecycle:
	// cleanup -> verify cleanup -> bootstrap -> create -> generate -> allindexes.
	assert.Equal(t, "oracle_prepare_sequence", cmd.CmdLine)
	assert.Len(t, cmd.Commands, 6, "Prepare should have 6 steps")

	// Step 1: Cleanup existing environment
	step1 := cmd.Commands[0]
	assert.Equal(t, "Cleanup Existing Environment", step1.StepName)
	assert.Contains(t, step1.CmdLine, "sqlplus")
	assert.Contains(t, step1.CmdLine, "sys/testpass@//localhost:1521/ORCL as sysdba")
	assert.Contains(t, step1.CmdLine, "/tmp/db-benchmind-cleanup-soe-")

	// Step 2: Cleanup verification gate
	step2 := cmd.Commands[1]
	assert.Equal(t, "Verify Cleanup State", step2.StepName)
	assert.Contains(t, step2.CmdLine, "sqlplus")
	assert.Contains(t, step2.CmdLine, "sys/testpass@//localhost:1521/ORCL as sysdba")
	verifySQLFile := strings.TrimSpace(step2.CmdLine[strings.LastIndex(step2.CmdLine, "@")+1:])
	verifySQL, readErr := os.ReadFile(verifySQLFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(verifySQL), "SOE cleanup verification")

	// Step 3: Bootstrap SOE user and tablespaces
	step3 := cmd.Commands[2]
	assert.Equal(t, "Bootstrap SOE", step3.StepName)
	assert.Contains(t, step3.CmdLine, "sqlplus")
	assert.Contains(t, step3.CmdLine, "sys/testpass@//localhost:1521/ORCL as sysdba")
	assert.Contains(t, step3.CmdLine, "/tmp/db-benchmind-cts-oracle-")

	// Step 4: oewizard -create
	step4Create := cmd.Commands[3]
	assert.Equal(t, "Create Schema", step4Create.StepName)
	assert.Contains(t, step4Create.CmdLine, "oewizard")
	assert.Contains(t, step4Create.CmdLine, "-cl")
	assert.Contains(t, step4Create.CmdLine, "-cs //localhost:1521/ORCL")
	assert.Contains(t, step4Create.CmdLine, "-dba system")
	assert.Contains(t, step4Create.CmdLine, "-dbap testpass")
	assert.Contains(t, step4Create.CmdLine, "-u soe")
	assert.Contains(t, step4Create.CmdLine, "-p soe")
	assert.Contains(t, step4Create.CmdLine, "-scale 0.1")
	assert.Contains(t, step4Create.CmdLine, "-ts SOE")
	assert.Contains(t, step4Create.CmdLine, "-normalfile")
	assert.Contains(t, step4Create.CmdLine, "-create")
	assert.Contains(t, step4Create.CmdLine, "-v")
	assert.Contains(t, step4Create.CmdLine, "-debugf /tmp/test/oewizard-create-debug.log")
	assert.NotContains(t, step4Create.CmdLine, "-c oewizard.xml")
	assert.NotContains(t, step4Create.CmdLine, "-version")
	assert.NotContains(t, step4Create.CmdLine, "-dt")
	assert.NotContains(t, step4Create.CmdLine, "-nopart")
	assert.NotContains(t, step4Create.CmdLine, "-nocompress")

	// Step 5: oewizard -generate
	step5Generate := cmd.Commands[4]
	assert.Equal(t, "Generate Data", step5Generate.StepName)
	assert.Contains(t, step5Generate.CmdLine, "oewizard")
	assert.Contains(t, step5Generate.CmdLine, "-cl")
	assert.Contains(t, step5Generate.CmdLine, "-cs //localhost:1521/ORCL")
	assert.Contains(t, step5Generate.CmdLine, "-dba system")
	assert.Contains(t, step5Generate.CmdLine, "-dbap testpass")
	assert.Contains(t, step5Generate.CmdLine, "-u soe")
	assert.Contains(t, step5Generate.CmdLine, "-p soe")
	assert.Contains(t, step5Generate.CmdLine, "-scale 0.1")
	assert.Contains(t, step5Generate.CmdLine, "-ts SOE")
	assert.Contains(t, step5Generate.CmdLine, "-normalfile")
	assert.Contains(t, step5Generate.CmdLine, "-generate")
	assert.Contains(t, step5Generate.CmdLine, "-tc 32")
	assert.Contains(t, step5Generate.CmdLine, "-v")
	assert.Contains(t, step5Generate.CmdLine, "-debugf /tmp/test/oewizard-generate-debug.log")
	assert.NotContains(t, step5Generate.CmdLine, "-c oewizard.xml")
	assert.NotContains(t, step5Generate.CmdLine, "-version")
	assert.NotContains(t, step5Generate.CmdLine, "-dt")
	assert.NotContains(t, step5Generate.CmdLine, "-nopart")
	assert.NotContains(t, step5Generate.CmdLine, "-nocompress")

	// Step 6: allindexes
	step6Indexes := cmd.Commands[5]
	assert.Equal(t, "Build Indexes", step6Indexes.StepName)
	assert.Contains(t, step6Indexes.CmdLine, "oewizard")
	assert.Contains(t, step6Indexes.CmdLine, "-cl")
	assert.Contains(t, step6Indexes.CmdLine, "-cs //localhost:1521/ORCL")
	assert.Contains(t, step6Indexes.CmdLine, "-dba system")
	assert.Contains(t, step6Indexes.CmdLine, "-dbap testpass")
	assert.Contains(t, step6Indexes.CmdLine, "-u soe")
	assert.Contains(t, step6Indexes.CmdLine, "-p soe")
	assert.Contains(t, step6Indexes.CmdLine, "-scale 0.1")
	assert.Contains(t, step6Indexes.CmdLine, "-ts SOE")
	assert.Contains(t, step6Indexes.CmdLine, "-normalfile")
	assert.Contains(t, step6Indexes.CmdLine, "-allindexes")
	assert.Contains(t, step6Indexes.CmdLine, "-v")
	assert.Contains(t, step6Indexes.CmdLine, "-debugf /tmp/test/oewizard-indexes-debug.log")
	assert.NotContains(t, step6Indexes.CmdLine, "-c oewizard.xml")
	assert.NotContains(t, step6Indexes.CmdLine, "-version")
	assert.NotContains(t, step6Indexes.CmdLine, "-dt")
	assert.NotContains(t, step6Indexes.CmdLine, "-nopart")
	assert.NotContains(t, step6Indexes.CmdLine, "-nocompress")
}

func TestSwingbenchAdapter_BuildPrepareCommand_UsesOracleConnectAsForAllAdministrativeSteps(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-sysdba",
			Name: "Oracle SYSDBA",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "sys",
		Password:    "manager",
		ConnectAs:   "sysdba",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"scale":        0.1,
			"threads":      8,
			"dba_username": "system",
			"dba_password": "manager",
		},
		WorkDir: "/tmp/test-sysdba",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, config)
	require.NoError(t, err)
	require.Len(t, cmd.Commands, 6)

	assert.Contains(t, cmd.Commands[0].CmdLine, "sys/manager@//localhost:1521/ORCL as sysdba")
	assert.Contains(t, cmd.Commands[1].CmdLine, "sys/manager@//localhost:1521/ORCL as sysdba")
	assert.Contains(t, cmd.Commands[2].CmdLine, "sys/manager@//localhost:1521/ORCL as sysdba")
	assert.Contains(t, cmd.Commands[3].CmdLine, "-dba system")
	assert.Contains(t, cmd.Commands[4].CmdLine, "-dba system")
	assert.Contains(t, cmd.Commands[5].CmdLine, "-dba system")
}

func TestSwingbenchAdapter_BuildPrepareCommand_UsesDefaultSystemCredentialsForOewizard(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-sys-auto",
			Name: "Oracle SYS Auto",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "sys",
		Password:    "manager",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"scale":   0.1,
			"threads": 2,
		},
		WorkDir: "/tmp/test-sys-auto",
	})
	require.NoError(t, err)
	require.Len(t, cmd.Commands, 6)
	assert.Contains(t, cmd.Commands[3].CmdLine, "-dba system")
	assert.Contains(t, cmd.Commands[4].CmdLine, "-dba system")
	assert.Contains(t, cmd.Commands[5].CmdLine, "-dba system")
	assert.Contains(t, cmd.Commands[3].CmdLine, "-dbap manager")
}

func TestSwingbenchAdapter_BuildPrepareCommand_StartsWithFullCleanupStep(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "test-conn-rebuild", Name: "Oracle Rebuild"},
		Host:           "192.168.134.129",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "sys",
		Password:       "Qwer1234",
		ConnectAs:      "sysdba",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"scale":        1.0,
			"threads":      8,
			"dba_username": "system",
			"dba_password": "Qwer1234",
		},
		WorkDir: "/tmp/rebuild",
	})
	require.NoError(t, err)
	require.Len(t, cmd.Commands, 6)

	first := cmd.Commands[0]
	assert.Equal(t, "Cleanup Existing Environment", first.StepName)
	assert.Contains(t, first.CmdLine, "@/tmp/db-benchmind-cleanup-soe-")
	sqlFile := strings.TrimSpace(first.CmdLine[strings.LastIndex(first.CmdLine, "@")+1:])
	content, readErr := os.ReadFile(sqlFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "drop user SOE cascade")
	assert.Contains(t, string(content), "drop tablespace SOE_IDX including contents and datafiles")
	assert.Contains(t, string(content), "drop tablespace SOE including contents and datafiles")
}

func TestSwingbenchAdapter_BuildPrepareCommand_VerifiesCleanupBeforeBootstrapAndCreate(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "test-conn-verify", Name: "Oracle Verify"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "sys",
		Password:       "manager",
		ConnectAs:      "sysdba",
	}

	cmd, err := adapter.BuildPrepareCommand(ctx, &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"scale":        1.0,
			"threads":      4,
			"dba_username": "system",
			"dba_password": "manager",
		},
		WorkDir: "/tmp/oracle-verify",
	})
	require.NoError(t, err)
	require.Len(t, cmd.Commands, 6)

	assert.Equal(t, "Cleanup Existing Environment", cmd.Commands[0].StepName)
	assert.Equal(t, "Verify Cleanup State", cmd.Commands[1].StepName)
	assert.Equal(t, "Bootstrap SOE", cmd.Commands[2].StepName)
	assert.Equal(t, "Create Schema", cmd.Commands[3].StepName)
	assert.Equal(t, "Generate Data", cmd.Commands[4].StepName)
	assert.Equal(t, "Build Indexes", cmd.Commands[5].StepName)

	verifySQLFile := strings.TrimSpace(cmd.Commands[1].CmdLine[strings.LastIndex(cmd.Commands[1].CmdLine, "@")+1:])
	verifySQL, readErr := os.ReadFile(verifySQLFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(verifySQL), "from dba_users")
	assert.Contains(t, string(verifySQL), "from dba_tablespaces")
	assert.Contains(t, string(verifySQL), "from dba_data_files")
	assert.Contains(t, string(verifySQL), "raise_application_error")
}

func TestResolveOracleWizardCredentials_PrefersExplicitDBACredentialsAndReportsSource(t *testing.T) {
	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "test-conn-creds", Name: "Oracle Creds"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "sys",
		Password:       "sys-pass",
		ConnectAs:      "sysdba",
	}

	user, pass, source, err := resolveOracleWizardCredentials(conn, &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"dba_username": "system",
			"dba_password": "system-pass",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "system", user)
	assert.Equal(t, "system-pass", pass)
	assert.Equal(t, "dba_username/dba_password", source)
}

func TestResolveOracleWizardCredentials_DefaultsSysConnectionToSystemWithSamePassword(t *testing.T) {
	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "test-conn-no-map", Name: "Oracle No Map"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "sys",
		Password:       "sys-pass",
		ConnectAs:      "sysdba",
	}

	user, pass, source, err := resolveOracleWizardCredentials(conn, &Config{Connection: conn})

	require.NoError(t, err)
	assert.Equal(t, "system", user)
	assert.Equal(t, "sys-pass", pass)
	assert.Equal(t, "connection credentials (default SYSTEM with connection password)", source)
}

func TestSwingbenchAdapter_BuildCleanupCommand_DropsUserAndTablespaces(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "test-conn-clean", Name: "Oracle Cleanup"},
		Host:           "192.168.134.129",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "sys",
		Password:       "Qwer1234",
		ConnectAs:      "sysdba",
	}

	cmd, err := adapter.BuildCleanupCommand(ctx, &Config{
		Connection: conn,
		Parameters: map[string]interface{}{},
		WorkDir:    "/tmp/cleanup",
	})
	require.NoError(t, err)
	assert.Contains(t, cmd.CmdLine, "sqlplus")
	assert.Contains(t, cmd.CmdLine, "@/tmp/db-benchmind-cleanup-soe-")
	sqlFile := strings.TrimSpace(cmd.CmdLine[strings.LastIndex(cmd.CmdLine, "@")+1:])
	content, readErr := os.ReadFile(sqlFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "drop user SOE cascade")
	assert.Contains(t, string(content), "drop tablespace SOE_IDX including contents and datafiles")
	assert.Contains(t, string(content), "drop tablespace SOE including contents and datafiles")
}

func TestSwingbenchAdapter_BuildCleanupCommand_IsIdempotentForMissingOrPartialObjects(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{ID: "test-conn-clean-idempotent", Name: "Oracle Cleanup Idempotent"},
		Host:           "localhost",
		Port:           1521,
		ServiceName:    "ORCL",
		Username:       "sys",
		Password:       "manager",
		ConnectAs:      "sysdba",
	}

	cmd, err := adapter.BuildCleanupCommand(ctx, &Config{
		Connection: conn,
		Parameters: map[string]interface{}{},
		WorkDir:    "/tmp/cleanup-idempotent",
	})
	require.NoError(t, err)

	sqlFile := strings.TrimSpace(cmd.CmdLine[strings.LastIndex(cmd.CmdLine, "@")+1:])
	content, readErr := os.ReadFile(sqlFile)
	require.NoError(t, readErr)
	sqlText := string(content)

	assert.Contains(t, sqlText, "whenever sqlerror continue")
	assert.Contains(t, sqlText, "Drop user skipped")
	assert.Contains(t, sqlText, "Drop tablespace SOE_IDX skipped")
	assert.Contains(t, sqlText, "Drop tablespace SOE skipped")
}

func TestSwingbenchAdapter_BuildPrepareCommand_RejectsNonSysAdministrativeUser(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-system",
			Name: "Oracle SYSTEM",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "system",
		Password:    "manager",
	}

	_, err := adapter.BuildPrepareCommand(ctx, &Config{
		Connection: conn,
		Parameters: map[string]interface{}{
			"scale":   0.1,
			"threads": 2,
		},
		WorkDir: "/tmp/test-system",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Oracle Swingbench prepare requires SYS")
}

// TestSwingbenchAdapter_BuildRunCommand tests building run command.
func TestSwingbenchAdapter_BuildRunCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name     string
		conn     connection.Connection
		params   map[string]interface{}
		validate func(t *testing.T, cmd *Command, err error)
	}{
		{
			name: "SOE benchmark with default parameters",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-conn-1",
					Name: "Test Oracle",
				},
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Username:    "testuser",
				Password:    "testpass",
			},
			params: map[string]interface{}{
				"virtual_users": 10,
				"time":          10,
				"config_file":   "/opt/benchtools/swingbench/configs/SOE_CPU_Bound.xml",
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				assert.Contains(t, cmd.CmdLine, "charbench")
				assert.Contains(t, cmd.CmdLine, "-c /opt/benchtools/swingbench/configs/SOE_CPU_Bound.xml")
				assert.Contains(t, cmd.CmdLine, "-cs //localhost:1521/ORCL")
				assert.Contains(t, cmd.CmdLine, "-u soe")
				assert.Contains(t, cmd.CmdLine, "-p soe")
				assert.Contains(t, cmd.CmdLine, "-uc 10")
				assert.Contains(t, cmd.CmdLine, "-rt 00:00:10")
				assert.Contains(t, cmd.CmdLine, "-a")
				assert.Contains(t, cmd.CmdLine, "-v trans,tpm,tps,resp,cpu,errs")
				assert.NotContains(t, cmd.CmdLine, "-dt")
				assert.NotContains(t, cmd.CmdLine, "-intermin")
				assert.NotContains(t, cmd.CmdLine, "-intermax")
				assert.NotContains(t, cmd.CmdLine, "-min")
				assert.NotContains(t, cmd.CmdLine, "-max")
			},
		},
		{
			name: "SOE benchmark with SID instead of service name",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-conn-2",
					Name: "Test Oracle SID",
				},
				Host:     "192.168.1.100",
				Port:     1521,
				SID:      "ORCLSID",
				Username: "testuser",
				Password: "testpass",
			},
			params: map[string]interface{}{
				"virtual_users": 20,
				"time":          5,
				"config_file":   "/opt/benchtools/swingbench/configs/SOE_Disk_Bound.xml",
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.NoError(t, err)
				// The connection string uses service name format //host:port/SID
				assert.Contains(t, cmd.CmdLine, "//192.168.1.100:1521/ORCLSID")
			},
		},
		{
			name: "Missing config_file parameter",
			conn: &connection.OracleConnection{
				BaseConnection: connection.BaseConnection{
					ID:   "test-conn-3",
					Name: "Test Oracle",
				},
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Username:    "testuser",
			},
			params: map[string]interface{}{
				"virtual_users": 5,
			},
			validate: func(t *testing.T, cmd *Command, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "config_file parameter is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Connection: tt.conn,
				Parameters: tt.params,
				WorkDir:    "/tmp/test",
			}

			cmd, err := adapter.BuildRunCommand(ctx, config)
			tt.validate(t, cmd, err)
		})
	}
}

func TestSwingbenchAdapter_BuildRunCommand_ManagedConfigUsesOfficialDefaults(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-4",
			Name: "Test Oracle Managed",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "system",
		Password:    "manager",
	}

	config := &Config{
		Connection: conn,
		Template: &template.Template{
			ID:   "tpl_swing_oe",
			Name: "Swingbench-Oracle-OE-Medium-64u-30m",
			ToolConfig: template.ToolConfig{
				Swingbench: template.SwingbenchConfig{
					Benchmark:      "orderEntry",
					Frontend:       "charbench",
					ConfigMode:     "managed",
					RunTimeSeconds: 120,
					UserCount:      8,
				},
			},
		},
		Parameters: map[string]interface{}{
			"virtual_users": 8,
			"time":          120,
		},
		WorkDir: "/tmp/db-benchmind-managed",
	}

	cmd, err := adapter.BuildRunCommand(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, cmd.CmdLine, "./charbench")
	assert.Contains(t, cmd.CmdLine, "-c ../configs/server_side_soe_v2.xml")
	assert.Contains(t, cmd.CmdLine, "-a")
	assert.Contains(t, cmd.CmdLine, "-u soe")
	assert.Contains(t, cmd.CmdLine, "-p soe")
	assert.Contains(t, cmd.CmdLine, "-v trans,tpm,tps,resp,cpu,errs")
	assert.Contains(t, cmd.CmdLine, "-rt 00:02:00")
	assert.NotContains(t, cmd.CmdLine, "-dt")
	assert.NotContains(t, cmd.CmdLine, "-intermin")
	assert.NotContains(t, cmd.CmdLine, "-intermax")
	assert.NotContains(t, cmd.CmdLine, "-min")
	assert.NotContains(t, cmd.CmdLine, "-max")
	assert.Contains(t, cmd.CmdLine, "-r /tmp/db-benchmind-managed/results.xml")
	assert.Equal(t, "/tmp/db-benchmind-managed/results.xml", cmd.ResultFile)
}

// TestSwingbenchAdapter_BuildRunCommand_NonOracle tests that non-Oracle databases fail.
func TestSwingbenchAdapter_BuildRunCommand_NonOracle(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.MySQLConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-1",
			Name: "Test MySQL",
		},
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "root",
	}

	config := &Config{
		Connection: conn,
		WorkDir:    "/tmp/test",
	}

	_, err := adapter.BuildRunCommand(ctx, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only supports Oracle")
}

// TestSwingbenchAdapter_BuildCleanupCommand tests building cleanup command.
func TestSwingbenchAdapter_BuildCleanupCommand(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	conn := &connection.OracleConnection{
		BaseConnection: connection.BaseConnection{
			ID:   "test-conn-1",
			Name: "Test Oracle",
		},
		Host:        "localhost",
		Port:        1521,
		ServiceName: "ORCL",
		Username:    "sys",
		Password:    "testpass",
		ConnectAs:   "sysdba",
	}

	config := &Config{
		Connection: conn,
		Parameters: map[string]interface{}{},
		WorkDir:    "/tmp/test",
	}

	cmd, err := adapter.BuildCleanupCommand(ctx, config)
	require.NoError(t, err)
	assert.Contains(t, cmd.CmdLine, "sqlplus")
	assert.Contains(t, cmd.CmdLine, "@/tmp/db-benchmind-cleanup-soe-")
	assert.Contains(t, cmd.CmdLine, "sys/testpass@//localhost:1521/ORCL as sysdba")
	sqlFile := strings.TrimSpace(cmd.CmdLine[strings.LastIndex(cmd.CmdLine, "@")+1:])
	content, readErr := os.ReadFile(sqlFile)
	require.NoError(t, readErr)
	assert.Contains(t, string(content), "drop user SOE cascade")
	assert.Contains(t, string(content), "drop tablespace SOE_IDX including contents and datafiles")
	assert.Contains(t, string(content), "drop tablespace SOE including contents and datafiles")
}

// TestSwingbenchAdapter_ParseRunOutput tests parsing swingbench output.
func TestSwingbenchAdapter_ParseRunOutput(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name     string
		stdout   string
		validate func(t *testing.T, result *Result)
	}{
		{
			name: "parse charbench output",
			stdout: `
Time     Users       TPM      TPS     Errors   NCR   UCD   BP    OP    PO    BO
10:58:35 [0/4]       0        0       0        0     0     0     0     0     0
10:58:37 [4/4]       0        0       0        0     0     248   414   0     0
10:58:38 [4/4]       8        8       0        0     0     32    213   0     19
`,
			validate: func(t *testing.T, result *Result) {
				assert.NotNil(t, result)
			},
		},
		{
			name:   "empty output",
			stdout: "",
			validate: func(t *testing.T, result *Result) {
				assert.NotNil(t, result)
				assert.Equal(t, 0.0, result.TPS)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.ParseRunOutput(ctx, tt.stdout, "")
			require.NoError(t, err)
			tt.validate(t, result)
		})
	}
}

func TestSwingbenchAdapter_ParseFinalResults_UsesHintedResultFile(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tempDir := t.TempDir()
	resultFile := filepath.Join(tempDir, "results.xml")
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Results>
  <TotalCompletedTransactions>120</TotalCompletedTransactions>
  <AverageTransactionsPerSecond>2.5</AverageTransactionsPerSecond>
  <TotalRunTime>0:00:48</TotalRunTime>
  <TotalFailedTransactions>3</TotalFailedTransactions>
</Results>`
	require.NoError(t, os.WriteFile(resultFile, []byte(xml), 0o644))

	result, err := adapter.ParseFinalResults(ctx, swingbenchResultFileHintPrefix+resultFile)
	require.NoError(t, err)
	assert.Equal(t, int64(120), result.TotalTransactions)
	assert.Equal(t, 2.5, result.TransactionsPerSec)
	assert.Equal(t, 150.0, result.AvgTPM)
	assert.Equal(t, 48.0, result.TotalTime)
	assert.Equal(t, int64(3), result.IgnoredErrors)
}

// TestSwingbenchAdapter_ValidateConfig tests configuration validation.
func TestSwingbenchAdapter_ValidateConfig(t *testing.T) {
	ctx := context.Background()
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid Oracle connection",
			config: &Config{
				Connection: &connection.OracleConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test-conn-1",
						Name: "Test Oracle",
					},
					Host:     "localhost",
					Port:     1521,
					SID:      "ORCL",
					Username: "testuser",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config is required",
		},
		{
			name: "nil connection",
			config: &Config{
				Connection: nil,
			},
			wantErr: true,
			errMsg:  "connection is required",
		},
		{
			name: "non-Oracle database",
			config: &Config{
				Connection: &connection.MySQLConnection{
					BaseConnection: connection.BaseConnection{
						ID:   "test-conn-1",
						Name: "Test MySQL",
					},
					Host:     "localhost",
					Port:     3306,
					Database: "testdb",
					Username: "root",
				},
			},
			wantErr: true,
			errMsg:  "only supports Oracle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateConfig(ctx, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSwingbenchAdapter_SupportsDatabase tests database support check.
func TestSwingbenchAdapter_SupportsDatabase(t *testing.T) {
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name   string
		dbType connection.DatabaseType
		want   bool
	}{
		{
			name:   "Oracle",
			dbType: connection.DatabaseTypeOracle,
			want:   true,
		},
		{
			name:   "MySQL",
			dbType: connection.DatabaseTypeMySQL,
			want:   false,
		},
		{
			name:   "PostgreSQL",
			dbType: connection.DatabaseTypePostgreSQL,
			want:   false,
		},
		{
			name:   "SQL Server",
			dbType: connection.DatabaseTypeSQLServer,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.SupportsDatabase(tt.dbType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestSwingbenchAdapter_buildConnectionString tests connection string building.
func TestSwingbenchAdapter_buildConnectionString(t *testing.T) {
	adapter := NewSwingbenchAdapter()

	tests := []struct {
		name          string
		conn          *connection.OracleConnection
		expectedParts []string
	}{
		{
			name: "service name",
			conn: &connection.OracleConnection{
				Host:        "localhost",
				Port:        1521,
				ServiceName: "ORCL",
				Username:    "testuser",
			},
			expectedParts: []string{"jdbc:oracle:thin:@//localhost:1521/ORCL"},
		},
		{
			name: "SID",
			conn: &connection.OracleConnection{
				Host:     "192.168.1.100",
				Port:     1521,
				SID:      "ORCLSID",
				Username: "testuser",
			},
			expectedParts: []string{"jdbc:oracle:thin:@192.168.1.100:1521:ORCLSID"},
		},
		{
			name: "fallback when no service name or SID",
			conn: &connection.OracleConnection{
				Host:     "localhost",
				Port:     1521,
				Username: "testuser",
			},
			expectedParts: []string{"jdbc:oracle:thin:@//localhost:1521/ORCL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.buildConnectionString(tt.conn)
			for _, part := range tt.expectedParts {
				assert.Contains(t, result, part)
			}
		})
	}
}
