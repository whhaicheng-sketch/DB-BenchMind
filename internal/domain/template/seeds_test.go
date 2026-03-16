package template

import (
	"strings"
	"testing"
)

func TestDefaultSeedTemplates_IncludeSmokeTemplatesForSupportedDatabases(t *testing.T) {
	templates := DefaultSeedTemplates()

	want := map[string]string{
		"mysql":      ToolSysbench,
		"postgresql": ToolSysbench,
		"oracle":     ToolSwingbench,
		"sqlserver":  ToolHammerDB,
	}

	found := map[string]*Template{}
	for _, tmpl := range templates {
		if tmpl.Scope != ScopeTest {
			continue
		}
		found[tmpl.DBFamily] = tmpl
	}

	for dbFamily, tool := range want {
		tmpl := found[dbFamily]
		if tmpl == nil {
			t.Fatalf("missing default test template for %s", dbFamily)
		}
		if tmpl.Tool != tool {
			t.Fatalf("template for %s uses tool %s, want %s", dbFamily, tmpl.Tool, tool)
		}
		if !tmpl.SupportsDatabase(dbFamily) {
			t.Fatalf("template for %s does not support its own database family", dbFamily)
		}
		if !containsTag(tmpl.Tags, "test") {
			t.Fatalf("template for %s must include 'test' tag: %v", dbFamily, tmpl.Tags)
		}
		if !tmpl.Phases.Prepare.Enabled || !tmpl.Phases.Run.Enabled || !tmpl.Phases.Cleanup.Enabled {
			t.Fatalf("template for %s must support prepare/run/cleanup", dbFamily)
		}
	}
}

func TestDefaultSeedTemplates_SmokeTemplatesDescribeUnifiedLifecycle(t *testing.T) {
	templates := DefaultSeedTemplates()

	for _, tmpl := range templates {
		if tmpl.Scope != ScopeTest {
			continue
		}

		switch tmpl.DBFamily {
		case "mysql", "postgresql", "oracle", "sqlserver":
			if tmpl.Description == "" {
				t.Fatalf("test template for %s must have description", tmpl.DBFamily)
			}
			if !(containsTextFold(tmpl.Description, "rebuild") || containsTextFold(tmpl.Description, "recreate")) {
				t.Fatalf("test template for %s should describe prepare rebuild semantics: %q", tmpl.DBFamily, tmpl.Description)
			}
			if !(containsTextFold(tmpl.Description, "cleanup") && (containsTextFold(tmpl.Description, "full teardown") || containsTextFold(tmpl.Description, "fully removes"))) {
				t.Fatalf("test template for %s should describe cleanup teardown semantics: %q", tmpl.DBFamily, tmpl.Description)
			}
		}
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
