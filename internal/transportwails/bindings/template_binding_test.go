package bindings

import (
	"encoding/json"
	"strings"
	"testing"

	domaintemplate "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

func TestTemplateBinding_DTOJSONOmitsRemovedMetadataFields(t *testing.T) {
	binding := &TemplateBinding{}
	tmpl := &domaintemplate.Template{
		ID:              "tpl-test",
		Name:            "Template Test",
		Description:     "test",
		Tool:            domaintemplate.ToolSysbench,
		DBFamily:        "mysql",
		ProfileType:     "test",
		SourceAlignment: "engineered_minimal",
		Goal:            "最小数据量跑通链路",
		PrepareConfig:   map[string]interface{}{"dataset": "0.1GB"},
		RunConfig:       map[string]interface{}{"threads": 8},
		CleanupConfig:   map[string]interface{}{"strategy": "sysbench cleanup"},
		Metrics:         []string{"TPS", "QPS"},
		Tags:            []string{"builtin", "smoke", "minimal"},
		TestPosition:    "smoke",
		WorkloadFamily:  "oltp-read-write",
		Version:         "1.0.0",
		Phases: domaintemplate.PhaseSet{
			Prepare: domaintemplate.PhaseConfig{Enabled: true},
			Warmup:  domaintemplate.PhaseConfig{Enabled: true},
			Run:     domaintemplate.PhaseConfig{Enabled: true, Required: true},
			Cleanup: domaintemplate.PhaseConfig{Enabled: true},
		},
		Runtime: domaintemplate.Runtime{
			Concurrency:     domaintemplate.Concurrency{Mode: "threads", Value: 1},
			DurationSeconds: 30,
		},
		ToolConfig: domaintemplate.ToolConfig{
			Sysbench: domaintemplate.SysbenchConfig{
				DBDriver:   "mysql",
				ScriptType: "oltp_read_write",
				Tables:     1,
				TableSize:  1000,
			},
		},
	}
	tmpl.Normalize()
	dto := binding.toDTO(tmpl)

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal dto: %v", err)
	}

	jsonText := string(data)
	for _, required := range []string{`"profile_type"`, `"source_alignment"`, `"prepare_config"`, `"run_config"`, `"cleanup_config"`, `"metrics"`, `"tags"`, `"test_position"`} {
		if !strings.Contains(jsonText, required) {
			t.Fatalf("dto json must include %s, got %s", required, jsonText)
		}
	}
	for _, forbidden := range []string{`"scope"`, `"tags"`, `"status"`, `"updatedAt"`} {
		if forbidden == `"tags"` {
			continue
		}
		if strings.Contains(jsonText, forbidden) {
			t.Fatalf("dto json must omit %s, got %s", forbidden, jsonText)
		}
	}
}
