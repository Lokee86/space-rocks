package jsonlstore

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

func scanReportFile(ctx context.Context, input io.Reader, id string) (storage.Report, error) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var found storage.Report
	matched := false
	for scanner.Scan() {
		if err := reportContextError(ctx); err != nil {
			return storage.Report{}, err
		}
		report, err := decodeReport(scanner.Bytes())
		if err != nil {
			return storage.Report{}, fmt.Errorf("jsonlstore: decode diagnostic report: %w", err)
		}
		if report.DiagnosticReportID == id {
			found, matched = report, true
		}
	}
	if err := scanner.Err(); err != nil {
		return storage.Report{}, fmt.Errorf("jsonlstore: scan diagnostic reports: %w", err)
	}
	if !matched {
		return storage.Report{}, ErrReportNotFound
	}
	return found, nil
}

func rewriteReports(ctx context.Context, input io.Reader, output io.Writer, cutoff time.Time) (int, int, error) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	kept, removed := 0, 0
	for scanner.Scan() {
		if err := reportContextError(ctx); err != nil {
			return kept, removed, err
		}
		report, err := decodeReport(scanner.Bytes())
		if err != nil {
			return kept, removed, fmt.Errorf("jsonlstore: decode diagnostic report: %w", err)
		}
		if report.CreatedAt.Before(cutoff) {
			removed++
			continue
		}
		line, err := encodeReport(report)
		if err != nil {
			return kept, removed, fmt.Errorf("jsonlstore: encode diagnostic report: %w", err)
		}
		if _, err := fmt.Fprintf(output, "%s\n", line); err != nil {
			return kept, removed, fmt.Errorf("jsonlstore: rewrite diagnostic reports: %w", err)
		}
		kept++
	}
	if err := scanner.Err(); err != nil {
		return kept, removed, err
	}
	return kept, removed, nil
}

func decodeReport(data []byte) (storage.Report, error) {
	var report storage.Report
	if err := json.Unmarshal(data, &report); err != nil {
		return storage.Report{}, err
	}
	if report.DiagnosticReportID == "" {
		return storage.Report{}, fmt.Errorf("diagnostic report id is empty")
	}
	if report.CreatedAt.IsZero() {
		return storage.Report{}, fmt.Errorf("diagnostic report created_at is zero")
	}
	report.RawJSON = append(json.RawMessage(nil), data...)
	return report, nil
}

func encodeReport(report storage.Report) ([]byte, error) {
	if len(report.RawJSON) == 0 {
		return nil, fmt.Errorf("raw diagnostic report payload is empty")
	}
	decoded, err := decodeReport(report.RawJSON)
	if err != nil {
		return nil, fmt.Errorf("raw diagnostic report payload is invalid: %w", err)
	}
	if decoded.DiagnosticReportID != report.DiagnosticReportID {
		return nil, fmt.Errorf("diagnostic report id does not match projection")
	}
	if !decoded.CreatedAt.Equal(report.CreatedAt) {
		return nil, fmt.Errorf("diagnostic report created_at does not match projection")
	}
	return append([]byte(nil), report.RawJSON...), nil
}

func reportContextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
