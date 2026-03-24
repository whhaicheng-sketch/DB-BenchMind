package template

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestDefaultSeedTemplates_SeedsAllBuiltinDatabaseProfiles(t *testing.T) {
	templates := DefaultSeedTemplates()

	if len(templates) != 12 {
		t.Fatalf("DefaultSeedTemplates() len = %d, want 12", len(templates))
	}

	expectedIDs := []string{
		"oracle_cpu_bound", "oracle_io_bound", "oracle_test",
		"mysql_cpu_bound", "mysql_io_bound", "mysql_test",
		"sqlserver_cpu_bound", "sqlserver_io_bound", "sqlserver_test",
		"postgresql_cpu_bound", "postgresql_io_bound", "postgresql_test",
	}

	seen := map[string]*Template{}
	for _, tmpl := range templates {
		if tmpl == nil {
			t.Fatal("seed template must not be nil")
		}
		seen[tmpl.ID] = tmpl
		if !tmpl.IsBuiltin {
			t.Fatalf("seed template %s must be builtin", tmpl.ID)
		}
		if !tmpl.IsReadOnly() {
			t.Fatalf("seed template %s must be readonly", tmpl.ID)
		}
	}

	for _, id := range expectedIDs {
		if _, ok := seen[id]; !ok {
			t.Fatalf("expected seeded template %s", id)
		}
	}
}

func TestDefaultSeedTemplates_EachDatabaseHasCpuIoAndTestProfiles(t *testing.T) {
	templates := DefaultSeedTemplates()

	profilesByDB := map[string]map[string]bool{}
	for _, tmpl := range templates {
		if profilesByDB[tmpl.DBFamily] == nil {
			profilesByDB[tmpl.DBFamily] = map[string]bool{}
		}
		profilesByDB[tmpl.DBFamily][tmpl.ProfileType] = true
	}

	for _, dbFamily := range []string{"oracle", "mysql", "sqlserver", "postgresql"} {
		if !profilesByDB[dbFamily]["cpu_bound"] || !profilesByDB[dbFamily]["io_bound"] || !profilesByDB[dbFamily]["test"] {
			t.Fatalf("database %s does not have cpu_bound/io_bound/test builtins: %#v", dbFamily, profilesByDB[dbFamily])
		}
	}
}

func TestDefaultSeedTemplates_EngineeredFlagsAndSmokeMetadata(t *testing.T) {
	templates := DefaultSeedTemplates()
	index := map[string]*Template{}
	for _, tmpl := range templates {
		index[tmpl.ID] = tmpl
	}

	for _, id := range []string{"sqlserver_cpu_bound", "sqlserver_io_bound", "postgresql_cpu_bound"} {
		tmpl := index[id]
		if tmpl == nil {
			t.Fatalf("missing template %s", id)
		}
		if tmpl.SourceAlignment != "engineered_split_from_baseline" {
			t.Fatalf("%s sourceAlignment = %s, want engineered_split_from_baseline", id, tmpl.SourceAlignment)
		}
	}

	if got := index["postgresql_io_bound"].SourceAlignment; got != "direct_from_doc_as_baseline" {
		t.Fatalf("postgresql_io_bound sourceAlignment = %s, want direct_from_doc_as_baseline", got)
	}

	for _, id := range []string{"oracle_test", "mysql_test", "sqlserver_test", "postgresql_test"} {
		tmpl := index[id]
		if tmpl == nil {
			t.Fatalf("missing template %s", id)
		}
		if tmpl.SourceAlignment != "engineered_minimal" {
			t.Fatalf("%s sourceAlignment = %s, want engineered_minimal", id, tmpl.SourceAlignment)
		}
		if tmpl.TestPosition != "smoke" {
			t.Fatalf("%s testPosition = %s, want smoke", id, tmpl.TestPosition)
		}
		if !slices.Contains(tmpl.Tags, "smoke") || !slices.Contains(tmpl.Tags, "minimal") {
			t.Fatalf("%s tags = %v, want smoke and minimal", id, tmpl.Tags)
		}
		if !(containsTextFold(tmpl.Goal, "最小") && containsTextFold(tmpl.Description, "smoke")) {
			t.Fatalf("%s should describe minimal smoke semantics, goal=%q description=%q", id, tmpl.Goal, tmpl.Description)
		}
	}
}

