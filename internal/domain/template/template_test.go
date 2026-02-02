// Package template provides unit tests for the template domain model.
package template

import (
	"testing"
)

// TestTemplate_Validate_ValidTemplate tests validation of a valid template.
func TestTemplate_Validate_ValidTemplate(t *testing.T) {
	tmpl := &Template{
		ID:            "sysbench-oltp-read-write",
		Name:          "Sysbench OLTP Read-Write",
		Description:   "Standard OLTP read-write mixed test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql", "postgresql"},
		Version:       "1.0.0",
		Parameters: map[string]Parameter{
			"threads": {
				Type:    ParameterTypeInteger,
				Label:   "Thread count",
				Default: 8,
				Min:     intPtr(1),
				Max:     intPtr(1024),
			},
			"time": {
				Type:    ParameterTypeInteger,
				Label:   "Runtime (seconds)",
				Default: 60,
				Min:     intPtr(10),
				Max:     intPtr(86400),
			},
		},
		CommandTemplate: CommandTemplate{
			Prepare: "sysbench {db_type} --tables={tables} prepare",
			Run:     "sysbench {db_type} --threads={threads} --time={time} run",
			Cleanup: "sysbench {db_type} cleanup",
		},
		OutputParser: OutputParser{
			Type: ParserTypeRegex,
			Patterns: map[string]string{
				"tps":         `transactions:\s*\(\s*(\d+\.?\d*)\s*per sec\.`,
				"latency_avg": `latency:\s*\(ms\).*?avg=\s*(\d+\.?\d*)`,
			},
		},
	}

	if err := tmpl.Validate(); err != nil {
		t.Errorf("Validate() on valid template returned error: %v", err)
	}
}

// TestTemplate_Validate_MissingID tests validation fails when ID is missing.
func TestTemplate_Validate_MissingID(t *testing.T) {
	tmpl := &Template{
		Name:          "Test Template",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
	}

	if err := tmpl.Validate(); err == nil {
		t.Error("Validate() with missing ID should return error")
	}
}

// TestTemplate_Validate_MissingName tests validation fails when Name is missing.
func TestTemplate_Validate_MissingName(t *testing.T) {
	tmpl := &Template{
		ID:            "test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
	}

	if err := tmpl.Validate(); err == nil {
		t.Error("Validate() with missing Name should return error")
	}
}

// TestTemplate_Validate_MissingTool tests validation fails when Tool is missing.
func TestTemplate_Validate_MissingTool(t *testing.T) {
	tmpl := &Template{
		ID:            "test",
		Name:          "Test",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
	}

	if err := tmpl.Validate(); err == nil {
		t.Error("Validate() with missing Tool should return error")
	}
}

// TestTemplate_Validate_EmptyDatabaseTypes tests validation fails when no database types.
func TestTemplate_Validate_EmptyDatabaseTypes(t *testing.T) {
	tmpl := &Template{
		ID:            "test",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
	}

	if err := tmpl.Validate(); err == nil {
		t.Error("Validate() with empty DatabaseTypes should return error")
	}
}

// TestTemplate_Validate_MissingRunCommand tests validation fails when Run command is missing.
func TestTemplate_Validate_MissingRunCommand(t *testing.T) {
	tmpl := &Template{
		ID:            "test",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: CommandTemplate{
			Prepare: "prepare",
			Cleanup: "cleanup",
		},
	}

	if err := tmpl.Validate(); err == nil {
		t.Error("Validate() with missing Run command should return error")
	}
}

