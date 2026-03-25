// Package template provides the core domain model for benchmark templates.
package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
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

const (
	ToolSysbench   = "sysbench"
	ToolSwingbench = "swingbench"
	ToolHammerDB   = "hammerdb"
)

var (
	validTools = map[string]struct{}{
		ToolSysbench:   {},
		ToolSwingbench: {},
		ToolHammerDB:   {},
	}
	validDBFamilies = map[string]struct{}{
		"mysql":      {},
		"postgresql": {},
		"oracle":     {},
		"sqlserver":  {},
		"mariadb":    {},
		"db2":        {},
	}
	validConcurrencyModes = map[string]struct{}{
		"threads":      {},
		"users":        {},
		"virtualUsers": {},
	}
	toolDBMatrix = map[string]map[string]struct{}{
		ToolSysbench: {
			"mysql":      {},
			"postgresql": {},
		},
		ToolSwingbench: {
			"oracle": {},
		},
		ToolHammerDB: {
			"oracle":     {},
			"sqlserver":  {},
			"db2":        {},
			"postgresql": {},
			"mysql":      {},
			"mariadb":    {},
		},
	}
	toolWorkloadMatrix = map[string]map[string]struct{}{
		ToolSysbench: {
			"oltp-read-write":   {},
			"oltp-read-only":    {},
			"oltp-write-only":   {},
			"oltp-point-select": {},
		},
		ToolSwingbench: {
			"order-entry":   {},
			"sales-history": {},
			"stress-test":   {},
		},
		ToolHammerDB: {
			"tproc-c": {},
			"tproc-h": {},
		},
	}
	toolConcurrencyMatrix = map[string]map[string]struct{}{
		ToolSysbench: {
			"threads": {},
		},
		ToolSwingbench: {
			"users": {},
		},
		ToolHammerDB: {
			"virtualUsers": {},
		},
	}
	allowedPhases = map[string]struct{}{
		"prepare": {},
		"warmup":  {},
		"run":     {},
		"cleanup": {},
	}
)

// Template represents a benchmark template.
// It keeps the current Templates canonical model while retaining legacy fields
// used by the execution path.
type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tool        string `json:"tool"`
	Version     string `json:"version"`
	CreatedAt   string `json:"createdAt,omitempty"`
	IsBuiltin   bool   `json:"is_builtin,omitempty"`
	ProfileType string `json:"profile_type,omitempty"`
	Goal        string `json:"goal,omitempty"`
	Readonly    bool   `json:"readonly,omitempty"`

	SourceAlignment string                 `json:"source_alignment,omitempty"`
	PrepareConfig   map[string]interface{} `json:"prepare_config,omitempty"`
	RunConfig       map[string]interface{} `json:"run_config,omitempty"`
	CleanupConfig   map[string]interface{} `json:"cleanup_config,omitempty"`
	Metrics         []string               `json:"metrics,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	TestPosition    string                 `json:"test_position,omitempty"`

	// Current canonical model used by Templates UI/backend CRUD.
	DBFamily       string        `json:"dbFamily,omitempty"`
	WorkloadFamily string        `json:"workloadFamily,omitempty"`
	Compatibility  Compatibility `json:"compatibility,omitempty"`
	Phases         PhaseSet      `json:"phases,omitempty"`
	Runtime        Runtime       `json:"runtime,omitempty"`
	ToolConfig     ToolConfig    `json:"toolConfig,omitempty"`

	// Legacy fields retained for existing benchmark execution code paths.
	DatabaseTypes   []string               `json:"database_types,omitempty"`
	Parameters      map[string]Parameter   `json:"parameters,omitempty"`
	CommandTemplate CommandTemplate        `json:"command_template,omitempty"`
	OutputParser    OutputParser           `json:"output_parser,omitempty"`
	CustomData      map[string]interface{} `json:"custom_data,omitempty"`
}

type Compatibility struct {
	SupportedDatabases []string `json:"supportedDatabases,omitempty"`
	SupportedVersions  []string `json:"supportedVersions,omitempty"`
	CompatibilityNotes string   `json:"compatibilityNotes,omitempty"`
	RequiresPrivileges []string `json:"requiresPrivileges,omitempty"`
	Constraints        []string `json:"constraints,omitempty"`
}

type PhaseConfig struct {
	Enabled  bool                   `json:"enabled"`
	Required bool                   `json:"required"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

