package autobench

import (
	"encoding/json"
	"time"
)

type ProfileType string

const (
	ProfileTest ProfileType = "test"
	ProfileCPU  ProfileType = "cpu_bound"
	ProfileIO   ProfileType = "io_bound"
)

type SuiteStatus string

const (
	SuiteStatusDraft          SuiteStatus = "draft"
	SuiteStatusReady          SuiteStatus = "ready"
	SuiteStatusRunning        SuiteStatus = "running"
	SuiteStatusPaused         SuiteStatus = "paused"
	SuiteStatusPartialSuccess SuiteStatus = "partial_success"
	SuiteStatusSuccess        SuiteStatus = "success"
	SuiteStatusFailed         SuiteStatus = "failed"
	SuiteStatusCancelled      SuiteStatus = "cancelled"
)

type SuiteItemStatus string

const (
	SuiteItemStatusPending        SuiteItemStatus = "pending"
	SuiteItemStatusValidating     SuiteItemStatus = "validating"
	SuiteItemStatusPreparing      SuiteItemStatus = "preparing"
	SuiteItemStatusPrechecking    SuiteItemStatus = "prechecking"
	SuiteItemStatusPrecheckFailed SuiteItemStatus = "precheck_failed"
	SuiteItemStatusRunning        SuiteItemStatus = "running"
	SuiteItemStatusCleaning       SuiteItemStatus = "cleaning"
	SuiteItemStatusSuccess        SuiteItemStatus = "success"
	SuiteItemStatusFailed         SuiteItemStatus = "failed"
	SuiteItemStatusSkipped        SuiteItemStatus = "skipped"
)

// IsTerminal returns true if the item status is terminal (no further transitions).
func (s SuiteItemStatus) IsTerminal() bool {
	return s == SuiteItemStatusSuccess ||
		s == SuiteItemStatusFailed ||
		s == SuiteItemStatusPrecheckFailed ||
		s == SuiteItemStatusSkipped
}

// IsFailed returns true if the item ended in a failure state.
func (s SuiteItemStatus) IsFailed() bool {
	return s == SuiteItemStatusFailed || s == SuiteItemStatusPrecheckFailed
}

type ExecutionMode string

const (
	ExecutionModeSerial ExecutionMode = "serial"
)

type FailurePolicy string

const (
	FailurePolicyContinueByConnection FailurePolicy = "continue_by_connection"
)

var DefaultProfileOrder = []ProfileType{ProfileTest, ProfileCPU, ProfileIO}

type ExecutionPolicy struct {
	Mode           ExecutionMode `json:"mode"`
	ProfileOrder   []ProfileType `json:"profile_order"`
	FailurePolicy  FailurePolicy `json:"failure_policy"`
	CleanupEnabled bool          `json:"cleanup_enabled"`
}

type Suite struct {
	ID                     string          `json:"id"`
	Name                   string          `json:"name"`
	SelectedConnectionIDs  []string        `json:"selected_connection_ids"`
	SelectedProfiles       []ProfileType   `json:"selected_profiles"`
	ExecutionPolicy        ExecutionPolicy `json:"execution_policy"`
	Status                 SuiteStatus     `json:"status"`
	StartedAt              *time.Time      `json:"started_at,omitempty"`
	EndedAt                *time.Time      `json:"ended_at,omitempty"`
	Items                  []SuiteItem     `json:"items,omitempty"`
	ReportPath             string          `json:"report_path,omitempty"`
	ReportJSONPath         string          `json:"report_json_path,omitempty"`
	SuiteManifestJSONPath  string          `json:"suite_manifest_json_path,omitempty"`
}

// SuiteManifest represents the suite_manifest.json structure for persistence.
type SuiteManifest struct {
	SchemaVersion         string                  `json:"schema_version"`
	SuiteID               string                  `json:"suite_id"`
	GeneratedAt           string                  `json:"generated_at"`
	SuiteInfo             SuiteManifestInfo       `json:"suite_info"`
	SelectedConnectionIDs []string                `json:"selected_connection_ids"`
	Items                 []SuiteManifestItem     `json:"items"`
	Statistics            SuiteManifestStatistics `json:"statistics"`
}

