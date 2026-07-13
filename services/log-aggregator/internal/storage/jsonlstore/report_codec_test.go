package jsonlstore

import (
	"bytes"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

func TestReportCodecRejectsInvalidProjectionPayload(t *testing.T) {
	created := time.Unix(10, 0).UTC()
	cases := []struct {
		name   string
		report storage.Report
	}{
		{"empty payload", storage.Report{DiagnosticReportID: "id", CreatedAt: created}},
		{"invalid json", storage.Report{DiagnosticReportID: "id", CreatedAt: created, RawJSON: []byte("{")}},
		{"missing id", storage.Report{CreatedAt: created, RawJSON: []byte(`{"created_at":"1970-01-01T00:00:10Z"}`)}},
		{"missing time", storage.Report{DiagnosticReportID: "id", RawJSON: []byte(`{"diagnostic_report_id":"id"}`)}},
		{"mismatched id", storage.Report{DiagnosticReportID: "other", CreatedAt: created, RawJSON: []byte(`{"diagnostic_report_id":"id","created_at":"1970-01-01T00:00:10Z"}`)}},
		{"mismatched time", storage.Report{DiagnosticReportID: "id", CreatedAt: created.Add(time.Second), RawJSON: []byte(`{"diagnostic_report_id":"id","created_at":"1970-01-01T00:00:10Z"}`)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := encodeReport(tc.report); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestReportCodecPreservesRawJSONRoundTrip(t *testing.T) {
	payload := []byte(`{"diagnostic_report_id":"id","created_at":"1970-01-01T00:00:10Z","extra":{"keep":true}}`)
	report, err := decodeReport(payload)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := encodeReport(report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, payload) {
		t.Fatalf("encoded payload changed: %s", encoded)
	}
}
