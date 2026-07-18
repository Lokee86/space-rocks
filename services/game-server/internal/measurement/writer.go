package measurement

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ReportWriter struct {
	Directory string
}

func NewReportWriter(directory string) ReportWriter {
	return ReportWriter{Directory: directory}
}

func WriteReport(directory string, report ServerReport) (string, error) {
	return NewReportWriter(directory).Write(report)
}

func (writer ReportWriter) Write(report ServerReport) (string, error) {
	if writer.Directory == "" {
		return "", fmt.Errorf("measurement report directory is empty")
	}
	if report.Version == 0 {
		report.Version = ReportVersion
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal measurement report: %w", err)
	}
	if err := os.MkdirAll(writer.Directory, 0o755); err != nil {
		return "", fmt.Errorf("create measurement report directory: %w", err)
	}

	runID := sanitizeRunID(report.Context.RunID)
	stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	name := fmt.Sprintf("server-measurement-v%d-%s-%s.json", report.Version, runID, stamp)
	path := filepath.Join(writer.Directory, name)
	temporary, err := os.CreateTemp(writer.Directory, ".server-measurement-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temporary measurement report: %w", err)
	}
	temporaryName := temporary.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(temporaryName)
		}
	}()

	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("write temporary measurement report: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("sync temporary measurement report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close temporary measurement report: %w", err)
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return "", fmt.Errorf("publish measurement report: %w", err)
	}
	cleanup = false
	return path, nil
}

func sanitizeRunID(value string) string {
	var builder strings.Builder
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' || character == '_' || character == '.' {
			builder.WriteRune(character)
		} else {
			builder.WriteByte('_')
		}
	}
	if builder.Len() == 0 {
		return "run"
	}
	return builder.String()
}
