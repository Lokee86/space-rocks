package realtime

import (
	"fmt"
	"sort"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

type EventLaneProjection struct {
	Batch EventBatchRecord
}

func ProjectEventLane(pending []game.PendingPresentationEvent, sequence int) EventLaneProjection {
	ids := make([]string, 0, len(pending))
	for _, event := range pending {
		ids = append(ids, event.EventID)
	}
	sort.Strings(ids)

	byID := make(map[string]game.PendingPresentationEvent, len(pending))
	for _, event := range pending {
		byID[event.EventID] = event
	}

	events := make([]EventRecord, 0, len(ids))
	for _, eventID := range ids {
		pendingEvent := byID[eventID]
		events = append(events, EventRecord{
			EventID: pendingEvent.EventID,
			Event:   pendingEvent.Event,
		})
	}

	return EventLaneProjection{
		Batch: EventBatchRecord{
			BatchID: sequenceBackedBatchID(sequence),
			Sequence: sequence,
			Events:   events,
		},
	}
}

func sequenceBackedBatchID(sequence int) string {
	return fmt.Sprintf("event-batch-%d", sequence)
}

func BuildEventBatchPacket(pending []game.PendingPresentationEvent, sequence int, serverSentMsec int) EventBatchPacket {
	projection := ProjectEventLane(pending, sequence)
	return EventBatchPacket{
		Type: PacketFamilyEventBatch,
		Metadata: Metadata{
			Lane:           LaneEvent,
			Sequence:       sequence,
			SnapshotID:     projection.Batch.BatchID,
			ServerSentMsec: serverSentMsec,
			SnapshotKind:   SnapshotKind("batch"),
			ChunkIndex:     0,
			ChunkCount:     1,
			IsFinalChunk:   true,
		},
		Batch: projection.Batch,
	}
}
