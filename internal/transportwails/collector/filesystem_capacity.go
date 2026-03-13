package collector

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/sijms/go-ora/v2"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	domaintask "github.com/whhaicheng/DB-BenchMind/internal/domain/task"
)

type FilesystemTarget struct {
	Key     string
	Label   string
	Path    string
	Status  string
	Message string
}

func DetectFilesystemTargets(ctx context.Context, conn connection.Connection) []FilesystemTarget {
	switch c := conn.(type) {
	case *connection.MySQLConnection:
		return detectMySQLFilesystemTargets(ctx, c)
	case *connection.OracleConnection:
		return detectOracleFilesystemTargets(ctx, c)
	default:
		return []FilesystemTarget{
			notDetectedTarget("data_disk", "Data Disk", "automatic data disk detection is unavailable for this database type"),
			notApplicableTarget("binlog_disk", "Binlog Disk", "not applicable"),
			notApplicableTarget("archive_log_disk", "Archive Log Disk", "not applicable"),
		}
	}
}

func CollectFilesystemCapacity(ctx context.Context, sshConfig *connection.SSHTunnelConfig, targets []FilesystemTarget) []domaintask.CapacityEntry {
	entries := make([]domaintask.CapacityEntry, 0, len(targets))
	for _, target := range targets {
		if target.Status != "" && target.Status != "ok" {
			entries = append(entries, domaintask.CapacityEntry{
				Key:       target.Key,
				Label:     target.Label,
				Status:    target.Status,
				Message:   target.Message,
				Threshold: thresholdForPercent(0),
				Path:      target.Path,
			})
			continue
		}
		entry, err := collectFilesystemEntry(ctx, sshConfig, target)
		if err != nil {
			entries = append(entries, domaintask.CapacityEntry{
				Key:       target.Key,
				Label:     target.Label,
				Status:    "unavailable",
				Message:   err.Error(),
				Threshold: thresholdForPercent(0),
				Path:      target.Path,
			})
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

func collectFilesystemEntry(_ context.Context, sshConfig *connection.SSHTunnelConfig, target FilesystemTarget) (domaintask.CapacityEntry, error) {
	command := fmt.Sprintf(
		"sh -lc 'target=$(readlink -f %s 2>/dev/null || printf %%s %s); df -Pk \"$target\" | awk \"NR==2 {print \\$6 \\\"\\t\\\" \\$3 \\\"\\t\\\" \\$2 \\\"\\t\\\" \\$4 \\\"\\t\\\" \\$5}\"'",
		shellQuote(target.Path),
		shellQuote(target.Path),
	)
	output, err := runSSHCommand(sshConfig, command)
	if err != nil {
		return domaintask.CapacityEntry{}, err
	}
	fields := strings.Fields(strings.TrimSpace(output))
	if len(fields) < 5 {
		return domaintask.CapacityEntry{}, fmt.Errorf("capacity probe returned unexpected output")
	}
	usedKB, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return domaintask.CapacityEntry{}, err
	}
	totalKB, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return domaintask.CapacityEntry{}, err
	}
	freeKB, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return domaintask.CapacityEntry{}, err
	}
	usePercent, err := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)
	if err != nil {
		return domaintask.CapacityEntry{}, err
	}
	return domaintask.CapacityEntry{
		Key:        target.Key,
		Label:      target.Label,
		UsedBytes:  usedKB * 1024,
		TotalBytes: totalKB * 1024,
		FreeBytes:  freeKB * 1024,
		UsePercent: usePercent,
		Threshold:  thresholdForPercent(usePercent),
		Status:     "ok",
		Path:       target.Path,
		MountPoint: fields[0],
	}, nil
}

func detectMySQLFilesystemTargets(ctx context.Context, conn *connection.MySQLConnection) []FilesystemTarget {
	targets := []FilesystemTarget{
		notDetectedTarget("data_disk", "Data Disk", "data directory not detected"),
		notDetectedTarget("binlog_disk", "Binlog Disk", "binlog directory not detected"),
		notApplicableTarget("archive_log_disk", "Archive Log Disk", "not applicable"),
	}
	db, err := sql.Open("mysql", conn.GetDSNWithPassword())
	if err != nil {
		targets[0].Message = fmt.Sprintf("open mysql connection: %v", err)
		targets[1].Message = fmt.Sprintf("open mysql connection: %v", err)
		return targets
	}
	defer db.Close()

	var dataDir sql.NullString
	var binlogBase sql.NullString
	query := "SELECT @@global.datadir, @@global.log_bin_basename"
	if err := db.QueryRowContext(ctx, query).Scan(&dataDir, &binlogBase); err != nil {
		targets[0].Message = fmt.Sprintf("query mysql storage paths: %v", err)
		targets[1].Message = fmt.Sprintf("query mysql storage paths: %v", err)
		return targets
	}
	if dataDir.Valid && strings.TrimSpace(dataDir.String) != "" {
		targets[0] = okTarget("data_disk", "Data Disk", strings.TrimSpace(dataDir.String))
	}
	if binlogBase.Valid && strings.TrimSpace(binlogBase.String) != "" {
		targets[1] = okTarget("binlog_disk", "Binlog Disk", filepath.Dir(strings.TrimSpace(binlogBase.String)))
	}
	return targets
}