type PhaseSet struct {
	Prepare PhaseConfig `json:"prepare"`
	Warmup  PhaseConfig `json:"warmup"`
	Run     PhaseConfig `json:"run"`
	Cleanup PhaseConfig `json:"cleanup"`
}

func (p *PhaseSet) UnmarshalJSON(data []byte) error {
	type phaseSetJSON struct {
		Prepare  PhaseConfig `json:"prepare"`
		Warmup   PhaseConfig `json:"warmup"`
		Run      PhaseConfig `json:"run"`
		Cleanup  PhaseConfig `json:"cleanup"`
		Build    PhaseConfig `json:"build"`
		Generate PhaseConfig `json:"generate"`
		Verify   PhaseConfig `json:"verify"`
		Delete   PhaseConfig `json:"delete"`
	}

	var payload phaseSetJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	p.Prepare = payload.Prepare
	p.Warmup = payload.Warmup
	p.Run = payload.Run
	p.Cleanup = payload.Cleanup

	absorbLegacyPhase(&p.Prepare, payload.Build)
	absorbLegacyPhase(&p.Prepare, payload.Generate)
	absorbLegacyPhase(&p.Cleanup, payload.Verify)
	absorbLegacyPhase(&p.Cleanup, payload.Delete)
	if p.Run.Enabled && !p.Prepare.Enabled && !p.Cleanup.Enabled && (payload.Build.Enabled || payload.Generate.Enabled || payload.Delete.Enabled) {
		p.Prepare.Enabled = true
		p.Cleanup.Enabled = true
	}
	p.normalize("")
	return nil
}

func absorbLegacyPhase(target *PhaseConfig, legacy PhaseConfig) {
	if !legacy.Enabled && !legacy.Required && len(legacy.Params) == 0 {
		return
	}
	if legacy.Enabled {
		target.Enabled = true
	}
	if legacy.Required {
		target.Required = true
	}
	if len(legacy.Params) == 0 {
		return
	}
	if target.Params == nil {
		target.Params = map[string]interface{}{}
	}
	for key, value := range legacy.Params {
		if _, exists := target.Params[key]; !exists {
			target.Params[key] = value
		}
	}
}

func (p PhaseSet) MarshalJSON() ([]byte, error) {
	type phaseSetJSON struct {
		Prepare PhaseConfig `json:"prepare"`
		Warmup  PhaseConfig `json:"warmup"`
		Run     PhaseConfig `json:"run"`
		Cleanup PhaseConfig `json:"cleanup"`
	}

	return json.Marshal(phaseSetJSON{
		Prepare: p.Prepare,
		Warmup:  p.Warmup,
		Run:     p.Run,
		Cleanup: p.Cleanup,
	})
}

type Runtime struct {
	Concurrency           Concurrency `json:"concurrency"`
	DurationSeconds       int         `json:"durationSeconds"`
	WarmupSeconds         int         `json:"warmupSeconds"`
	RampUpSeconds         int         `json:"rampUpSeconds"`
	ReportIntervalSeconds int         `json:"reportIntervalSeconds"`
	Percentile            int         `json:"percentile"`
	Iterations            int         `json:"iterations"`
	RateLimit             int         `json:"rateLimit"`
	ValidationEnabled     bool        `json:"validationEnabled"`
	Notes                 string      `json:"notes,omitempty"`
}

type Concurrency struct {
	Mode  string `json:"mode"`
	Value int    `json:"value"`
}

type ToolConfig struct {
	Sysbench   SysbenchConfig   `json:"sysbench"`
	Swingbench SwingbenchConfig `json:"swingbench"`
	HammerDB   HammerDBConfig   `json:"hammerdb"`
}

type SysbenchConfig struct {
	DBDriver     string `json:"dbDriver,omitempty"`
	ScriptType   string `json:"scriptType,omitempty"`
	Tables       int    `json:"tables,omitempty"`
	TableSize    int    `json:"tableSize,omitempty"`
	ReportChecks bool   `json:"reportChecks"`
	ExtraCLIArgs string `json:"extraCliArgs,omitempty"`
}

