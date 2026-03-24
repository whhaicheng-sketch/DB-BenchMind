package template

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type builtinTemplateDefinition struct {
	ID              string                 `json:"id"`
	DB              string                 `json:"db"`
	Profile         string                 `json:"profile"`
	Tool            string                 `json:"tool"`
	SourceAlignment string                 `json:"source_alignment"`
	Goal            string                 `json:"goal"`
	Prepare         map[string]interface{} `json:"prepare"`
	Run             map[string]interface{} `json:"run"`
	Cleanup         map[string]interface{} `json:"cleanup"`
	TestPosition    string                 `json:"test_position"`
}

// DefaultSeedTemplates returns the builtin templates synchronized at startup.
func DefaultSeedTemplates() []*Template {
	var defs []builtinTemplateDefinition
	if err := json.Unmarshal([]byte(defaultBuiltinTemplateJSON), &defs); err != nil {
		panic(fmt.Errorf("parse builtin template seed json: %w", err))
	}

	now := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	templates := make([]*Template, 0, len(defs))
	for _, def := range defs {
		templates = append(templates, mustTemplate(buildBuiltinTemplate(def, now)))
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})
	return templates
}

func buildBuiltinTemplate(def builtinTemplateDefinition, createdAt string) *Template {
	dbFamily := strings.ToLower(def.DB)
	tool := mapBuiltinTool(def.Tool)
	workloadFamily := mapBuiltinWorkload(dbFamily, tool)
	runtime := buildRuntime(def, tool)
	description := buildBuiltinDescription(def)
	tags := buildBuiltinTags(def)

	tmpl := &Template{
		ID:              def.ID,
		Name:            formatBuiltinName(dbFamily, def.Profile),
		Description:     description,
		Tool:            tool,
		Version:         "1.0.0",
		CreatedAt:       createdAt,
		IsBuiltin:       true,
		Readonly:        true,
		ProfileType:     def.Profile,
		Goal:            def.Goal,
		SourceAlignment: def.SourceAlignment,
		PrepareConfig:   cloneConfigMap(def.Prepare),
		RunConfig:       cloneConfigMap(def.Run),
		CleanupConfig:   cloneConfigMap(def.Cleanup),
		Metrics:         extractMetrics(def),
		Tags:            tags,
		TestPosition:    def.TestPosition,
		DBFamily:        dbFamily,
		WorkloadFamily:  workloadFamily,
		Compatibility: Compatibility{
			SupportedDatabases: []string{dbFamily},
			CompatibilityNotes: buildCompatibilityNotes(def),
			Constraints:        buildConstraints(def),
		},
		Phases: PhaseSet{
			Prepare: PhaseConfig{Enabled: true, Params: cloneConfigMap(def.Prepare)},
			Warmup:  PhaseConfig{Enabled: true, Params: map[string]interface{}{}},
			Run:     PhaseConfig{Enabled: true, Required: true, Params: cloneConfigMap(def.Run)},
			Cleanup: PhaseConfig{Enabled: true, Params: cloneConfigMap(def.Cleanup)},
		},
		Runtime:    runtime,
		ToolConfig: buildToolConfig(def, dbFamily, workloadFamily, runtime),
	}
	tmpl.Normalize()
	return tmpl
}

func mapBuiltinTool(tool string) string {
	switch strings.ToLower(tool) {
	case "sysbench oltp_read_write.lua":
		return ToolSysbench
	case "swingbench oe":
		return ToolSwingbench
	case "hammerdb tproc-c":
		return ToolHammerDB
	default:
		panic(fmt.Sprintf("unsupported builtin tool: %s", tool))
	}
}

func mapBuiltinWorkload(dbFamily, tool string) string {
	switch tool {
	case ToolSysbench:
		return "oltp-read-write"
	case ToolSwingbench:
		return "order-entry"
	case ToolHammerDB:
		return "tproc-c"
	default:
		panic(fmt.Sprintf("unsupported builtin workload mapping for %s/%s", dbFamily, tool))
	}
}