// SuiteManifestInfo contains suite metadata.
type SuiteManifestInfo struct {
	Name           string        `json:"name"`
	ExecutionMode  ExecutionMode `json:"execution_mode"`
	FailurePolicy  FailurePolicy `json:"failure_policy"`
	CleanupEnabled bool          `json:"cleanup_enabled"`
}

// SuiteManifestItem represents a single item in the manifest.
type SuiteManifestItem struct {
	ID           string      `json:"id"`
	ConnectionID string      `json:"connection_id"`
	DatabaseType string      `json:"database_type,omitempty"`
	ProfileType  ProfileType `json:"profile_type"`
	TemplateID   string      `json:"template_id,omitempty"`
	Status       string      `json:"status"`
	ReportID     string      `json:"report_id,omitempty"`
	StartedAt    string      `json:"started_at,omitempty"`
	EndedAt      string      `json:"ended_at,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
	PhaseTimings []PhaseTiming `json:"phase_timings,omitempty"`
}

// SuiteManifestStatistics contains aggregate statistics.
type SuiteManifestStatistics struct {
	TotalItems     int `json:"total_items"`
	Pending        int `json:"pending"`
	Running        int `json:"running"`
	Success        int `json:"success"`
	Failed         int `json:"failed"`
	PrecheckFailed int `json:"precheck_failed"`
	Skipped        int `json:"skipped"`
}

// PhaseTiming records the duration of a single execution phase for a suite item.
type PhaseTiming struct {
	Phase      string `json:"phase"`
	DurationMs int64  `json:"duration_ms,omitempty"`
}

type SuiteItem struct {
	ID             string                 `json:"id"`
	SuiteID        string                 `json:"suite_id"`
	ConnectionID   string                 `json:"connection_id"`
	DatabaseType   string                 `json:"database_type,omitempty"`
	ProfileType    ProfileType            `json:"profile_type"`
	TemplateID     string                 `json:"template_id,omitempty"`
	Status         SuiteItemStatus        `json:"status"`
	PhaseStatus    string                 `json:"phase_status,omitempty"`
	LinkedTaskID   string                 `json:"linked_task_id,omitempty"`
	RunID          string                 `json:"run_id,omitempty"`
	MetricsSummary map[string]interface{} `json:"metrics_summary,omitempty"`
	LogRefs        []string               `json:"log_refs,omitempty"`
	ErrorSummary   string                 `json:"error_summary,omitempty"`
	ReportID       string                 `json:"report_id,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	EndedAt        *time.Time             `json:"ended_at,omitempty"`
	PhaseTimings   []PhaseTiming          `json:"phase_timings,omitempty"`
}

type SuiteResult struct {
	SuiteID            string                   `json:"suite_id"`
	Status             SuiteStatus              `json:"status"`
	CompletedItemCount int                      `json:"completed_item_count"`
	SuccessItemCount   int                      `json:"success_item_count"`
	FailedItemCount    int                      `json:"failed_item_count"`
	SkippedItemCount   int                      `json:"skipped_item_count"`
	ConnectionRows     []map[string]interface{} `json:"connection_rows,omitempty"`
}

type SuiteReport struct {
	SuiteID         string                     `json:"suite_id"`
	GeneratedAt     time.Time                  `json:"generated_at"`
	Summary         SuiteReportSummary         `json:"summary"`
	ConnectionRows  []SuiteReportConnectionRow `json:"connection_rows,omitempty"`
	Failures        []SuiteReportFailure       `json:"failures,omitempty"`
	Recommendations []string                   `json:"recommendations,omitempty"`
	ArtifactPaths   SuiteReportArtifactPaths   `json:"artifact_paths,omitempty"`
}

type SuiteReportSummary struct {
	Status             SuiteStatus `json:"status"`
	TotalItems         int         `json:"total_items"`
	CompletedItemCount int         `json:"completed_item_count"`
	SuccessItemCount   int         `json:"success_item_count"`
	FailedItemCount    int         `json:"failed_item_count"`
	SkippedItemCount   int         `json:"skipped_item_count"`
}

