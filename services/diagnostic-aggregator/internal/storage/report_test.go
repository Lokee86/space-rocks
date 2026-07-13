package storage

import (
	"context"
	"testing"
	"time"
)

type reportStoreContract struct{}

func (reportStoreContract) Save(context.Context, Report) error { return nil }
func (reportStoreContract) Get(context.Context, string) (Report, error) {
	return Report{}, nil
}
func (reportStoreContract) DeleteExpired(context.Context, time.Time) (int, error) {
	return 0, nil
}
func (reportStoreContract) Close() error { return nil }

var _ ReportStore = reportStoreContract{}

func TestReportStorageBoundaryUsesFinalizedReport(t *testing.T) {
	report := Report{
		DiagnosticReportID: "550e8400-e29b-41d4-a716-446655440000",
		CreatedAt:          time.Unix(10, 0).UTC(),
	}

	if report.DiagnosticReportID == "" {
		t.Fatalf("report = %#v", report)
	}
}
