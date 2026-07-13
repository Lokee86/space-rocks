package diagnostics

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDiagnosticReportStoreClosed = errors.New("diagnostics: diagnostic report store is closed")
	ErrDiagnosticReportNotFound    = errors.New("diagnostics: diagnostic report not found")
	ErrDiagnosticReportDuplicate   = errors.New("diagnostics: diagnostic report already exists")
)

// BundleStore is the backend-neutral persistence seam for diagnostic bundles.
type BundleStore interface {
	Save(context.Context, Bundle) error
	Get(context.Context, string) (Bundle, error)
}

// DiagnosticReportStore is the complete storage contract for diagnostic
// reports. Its inherited Save method stores immutable bundles and returns
// ErrDiagnosticReportDuplicate for duplicate diagnostic report IDs. Its
// inherited Get method returns ErrDiagnosticReportNotFound for a missing valid
// ID. Implementations return context cancellation when observed. DeleteExpired
// removes reports created before cutoff and returns the number deleted; it
// succeeds when nothing matches. Close is idempotent, and operations after
// Close return ErrDiagnosticReportStoreClosed.
type DiagnosticReportStore interface {
	BundleStore
	DeleteExpired(context.Context, time.Time) (uint64, error)
	Close() error
}
