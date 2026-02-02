// Package template provides the core domain model for benchmark templates.
// Templates define how to execute different benchmark tools (sysbench, swingbench, hammerdb, tpcc).
package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrTemplateInvalid is returned when template validation fails.
	ErrTemplateInvalid = errors.New("template validation failed")

	// ErrInvalidParameterType is returned when parameter type is invalid.
	ErrInvalidParameterType = errors.New("invalid parameter type")

	// ErrInvalidCommand is returned when command template is invalid.
	ErrInvalidCommand = errors.New("invalid command template")

	// ErrInvalidParser is returned when output parser configuration is invalid.
	ErrInvalidParser = errors.New("invalid output parser")
)

// Template represents a benchmark template with all configuration needed to execute it.
// Implements: REQ-TMPL-002 (display template details)
type Template struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Tool            string                 `json:"tool"`
	DatabaseTypes   []string               `json:"database_types"`
	Version         string                 `json:"version"`
	Parameters      map[string]Parameter   `json:"parameters"`
	CommandTemplate CommandTemplate        `json:"command_template"`
	OutputParser    OutputParser           `json:"output_parser"`
	CustomData      map[string]interface{} `json:"custom_data,omitempty"`
}

// Parameter defines a configurable parameter for a template.
// Implements: REQ-TMPL-002 (display parameter configuration)
type Parameter struct {
	Type    ParameterType          `json:"type"`    // integer, string, boolean, enum
	Label   string                 `json:"label"`   // Display label
	Default interface{}            `json:"default"` // Default value
	Min     *int                   `json:"min,omitempty"`
	Max     *int                   `json:"max,omitempty"`
	Options []string               `json:"options,omitempty"` // For enum type
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// ParameterType represents the type of a parameter.
type ParameterType string

const (
	// ParameterTypeInteger is for integer parameters.
	ParameterTypeInteger ParameterType = "integer"
	// ParameterTypeString is for string parameters.
	ParameterTypeString ParameterType = "string"
	// ParameterTypeBoolean is for boolean parameters.
	ParameterTypeBoolean ParameterType = "boolean"
	// ParameterTypeEnum is for enum parameters with predefined options.
	ParameterTypeEnum ParameterType = "enum"
)

// CommandTemplate contains command templates for different execution phases.
// Implements: REQ-EXEC-002 (prepare → warmup → run → cleanup)
type CommandTemplate struct {
	Prepare string `json:"prepare"` // Data preparation command
	Run     string `json:"run"`     // Main benchmark command
	Cleanup string `json:"cleanup"` // Cleanup command
}

// OutputParser defines how to parse benchmark tool output.
type OutputParser struct {
	Type     ParserType             `json:"type"`
	Patterns map[string]string      `json:"patterns,omitempty"` // Regex patterns
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// ParserType represents the type of output parser.
type ParserType string

const (
	// ParserTypeRegex uses regex patterns to extract metrics.
	ParserTypeRegex ParserType = "regex"
	// ParserTypeJSON parses JSON output.
	ParserTypeJSON ParserType = "json"
	// ParserTypeCSV parses CSV output.
	ParserTypeCSV ParserType = "csv"
)

// Validate validates the template for correctness.
// Returns an error if any validation rule fails.
// Implements: REQ-TMPL-004 (validate imported templates)
func (t *Template) Validate() error {
	// Validate required fields
	if t.ID == "" {
		return fmt.Errorf("%w: ID is required", ErrTemplateInvalid)
	}
	if t.Name == "" {
		return fmt.Errorf("%w: Name is required", ErrTemplateInvalid)
	}
	if t.Tool == "" {
		return fmt.Errorf("%w: Tool is required", ErrTemplateInvalid)
	}
	if len(t.DatabaseTypes) == 0 {
		return fmt.Errorf("%w: At least one database type is required", ErrTemplateInvalid)
	}

	// Validate command templates
	if t.CommandTemplate.Run == "" {
		return fmt.Errorf("%w: Run command is required", ErrInvalidCommand)
	}

	// Validate parameters
	for name, param := range t.Parameters {
		if err := param.Validate(); err != nil {
			return fmt.Errorf("parameter '%s': %w", name, err)
		}
	}

	// Validate output parser
	if err := t.OutputParser.Validate(); err != nil {
		return fmt.Errorf("output parser: %w", err)
	}

	return nil
}

// SupportsDatabase checks if the template supports a specific database type.
// Implements: REQ-EXEC-001 (pre-check tool compatibility)
func (t *Template) SupportsDatabase(dbType string) bool {
	dbType = strings.ToLower(strings.TrimSpace(dbType))
	for _, supported := range t.DatabaseTypes {
		if strings.ToLower(supported) == dbType {
			return true
		}
	}
	return false
}

// GetParameter returns a parameter by name, or error if not found.
func (t *Template) GetParameter(name string) (Parameter, error) {
	param, ok := t.Parameters[name]
	if !ok {
		return Parameter{}, fmt.Errorf("parameter '%s' not found", name)
	}
	return param, nil
}

// HasParameter checks if a parameter exists in the template.
func (t *Template) HasParameter(name string) bool {
	_, ok := t.Parameters[name]
	return ok
}

// Validate validates a parameter definition.
func (p *Parameter) Validate() error {
	if p.Label == "" {
		return fmt.Errorf("%w: label is required", ErrInvalidParameterType)
	}

	switch p.Type {
	case ParameterTypeInteger:
		if p.Min != nil && p.Max != nil && *p.Min > *p.Max {
			return fmt.Errorf("%w: min (%d) > max (%d)", ErrInvalidParameterType, *p.Min, *p.Max)
		}
	case ParameterTypeEnum:
		if len(p.Options) == 0 {
			return fmt.Errorf("%w: enum type requires options", ErrInvalidParameterType)
		}
	case ParameterTypeString, ParameterTypeBoolean:
		// No additional validation needed
	default:
		return fmt.Errorf("%w: unknown type '%s'", ErrInvalidParameterType, p.Type)
	}

	return nil
}

// ValidateDefaultValue checks if the default value is valid for this parameter.
func (p *Parameter) ValidateDefaultValue() error {
	if p.Default == nil {
		return nil // No default value is OK
	}

	switch p.Type {
	case ParameterTypeInteger:
		if _, ok := p.Default.(int); !ok {
			if f, ok := p.Default.(float64); ok {
				// JSON unmarshaling converts numbers to float64
				p.Default = int(f)
				return nil
			}
			return fmt.Errorf("default value for integer parameter must be an integer")
		}
		if p.Min != nil {
			min := *p.Min
			if val, ok := p.Default.(int); ok && val < min {
				return fmt.Errorf("default value (%d) < min (%d)", val, min)
			}
		}
		if p.Max != nil {
			max := *p.Max
			if val, ok := p.Default.(int); ok && val > max {
				return fmt.Errorf("default value (%d) > max (%d)", val, max)
			}
		}
	case ParameterTypeString:
		if _, ok := p.Default.(string); !ok {
			return fmt.Errorf("default value for string parameter must be a string")
		}
	case ParameterTypeBoolean:
		if _, ok := p.Default.(bool); !ok {
			return fmt.Errorf("default value for boolean parameter must be a boolean")
		}
	case ParameterTypeEnum:
		strVal, ok := p.Default.(string)
		if !ok {
			return fmt.Errorf("default value for enum parameter must be a string")
		}
		found := false
		for _, opt := range p.Options {
			if opt == strVal {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default value '%s' is not in options", strVal)
		}
	}

	return nil
}

// Validate validates the output parser configuration.
func (op *OutputParser) Validate() error {
	switch op.Type {
	case ParserTypeRegex:
		// Validate all regex patterns compile correctly
		for name, pattern := range op.Patterns {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("%w: invalid regex for '%s': %w", ErrInvalidParser, name, err)
			}
		}
	case ParserTypeJSON, ParserTypeCSV:
		// No additional validation needed
	default:
		return fmt.Errorf("%w: unknown parser type '%s'", ErrInvalidParser, op.Type)
	}
	return nil
}

// ToJSON serializes the template to JSON.
func (t *Template) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

// FromJSON deserializes a template from JSON.
func FromJSON(data []byte) (*Template, error) {
	var tmpl Template
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}
	if err := tmpl.Validate(); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// GetParameterDefault returns the default value for a parameter, or error if not found.
func (t *Template) GetParameterDefault(name string) (interface{}, error) {
	param, err := t.GetParameter(name)
	if err != nil {
		return nil, err
	}
	return param.Default, nil
}

// Clone creates a deep copy of the template.
func (t *Template) Clone() (*Template, error) {
	data, err := t.ToJSON()
	if err != nil {
		return nil, err
	}
	return FromJSON(data)
}