func TestDefaultSeedTemplates_JSONOmitsRemovedMetadataAndCarriesBuiltinTemplateMetadata(t *testing.T) {
	templates := DefaultSeedTemplates()
	payload, err := json.Marshal(templates)
	if err != nil {
		t.Fatalf("marshal templates: %v", err)
	}

	jsonText := string(payload)
	for _, required := range []string{`"profile_type"`, `"source_alignment"`, `"prepare_config"`, `"run_config"`, `"cleanup_config"`, `"metrics"`, `"tags"`} {
		if !strings.Contains(jsonText, required) {
			t.Fatalf("seed templates must expose %s in json payload: %s", required, jsonText)
		}
	}
	for _, forbidden := range []string{`"updatedAt"`, `"scope"`, `"status"`} {
		if strings.Contains(jsonText, forbidden) {
			t.Fatalf("seed templates should not expose %s metadata: %s", forbidden, jsonText)
		}
	}
}

func TestDefaultSeedTemplates_TestProfilesUseSixtySecondRuntime(t *testing.T) {
	templates := DefaultSeedTemplates()
	index := map[string]*Template{}
	for _, tmpl := range templates {
		index[tmpl.ID] = tmpl
	}

	for _, id := range []string{"oracle_test", "mysql_test", "sqlserver_test", "postgresql_test"} {
		tmpl := index[id]
		if tmpl == nil {
			t.Fatalf("missing template %s", id)
		}
		if tmpl.Runtime.DurationSeconds != 60 {
			t.Fatalf("%s runtime duration = %d, want 60", id, tmpl.Runtime.DurationSeconds)
		}
		if tmpl.Tool == ToolSwingbench {
			if tmpl.ToolConfig.Swingbench.RunTimeSeconds != 60 {
				t.Fatalf("%s swingbench runtime = %d, want 60", id, tmpl.ToolConfig.Swingbench.RunTimeSeconds)
			}
		}
	}

	for _, id := range []string{"oracle_cpu_bound", "oracle_io_bound", "mysql_cpu_bound", "mysql_io_bound", "sqlserver_cpu_bound", "sqlserver_io_bound", "postgresql_cpu_bound", "postgresql_io_bound"} {
		tmpl := index[id]
		if tmpl == nil {
			t.Fatalf("missing template %s", id)
		}
		if tmpl.Runtime.DurationSeconds == 60 {
			t.Fatalf("%s runtime duration unexpectedly changed to 60", id)
		}
	}
}

func TestDefaultSeedTemplates_OracleTestUsesManagedXMLPathInsteadOfNarrativeText(t *testing.T) {
	templates := DefaultSeedTemplates()
	index := map[string]*Template{}
	for _, tmpl := range templates {
		index[tmpl.ID] = tmpl
	}

	tmpl := index["oracle_test"]
	if tmpl == nil {
		t.Fatal("missing oracle_test")
	}

	got := tmpl.ToolConfig.Swingbench.XMLOverrides
	if got == "" {
		t.Fatal("oracle_test xmlOverrides must not be empty")
	}
	if got != "../configs/server_side_soe_v2.xml" {
		t.Fatalf("oracle_test xmlOverrides = %q, want managed XML path", got)
	}
	if strings.Contains(got, "沿用") || strings.Contains(got, "cpu_bound") || strings.Contains(got, "9:1") {
		t.Fatalf("oracle_test xmlOverrides should be a usable config path, got %q", got)
	}
}

func containsTextFold(text, target string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(target))
}