type SwingbenchConfig struct {
	Benchmark       string `json:"benchmark,omitempty"`
	Frontend        string `json:"frontend,omitempty"`
	ConfigMode      string `json:"configMode,omitempty"`
	WizardOperation string `json:"wizardOperation,omitempty"`
	UserCount       int    `json:"userCount,omitempty"`
	RunTimeSeconds  int    `json:"runTimeSeconds,omitempty"`
	MinThinkTime    int    `json:"minThinkTime,omitempty"`
	MaxThinkTime    int    `json:"maxThinkTime,omitempty"`
	XMLOverrides    string `json:"xmlOverrides,omitempty"`
}

type HammerDBConfig struct {
	Benchmark      string `json:"benchmark,omitempty"`
	VirtualUsers   int    `json:"virtualUsers,omitempty"`
	Warehouses     int    `json:"warehouses,omitempty"`
	ScaleFactor    int    `json:"scaleFactor,omitempty"`
	TimeProfile    bool   `json:"timeProfile"`
	StepTesting    bool   `json:"stepTesting"`
	XMLConnectPool bool   `json:"xmlConnectPool"`
	AdvancedNotes  string `json:"advancedNotes,omitempty"`
}

// Parameter defines a configurable legacy parameter.
type Parameter struct {
	Type    ParameterType          `json:"type"`
	Label   string                 `json:"label"`
	Default interface{}            `json:"default"`
	Min     *int                   `json:"min,omitempty"`
	Max     *int                   `json:"max,omitempty"`
	Options []string               `json:"options,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

type ParameterType string

const (
	ParameterTypeInteger ParameterType = "integer"
	ParameterTypeNumber  ParameterType = "number"
	ParameterTypeString  ParameterType = "string"
	ParameterTypeBoolean ParameterType = "boolean"
	ParameterTypeEnum    ParameterType = "enum"
)

type CommandTemplate struct {
	Prepare string `json:"prepare,omitempty"`
	Run     string `json:"run,omitempty"`
	Cleanup string `json:"cleanup,omitempty"`
}

type OutputParser struct {
	Type     ParserType             `json:"type,omitempty"`
	Patterns map[string]string      `json:"patterns,omitempty"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

type ParserType string

const (
	ParserTypeRegex ParserType = "regex"
	ParserTypeJSON  ParserType = "json"
	ParserTypeCSV   ParserType = "csv"
)

// NewPhaseSet returns the default phase configuration.
func NewPhaseSet() PhaseSet {
	return PhaseSet{
		Prepare: PhaseConfig{Params: map[string]interface{}{}},
		Warmup:  PhaseConfig{Params: map[string]interface{}{}},
		Run:     PhaseConfig{Enabled: true, Required: true, Params: map[string]interface{}{}},
		Cleanup: PhaseConfig{Params: map[string]interface{}{}},
	}
}

// Normalize fills defaults and compatibility aliases for the canonical model.
func (t *Template) Normalize() {
	now := time.Now().UTC().Format(time.RFC3339)
	if t.Version == "" {
		t.Version = "0.1.0"
	}
	if t.CreatedAt == "" {
		t.CreatedAt = now
	}
	if t.DBFamily != "" {
		t.DatabaseTypes = []string{t.DBFamily}
	}
	if len(t.Compatibility.SupportedDatabases) == 0 && t.DBFamily != "" {
		t.Compatibility.SupportedDatabases = []string{t.DBFamily}
	}
	t.Readonly = t.IsBuiltin
	t.Phases.normalize(t.Tool)
	t.Runtime.normalize()
	t.ToolConfig.normalize(t.Tool, t.DBFamily, t.WorkloadFamily, t.Runtime.Concurrency.Value)
}

func (p *PhaseSet) normalize(tool string) {
	defaults := NewPhaseSet()
	mergePhase := func(target *PhaseConfig, fallback PhaseConfig) {
		if target.Params == nil {
			target.Params = map[string]interface{}{}
		}
		if target.Required {
			target.Enabled = true
		}
		if !target.Enabled && !target.Required && len(target.Params) == 0 && fallback.Required {
			target.Required = fallback.Required
			target.Enabled = fallback.Enabled
		}
	}
	mergePhase(&p.Prepare, defaults.Prepare)
	mergePhase(&p.Warmup, defaults.Warmup)
	mergePhase(&p.Run, defaults.Run)
	mergePhase(&p.Cleanup, defaults.Cleanup)
	if !p.Run.Enabled {
		p.Run.Enabled = true
	}
	p.Run.Required = true
}

func (r *Runtime) normalize() {
	if r.Concurrency.Mode == "" {
		r.Concurrency.Mode = "threads"
	}
	if r.Concurrency.Value == 0 {
		r.Concurrency.Value = 16
	}
	if r.DurationSeconds == 0 {
		r.DurationSeconds = 300
	}
	if r.WarmupSeconds < 0 {
		r.WarmupSeconds = 0
	}
	if r.RampUpSeconds < 0 {
		r.RampUpSeconds = 0
	}
	if r.ReportIntervalSeconds == 0 {
		r.ReportIntervalSeconds = 10
	}
	if r.Percentile == 0 {
		r.Percentile = 95
	}
}

func (c *ToolConfig) normalize(tool, dbFamily, workload string, concurrency int) {
	if c.Sysbench.DBDriver == "" {
		if dbFamily == "postgresql" {
			c.Sysbench.DBDriver = "pgsql"
		} else {
			c.Sysbench.DBDriver = "mysql"
		}
	}
	if c.Sysbench.Tables == 0 {
		c.Sysbench.Tables = 10
	}
	if c.Sysbench.TableSize == 0 {
		c.Sysbench.TableSize = 100000
	}
	if c.Swingbench.UserCount == 0 && concurrency > 0 {
		c.Swingbench.UserCount = concurrency
	}
	if c.Swingbench.RunTimeSeconds == 0 {
		c.Swingbench.RunTimeSeconds = 1800
	}
	if c.HammerDB.Benchmark == "" && workload != "" {
		c.HammerDB.Benchmark = workload
	}
	if c.HammerDB.VirtualUsers == 0 && concurrency > 0 {
		c.HammerDB.VirtualUsers = concurrency
	}
	if workload == "tproc-c" && c.HammerDB.Warehouses == 0 {
		c.HammerDB.Warehouses = 10
	}
	if workload == "tproc-h" && c.HammerDB.ScaleFactor == 0 {
		c.HammerDB.ScaleFactor = 10
	}
}

// IsReadOnly returns true when the template cannot be updated/deleted directly.
func (t *Template) IsReadOnly() bool {
	return t.IsBuiltin || t.Readonly
}

// SupportsDatabase checks if the template supports a specific database type.
func (t *Template) SupportsDatabase(dbType string) bool {
	dbType = strings.ToLower(strings.TrimSpace(dbType))
	for _, supported := range t.DatabaseTypes {
		if strings.ToLower(strings.TrimSpace(supported)) == dbType {
			return true
		}
	}
	if t.DBFamily != "" {
		return strings.EqualFold(t.DBFamily, dbType)
	}
	return false
}

// Validate validates the template.
func (t *Template) Validate() error {
	if t.usesCanonicalModel() {
		return t.validateCanonical()
	}
	return t.validateLegacy()
}

func (t *Template) usesCanonicalModel() bool {
	return t.DBFamily != "" ||
		t.WorkloadFamily != "" ||
		t.Runtime.DurationSeconds != 0 ||
		t.Runtime.Concurrency.Value != 0 ||
		t.Runtime.Concurrency.Mode != "" ||
		len(t.Compatibility.SupportedDatabases) > 0 ||
		len(t.Compatibility.SupportedVersions) > 0 ||
		t.ToolConfig.Sysbench.ScriptType != "" ||
		t.ToolConfig.Swingbench.Benchmark != "" ||
		t.ToolConfig.HammerDB.Benchmark != ""
}

func (t *Template) validateCanonical() error {
	t.Normalize()

	if strings.TrimSpace(t.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrTemplateInvalid)
	}
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrTemplateInvalid)
	}
	if _, ok := validTools[t.Tool]; !ok {
		return fmt.Errorf("%w: invalid tool '%s'", ErrTemplateInvalid, t.Tool)
	}
	if _, ok := validDBFamilies[t.DBFamily]; !ok {
		return fmt.Errorf("%w: invalid dbFamily '%s'", ErrTemplateInvalid, t.DBFamily)
	}
	if _, ok := toolDBMatrix[t.Tool][t.DBFamily]; !ok {
		return fmt.Errorf("%w: tool '%s' does not support dbFamily '%s'", ErrTemplateInvalid, t.Tool, t.DBFamily)
	}
	if _, ok := toolWorkloadMatrix[t.Tool][t.WorkloadFamily]; !ok {
		return fmt.Errorf("%w: tool '%s' does not support workloadFamily '%s'", ErrTemplateInvalid, t.Tool, t.WorkloadFamily)
	}
	if _, ok := validConcurrencyModes[t.Runtime.Concurrency.Mode]; !ok {
		return fmt.Errorf("%w: invalid concurrency mode '%s'", ErrTemplateInvalid, t.Runtime.Concurrency.Mode)
	}
	if _, ok := toolConcurrencyMatrix[t.Tool][t.Runtime.Concurrency.Mode]; !ok {
		return fmt.Errorf("%w: tool '%s' does not support concurrency mode '%s'", ErrTemplateInvalid, t.Tool, t.Runtime.Concurrency.Mode)
	}
	if t.Runtime.Concurrency.Value < 1 {
		return fmt.Errorf("%w: concurrency value must be >= 1", ErrTemplateInvalid)
	}
	if t.Runtime.DurationSeconds < 1 {
		return fmt.Errorf("%w: durationSeconds must be >= 1", ErrTemplateInvalid)
	}
	if t.Runtime.WarmupSeconds < 0 || t.Runtime.RampUpSeconds < 0 || t.Runtime.ReportIntervalSeconds < 0 || t.Runtime.Iterations < 0 || t.Runtime.RateLimit < 0 {
		return fmt.Errorf("%w: runtime values cannot be negative", ErrTemplateInvalid)
	}
	if !t.Phases.Run.Enabled {
		return fmt.Errorf("%w: run phase is required", ErrTemplateInvalid)
	}
	if err := t.validatePhaseSet(); err != nil {
		return err
	}
	if err := t.validateToolConfig(); err != nil {
		return err
	}
	return nil
}

