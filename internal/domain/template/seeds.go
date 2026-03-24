package template

import "time"

// DefaultSeedTemplates returns the only template that should exist after
// startup synchronization: the default MySQL sysbench smoke test template.
func DefaultSeedTemplates() []*Template {
	now := time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	return []*Template{
		mustTemplate(&Template{
			ID:             "tpl_test_mysql_sysbench",
			Name:           "MySQL - Sysbench Test",
			Description:    "Minimal MySQL sysbench smoke template where prepare rebuilds the environment, run can be repeated, and cleanup fully removes the benchmark state.",
			Tool:           ToolSysbench,
			DBFamily:       "mysql",
			WorkloadFamily: "oltp-read-write",
			IsBuiltin:      true,
			Version:        "1.0.0",
			CreatedAt:      now,
			Compatibility: Compatibility{
				SupportedDatabases: []string{"mysql"},
				CompatibilityNotes: "Minimal dataset for functional validation, not performance baseline testing.",
				RequiresPrivileges: []string{"CREATE", "DROP"},
				Constraints:        []string{"Use for smoke validation and task-chain verification."},
			},
			Phases: PhaseSet{
				Prepare: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Warmup:  PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
				Run:     PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
				Cleanup: PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			},
			Runtime: Runtime{
				Concurrency:           Concurrency{Mode: "threads", Value: 1},
				DurationSeconds:       30,
				WarmupSeconds:         0,
				RampUpSeconds:         3,
				ReportIntervalSeconds: 1,
				Percentile:            95,
				ValidationEnabled:     true,
				Notes:                 "Lowest-volume sysbench workflow intended to finish quickly and exercise the full chain.",
			},
			ToolConfig: ToolConfig{
				Sysbench: SysbenchConfig{
					DBDriver:     "mysql",
					ScriptType:   "oltp_read_write",
					Tables:       1,
					TableSize:    1000,
					ReportChecks: true,
				},
			},
		}),
	}
}

func mustTemplate(tmpl *Template) *Template {
	tmpl.Normalize()
	if err := tmpl.Validate(); err != nil {
		panic(err)
	}
	return tmpl
}
