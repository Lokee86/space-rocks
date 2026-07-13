package runtime

import (
	"context"
	"encoding/json"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/health"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/ingest"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/query"
)

type InstrumentedIngestor struct {
	next  ingest.BatchIngestor
	state *health.State
}

func NewInstrumentedIngestor(next ingest.BatchIngestor, state *health.State) ingest.BatchIngestor {
	return &InstrumentedIngestor{next: next, state: state}
}

func (i *InstrumentedIngestor) IngestBatch(ctx context.Context, events []json.RawMessage) (ingest.BatchResult, error) {
	i.state.AddBatchesReceived(1)
	result, err := i.next.IngestBatch(ctx, events)
	if err != nil {
		return result, err
	}
	i.state.AddEventsAccepted(result.Accepted)
	i.state.AddEventsRejected(len(result.Rejected))
	for _, rejection := range result.Rejected {
		if rejection.Code == "redaction_rejected" {
			i.state.AddEventsRedacted(1)
		}
	}
	return result, nil
}

type InstrumentedQuerier struct {
	next  query.EventQuerier
	state *health.State
}

func NewInstrumentedQuerier(next query.EventQuerier, state *health.State) query.EventQuerier {
	return &InstrumentedQuerier{next: next, state: state}
}

func (q *InstrumentedQuerier) Query(ctx context.Context, filter query.Filter) (query.Result, error) {
	result, err := q.next.Query(ctx, filter)
	if err != nil {
		q.state.AddQueryFailures(1)
	}
	return result, err
}
