package bindings

import (
	"context"
	"time"

	domaintask "github.com/whhaicheng/DB-BenchMind/internal/domain/task"
	"github.com/whhaicheng/DB-BenchMind/internal/transportwails/collector"
)

const capacityRefreshInterval = 5 * time.Second

func shouldRefreshCapacity(last time.Time) bool {
	return last.IsZero() || time.Since(last) >= capacityRefreshInterval
}

func refreshCapacity(task *domaintask.ExecutionTask, execCtx *taskExecutionContext) {
	if task == nil || execCtx == nil || execCtx.connection == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task.Metrics.Capacity.OracleStorage = collector.CollectOracleStorageCapacity(ctx, execCtx.connection)

	sshConfig := sshConfigFromConnection(execCtx.connection)
	if sshConfig == nil || !sshConfig.Enabled || !task.Readiness.SSHAvailable {
		task.Metrics.Capacity.Filesystem = []domaintask.CapacityEntry{
			{
				Key:       "data_disk",
				Label:     "Data Disk",
				Status:    "unavailable",
				Message:   task.Readiness.SSHMessage,
				Threshold: "safe",
			},
			{
				Key:       "binlog_disk",
				Label:     "Binlog Disk",
				Status:    "unavailable",
				Message:   task.Readiness.SSHMessage,
				Threshold: "safe",
			},
			{
				Key:       "archive_log_disk",
				Label:     "Archive Log Disk",
				Status:    "unavailable",
				Message:   task.Readiness.SSHMessage,
				Threshold: "safe",
			},
		}
		return
	}

	targets := collector.DetectFilesystemTargets(ctx, execCtx.connection)
	task.Metrics.Capacity.Filesystem = collector.CollectFilesystemCapacity(ctx, sshConfig, targets)
}
