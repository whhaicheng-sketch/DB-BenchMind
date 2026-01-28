// Package tool provides tool detection and version checking.
// Implements: Phase 7 - Tool Detection
package tool

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/config"
)

var (
	// detectorMutex protects tool detection operations.
	detectorMutex sync.Mutex
)

// Detector provides tool detection capabilities.
type Detector struct{}

// NewDetector creates a new tool detector.
func NewDetector() *Detector {
	return &Detector{}
}

// DetectTool checks if a tool is available on the system.
// Returns the path to the tool if found, empty string otherwise.
func (d *Detector) DetectTool(ctx context.Context, toolType config.ToolType) (string, error) {
	detectorMutex.Lock()
	defer detectorMutex.Unlock()

	// Determine tool executable name based on type
	executable := d.getExecutableName(toolType)

	// First try looking in PATH
	path, err := exec.LookPath(executable)
	if err == nil {
		return path, nil
	}

	// Not found in PATH
	return "", fmt.Errorf("%w: %s not found in PATH", config.ErrToolNotFound, executable)
}

// GetToolVersion detects the version of an installed tool.
// Returns the version string if successful.
func (d *Detector) GetToolVersion(ctx context.Context, toolType config.ToolType) (string, error) {
	detectorMutex.Lock()
	defer detectorMutex.Unlock()

	// Get version command for tool
	cmdArgs := d.getVersionCommand(toolType)
	if cmdArgs == nil {
		return "", fmt.Errorf("unsupported tool type: %s", toolType)
	}

	// Execute command
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("execute version command: %w", err)
	}

	// Parse version from output
	version := d.parseVersion(toolType, string(output))
	if version == "" {
		return "", fmt.Errorf("failed to parse version from output: %s", string(output))
	}

	return version, nil
}

// DetectAllTools detects all benchmark tools on the system.
// Returns a map of tool type to detected information.
func (d *Detector) DetectAllTools(ctx context.Context) map[config.ToolType]*ToolInfo {
	detectorMutex.Lock()
	defer detectorMutex.Unlock()

	results := make(map[config.ToolType]*ToolInfo)
	var wg sync.WaitGroup
	var mu sync.Mutex

	tools := []config.ToolType{
		config.ToolTypeSysbench,
		config.ToolTypeSwingbench,
		config.ToolTypeHammerDB,
	}

	for _, toolType := range tools {
		wg.Add(1)
		go func(tt config.ToolType) {
			defer wg.Done()

			info := &ToolInfo{
				Type:    tt,
				Found:   false,
				Version: "",
				Path:    "",
			}

			// Check if tool exists
			path, err := exec.LookPath(d.getExecutableName(tt))
			if err == nil {
				info.Found = true
				info.Path = path

				// Try to get version
				cmdArgs := d.getVersionCommand(tt)
				if cmdArgs != nil {
					// Use absolute path
					cmdArgs[0] = path
					cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
					output, err := cmd.Output()
					if err == nil {
						info.Version = d.parseVersion(tt, string(output))
					}
				}
			}

			mu.Lock()
			results[tt] = info
			mu.Unlock()
		}(toolType)
	}

	wg.Wait()
	return results
}

// ToolInfo contains information about a detected tool.
type ToolInfo struct {
	Type    config.ToolType `json:"type"`
	Found   bool            `json:"found"`
	Path    string          `json:"path,omitempty"`
	Version string          `json:"version,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// getExecutableName returns the executable name for a tool type.
func (d *Detector) getExecutableName(toolType config.ToolType) string {
	switch toolType {
	case config.ToolTypeSysbench:
		return "sysbench"
	case config.ToolTypeSwingbench:
		return "swingbench"
	case config.ToolTypeHammerDB:
		if runtime.GOOS == "windows" {
			return "hammerdbcli.bat"
		}
		return "hammerdbcli"
	default:
		return ""
	}
}

// getVersionCommand returns the command to get version for a tool.
func (d *Detector) getVersionCommand(toolType config.ToolType) []string {
	switch toolType {
	case config.ToolTypeSysbench:
		return []string{"sysbench", "--version"}
	case config.ToolTypeSwingbench:
		// Swingbench doesn't have a standard version flag
		// We'll try running it and parsing the output
		return []string{"swingbench", "-h"}
	case config.ToolTypeHammerDB:
		if runtime.GOOS == "windows" {
			return []string{"hammerdbcli", "v"}
		}
		return []string{"hammerdbcli", "v"}
	default:
		return nil
	}
}

// parseVersion parses version string from tool output.
func (d *Detector) parseVersion(toolType config.ToolType, output string) string {
	switch toolType {
	case config.ToolTypeSysbench:
		// Output: "sysbench 1.0.20"
		parts := strings.Fields(output)
		if len(parts) >= 2 && parts[0] == "sysbench" {
			return parts[1]
		}

	case config.ToolTypeSwingbench:
		// Try to find version in output like "Swingbench V2.5.1234"
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Swingbench") && strings.Contains(line, "V") {
				// Extract version after "V"
				idx := strings.Index(line, "V")
				if idx != -1 && idx+1 < len(line) {
					version := strings.TrimSpace(line[idx+1:])
					// Take first word as version
					parts := strings.Fields(version)
					if len(parts) > 0 {
						return parts[0]
					}
				}
			}
		}

	case config.ToolTypeHammerDB:
		// Output: "HammerDB CLI v4.6"
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "HammerDB") && strings.Contains(line, "v") {
				// Extract version after "v"
				idx := strings.Index(line, "v")
				if idx != -1 && idx+1 < len(line) {
					version := strings.TrimSpace(line[idx+1:])
					// Take first word as version
					parts := strings.Fields(version)
					if len(parts) > 0 {
						return parts[0]
					}
				}
			}
		}
	}

	return ""
}

// CheckAvailability checks if a tool at the given path is available and executable.
func (d *Detector) CheckAvailability(path string) error {
	// Check if file exists
	info, err := exec.LookPath(path)
	if err != nil {
		return fmt.Errorf("%w: %s", config.ErrToolNotFound, path)
	}

	// Verify it's the same path
	if info != path {
		return fmt.Errorf("path mismatch: expected %s, found %s", path, info)
	}

	return nil
}

// DetectToolsAsync detects all tools asynchronously and returns results through a channel.
func (d *Detector) DetectToolsAsync(ctx context.Context) <-chan *ToolInfo {
	resultCh := make(chan *ToolInfo, 3)

	go func() {
		defer close(resultCh)

		tools := []config.ToolType{
			config.ToolTypeSysbench,
			config.ToolTypeSwingbench,
			config.ToolTypeHammerDB,
		}

		for _, toolType := range tools {
			info := &ToolInfo{
				Type:  toolType,
				Found: false,
			}

			// Check if tool exists
			path, err := d.DetectTool(ctx, toolType)
			if err == nil {
				info.Found = true
				info.Path = path

				// Try to get version
				version, err := d.GetToolVersion(ctx, toolType)
				if err == nil {
					info.Version = version
				}
			} else {
				info.Error = err.Error()
			}

			resultCh <- info
		}
	}()

	return resultCh
}
