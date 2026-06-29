package realtime

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestProjectEventLanePreservesIdentityAndSource(t *testing.T) {
	pending := []game.PendingPresentationEvent{
		{EventID: "evt-b", Event: game.EventState{Type: "bullet_blast", X: 2, Y: 3}},
		{EventID: "evt-a", Event: game.EventState{Type: "ship_death", PlayerID: "player-1", Lives: 2, RespawnDelay: 1.25}},
	}
	before := append([]game.PendingPresentationEvent(nil), pending...)

	projection := ProjectEventLane(pending, 7)

	if projection.Batch.Sequence != 7 {
		t.Fatalf("expected sequence 7, got %d", projection.Batch.Sequence)
	}
	if projection.Batch.BatchID != "event-batch-7" {
		t.Fatalf("expected sequence-backed batch ID, got %q", projection.Batch.BatchID)
	}
	if len(projection.Batch.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(projection.Batch.Events))
	}
	if projection.Batch.Events[0].EventID != "evt-a" || projection.Batch.Events[1].EventID != "evt-b" {
		t.Fatalf("expected events sorted by ID, got %#v", projection.Batch.Events)
	}
	if projection.Batch.Events[0].Event.Type != "ship_death" || projection.Batch.Events[0].Event.PlayerID != "player-1" || projection.Batch.Events[0].Event.Lives != 2 || projection.Batch.Events[0].Event.RespawnDelay != 1.25 {
		t.Fatalf("expected ship death payload to be preserved, got %#v", projection.Batch.Events[0].Event)
	}
	if projection.Batch.Events[1].Event.Type != "bullet_blast" || projection.Batch.Events[1].Event.X != 2 || projection.Batch.Events[1].Event.Y != 3 {
		t.Fatalf("expected bullet blast payload to be preserved, got %#v", projection.Batch.Events[1].Event)
	}
	if pending[0] != before[0] || pending[1] != before[1] {
		t.Fatalf("expected source slice to remain unchanged, got %#v before %#v", pending, before)
	}
}

func TestBuildEventBatchPacketUsesMetadataAndPreservesEventIDs(t *testing.T) {
	pending := []game.PendingPresentationEvent{
		{EventID: "evt-b", Event: game.EventState{Type: "bullet_blast"}},
		{EventID: "evt-a", Event: game.EventState{Type: "ship_death"}},
	}
	before := append([]game.PendingPresentationEvent(nil), pending...)

	packet := BuildEventBatchPacket(pending, 11, 1234)

	if packet.Type != PacketFamilyEventBatch {
		t.Fatalf("expected event batch packet type, got %q", packet.Type)
	}
	if packet.Metadata.Lane != LaneEvent || packet.Metadata.Sequence != 11 || packet.Metadata.SnapshotID != "event-batch-11" || packet.Metadata.ServerSentMsec != 1234 || packet.Metadata.SnapshotKind != SnapshotKind("batch") || packet.Metadata.ChunkIndex != 0 || packet.Metadata.ChunkCount != 1 || !packet.Metadata.IsFinalChunk {
		t.Fatalf("expected event batch metadata to be populated, got %#v", packet.Metadata)
	}
	if len(packet.Batch.Events) != 2 || packet.Batch.Events[0].EventID != "evt-a" || packet.Batch.Events[1].EventID != "evt-b" {
		t.Fatalf("expected event IDs sorted and preserved, got %#v", packet.Batch.Events)
	}
	if packet.Batch.Events[0].Event.Type != "ship_death" || packet.Batch.Events[1].Event.Type != "bullet_blast" {
		t.Fatalf("expected packet event payloads to be preserved, got %#v", packet.Batch.Events)
	}
	if pending[0] != before[0] || pending[1] != before[1] {
		t.Fatalf("expected source slice to remain unchanged after packet build, got %#v before %#v", pending, before)
	}
}
func TestSuccessiveEventBatchPacketsUseDifferentBatchIDs(t *testing.T) {
	pending := []game.PendingPresentationEvent{
		{EventID: "evt-a", Event: game.EventState{Type: "bullet_blast"}},
	}

	state := NewRealtimeSessionState("player-1")
	first := BuildEventBatchPacket(pending, 0, 1234)
	state.UpdateLane(LaneEvent, AdvanceMetadataForSuccessfulWrite(LaneEvent, first.Metadata))
	laneState, ok := state.LaneState(LaneEvent)
	if !ok {
		t.Fatal("expected event lane state after first batch write")
	}
	second := BuildEventBatchPacket(pending, laneState.Sequence, 1234)

	if first.Batch.BatchID != "event-batch-0" {
		t.Fatalf("first batch id = %q, want event-batch-0", first.Batch.BatchID)
	}
	if second.Batch.BatchID != "event-batch-1" {
		t.Fatalf("second batch id = %q, want event-batch-1", second.Batch.BatchID)
	}
	if first.Batch.BatchID == second.Batch.BatchID {
		t.Fatalf("expected successive event batches to use different ids, got %q and %q", first.Batch.BatchID, second.Batch.BatchID)
	}
}
