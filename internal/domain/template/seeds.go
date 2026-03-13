package template

import "time"

// DefaultSeedTemplates returns built-in and readonly shared templates used by
// the Templates module before user-created templates exist.
func DefaultSeedTemplates() []*Template {
	now := time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	return []*Template{
		mustTemplate(&Template{
			ID:             "tpl_sys_mysql_rw",
			Name:           "Sysbench-MySQL-OLTP_RW-10x100k-32th-300s",
			Description:    "Built-in MySQL OLTP read/write baseline for fast smoke validation.",
			Tool:           ToolSysbench,
			DBFamily:       "mysql",
			WorkloadFamily: "oltp-read-write",
			Scope:          ScopeBuiltin,
			Status:         StatusReady,
			Tags:           []string{"mysql", "baseline", "oltp"},
			Version:        "1.0.0",
			CreatedAt:      now,
			UpdatedAt:      now,
			Compatibility: Compatibility{
				SupportedDatabases: []string{"mysql"},
				SupportedVersions:  []string{"MySQL 8.0+", "MariaDB 10.6+"},
				CompatibilityNotes: "Prepared for transactional OLTP verification and regression checks.",
				RequiresPrivileges: []string{"CREATE", "DROP"},
				Constraints:        []string{"Connection binding happens in Tasks & Monitor."},
			},
			Phases: PhaseSet{
				Prepare: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Warmup:  PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Run:     PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
				Cleanup: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			},
			Runtime: Runtime{
				Concurrency:           Concurrency{Mode: "threads", Value: 32},
				DurationSeconds:       300,
				WarmupSeconds:         30,
				RampUpSeconds:         15,
				ReportIntervalSeconds: 10,
				Percentile:            95,
				ValidationEnabled:     true,
				Notes:                 "Recommended as default baseline before environment-specific tuning.",
			},
			ToolConfig: ToolConfig{
				Sysbench: SysbenchConfig{
					DBDriver:     "mysql",
					ScriptType:   "oltp_read_write",
					Tables:       10,
					TableSize:    100000,
					ReportChecks: true,
					ExtraCLIArgs: "--db-ps-mode=disable",
				},
			},
		}),
		mustTemplate(&Template{
			ID:             "tpl_swing_oe",
			Name:           "Swingbench-Oracle-OE-Medium-64u-30m",
			Description:    "Built-in OrderEntry scenario aligned to Oracle 11g compatible target.",
			Tool:           ToolSwingbench,
			DBFamily:       "oracle",
			WorkloadFamily: "order-entry",
			Scope:          ScopeBuiltin,
			Status:         StatusReady,
			Tags:           []string{"oracle", "orderentry", "baseline"},
			Version:        "1.0.0",
			CreatedAt:      now,
			UpdatedAt:      now,
			Phases: PhaseSet{
				Build:    PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Generate: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Warmup:   PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Run:      PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
			},
			Runtime: Runtime{
				Concurrency:           Concurrency{Mode: "users", Value: 64},
				DurationSeconds:       1800,
				WarmupSeconds:         120,
				RampUpSeconds:         60,
				ReportIntervalSeconds: 30,
				Percentile:            95,
			},
			ToolConfig: ToolConfig{
				Swingbench: SwingbenchConfig{
					Benchmark:       "orderEntry",
					Frontend:        "charbench",
					ConfigMode:      "managed",
					WizardOperation: "generate",
					UserCount:       64,
					RunTimeSeconds:  1800,
					MaxThinkTime:    2,
				},
			},
		}),
		mustTemplate(&Template{
			ID:             "tpl_hammer_oracle_c",
			Name:           "HammerDB-Oracle-TPROC-C-1000W-96vu-30m",
			Description:    "Built-in Oracle TPROC-C workflow with build, run and cleanup stages.",
			Tool:           ToolHammerDB,
			DBFamily:       "oracle",
			WorkloadFamily: "tproc-c",
			Scope:          ScopeBuiltin,
			Status:         StatusReady,
			Tags:           []string{"oracle", "tproc-c", "workflow"},
			Version:        "1.0.0",
			CreatedAt:      now,
			UpdatedAt:      now,
			Phases: PhaseSet{
				Build:   PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Prepare: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Run:     PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
				Cleanup: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Delete:  PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			},
			Runtime: Runtime{
				Concurrency:           Concurrency{Mode: "virtualUsers", Value: 96},
				DurationSeconds:       1800,
				WarmupSeconds:         120,
				RampUpSeconds:         60,
				ReportIntervalSeconds: 20,
				Percentile:            95,
				Iterations:            1,
			},
			ToolConfig: ToolConfig{
				HammerDB: HammerDBConfig{
					Benchmark:    "tproc-c",
					VirtualUsers: 96,
					Warehouses:   1000,
					ScaleFactor:  10,
					TimeProfile:  true,
				},
			},
		}),
		mustTemplate(&Template{
			ID:             "tpl_shared_pg_ro",
			Name:           "Shared-PostgreSQL-ReadOnly-Validation",
			Description:    "Readonly shared template for PostgreSQL read-only validation across teams.",
			Tool:           ToolSysbench,
			DBFamily:       "postgresql",
			WorkloadFamily: "oltp-read-only",
			Scope:          ScopeReadonlyShared,
			Status:         StatusReady,
			Tags:           []string{"shared", "postgresql", "readonly"},
			Version:        "1.0.0",
			CreatedAt:      now,
			UpdatedAt:      now,
			Phases: PhaseSet{
				Prepare: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Run:     PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
				Cleanup: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			},
			Runtime: Runtime{
				Concurrency:           Concurrency{Mode: "threads", Value: 24},
				DurationSeconds:       240,
				ReportIntervalSeconds: 10,
				Percentile:            95,
			},
			ToolConfig: ToolConfig{
				Sysbench: SysbenchConfig{
					DBDriver:   "pgsql",
					ScriptType: "oltp_read_only",
					Tables:     8,
					TableSize:  50000,
				},
			},
		}),
	}
}

func mustTemplate(t *Template) *Template {
	t.Normalize()
	if err := t.Validate(); err != nil {
		panic(err)
	}
	return t
}
