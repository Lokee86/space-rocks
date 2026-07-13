package diagnostics

import (
	"context"
	"errors"
)

var (
	ErrInvalidDiagnosticReportID = errors.New("diagnostics: invalid diagnostic report id")
	ErrBundleNotFound            = errors.New("diagnostics: stored bundle not found")
)

type InvalidDiagnosticReportIDError struct{ DiagnosticReportID string }

func (e *InvalidDiagnosticReportIDError) Error() string {
	return "diagnostics: invalid diagnostic report id " + e.DiagnosticReportID
}

func (*InvalidDiagnosticReportIDError) Unwrap() error { return ErrInvalidDiagnosticReportID }

type BundleNotFoundError struct{ DiagnosticReportID string }

func (e *BundleNotFoundError) Error() string {
	return "diagnostics: stored bundle not found " + e.DiagnosticReportID
}
func (*BundleNotFoundError) Unwrap() error   { return ErrBundleNotFound }

type Service struct {
	Builder Builder
	Store   BundleStore
}

func (s Service) Create(ctx context.Context, traceID string, limit int) (Bundle, error) {
	if s.Store == nil {
		return Bundle{}, errors.New("diagnostics: nil bundle store")
	}
	bundle, err := s.Builder.Build(ctx, traceID, limit)
	if err != nil {
		return Bundle{}, err
	}
	if err := s.Store.Save(ctx, cloneBundle(bundle)); err != nil {
		return Bundle{}, err
	}
	return cloneBundle(bundle), nil
}

func (s Service) Get(ctx context.Context, diagnosticReportID string) (Bundle, error) {
	if !validUUID(diagnosticReportID) {
		return Bundle{}, &InvalidDiagnosticReportIDError{DiagnosticReportID: diagnosticReportID}
	}
	if s.Store == nil {
		return Bundle{}, errors.New("diagnostics: nil bundle store")
	}
	bundle, err := s.Store.Get(ctx, diagnosticReportID)
	if err != nil {
		return Bundle{}, err
	}
	return cloneBundle(bundle), nil
}

func cloneBundle(bundle Bundle) Bundle {
	bundle.Events = append([]Event(nil), bundle.Events...)
	for index := range bundle.Events {
		bundle.Events[index].Payload = copyPayload(bundle.Events[index].Payload)
	}
	bundle.Services = append([]string(nil), bundle.Services...)
	bundle.ServiceInstances = append([]string(nil), bundle.ServiceInstances...)
	bundle.Environments = append([]string(nil), bundle.Environments...)
	bundle.Builds = append([]string(nil), bundle.Builds...)
	bundle.RequestIDs = append([]string(nil), bundle.RequestIDs...)
	bundle.SessionIDs = append([]string(nil), bundle.SessionIDs...)
	bundle.RoomIDs = append([]string(nil), bundle.RoomIDs...)
	bundle.MatchIDs = append([]string(nil), bundle.MatchIDs...)
	bundle.PlayerIDs = append([]string(nil), bundle.PlayerIDs...)
	bundle.AccountIDs = append([]string(nil), bundle.AccountIDs...)
	return bundle
}