func detectOracleFilesystemTargets(ctx context.Context, conn *connection.OracleConnection) []FilesystemTarget {
	targets := []FilesystemTarget{
		notDetectedTarget("data_disk", "Data Disk", "datafile path not detected"),
		notApplicableTarget("binlog_disk", "Binlog Disk", "not applicable"),
		notDetectedTarget("archive_log_disk", "Archive Log Disk", "archive log path not detected"),
	}
	db, err := sql.Open("oracle", conn.GetDSNWithPassword())
	if err != nil {
		targets[0].Message = fmt.Sprintf("open oracle connection: %v", err)
		targets[2].Message = fmt.Sprintf("open oracle connection: %v", err)
		return targets
	}
	defer db.Close()

	var dataFile sql.NullString
	dataSQL := `
SELECT file_name
  FROM (
    SELECT file_name, CASE WHEN tablespace_name = 'SOE' THEN 0 ELSE 1 END AS priority
      FROM dba_data_files
     ORDER BY priority, file_id
  )
 WHERE rownum = 1`
	if err := db.QueryRowContext(ctx, dataSQL).Scan(&dataFile); err == nil && dataFile.Valid {
		if dir, ok := normalizeOraclePath(dataFile.String); ok {
			targets[0] = okTarget("data_disk", "Data Disk", dir)
		} else {
			targets[0].Message = "oracle datafile path is not a filesystem path"
		}
	} else if err != nil {
		targets[0].Message = fmt.Sprintf("query oracle datafile path: %v", err)
	}

	archivePath, archiveErr := detectOracleArchiveLogPath(ctx, db)
	if archiveErr != nil {
		targets[2].Message = archiveErr.Error()
	} else if archivePath != "" {
		targets[2] = okTarget("archive_log_disk", "Archive Log Disk", archivePath)
	}
	return targets
}

func detectOracleArchiveLogPath(ctx context.Context, db *sql.DB) (string, error) {
	var fra sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT value FROM v$parameter WHERE name = 'db_recovery_file_dest'").Scan(&fra); err == nil {
		if fra.Valid && strings.TrimSpace(fra.String) != "" {
			if dir, ok := normalizeOraclePath(fra.String); ok {
				return dir, nil
			}
			return "", fmt.Errorf("archive log path is not a filesystem path")
		}
	}

	rows, err := db.QueryContext(ctx, `
SELECT value
  FROM v$parameter
 WHERE name LIKE 'log_archive_dest_%'
   AND value IS NOT NULL
 ORDER BY name`)
	if err != nil {
		return "", fmt.Errorf("query archive log path: %w", err)
	}
	defer rows.Close()

	re := regexp.MustCompile(`(?i)LOCATION=([^, ]+)`)
	for rows.Next() {
		var value sql.NullString
		if err := rows.Scan(&value); err != nil {
			return "", err
		}
		match := re.FindStringSubmatch(strings.TrimSpace(value.String))
		if len(match) < 2 {
			continue
		}
		if dir, ok := normalizeOraclePath(match[1]); ok {
			return dir, nil
		}
		return "", fmt.Errorf("archive log path is not a filesystem path")
	}
	return "", fmt.Errorf("archive log path not detected")
}

func normalizeOraclePath(value string) (string, bool) {
	path := strings.Trim(strings.TrimSpace(value), `"'`)
	if path == "" || strings.HasPrefix(path, "+") {
		return "", false
	}
	if strings.HasPrefix(strings.ToUpper(path), "USE_DB_RECOVERY_FILE_DEST") {
		return "", false
	}
	if strings.HasSuffix(path, "/") {
		return path, true
	}
	if strings.Contains(filepath.Base(path), ".") {
		return filepath.Dir(path), true
	}
	return path, strings.HasPrefix(path, "/")
}

func okTarget(key, label, path string) FilesystemTarget {
	return FilesystemTarget{Key: key, Label: label, Path: path, Status: "ok"}
}

func notDetectedTarget(key, label, message string) FilesystemTarget {
	return FilesystemTarget{Key: key, Label: label, Status: "not_detected", Message: message}
}

func notApplicableTarget(key, label, message string) FilesystemTarget {
	return FilesystemTarget{Key: key, Label: label, Status: "not_applicable", Message: message}
}

func thresholdForPercent(usePercent float64) string {
	switch {
	case usePercent > 90:
		return "danger"
	case usePercent >= 75:
		return "warning"
	default:
		return "safe"
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
