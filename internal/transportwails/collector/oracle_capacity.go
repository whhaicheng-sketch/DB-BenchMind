package collector

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/sijms/go-ora/v2"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
	domaintask "github.com/whhaicheng/DB-BenchMind/internal/domain/task"
)

func CollectOracleStorageCapacity(ctx context.Context, conn connection.Connection) []domaintask.CapacityEntry {
	if conn.GetType() != connection.DatabaseTypeOracle {
		return []domaintask.CapacityEntry{
			{Key: "soe", Label: "SOE", Status: "not_applicable", Message: "not applicable", Threshold: thresholdForPercent(0)},
			{Key: "temp", Label: "TEMP", Status: "not_applicable", Message: "not applicable", Threshold: thresholdForPercent(0)},
			{Key: "undo", Label: "UNDO", Status: "not_applicable", Message: "not applicable", Threshold: thresholdForPercent(0)},
		}
	}

	oracleConn, ok := conn.(*connection.OracleConnection)
	if !ok {
		return []domaintask.CapacityEntry{
			{Key: "soe", Label: "SOE", Status: "unavailable", Message: "oracle connection is unavailable", Threshold: thresholdForPercent(0)},
			{Key: "temp", Label: "TEMP", Status: "unavailable", Message: "oracle connection is unavailable", Threshold: thresholdForPercent(0)},
			{Key: "undo", Label: "UNDO", Status: "unavailable", Message: "oracle connection is unavailable", Threshold: thresholdForPercent(0)},
		}
	}

	db, err := sql.Open("oracle", oracleConn.GetDSNWithPassword())
	if err != nil {
		return unavailableOracleEntries(fmt.Sprintf("open oracle connection: %v", err))
	}
	defer db.Close()

	return []domaintask.CapacityEntry{
		queryOracleCapacityEntry(ctx, db, "soe", "SOE", oracleSOESQL),
		queryOracleCapacityEntry(ctx, db, "temp", "TEMP", oracleTempSQL),
		queryOracleCapacityEntry(ctx, db, "undo", "UNDO", oracleUndoSQL),
	}
}

const oracleSOESQL = `
SELECT NVL(total_bytes, 0), NVL(used_bytes, 0), NVL(free_bytes, 0)
  FROM (
    SELECT SUM(df.bytes) AS total_bytes,
           SUM(df.bytes) - NVL(fs.free_bytes, 0) AS used_bytes,
           NVL(fs.free_bytes, 0) AS free_bytes
      FROM (SELECT SUM(bytes) AS bytes FROM dba_data_files WHERE tablespace_name = 'SOE') df
      LEFT JOIN (SELECT SUM(bytes) AS free_bytes FROM dba_free_space WHERE tablespace_name = 'SOE') fs ON 1 = 1
  )`

const oracleTempSQL = `
SELECT NVL(total_bytes, 0), NVL(used_bytes, 0), NVL(free_bytes, 0)
  FROM (
    SELECT SUM(tf.bytes) AS total_bytes,
           NVL(h.used_bytes, 0) AS used_bytes,
           NVL(h.free_bytes, 0) AS free_bytes
      FROM dba_temp_files tf
      LEFT JOIN (
        SELECT SUM(bytes_used) AS used_bytes, SUM(bytes_free) AS free_bytes
          FROM v$temp_space_header
      ) h ON 1 = 1
     WHERE tf.tablespace_name IN (
       SELECT tablespace_name FROM dba_tablespaces WHERE contents = 'TEMPORARY'
     )
  )`

const oracleUndoSQL = `
SELECT NVL(total_bytes, 0), NVL(used_bytes, 0), NVL(free_bytes, 0)
  FROM (
    SELECT SUM(df.bytes) AS total_bytes,
           SUM(df.bytes) - NVL(fs.free_bytes, 0) AS used_bytes,
           NVL(fs.free_bytes, 0) AS free_bytes
      FROM dba_data_files df
      LEFT JOIN (
        SELECT tablespace_name, SUM(bytes) AS free_bytes
          FROM dba_free_space
         GROUP BY tablespace_name
      ) fs ON fs.tablespace_name = df.tablespace_name
     WHERE df.tablespace_name IN (
       SELECT tablespace_name FROM dba_tablespaces WHERE contents = 'UNDO'
     )
  )`

func queryOracleCapacityEntry(ctx context.Context, db *sql.DB, key string, label string, query string) domaintask.CapacityEntry {
	var totalBytes float64
	var usedBytes float64
	var freeBytes float64
	if err := db.QueryRowContext(ctx, query).Scan(&totalBytes, &usedBytes, &freeBytes); err != nil {
		return domaintask.CapacityEntry{
			Key:       key,
			Label:     label,
			Status:    "unavailable",
			Message:   err.Error(),
			Threshold: thresholdForPercent(0),
		}
	}
	if totalBytes <= 0 {
		return domaintask.CapacityEntry{
			Key:       key,
			Label:     label,
			Status:    "unavailable",
			Message:   "tablespace not found",
			Threshold: thresholdForPercent(0),
		}
	}
	usePercent := (usedBytes / totalBytes) * 100
	return domaintask.CapacityEntry{
		Key:        key,
		Label:      label,
		UsedBytes:  usedBytes,
		TotalBytes: totalBytes,
		FreeBytes:  freeBytes,
		UsePercent: usePercent,
		Threshold:  thresholdForPercent(usePercent),
		Status:     "ok",
	}
}

func unavailableOracleEntries(message string) []domaintask.CapacityEntry {
	return []domaintask.CapacityEntry{
		{Key: "soe", Label: "SOE", Status: "unavailable", Message: message, Threshold: thresholdForPercent(0)},
		{Key: "temp", Label: "TEMP", Status: "unavailable", Message: message, Threshold: thresholdForPercent(0)},
		{Key: "undo", Label: "UNDO", Status: "unavailable", Message: message, Threshold: thresholdForPercent(0)},
	}
}
