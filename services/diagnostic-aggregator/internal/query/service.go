package query

import (
	"context"
	"encoding/json"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

// Service owns the storage-backed query boundary. Archive and file selection
// remain behind storage.EventStore; query does not couple to the JSONL backend.
type Service struct {
	store storage.EventStore
}

var _ EventQuerier = (*Service)(nil)

func NewService(store storage.EventStore) *Service {
	return &Service{store: store}
}

func (service *Service) Query(ctx context.Context, filter Filter) (Result, error) {
	result, err := service.store.Query(ctx, storage.Query{
		From: filter.From, To: filter.To, SchemaVersion: filter.SchemaVersion,
		Environment: filter.Environment, BuildVersion: filter.BuildVersion,
		Service: filter.Service, ServiceInstanceID: filter.ServiceInstanceID,
		Category: filter.Category, Level: filter.Level, Event: filter.Event,
		TraceID: filter.TraceID, RequestID: filter.RequestID, SessionID: filter.SessionID,
		RoomID: filter.RoomID, MatchID: filter.MatchID, PlayerID: filter.PlayerID,
		AccountID: filter.AccountID, EventID: filter.EventID,
		DiagnosticReportID: filter.DiagnosticReportID, AuditEventID: filter.AuditEventID,
		IdempotencyKey: filter.IdempotencyKey, AuditRequired: filter.AuditRequired,
		Limit: filter.Limit,
	})
	if err != nil {
		return Result{}, err
	}
	events := make([]json.RawMessage, len(result.Records))
	for index, record := range result.Records {
		events[index] = append(json.RawMessage(nil), record.Payload...)
	}
	return Result{Events: events, Total: result.Total, Limited: result.Limited}, nil
}