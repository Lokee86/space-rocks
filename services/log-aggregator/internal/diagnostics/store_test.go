package diagnostics

import (
	"context"
	"errors"
	"testing"
	"time"
)

type contractStore struct{}

func (contractStore) Save(context.Context, Bundle) error { return nil }
func (contractStore) Get(context.Context, string) (Bundle, error) {
	return Bundle{}, nil
}
func (contractStore) DeleteExpired(context.Context, time.Time) (uint64, error) {
	return 0, nil
}
func (contractStore) Close() error { return nil }

var _ BundleStore = contractStore{}

type reportContractStore struct{}

func (reportContractStore) Save(context.Context, DiagnosticReport) error { return nil }
func (reportContractStore) Get(context.Context, string) (DiagnosticReport, error) {
	return DiagnosticReport{}, nil
}
func (reportContractStore) DeleteExpired(context.Context, time.Time) (uint64, error) {
	return 0, nil
}
func (reportContractStore) Close() error { return nil }

var _ DiagnosticReportStore = reportContractStore{}

func TestDiagnosticReportStorageSentinels(t *testing.T) {
	if ErrBundleNotFound != ErrDiagnosticReportNotFound {
		t.Fatal("ErrBundleNotFound must alias ErrDiagnosticReportNotFound")
	}
	if !errors.Is((&BundleNotFoundError{DiagnosticReportID: "id"}), ErrDiagnosticReportNotFound) {
		t.Fatal("BundleNotFoundError must unwrap to the canonical missing sentinel")
	}
}
