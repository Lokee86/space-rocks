package diagnostics

import "context"

// BundleStore is the backend-neutral persistence seam for diagnostic bundles.
type BundleStore interface {
	Save(context.Context, Bundle) error
	Get(context.Context, string) (Bundle, error)
}