func (t *Template) validatePhaseSet() error {
	phaseChecks := map[string]PhaseConfig{
		"prepare": t.Phases.Prepare,
		"warmup":  t.Phases.Warmup,
		"run":     t.Phases.Run,
		"cleanup": t.Phases.Cleanup,
	}
	for phase, cfg := range phaseChecks {
		if _, ok := allowedPhases[phase]; !ok {
			return fmt.Errorf("%w: invalid phase '%s'", ErrTemplateInvalid, phase)
		}
		if cfg.Required && !cfg.Enabled {
			return fmt.Errorf("%w: required phase '%s' must be enabled", ErrTemplateInvalid, phase)
		}
	}
	return nil
}

func (t *Template) validateToolConfig() error {
	switch t.Tool {
	case ToolSysbench:
		if strings.TrimSpace(t.ToolConfig.Sysbench.ScriptType) == "" {
			return fmt.Errorf("%w: sysbench scriptType is required", ErrTemplateInvalid)
		}
		if t.ToolConfig.Sysbench.Tables < 1 || t.ToolConfig.Sysbench.TableSize < 1 {
			return fmt.Errorf("%w: sysbench tables and tableSize must be >= 1", ErrTemplateInvalid)
		}
	case ToolSwingbench:
		if strings.TrimSpace(t.ToolConfig.Swingbench.Benchmark) == "" {
			return fmt.Errorf("%w: swingbench benchmark is required", ErrTemplateInvalid)
		}
		if t.ToolConfig.Swingbench.UserCount < 1 {
			return fmt.Errorf("%w: swingbench userCount must be >= 1", ErrTemplateInvalid)
		}
	case ToolHammerDB:
		if strings.TrimSpace(t.ToolConfig.HammerDB.Benchmark) == "" {
			return fmt.Errorf("%w: hammerdb benchmark is required", ErrTemplateInvalid)
		}
		if t.ToolConfig.HammerDB.VirtualUsers < 1 {
			return fmt.Errorf("%w: hammerdb virtualUsers must be >= 1", ErrTemplateInvalid)
		}
		if t.WorkloadFamily == "tproc-c" && t.ToolConfig.HammerDB.Warehouses < 1 {
			return fmt.Errorf("%w: hammerdb warehouses must be >= 1", ErrTemplateInvalid)
		}
		if t.WorkloadFamily == "tproc-h" && t.ToolConfig.HammerDB.ScaleFactor < 1 {
			return fmt.Errorf("%w: hammerdb scaleFactor must be >= 1", ErrTemplateInvalid)
		}
	}
	return nil
}