func buildRuntime(def builtinTemplateDefinition, tool string) Runtime {
	runtime := Runtime{
		DurationSeconds:       extractDurationSeconds(def),
		ReportIntervalSeconds: extractReportInterval(def),
		Percentile:            95,
		ValidationEnabled:     def.Profile == "test",
		Notes:                 buildRuntimeNotes(def),
	}

	switch tool {
	case ToolSysbench:
		runtime.Concurrency = Concurrency{Mode: "threads", Value: extractConcurrency(def.Run, "threads")}
	case ToolSwingbench:
		runtime.Concurrency = Concurrency{Mode: "users", Value: extractConcurrency(def.Run, "default_users")}
	case ToolHammerDB:
		runtime.Concurrency = Concurrency{Mode: "virtualUsers", Value: extractConcurrency(def.Run, "virtual_users")}
	}
	return runtime
}

func buildToolConfig(def builtinTemplateDefinition, dbFamily, workloadFamily string, runtime Runtime) ToolConfig {
	cfg := ToolConfig{}
	switch mapBuiltinTool(def.Tool) {
	case ToolSysbench:
		driver := "mysql"
		if dbFamily == "postgresql" {
			driver = "pgsql"
		}
		cfg.Sysbench = SysbenchConfig{
			DBDriver:     driver,
			ScriptType:   "oltp_read_write",
			Tables:       extractPrepareInt(def.Prepare, "tables", 1),
			TableSize:    extractPrepareInt(def.Prepare, "table_size", 1000),
			ReportChecks: true,
		}
	case ToolSwingbench:
		cfg.Swingbench = SwingbenchConfig{
			Benchmark:       "orderEntry",
			Frontend:        "charbench",
			ConfigMode:      "managed",
			WizardOperation: "generate",
			UserCount:       runtime.Concurrency.Value,
			RunTimeSeconds:  runtime.DurationSeconds,
			XMLOverrides:    buildSwingbenchOverrides(def),
		}
	case ToolHammerDB:
		cfg.HammerDB = HammerDBConfig{
			Benchmark:    workloadFamily,
			VirtualUsers: runtime.Concurrency.Value,
			Warehouses:   extractPrepareInt(def.Prepare, "warehouses", 10),
			TimeProfile:  true,
			AdvancedNotes: firstString(
				stringValue(def.Run, "tuning_rule"),
				stringValue(def.Prepare, "notes"),
			),
		}
	}
	return cfg
}

func buildBuiltinDescription(def builtinTemplateDefinition) string {
	source := def.SourceAlignment
	if source == "engineered_split_from_baseline" {
		source = "engineered split from baseline"
	}
	if source == "engineered_minimal" {
		return fmt.Sprintf("%s Built-in smoke template for prepare -> run -> cleanup verification with minimal initialization; not for performance conclusions.", def.Goal)
	}
	return fmt.Sprintf("%s Built-in %s template.", def.Goal, source)
}

func buildCompatibilityNotes(def builtinTemplateDefinition) string {
	switch def.SourceAlignment {
	case "engineered_split_from_baseline":
		return "This built-in template is an engineered split from an existing baseline, not a native dual-model definition in the original document."
	case "engineered_minimal":
		return "This built-in test template is engineered for minimal smoke validation and should not be used for performance conclusions."
	case "direct_from_doc_as_baseline":
		return "This built-in template keeps the original document baseline and exposes it as the io_bound model."
	default:
		return "This built-in template is mapped directly from the source document."
	}
}

func buildConstraints(def builtinTemplateDefinition) []string {
	constraints := []string{}
	if tuning := stringValue(def.Run, "tuning_rule"); tuning != "" {
		constraints = append(constraints, tuning)
	}
	if notes := stringValue(def.Prepare, "notes"); notes != "" {
		constraints = append(constraints, notes)
	}
	if def.Profile == "test" {
		constraints = append(constraints, "Use only for smoke validation of prepare -> run -> cleanup.")
	}
	return constraints
}

func buildRuntimeNotes(def builtinTemplateDefinition) string {
	return strings.TrimSpace(strings.Join([]string{
		stringValue(def.Run, "duration"),
		stringValue(def.Run, "tuning_rule"),
		stringValue(def.Prepare, "notes"),
	}, " "))
}