// TestTemplate_Validate_InvalidParameter tests validation of invalid parameters.
func TestTemplate_Validate_InvalidParameter(t *testing.T) {
	tests := []struct {
		name    string
		param   Parameter
		wantErr bool
	}{
		{
			name: "missing label",
			param: Parameter{
				Type: ParameterTypeInteger,
			},
			wantErr: true,
		},
		{
			name: "integer min > max",
			param: Parameter{
				Type:  ParameterTypeInteger,
				Label: "Test",
				Min:   intPtr(100),
				Max:   intPtr(50),
			},
			wantErr: true,
		},
		{
			name: "enum without options",
			param: Parameter{
				Type:  ParameterTypeEnum,
				Label: "Test",
			},
			wantErr: true,
		},
		{
			name: "valid integer parameter",
			param: Parameter{
				Type:    ParameterTypeInteger,
				Label:   "Thread count",
				Default: 8,
				Min:     intPtr(1),
				Max:     intPtr(1024),
			},
			wantErr: false,
		},
		{
			name: "valid enum parameter",
			param: Parameter{
				Type:    ParameterTypeEnum,
				Label:   "Mode",
				Default: "fast",
				Options: []string{"fast", "slow", "medium"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.param.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parameter.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTemplate_Validate_InvalidRegex tests validation fails with invalid regex patterns.
func TestTemplate_Validate_InvalidRegex(t *testing.T) {
	tmpl := &Template{
		ID:            "test",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
		OutputParser: OutputParser{
			Type: ParserTypeRegex,
			Patterns: map[string]string{
				"tps": "[invalid(regex", // Unclosed bracket
			},
		},
	}

	if err := tmpl.Validate(); err == nil {
		t.Error("Validate() with invalid regex should return error")
	}
}

// TestTemplate_SupportsDatabase tests database type support checking.
func TestTemplate_SupportsDatabase(t *testing.T) {
	tmpl := &Template{
		ID:            "test",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql", "postgresql"},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
	}

	tests := []struct {
		name   string
		dbType string
		want   bool
	}{
		{"mysql supported", "mysql", true},
		{"MySQL supported (case insensitive)", "MySQL", true},
		{"postgresql supported", "postgresql", true},
		{"PostgreSQL supported (case insensitive)", "PostgreSQL", true},
		{"oracle not supported", "oracle", false},
		{"sqlserver not supported", "sqlserver", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tmpl.SupportsDatabase(tt.dbType); got != tt.want {
				t.Errorf("SupportsDatabase(%s) = %v, want %v", tt.dbType, got, tt.want)
			}
		})
	}
}

// TestTemplate_GetParameter tests retrieving parameters by name.
func TestTemplate_GetParameter(t *testing.T) {
	tmpl := &Template{
		ID:            "test",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
		Parameters: map[string]Parameter{
			"threads": {
				Type:    ParameterTypeInteger,
				Label:   "Threads",
				Default: 8,
			},
		},
	}

	t.Run("existing parameter", func(t *testing.T) {
		param, err := tmpl.GetParameter("threads")
		if err != nil {
			t.Errorf("GetParameter() returned error: %v", err)
		}
		if param.Label != "Threads" {
			t.Errorf("GetParameter() Label = %s, want 'Threads'", param.Label)
		}
	})

	t.Run("non-existing parameter", func(t *testing.T) {
		_, err := tmpl.GetParameter("nonexistent")
		if err == nil {
			t.Error("GetParameter() with non-existing name should return error")
		}
	})
}

// TestTemplate_HasParameter tests checking parameter existence.
func TestTemplate_HasParameter(t *testing.T) {
	tmpl := &Template{
		ID:            "test",
		Name:          "Test",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
		Parameters: map[string]Parameter{
			"threads": {
				Type:    ParameterTypeInteger,
				Label:   "Threads",
				Default: 8,
			},
		},
	}

	if !tmpl.HasParameter("threads") {
		t.Error("HasParameter(threads) should return true")
	}

	if tmpl.HasParameter("nonexistent") {
		t.Error("HasParameter(nonexistent) should return false")
	}
}

// TestParameter_ValidateDefaultValue tests default value validation.
func TestParameter_ValidateDefaultValue(t *testing.T) {
	tests := []struct {
		name    string
		param   Parameter
		wantErr bool
	}{
		{
			name: "nil default is OK",
			param: Parameter{
				Type:  ParameterTypeInteger,
				Label: "Test",
			},
			wantErr: false,
		},
		{
			name: "integer with valid default",
			param: Parameter{
				Type:    ParameterTypeInteger,
				Label:   "Threads",
				Default: 8,
				Min:     intPtr(1),
				Max:     intPtr(1024),
			},
			wantErr: false,
		},
		{
			name: "integer default < min",
			param: Parameter{
				Type:    ParameterTypeInteger,
				Label:   "Threads",
				Default: 0,
				Min:     intPtr(1),
				Max:     intPtr(1024),
			},
			wantErr: true,
		},
		{
			name: "integer default > max",
			param: Parameter{
				Type:    ParameterTypeInteger,
				Label:   "Threads",
				Default: 2000,
				Min:     intPtr(1),
				Max:     intPtr(1024),
			},
			wantErr: true,
		},
		{
			name: "string with valid default",
			param: Parameter{
				Type:    ParameterTypeString,
				Label:   "Name",
				Default: "test",
			},
			wantErr: false,
		},
		{
			name: "string with wrong default type",
			param: Parameter{
				Type:    ParameterTypeString,
				Label:   "Name",
				Default: 123,
			},
			wantErr: true,
		},
		{
			name: "boolean with valid default",
			param: Parameter{
				Type:    ParameterTypeBoolean,
				Label:   "Enabled",
				Default: true,
			},
			wantErr: false,
		},
		{
			name: "enum with valid default",
			param: Parameter{
				Type:    ParameterTypeEnum,
				Label:   "Mode",
				Default: "fast",
				Options: []string{"fast", "slow"},
			},
			wantErr: false,
		},
		{
			name: "enum with invalid default",
			param: Parameter{
				Type:    ParameterTypeEnum,
				Label:   "Mode",
				Default: "invalid",
				Options: []string{"fast", "slow"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.param.ValidateDefaultValue()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDefaultValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOutputParser_Validate tests output parser validation.
func TestOutputParser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		parser  OutputParser
		wantErr bool
	}{
		{
			name: "valid regex parser",
			parser: OutputParser{
				Type: ParserTypeRegex,
				Patterns: map[string]string{
					"tps": `transactions:\s*\(\s*(\d+\.?\d*)\)`,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid regex pattern",
			parser: OutputParser{
				Type: ParserTypeRegex,
				Patterns: map[string]string{
					"tps": "[invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "json parser (no validation)",
			parser: OutputParser{
				Type: ParserTypeJSON,
			},
			wantErr: false,
		},
		{
			name: "csv parser (no validation)",
			parser: OutputParser{
				Type: ParserTypeCSV,
			},
			wantErr: false,
		},
		{
			name: "unknown parser type",
			parser: OutputParser{
				Type: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parser.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OutputParser.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTemplate_ToJSON_FromJSON tests JSON serialization and deserialization.
func TestTemplate_ToJSON_FromJSON(t *testing.T) {
	original := &Template{
		ID:            "test-template",
		Name:          "Test Template",
		Description:   "A test template",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql", "postgresql"},
		Version:       "1.0.0",
		Parameters: map[string]Parameter{
			"threads": {
				Type:    ParameterTypeInteger,
				Label:   "Threads",
				Default: 8,
			},
		},
		CommandTemplate: CommandTemplate{
			Run: "sysbench oltp run",
		},
		OutputParser: OutputParser{
			Type: ParserTypeRegex,
		},
	}

	// Serialize
	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() returned error: %v", err)
	}

	// Deserialize
	restored, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() returned error: %v", err)
	}

	// Verify
	if restored.ID != original.ID {
		t.Errorf("ID = %s, want %s", restored.ID, original.ID)
	}
	if restored.Name != original.Name {
		t.Errorf("Name = %s, want %s", restored.Name, original.Name)
	}
	if len(restored.Parameters) != len(original.Parameters) {
		t.Errorf("Parameters count = %d, want %d", len(restored.Parameters), len(original.Parameters))
	}
}

// TestTemplate_Clone tests template cloning.
func TestTemplate_Clone(t *testing.T) {
	original := &Template{
		ID:            "test-template",
		Name:          "Test Template",
		Tool:          "sysbench",
		DatabaseTypes: []string{"mysql"},
		CommandTemplate: CommandTemplate{
			Run: "run command",
		},
		OutputParser: OutputParser{
			Type: ParserTypeRegex,
		},
	}

	cloned, err := original.Clone()
	if err != nil {
		t.Fatalf("Clone() returned error: %v", err)
	}

	// Verify equality
	if cloned.ID != original.ID {
		t.Errorf("Cloned ID = %s, want %s", cloned.ID, original.ID)
	}

	// Modify original and verify clone is independent
	original.Name = "Modified"
	if cloned.Name == original.Name {
		t.Error("Clone is not independent from original")
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