func (t *Template) validateLegacy() error {
	if strings.TrimSpace(t.ID) == "" {
		return fmt.Errorf("%w: ID is required", ErrTemplateInvalid)
	}
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("%w: Name is required", ErrTemplateInvalid)
	}
	if strings.TrimSpace(t.Tool) == "" {
		return fmt.Errorf("%w: Tool is required", ErrTemplateInvalid)
	}
	if len(t.DatabaseTypes) == 0 {
		return fmt.Errorf("%w: At least one database type is required", ErrTemplateInvalid)
	}
	if strings.TrimSpace(t.CommandTemplate.Run) == "" {
		return fmt.Errorf("%w: Run command is required", ErrInvalidCommand)
	}
	for name, param := range t.Parameters {
		if err := param.Validate(); err != nil {
			return fmt.Errorf("parameter '%s': %w", name, err)
		}
	}
	if err := t.OutputParser.Validate(); err != nil {
		return fmt.Errorf("output parser: %w", err)
	}
	return nil
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
	case ParameterTypeNumber:
		if p.Min != nil && p.Max != nil && *p.Min > *p.Max {
			return fmt.Errorf("%w: min (%d) > max (%d)", ErrInvalidParameterType, *p.Min, *p.Max)
		}
	case ParameterTypeEnum:
		if len(p.Options) == 0 {
			return fmt.Errorf("%w: enum type requires options", ErrInvalidParameterType)
		}
	case ParameterTypeString, ParameterTypeBoolean:
	default:
		return fmt.Errorf("%w: unknown type '%s'", ErrInvalidParameterType, p.Type)
	}

	return nil
}

