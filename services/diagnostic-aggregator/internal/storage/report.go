package storage

import (
	"context"
	"encoding/json"
	"time"
)

// Report is the narrow persistence projection for a finalized diagnostic report.
type Report struct {
	DiagnosticReportID string          `json:"diagnostic_report_id"`
	CreatedAt          time.Time       `json:"created_at"`
	RawJSON            json.RawMessage `json:"-"`
}

type ReportStore interface {
	Save(context.Context, Report) error
	Get(context.Context, string) (Report, error)
	DeleteExpired(context.Context, time.Time) (int, error)
	Close() error
}
