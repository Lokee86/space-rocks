package realtime

import (
	"reflect"
	"testing"
)

func TestTypedPayloadRuntimeConsumersPreferPayload(t *testing.T) {
	payload := WorldFullPacket{
		Type:     PacketFamilyWorldFull,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 7, SnapshotID: "world-7"},
	}
	candidate := mustRealtimeLaneCandidate(payload, nil)

	wire := mustWireLanePacket(t, candidate)
	if wire["type"] != PacketFamilyWorldFull {
		t.Fatalf("wire type = %#v, want %q", wire["type"], PacketFamilyWorldFull)
	}
	metadata, ok := candidate.Metadata()
	if !ok || metadata.Lane != LaneWorld || metadata.Sequence != 7 {
		t.Fatalf("metadata = %#v, ok=%v", metadata, ok)
	}
	record := scheduleRecordForCandidate(3, candidate)
	if record.Lane != LaneWorld || record.RecordKind != string(RealtimeLaneCandidateKindFull) || !reflect.DeepEqual(record.PayloadRef, payload) {
		t.Fatalf("schedule record = %#v", record)
	}
}

func TestTypedHotLaneChunkReplacementRemainsTyped(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(BulletWireDeltaPacket{
		Type:          PacketFamilyBulletDelta,
		Metadata:      Metadata{Lane: LaneBullets, Sequence: 4, ChunkCount: 1, IsFinalChunk: true},
		BulletUpdates: []map[string]any{{"id": "bullet-1", "x": 12}},
	}, nil)

	chunks := mustExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) != 1 {
		t.Fatalf("chunk count = %d, want 1", len(chunks))
	}
	packet, ok := chunks[0].Payload.(BulletWireDeltaPacket)
	if !ok {
		t.Fatalf("typed chunk payload = %T, want BulletWireDeltaPacket", chunks[0].Payload)
	}
	if len(packet.BulletUpdates) != 1 || packet.Metadata.ChunkIndex != 0 {
		t.Fatalf("typed chunk = %#v", chunks[0])
	}
}