// ValidateDefaultValue checks if the default value is valid for this parameter.
func (p *Parameter) ValidateDefaultValue() error {
	if p.Default == nil {
		return nil
	}

	switch p.Type {
	case ParameterTypeInteger:
		switch v := p.Default.(type) {
		case int:
			if p.Min != nil && v < *p.Min {
				return fmt.Errorf("default value (%d) < min (%d)", v, *p.Min)
			}
			if p.Max != nil && v > *p.Max {
				return fmt.Errorf("default value (%d) > max (%d)", v, *p.Max)
			}
		case float64:
			p.Default = int(v)
		default:
			return fmt.Errorf("default value for integer parameter must be an integer")
		}
	case ParameterTypeNumber:
		switch v := p.Default.(type) {
		case int:
			if p.Min != nil && v < *p.Min {
				return fmt.Errorf("default value (%d) < min (%d)", v, *p.Min)
			}
			if p.Max != nil && v > *p.Max {
				return fmt.Errorf("default value (%d) > max (%d)", v, *p.Max)
			}
			p.Default = float64(v)
		case float64:
			if p.Min != nil && v < float64(*p.Min) {
				return fmt.Errorf("default value (%v) < min (%d)", v, *p.Min)
			}
			if p.Max != nil && v > float64(*p.Max) {
				return fmt.Errorf("default value (%v) > max (%d)", v, *p.Max)
			}
		default:
			return fmt.Errorf("default value for number parameter must be numeric")
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
	case "", ParserTypeRegex:
		for name, pattern := range op.Patterns {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("%w: invalid regex for '%s': %w", ErrInvalidParser, name, err)
			}
		}
	case ParserTypeJSON, ParserTypeCSV:
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
