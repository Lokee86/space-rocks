package realtime

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

type EventLaneProjection struct {
	Batch EventBatchRecord
}

func ProjectEventLane(pending []game.PendingPresentationEvent, sequence int) EventLaneProjection {
	events := make([]EventRecord, 0, len(pending))
	for _, pendingEvent := range pending {
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
