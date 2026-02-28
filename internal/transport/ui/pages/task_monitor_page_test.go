// Package pages provides tests for task monitor page.
package pages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// TestLoadTemplatesData_AllDatabases tests that templates are loaded for all database types.
func TestLoadTemplatesData_AllDatabases(t *testing.T) {
	// Create a minimal TaskMonitorPage for testing
	page := &TaskMonitorPage{
		templates:   []templateInfo{},
		connections: make(map[string]connection.Connection),
	}

	// Load templates
	templates := page.loadTemplatesData()

	// Define expected template counts per database type
	expectedCounts := map[string]int{
		"MySQL":       3, // Test, CPU Bound, Disk Bound
		"PostgreSQL":  3, // Test, CPU Bound, Disk Bound
		"Oracle":      3, // Test, CPU Bound, Disk Bound (Swingbench)
		"SQL Server":  2, // Test, TPROC-C (HammerDB)
	}

	// Count templates by database type
	actualCounts := make(map[string]int)
	for _, tmpl := range templates {
		if tmpl.IsBuiltin {
			actualCounts[tmpl.DBType]++
		}
	}

	// Verify each database type has expected number of templates
	for dbType, expectedCount := range expectedCounts {
		actualCount := actualCounts[dbType]
		assert.Equal(t, expectedCount, actualCount,
			"Expected %d templates for %s, got %d", expectedCount, dbType, actualCount)
	}
}

// TestLoadTemplatesData_DefaultTemplates tests that each database type has a default template.
func TestLoadTemplatesData_DefaultTemplates(t *testing.T) {
	page := &TaskMonitorPage{
		templates:   []templateInfo{},
		connections: make(map[string]connection.Connection),
	}

	templates := page.loadTemplatesData()

	// Check that each database type has exactly one default template
	expectedDefaults := map[string]bool{
		"MySQL":       true,
		"PostgreSQL":  true,
		"Oracle":      true,
		"SQL Server":  true,
	}

	defaultCounts := make(map[string]int)
	for _, tmpl := range templates {
		if tmpl.IsBuiltin && tmpl.IsDefault {
			defaultCounts[tmpl.DBType]++
		}
	}

	for dbType := range expectedDefaults {
		count := defaultCounts[dbType]
		assert.Equal(t, 1, count,
			"Expected exactly 1 default template for %s, got %d", dbType, count)
	}
}

// TestLoadTemplatesData_OracleTemplates tests Oracle Swingbench template properties.
func TestLoadTemplatesData_OracleTemplates(t *testing.T) {
	page := &TaskMonitorPage{
		templates:   []templateInfo{},
		connections: make(map[string]connection.Connection),
	}

	templates := page.loadTemplatesData()

	// Find Oracle templates
	var testTemplate, cpuBoundTemplate, diskBoundTemplate *templateInfo
	for i := range templates {
		tmpl := &templates[i]
		if tmpl.DBType == "Oracle" && tmpl.IsBuiltin {
			switch tmpl.ID {
			case "swingbench-oracle-test":
				testTemplate = tmpl
			case "swingbench-oracle-cpu-bound":
				cpuBoundTemplate = tmpl
			case "swingbench-oracle-disk-bound":
				diskBoundTemplate = tmpl
			}
		}
	}

	// Verify Test template
	assert.NotNil(t, testTemplate, "Test template not found")
	assert.Equal(t, "swingbench", testTemplate.Tool)
	assert.NotNil(t, testTemplate.Weights)
	assert.Equal(t, 0.1, testTemplate.Weights.Scale) // 100MB

	// Verify CPU Bound template
	assert.NotNil(t, cpuBoundTemplate, "CPU Bound template not found")
	assert.Equal(t, "swingbench", cpuBoundTemplate.Tool)
	assert.NotNil(t, cpuBoundTemplate.Weights)
	assert.Equal(t, 1.0, cpuBoundTemplate.Weights.Scale) // 1GB
	assert.Equal(t, 85, cpuBoundTemplate.Weights.BrowseProducts) // 85% read

	// Verify Disk Bound template
	assert.NotNil(t, diskBoundTemplate, "Disk Bound template not found")
	assert.Equal(t, "swingbench", diskBoundTemplate.Tool)
	assert.NotNil(t, diskBoundTemplate.Weights)
	assert.Equal(t, 100.0, diskBoundTemplate.Weights.Scale) // 100GB
}

// TestLoadTemplatesData_SQLServerTemplates tests SQL Server HammerDB template properties.
func TestLoadTemplatesData_SQLServerTemplates(t *testing.T) {
	page := &TaskMonitorPage{
		templates:   []templateInfo{},
		connections: make(map[string]connection.Connection),
	}

	templates := page.loadTemplatesData()

	// Find SQL Server templates
	var testTemplate, tprocCTemplate *templateInfo
	for i := range templates {
		tmpl := &templates[i]
		if tmpl.DBType == "SQL Server" && tmpl.IsBuiltin {
			switch tmpl.ID {
			case "hammerdb-sqlserver-test":
				testTemplate = tmpl
			case "hammerdb-sqlserver-tproc-c":
				tprocCTemplate = tmpl
			}
		}
	}

	// Verify Test template
	assert.NotNil(t, testTemplate, "Test template not found")
	assert.Equal(t, "hammerdb", testTemplate.Tool)
	assert.NotNil(t, testTemplate.HammerParams)
	assert.Equal(t, 1, testTemplate.HammerParams.Warehouses)
	assert.Equal(t, "timed", testTemplate.HammerParams.Driver)
	assert.Equal(t, false, testTemplate.HammerParams.AllWarehouse)

	// Verify TPROC-C template
	assert.NotNil(t, tprocCTemplate, "TPROC-C template not found")
	assert.Equal(t, "hammerdb", tprocCTemplate.Tool)
	assert.NotNil(t, tprocCTemplate.HammerParams)
	assert.Equal(t, 100, tprocCTemplate.HammerParams.Warehouses)
	assert.Equal(t, "timed", tprocCTemplate.HammerParams.Driver)
	assert.Equal(t, true, tprocCTemplate.HammerParams.AllWarehouse)
}

// TestBuildBenchmarkTask_Sysbench tests building benchmark task for Sysbench templates.
func TestBuildBenchmarkTask_Sysbench(t *testing.T) {
	// This would require setting up a full TaskMonitorPage with mock use cases
	// For now, we'll skip this as it requires more complex setup
	t.Skip("Requires full page setup with mock use cases")
}

// TestBuildBenchmarkTask_Oracle tests building benchmark task for Oracle Swingbench templates.
func TestBuildBenchmarkTask_Oracle(t *testing.T) {
	t.Skip("Requires full page setup with mock use cases")
}

// TestBuildBenchmarkTask_SQLServer tests building benchmark task for SQL Server HammerDB templates.
func TestBuildBenchmarkTask_SQLServer(t *testing.T) {
	t.Skip("Requires full page setup with mock use cases")
}
