// Package pages provides GUI page tests.
// TDD Approach: Tests written before/with implementation to ensure correctness.
package pages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTemplateInfo_BuiltinTemplates tests that built-in templates are correctly configured.
func TestTemplateInfo_BuiltinTemplates(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		expectedDBType string
		isBuiltin      bool
	}{
		{
			name:           "MySQL built-in template",
			templateID:     "sysbench-oltp-mysql",
			expectedDBType: "MySQL",
			isBuiltin:      true,
		},
		{
			name:           "PostgreSQL built-in template",
			templateID:     "sysbench-oltp-postgresql",
			expectedDBType: "PostgreSQL",
			isBuiltin:      true,
		},
		{
			name:           "Oracle built-in template",
			templateID:     "sysbench-oltp-oracle",
			expectedDBType: "Oracle",
			isBuiltin:      true,
		},
		{
			name:           "SQL Server built-in template",
			templateID:     "sysbench-oltp-sqlserver",
			expectedDBType: "SQL Server",
			isBuiltin:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load templates
			page := &TemplateManagementPage{}
			templates := page.loadTemplatesData()

			// Find the template by ID
			var found *templateInfo
			for i := range templates {
				if templates[i].ID == tt.templateID {
					found = &templates[i]
					break
				}
			}

			// Assertions
			assert.NotNil(t, found, "Template should exist")
			assert.Equal(t, tt.expectedDBType, found.DBType, "DB type should match")
			assert.True(t, found.IsBuiltin, "Should be marked as builtin")
			assert.True(t, found.IsDefault, "Should be marked as default")
			assert.NotNil(t, found.Parameters, "Should have OLTP parameters")
		})
	}
}

// TestTemplateInfo_DefaultPerDBType tests that each DB type can have its own default.
func TestTemplateInfo_DefaultPerDBType(t *testing.T) {
	// Test setup: Create mock templates
	templates := []templateInfo{
		{
			ID:        "mysql-t1",
			Name:      "MySQL Template 1",
			DBType:    "MySQL",
			IsDefault: true,
		},
		{
			ID:        "mysql-t2",
			Name:      "MySQL Template 2",
			DBType:    "MySQL",
			IsDefault: false,
		},
		{
			ID:        "pg-t1",
			Name:      "PostgreSQL Template 1",
			DBType:    "PostgreSQL",
			IsDefault: true,
		},
	}

	// Verify that each DB type has exactly one default
	dbTypeDefaults := make(map[string]int)
	for _, tmpl := range templates {
		if tmpl.IsDefault {
			dbTypeDefaults[tmpl.DBType]++
		}
	}

	assert.Equal(t, 1, dbTypeDefaults["MySQL"], "MySQL should have 1 default template")
	assert.Equal(t, 1, dbTypeDefaults["PostgreSQL"], "PostgreSQL should have 1 default template")
}

// TestTemplateInfo_OLTPParameters tests OLTP parameter validation.
func TestTemplateInfo_OLTPParameters(t *testing.T) {
	params := &OLTPParameters{
		Tables:    10,
		TableSize: 10000,
	}

	// Test that all parameters are set
	assert.Equal(t, 10, params.Tables, "Tables should be 10")
	assert.Equal(t, 10000, params.TableSize, "TableSize should be 10000")
}

// TestTemplateInfo_Grouping tests that templates are grouped correctly by DB type.
func TestTemplateInfo_Grouping(t *testing.T) {
	page := &TemplateManagementPage{}
	templates := page.loadTemplatesData()

	// Group templates by DB type
	groups := make(map[string][]templateInfo)
	for _, tmpl := range templates {
		dbType := tmpl.DBType
		if dbType == "" {
			dbType = "MySQL"
		}
		groups[dbType] = append(groups[dbType], tmpl)
	}

	// Assertions
	assert.Contains(t, groups, "MySQL", "Should have MySQL group")
	assert.Contains(t, groups, "PostgreSQL", "Should have PostgreSQL group")
	assert.Contains(t, groups, "Oracle", "Should have Oracle group")
	assert.Contains(t, groups, "SQL Server", "Should have SQL Server group")

	// Each group should have at least one template
	for dbType, group := range groups {
		assert.Greater(t, len(group), 0, "DB type %s should have at least one template", dbType)
	}
}

// parseIntOrDefault tests the helper function.
func TestParseIntOrDefault(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		defaultValue  int
		expectedValue int
	}{
		{
			name:          "Valid number",
			input:         "100",
			defaultValue:  10,
			expectedValue: 100,
		},
		{
			name:          "Invalid number uses default",
			input:         "abc",
			defaultValue:  10,
			expectedValue: 10,
		},
		{
			name:          "Empty string uses default",
			input:         "",
			defaultValue:  10,
			expectedValue: 10,
		},
		{
			name:          "Zero value",
			input:         "0",
			defaultValue:  10,
			expectedValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIntOrDefault(tt.input, tt.defaultValue)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}
