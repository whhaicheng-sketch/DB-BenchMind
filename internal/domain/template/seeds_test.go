package template

import (
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
	if tmpl.Scope != ScopeTest {
		t.Fatalf("default template scope = %s, want %s", tmpl.Scope, ScopeTest)
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
	if !containsTag(tmpl.Tags, "test") {
		t.Fatalf("default template must include 'test' tag: %v", tmpl.Tags)
	}
	if !tmpl.Phases.Prepare.Enabled || !tmpl.Phases.Run.Enabled || !tmpl.Phases.Cleanup.Enabled {
		t.Fatal("default template must support prepare/run/cleanup")
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
}

func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

func containsTextFold(text, target string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(target))
}
