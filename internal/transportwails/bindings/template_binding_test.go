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
		ID:             "tpl-test",
		Name:           "Template Test",
		Description:    "test",
		Tool:           domaintemplate.ToolSysbench,
		DBFamily:       "mysql",
		WorkloadFamily: "oltp-read-write",
		Version:        "1.0.0",
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
	for _, forbidden := range []string{`"scope"`, `"tags"`, `"status"`, `"updatedAt"`} {
		if strings.Contains(jsonText, forbidden) {
			t.Fatalf("dto json must omit %s, got %s", forbidden, jsonText)
		}
	}
}
