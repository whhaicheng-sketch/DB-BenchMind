package template

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultSeedTemplates_OnlyKeepSingleDefaultTestTemplate(t *testing.T) {
	templates := DefaultSeedTemplates()

	if len(templates) != 1 {
		t.Fatalf("DefaultSeedTemplates() len = %d, want 1", len(templates))
	}

	tmpl := templates[0]
	if tmpl.ID != "tpl_test_mysql_sysbench" {
		t.Fatalf("default template id = %s, want tpl_test_mysql_sysbench", tmpl.ID)
	}
	if tmpl.Tool != ToolSysbench {
		t.Fatalf("default template tool = %s, want %s", tmpl.Tool, ToolSysbench)
	}
	if tmpl.DBFamily != "mysql" {
		t.Fatalf("default template dbFamily = %s, want mysql", tmpl.DBFamily)
	}
	if !tmpl.SupportsDatabase("mysql") {
		t.Fatal("default template must support mysql")
	}
	if !tmpl.Phases.Prepare.Enabled || !tmpl.Phases.Warmup.Enabled || !tmpl.Phases.Run.Enabled || !tmpl.Phases.Cleanup.Enabled {
		t.Fatal("default template must support prepare/warmup/run/cleanup")
	}
}

func TestDefaultSeedTemplates_SmokeTemplateDescribesUnifiedLifecycle(t *testing.T) {
	templates := DefaultSeedTemplates()

	if len(templates) != 1 {
		t.Fatalf("DefaultSeedTemplates() len = %d, want 1", len(templates))
	}

	tmpl := templates[0]
	if tmpl.Description == "" {
		t.Fatal("default test template must have description")
	}
	if !(containsTextFold(tmpl.Description, "rebuild") || containsTextFold(tmpl.Description, "recreate")) {
		t.Fatalf("default test template should describe prepare rebuild semantics: %q", tmpl.Description)
	}
	if !(containsTextFold(tmpl.Description, "cleanup") && (containsTextFold(tmpl.Description, "full teardown") || containsTextFold(tmpl.Description, "fully removes"))) {
		t.Fatalf("default test template should describe cleanup teardown semantics: %q", tmpl.Description)
	}
}

func TestDefaultSeedTemplates_DefaultTestDoesNotCarryLegacyCrossDatabaseEntries(t *testing.T) {
	templates := DefaultSeedTemplates()

	if len(templates) != 1 {
		t.Fatalf("DefaultSeedTemplates() len = %d, want 1", len(templates))
	}

	tmpl := templates[0]
	if tmpl.ID != "tpl_test_mysql_sysbench" {
		t.Fatalf("default template id = %s, want tpl_test_mysql_sysbench", tmpl.ID)
	}
	if len(tmpl.Parameters) != 0 {
		t.Fatalf("default test template should not carry legacy parameters, got %v", tmpl.Parameters)
	}
	if tmpl.CreatedAt == "" {
		t.Fatal("default test template should keep created_at for persistence")
	}
	payload, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatalf("marshal template: %v", err)
	}
	for _, forbidden := range []string{`"updatedAt"`, `"scope"`, `"status"`, `"tags"`} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("default test template should not expose %s metadata: %s", forbidden, string(payload))
		}
	}
}

func containsTextFold(text, target string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(target))
}