func buildSwingbenchOverrides(def builtinTemplateDefinition) string {
	workloadMix := stringValue(def.Run, "workload_mix")
	if workloadMix == "" {
		return ""
	}
	return fmt.Sprintf("workload_mix=%s", workloadMix)
}

func buildBuiltinTags(def builtinTemplateDefinition) []string {
	tags := []string{"builtin", strings.ToLower(def.DB), def.Profile}
	if def.TestPosition != "" {
		tags = append(tags, def.TestPosition)
	}
	switch def.SourceAlignment {
	case "engineered_split_from_baseline":
		tags = append(tags, "engineered", "split")
	case "engineered_minimal":
		tags = append(tags, "engineered", "minimal")
	case "direct_from_doc", "direct_from_doc_as_baseline":
		tags = append(tags, "direct_from_doc")
	}
	if def.Profile == "test" {
		tags = append(tags, "smoke", "minimal")
	} else {
		tags = append(tags, "performance")
	}
	return tags
}

func formatBuiltinName(dbFamily, profile string) string {
	label := map[string]string{
		"oracle":     "Oracle",
		"mysql":      "MySQL",
		"sqlserver":  "SQL Server",
		"postgresql": "PostgreSQL",
	}[dbFamily]

	profileLabel := map[string]string{
		"cpu_bound": "CPU Bound",
		"io_bound":  "IO Bound",
		"test":      "Test",
	}[profile]
	return fmt.Sprintf("%s %s", label, profileLabel)
}

func extractMetrics(def builtinTemplateDefinition) []string {
	runMetrics, _ := def.Run["metrics"].([]interface{})
	if len(runMetrics) == 0 {
		return nil
	}
	metrics := make([]string, 0, len(runMetrics))
	for _, metric := range runMetrics {
		if text, ok := metric.(string); ok && strings.TrimSpace(text) != "" {
			metrics = append(metrics, text)
		}
	}
	return metrics
}

func extractPrepareInt(prepare map[string]interface{}, key string, fallback int) int {
	params, _ := prepare["params"].(map[string]interface{})
	if params == nil {
		return fallback
	}
	if value, ok := params[key]; ok {
		if intValue, ok := toInt(value); ok {
			return intValue
		}
	}
	return fallback
}

func extractConcurrency(run map[string]interface{}, key string) int {
	if value, ok := run[key]; ok {
		if intValue, ok := toInt(value); ok && intValue > 0 {
			return intValue
		}
	}
	params, _ := run["params"].(map[string]interface{})
	if value, ok := params[key]; ok {
		if intValue, ok := toInt(value); ok && intValue > 0 {
			return intValue
		}
	}
	return 1
}

func extractDurationSeconds(def builtinTemplateDefinition) int {
	if value := extractConcurrency(def.Run, "time_sec"); value > 1 {
		return value
	}
	if text := stringValue(def.Run, "duration"); text != "" {
		switch {
		case strings.Contains(text, "180"):
			return 180
		case strings.Contains(text, "300"):
			return 300
		case strings.Contains(text, "900"):
			return 900
		case strings.Contains(text, "10 分钟"), strings.Contains(strings.ToLower(text), "10 minute"):
			return 600
		}
	}
	return 300
}

func extractReportInterval(def builtinTemplateDefinition) int {
	if value := extractConcurrency(def.Run, "report_interval"); value > 0 {
		return value
	}
	return 10
}

func stringValue(values map[string]interface{}, key string) string {
	if raw, ok := values[key]; ok {
		if text, ok := raw.(string); ok {
			return text
		}
	}
	return ""
}

func cloneConfigMap(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return map[string]interface{}{}
	}
	data, err := json.Marshal(src)
	if err != nil {
		panic(err)
	}
	var cloned map[string]interface{}
	if err := json.Unmarshal(data, &cloned); err != nil {
		panic(err)
	}
	return cloned
}

