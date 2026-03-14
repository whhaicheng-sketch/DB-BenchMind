package task

import "time"

type Action string

const (
	ActionPrepare      Action = "prepare"
	ActionRun          Action = "run"
	ActionCleanup      Action = "cleanup"
	ActionFullPipeline Action = "full_pipeline"
)

type Status string

const (
	StatusStarting  Status = "starting"
	StatusPreparing Status = "preparing"
	StatusRunning   Status = "running"
	StatusCleaning  Status = "cleaning"
	StatusStopping  Status = "stopping"
	StatusStopped   Status = "stopped"
	StatusSuccess   Status = "success"
	StatusFailed    Status = "failed"
)

type Phase string

const (
	PhaseNone    Phase = "none"
	PhasePrepare Phase = "prepare"
	PhaseRun     Phase = "run"
	PhaseCleanup Phase = "cleanup"
)

type PhaseRecord struct {
	Phase     Phase      `json:"phase"`
	Status    string     `json:"status"`
	RunID     string     `json:"run_id,omitempty"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Message   string     `json:"message,omitempty"`
}

type TemplateSnapshot struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Tool           string                 `json:"tool"`
	DBFamily       string                 `json:"db_family"`
	WorkloadFamily string                 `json:"workload_family"`
	Phases         map[string]bool        `json:"phases"`
	Parameters     map[string]interface{} `json:"parameters"`
}

type ConnectionSnapshot struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database,omitempty"`
	Username    string `json:"username"`
	SSHEnabled  bool   `json:"ssh_enabled"`
	SSHHost     string `json:"ssh_host,omitempty"`
	SSHPort     int    `json:"ssh_port,omitempty"`
	SSHUsername string `json:"ssh_username,omitempty"`
}

type Readiness struct {
	TemplateSelected   bool   `json:"template_selected"`
	ConnectionSelected bool   `json:"connection_selected"`
	ActionSupported    bool   `json:"action_supported"`
	RuntimeValid       bool   `json:"runtime_valid"`
	DBValid            bool   `json:"db_valid"`
	DBMessage          string `json:"db_message,omitempty"`
	SSHAvailable       bool   `json:"ssh_available"`
	SSHChecked         bool   `json:"ssh_checked"`
	SSHMessage         string `json:"ssh_message,omitempty"`
}

type MetricSeriesPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type MetricSummary struct {
	Current float64             `json:"current"`
	Avg     float64             `json:"avg"`
	Max     float64             `json:"max"`
	Series  []MetricSeriesPoint `json:"series"`
}

type SystemMetricSummary struct {
	Current float64             `json:"current"`
	Series  []MetricSeriesPoint `json:"series"`
}

type UnifiedMetrics struct {
	TPS                MetricSummary       `json:"tps"`
	TPM                MetricSummary       `json:"tpm"`
	CPUUser            SystemMetricSummary `json:"cpu_user"`
	CPUSys             SystemMetricSummary `json:"cpu_sys"`
	CPUIOWait          SystemMetricSummary `json:"cpu_iowait"`
	CPUSteal           SystemMetricSummary `json:"cpu_steal"`
	DiskReadBps        SystemMetricSummary `json:"disk_read_bps"`
	DiskWriteBps       SystemMetricSummary `json:"disk_write_bps"`
	DiskReadLatencyMs  SystemMetricSummary `json:"disk_read_latency_ms"`
	DiskWriteLatencyMs SystemMetricSummary `json:"disk_write_latency_ms"`
	SystemEnabled      bool                `json:"system_enabled"`
	SystemMessage      string              `json:"system_message,omitempty"`
}

type LogLine struct {
	Timestamp string `json:"timestamp"`
	Phase     Phase  `json:"phase"`
	Stream    string `json:"stream"`
	Content   string `json:"content"`
}

type ExecutionTask struct {
	ID                 string                 `json:"id"`
	PreviewToken       string                 `json:"preview_token,omitempty"`
	Name               string                 `json:"name"`
	Action             Action                 `json:"action"`
	Status             Status                 `json:"status"`
	CurrentPhase       Phase                  `json:"current_phase"`
	TemplateSnapshot   TemplateSnapshot       `json:"template_snapshot"`
	ConnectionSnapshot ConnectionSnapshot     `json:"connection_snapshot"`
	ResolvedParams     map[string]interface{} `json:"resolved_params"`
	Readiness          Readiness              `json:"readiness"`
	PhaseHistory       []PhaseRecord          `json:"phase_history"`
	Metrics            UnifiedMetrics         `json:"metrics"`
	LogTail            []LogLine              `json:"log_tail"`
	BenchmarkTool      string                 `json:"benchmark_tool"`
	CreatedAt          time.Time              `json:"created_at"`
	StartedAt          *time.Time             `json:"started_at,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	RunLogPaths        map[string]string      `json:"run_log_paths,omitempty"`
	SystemLogPaths     map[string]string      `json:"system_log_paths,omitempty"`
}
