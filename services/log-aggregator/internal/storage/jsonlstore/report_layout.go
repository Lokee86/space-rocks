package jsonlstore

import (
	"fmt"
	"path/filepath"
	"time"
)

// reportLayout owns the report-specific active, archive, and quarantine paths.
type reportLayout struct{ root string }

func newReportLayout(root string) reportLayout {
	return reportLayout{root: filepath.Join(root, "reports")}
}
func (l reportLayout) activeDir() string     { return filepath.Join(l.root, "active") }
func (l reportLayout) archiveDir() string    { return filepath.Join(l.root, "archive") }
func (l reportLayout) quarantineDir() string { return filepath.Join(l.root, "quarantine") }
func (l reportLayout) activePath() string {
	return filepath.Join(l.activeDir(), "diagnostic-reports.jsonl.open")
}
func (l reportLayout) archivePath(start, end time.Time, sequence uint64, compressed bool) string {
	ext := ".jsonl"
	if compressed {
		ext += ".gz"
	}
	name := fmt.Sprintf("diagnostic-reports-%s-%s-%06d%s", formatTimestamp(start), formatTimestamp(end), sequence, ext)
	start = start.UTC()
	return filepath.Join(l.archiveDir(), start.Format("2006"), start.Format("01"), start.Format("02"), name)
}