type SuiteReportConnectionRow struct {
	ConnectionID       string        `json:"connection_id"`
	DatabaseType       string        `json:"database_type,omitempty"`
	Status             SuiteStatus   `json:"status"`
	ProfileTypes       []ProfileType `json:"profile_types,omitempty"`
	CompletedItemCount int           `json:"completed_item_count"`
	SuccessItemCount   int           `json:"success_item_count"`
	FailedItemCount    int           `json:"failed_item_count"`
	SkippedItemCount   int           `json:"skipped_item_count"`
}

type SuiteReportFailure struct {
	SuiteItemID  string      `json:"suite_item_id"`
	ConnectionID string      `json:"connection_id"`
	ProfileType  ProfileType `json:"profile_type"`
	LinkedTaskID string      `json:"linked_task_id,omitempty"`
	ErrorSummary string      `json:"error_summary"`
}

type SuiteReportArtifactPaths struct {
	JSON string `json:"json,omitempty"`
	HTML string `json:"html,omitempty"`
}

func (r SuiteReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToManifest converts a Suite to a SuiteManifest for persistence.
func (s Suite) ToManifest() SuiteManifest {
	items := make([]SuiteManifestItem, len(s.Items))
	for i, item := range s.Items {
		items[i] = item.ToManifestItem()
	}

	return SuiteManifest{
		SchemaVersion:         "v1",
		SuiteID:               s.ID,
		GeneratedAt:           time.Now().Format(time.RFC3339),
		SuiteInfo: SuiteManifestInfo{
			Name:           s.Name,
			ExecutionMode:  s.ExecutionPolicy.Mode,
			FailurePolicy:  s.ExecutionPolicy.FailurePolicy,
			CleanupEnabled: s.ExecutionPolicy.CleanupEnabled,
		},
		SelectedConnectionIDs: append([]string(nil), s.SelectedConnectionIDs...),
		Items:                 items,
		Statistics:            calculateStatistics(s.Items),
	}
}

// ToManifestItem converts a SuiteItem to a SuiteManifestItem.
func (item SuiteItem) ToManifestItem() SuiteManifestItem {
	startedAt := ""
	if item.StartedAt != nil {
		startedAt = item.StartedAt.Format(time.RFC3339)
	}
	endedAt := ""
	if item.EndedAt != nil {
		endedAt = item.EndedAt.Format(time.RFC3339)
	}

	return SuiteManifestItem{
		ID:           item.ID,
		ConnectionID: item.ConnectionID,
		DatabaseType: item.DatabaseType,
		ProfileType:  item.ProfileType,
		TemplateID:   item.TemplateID,
		Status:       string(item.Status),
		ReportID:     item.ReportID,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		ErrorMessage: item.ErrorSummary,
		PhaseTimings: append([]PhaseTiming(nil), item.PhaseTimings...),
	}
}

// calculateStatistics computes statistics from suite items.
func calculateStatistics(items []SuiteItem) SuiteManifestStatistics {
	stats := SuiteManifestStatistics{
		TotalItems: len(items),
	}
	for _, item := range items {
		switch item.Status {
		case SuiteItemStatusPending:
			stats.Pending++
		case SuiteItemStatusRunning, SuiteItemStatusValidating, SuiteItemStatusPreparing, SuiteItemStatusPrechecking, SuiteItemStatusCleaning:
			stats.Running++
		case SuiteItemStatusSuccess:
			stats.Success++
		case SuiteItemStatusFailed:
			stats.Failed++
		case SuiteItemStatusPrecheckFailed:
			stats.PrecheckFailed++
		case SuiteItemStatusSkipped:
			stats.Skipped++
		}
	}
	return stats
}

// ToJSON returns the manifest as formatted JSON.
func (m SuiteManifest) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

func NewSuite(name string, connectionIDs []string, profiles []ProfileType) Suite {
	return Suite{
		Name:                  name,
		SelectedConnectionIDs: append([]string(nil), connectionIDs...),
		SelectedProfiles:      append([]ProfileType(nil), profiles...),
		ExecutionPolicy: ExecutionPolicy{
			Mode:           ExecutionModeSerial,
			ProfileOrder:   append([]ProfileType(nil), DefaultProfileOrder...),
			FailurePolicy:  FailurePolicyContinueByConnection,
			CleanupEnabled: true,
		},
		Status: SuiteStatusDraft,
	}
}