func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		i, err := v.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mustTemplate(tmpl *Template) *Template {
	tmpl.Normalize()
	if err := tmpl.Validate(); err != nil {
		panic(err)
	}
	return tmpl
}

const defaultBuiltinTemplateJSON = `[
  {
    "id": "oracle_cpu_bound",
    "db": "oracle",
    "profile": "cpu_bound",
    "tool": "Swingbench OE",
    "source_alignment": "direct_from_doc",
    "goal": "面向查询占比高、热点集中、偏 CPU 的 Oracle 业务压测",
    "prepare": {
      "dataset": "1-10GB，默认 1GB",
      "notes": "如无特殊要求按 1GB 初始化；若与 io_bound 使用相同数据量，可复用初始化数据",
      "pre_steps": [
        "确保 SOE 表空间存在且容量足够",
        "使用 oewizard 创建/校验 OE schema"
      ]
    },
    "run": {
      "workload_mix": "查询:事务 = 9:1",
      "txn_mix": {
        "Customer Register": 1,
        "Update Customer": 1,
        "Browse Products": 85,
        "Order Products": 5,
        "Process Orders": 3,
        "Browse Orders": 5
      },
      "default_users": 256,
      "tuning_rule": "从 256 起逐步升高并发；以响应时间接近 30ms、且低于 40ms 时的稳定 TPS/TPM 作为最佳值",
      "duration": "单轮建议 10 分钟，TPS/TPM 图稳定后取值",
      "metrics": ["TPS", "TPM", "Response Time"]
    },
    "cleanup": {
      "strategy": "删除/重建 SOE schema；必要时保留表空间供下一轮复用"
    },
    "test_position": "performance"
  },
  {
    "id": "oracle_io_bound",
    "db": "oracle",
    "profile": "io_bound",
    "tool": "Swingbench OE",
    "source_alignment": "direct_from_doc",
    "goal": "面向事务占比提高、资源争用更强、偏 I/O 的 Oracle 业务压测",
    "prepare": {
      "dataset": "100GB 推荐；如仅做演示可降到 1GB",
      "notes": "100GB 初始化约 2 小时；正式性能压测优先使用 100GB",
      "pre_steps": [
        "确保 SOE 表空间容量足够（100GB 模式通常需扩容）",
        "使用 oewizard 创建/校验 OE schema"
      ]
    },
    "run": {
      "workload_mix": "查询:事务 = 5:5",
      "txn_mix": {
        "Customer Register": 10,
        "Update Customer": 10,
        "Browse Products": 35,
        "Order Products": 35,
        "Process Orders": 5,
        "Browse Orders": 5
      },
      "default_users": 256,
      "tuning_rule": "从 256 起逐步升高并发；以响应时间接近 30ms、且低于 40ms 时的稳定 TPS/TPM 作为最佳值",
      "duration": "单轮建议 10 分钟，TPS/TPM 图稳定后取值",
      "metrics": ["TPS", "TPM", "Response Time"]
    },
    "cleanup": {
      "strategy": "删除/重建 SOE schema；大数据量场景建议完整清理后再切模型"
    },
    "test_position": "performance"
  },
  {
    "id": "oracle_test",
    "db": "oracle",
    "profile": "test",
    "tool": "Swingbench OE",
    "source_alignment": "engineered_minimal",
    "goal": "最小数据量跑通 Oracle prepare/run/cleanup 链路",
    "prepare": {
      "dataset": "1GB",
      "notes": "沿用文档默认的无特殊要求初始化量，优先保证快速完成",
      "pre_steps": [
        "确保 SOE 表空间存在",
        "使用 oewizard 创建 schema"
      ]
    },
    "run": {
      "workload_mix": "沿用 cpu_bound 的 9:1",
      "txn_mix": {
        "Customer Register": 1,
        "Update Customer": 1,
        "Browse Products": 85,
        "Order Products": 5,
        "Process Orders": 3,
        "Browse Orders": 5
      },
      "default_users": 32,
      "tuning_rule": "不追求峰值，只要求稳定产生 TPS/TPM",
      "duration": "180 秒",
      "metrics": ["TPS", "TPM"]
    },
    "cleanup": {
      "strategy": "删除 SOE schema 或保留供重复 smoke 使用"
    },
    "test_position": "smoke"
  },
  {
    "id": "mysql_cpu_bound",
    "db": "mysql",
    "profile": "cpu_bound",
    "tool": "Sysbench oltp_read_write.lua",
    "source_alignment": "direct_from_doc",
    "goal": "面向热数据比例更高、偏 CPU 的 MySQL 混合读写压测",
    "prepare": {
      "dataset": "约 25GB",
      "params": {
        "tables": 10,
        "table_size": 10000000,
        "threads": 256
      },
      "notes": "文档推荐模型 1"
    },
    "run": {
      "params": {
        "threads": 256,
        "time_sec": 600,
        "report_interval": 10,
        "forced_shutdown": 1,
        "db_ps_mode": "disable"
      },
      "metrics": ["TPS", "QPS", "95% Latency", "Errors/s"]
    },
    "cleanup": {
      "strategy": "sysbench cleanup；数据库重启后建议重新导数"
    },
    "test_position": "performance"
  },
  {
    "id": "mysql_io_bound",
    "db": "mysql",
    "profile": "io_bound",
    "tool": "Sysbench oltp_read_write.lua",
    "source_alignment": "direct_from_doc",
    "goal": "面向更大初始化数据、更多物理 I/O 的 MySQL 混合读写压测",
    "prepare": {
      "dataset": "约 125GB",
      "params": {
        "tables": 50,
        "table_size": 10000000,
        "threads": 256
      },
      "notes": "文档推荐模型 2"
    },
    "run": {
      "params": {
        "threads": 256,
        "time_sec": 2400,
        "report_interval": 10,
        "forced_shutdown": 1,
        "db_ps_mode": "disable"
      },
      "metrics": ["TPS", "QPS", "95% Latency", "Errors/s"]
    },
    "cleanup": {
      "strategy": "sysbench cleanup；数据库重启后建议重新导数"
    },
    "test_position": "performance"
  },
  {
    "id": "mysql_test",
    "db": "mysql",
    "profile": "test",
    "tool": "Sysbench oltp_read_write.lua",
    "source_alignment": "engineered_minimal",
    "goal": "最小数据量跑通 MySQL prepare/run/cleanup 链路",
    "prepare": {
      "dataset": "约 0.1GB",
      "params": {
        "tables": 4,
        "table_size": 100000,
        "threads": 8
      },
      "notes": "工程化最小规模，不用于性能结论"
    },
    "run": {
      "params": {
        "threads": 8,
        "time_sec": 120,
        "report_interval": 10,
        "forced_shutdown": 1,
        "db_ps_mode": "disable"
      },
      "metrics": ["TPS", "QPS"]
    },
    "cleanup": {
      "strategy": "sysbench cleanup"
    },
    "test_position": "smoke"
  },
  {
    "id": "sqlserver_cpu_bound",
    "db": "sqlserver",
    "profile": "cpu_bound",
    "tool": "HammerDB TPROC-C",
    "source_alignment": "engineered_split_from_baseline",
    "goal": "在较小仓库集下提升用户并发，使更多热点数据停留在内存中，偏 CPU 压测",
    "prepare": {
      "dataset": "约 5GB",
      "params": {
        "warehouses": 20
      },
      "notes": "由文档 100 仓库基线工程化缩小，便于形成更明显的 CPU 偏向"
    },
    "run": {
      "params": {
        "virtual_users": 120,
        "time_sec": 600
      },
      "tuning_rule": "优先把 Response Time 控制在 30-40ms 区间内",
      "metrics": ["TPM", "TPS", "Response Time"]
    },
    "cleanup": {
      "strategy": "删除 tpcc schema/数据库并重建"
    },
    "test_position": "performance"
  },
  {
    "id": "sqlserver_io_bound",
    "db": "sqlserver",
    "profile": "io_bound",
    "tool": "HammerDB TPROC-C",
    "source_alignment": "engineered_split_from_baseline",
    "goal": "在更大仓库规模下增加数据访问跨度，使工作集更难完全缓存，偏 I/O 压测",
    "prepare": {
      "dataset": "约 50GB",
      "params": {
        "warehouses": 200
      },
      "notes": "由文档 100 仓库基线工程化放大，保留 TPROC-C 语义"
    },
    "run": {
      "params": {
        "virtual_users": 100,
        "time_sec": 900
      },
      "tuning_rule": "优先把 Response Time 控制在 30-40ms 区间内",
      "metrics": ["TPM", "TPS", "Response Time"]
    },
    "cleanup": {
      "strategy": "删除 tpcc schema/数据库并重建"
    },
    "test_position": "performance"
  },
  {
    "id": "sqlserver_test",
    "db": "sqlserver",
    "profile": "test",
    "tool": "HammerDB TPROC-C",
    "source_alignment": "engineered_minimal",
    "goal": "最小数据量跑通 SQL Server prepare/run/cleanup 链路",
    "prepare": {
      "dataset": "约 2.5GB",
      "params": {
        "warehouses": 10
      },
      "notes": "相对 100 仓库基线缩小 10 倍，保证初始化更快"
    },
    "run": {
      "params": {
        "virtual_users": 10,
        "time_sec": 300
      },
      "metrics": ["TPM", "TPS"]
    },
    "cleanup": {
      "strategy": "删除 tpcc schema/数据库"
    },
    "test_position": "smoke"
  },
  {
    "id": "postgresql_cpu_bound",
    "db": "postgresql",
    "profile": "cpu_bound",
    "tool": "Sysbench oltp_read_write.lua",
    "source_alignment": "engineered_split_from_baseline",
    "goal": "缩小数据集、保持较高线程数，让更多访问命中缓存，偏 CPU 压测",
    "prepare": {
      "dataset": "约 4GB",
      "params": {
        "tables": 16,
        "table_size": 1000000,
        "threads": 16
      },
      "notes": "由文档 64x1000w 的单一基线工程化拆分"
    },
    "run": {
      "params": {
        "threads": 64,
        "time_sec": 600,
        "report_interval": 5,
        "forced_shutdown": 1,
        "max_requests": 0
      },
      "metrics": ["TPS", "QPS", "95% Latency"]
    },
    "cleanup": {
      "strategy": "sysbench cleanup"
    },
    "test_position": "performance"
  },
  {
    "id": "postgresql_io_bound",
    "db": "postgresql",
    "profile": "io_bound",
    "tool": "Sysbench oltp_read_write.lua",
    "source_alignment": "direct_from_doc_as_baseline",
    "goal": "沿用文档大数据集基线，形成明显的 I/O 压力",
    "prepare": {
      "dataset": "约 160GB",
      "params": {
        "tables": 64,
        "table_size": 10000000,
        "threads": 20
      },
      "notes": "文档原始基线"
    },
    "run": {
      "params": {
        "threads": 64,
        "time_sec": 180,
        "report_interval": 5,
        "forced_shutdown": 1,
        "max_requests": 0
      },
      "metrics": ["TPS", "QPS", "95% Latency"]
    },
    "cleanup": {
      "strategy": "sysbench cleanup"
    },
    "test_position": "performance"
  },
  {
    "id": "postgresql_test",
    "db": "postgresql",
    "profile": "test",
    "tool": "Sysbench oltp_read_write.lua",
    "source_alignment": "engineered_minimal",
    "goal": "最小数据量跑通 PostgreSQL prepare/run/cleanup 链路",
    "prepare": {
      "dataset": "约 0.1GB",
      "params": {
        "tables": 4,
        "table_size": 100000,
        "threads": 8
      },
      "notes": "工程化最小规模，不用于性能结论"
    },
    "run": {
      "params": {
        "threads": 8,
        "time_sec": 120,
        "report_interval": 5,
        "forced_shutdown": 1,
        "max_requests": 0
      },
      "metrics": ["TPS", "QPS"]
    },
    "cleanup": {
      "strategy": "sysbench cleanup"
    },
    "test_position": "smoke"
  }
]`
