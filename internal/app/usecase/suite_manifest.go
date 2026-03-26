// Package usecase provides suite manifest persistence logic.
package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	domainautobench "github.com/whhaicheng/DB-BenchMind/internal/domain/autobench"
)

// SuiteManifestWriter handles persistence of suite_manifest.json.
type SuiteManifestWriter struct {
	reportsDir string
}

// NewSuiteManifestWriter creates a new manifest writer.
func NewSuiteManifestWriter(reportsDir string) *SuiteManifestWriter {
	return &SuiteManifestWriter{reportsDir: reportsDir}
}

// WriteManifest writes the suite manifest to disk.
// Returns the path to the written manifest file.
func (w *SuiteManifestWriter) WriteManifest(ctx context.Context, suite *domainautobench.Suite) (string, error) {
	if suite == nil {
		return "", fmt.Errorf("suite is nil")
	}

	manifest := suite.ToManifest()
	manifestPath := w.getManifestPath(suite.ID)

	// Ensure directory exists
	suiteDir := filepath.Dir(manifestPath)
	if err := os.MkdirAll(suiteDir, 0755); err != nil {
		return "", fmt.Errorf("create suite directory: %w", err)
	}

	// Marshal to JSON
	data, err := manifest.ToJSON()
	if err != nil {
		return "", fmt.Errorf("marshal manifest: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpPath := manifestPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return "", fmt.Errorf("write manifest temp file: %w", err)
	}

	if err := os.Rename(tmpPath, manifestPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("rename manifest file: %w", err)
	}

	return manifestPath, nil
}

// ReadManifest reads a suite manifest from disk.
func (w *SuiteManifestWriter) ReadManifest(ctx context.Context, suiteID string) (*domainautobench.SuiteManifest, error) {
	manifestPath := w.getManifestPath(suiteID)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest file: %w", err)
	}

	var manifest domainautobench.SuiteManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// UpdateManifestItem updates a specific item in the manifest.
// This is a convenience method that reads, updates, and writes the manifest.
func (w *SuiteManifestWriter) UpdateManifestItem(
	ctx context.Context,
	suiteID string,
	itemID string,
	updateFn func(*domainautobench.SuiteManifestItem) error,
) error {
	manifest, err := w.ReadManifest(ctx, suiteID)
	if err != nil {
		return fmt.Errorf("read manifest for update: %w", err)
	}

	// Find and update the item
	found := false
	for i := range manifest.Items {
		if manifest.Items[i].ID == itemID {
			if err := updateFn(&manifest.Items[i]); err != nil {
				return fmt.Errorf("update item %s: %w", itemID, err)
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("item %s not found in manifest", itemID)
	}

	// Recalculate statistics
	manifest.Statistics = w.calculateStatistics(manifest.Items)

	// Write back
	manifestPath := w.getManifestPath(suiteID)
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal updated manifest: %w", err)
	}

	tmpPath := manifestPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write updated manifest: %w", err)
	}

	if err := os.Rename(tmpPath, manifestPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename updated manifest: %w", err)
	}

	return nil
}

// getManifestPath returns the path to the suite_manifest.json file.
func (w *SuiteManifestWriter) getManifestPath(suiteID string) string {
	return filepath.Join(w.reportsDir, suiteID, "suite_manifest.json")
}

// calculateStatistics computes statistics from manifest items.
func (w *SuiteManifestWriter) calculateStatistics(items []domainautobench.SuiteManifestItem) domainautobench.SuiteManifestStatistics {
	stats := domainautobench.SuiteManifestStatistics{
		TotalItems: len(items),
	}
	for _, item := range items {
		switch item.Status {
		case "pending":
			stats.Pending++
		case "running", "validating", "preparing", "cleaning":
			stats.Running++
		case "success":
			stats.Success++
		case "failed":
			stats.Failed++
		case "skipped":
			stats.Skipped++
		}
	}
	return stats
}
