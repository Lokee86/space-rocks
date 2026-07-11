package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
)

func TestCandidateWriteDiagnosticsForReturnsMetadataForFullCandidate(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1"}
	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 4, BaselineID: "world-baseline", SnapshotID: "world-snapshot", SnapshotKind: SnapshotKind("full"), ChunkIndex: 1, ChunkCount: 2, IsFinalChunk: false})
	full := mustWorldWireFull(t, snapshot, 4)
	candidate := mustRealtimeLaneCandidate(full, nil)

	diagnostics := CandidateWriteDiagnosticsFor(candidate, state, 0)
	if diagnostics.PacketFamily == "" || diagnostics.Lane == "" || diagnostics.Kind == "" {
		t.Fatalf("expected non-empty family/lane/kind, got %#v", diagnostics)
	}
	if got, want := diagnostics.Sequence, full.Metadata.Sequence; got != want {
		t.Fatalf("diagnostics sequence = %d, want %d", got, want)
	}
	if got, want := diagnostics.BaselineID, full.Metadata.BaselineID; got != want {
		t.Fatalf("diagnostics baseline id = %q, want %q", got, want)
	}
	if got, want := diagnostics.SnapshotID, full.Metadata.SnapshotID; got != want {
		t.Fatalf("diagnostics snapshot id = %q, want %q", got, want)
	}
	if got, want := diagnostics.SnapshotKind, full.Metadata.SnapshotKind; got != want {
		t.Fatalf("diagnostics snapshot kind = %q, want %q", got, want)
	}
	if got, want := diagnostics.ChunkIndex, full.Metadata.ChunkIndex; got != want {
		t.Fatalf("diagnostics chunk index = %d, want %d", got, want)
	}
	if got, want := diagnostics.ChunkCount, full.Metadata.ChunkCount; got != want {
		t.Fatalf("diagnostics chunk count = %d, want %d", got, want)
	}
	if got, want := diagnostics.IsFinalChunk, full.Metadata.IsFinalChunk; got != want {
		t.Fatalf("diagnostics final chunk = %v, want %v", got, want)
	}
}

func TestCandidateWriteDiagnosticsForReturnsMetadataForDeltaCandidateAndFallsBackWithoutMetadata(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneSession, Metadata{Lane: LaneSession, Sequence: 8, BaselineID: "session-baseline", SnapshotID: "session-snapshot", SnapshotKind: SnapshotKind("delta"), ChunkIndex: 0, ChunkCount: 1, IsFinalChunk: true})
	candidate := mustRealtimeLaneCandidate(SessionWireLaneDelta{Metadata: Metadata{Lane: LaneSession, Sequence: 8, BaselineID: "session-baseline", SnapshotID: "session-snapshot", SnapshotKind: SnapshotKind("delta"), ChunkIndex: 0, ChunkCount: 1, IsFinalChunk: true}}, nil)

	diagnostics := CandidateWriteDiagnosticsFor(candidate, state, 0)
	if diagnostics.PacketFamily == "" || diagnostics.Lane == "" || diagnostics.Kind == "" {
		t.Fatalf("expected non-empty family/lane/kind, got %#v", diagnostics)
	}
	if got, want := diagnostics.Sequence, 8; got != want {
		t.Fatalf("diagnostics sequence = %d, want %d", got, want)
	}
	if got, want := diagnostics.SnapshotID, "session-snapshot"; got != want {
		t.Fatalf("diagnostics snapshot id = %q, want %q", got, want)
	}
	if got, want := diagnostics.SnapshotKind, SnapshotKind("delta"); got != want {
		t.Fatalf("diagnostics snapshot kind = %q, want %q", got, want)
	}
	if !diagnostics.IsFinalChunk {
		t.Fatalf("expected final chunk diagnostics, got %#v", diagnostics)
	}

	fallback := CandidateWriteDiagnosticsFor(mustRealtimeLaneCandidate(EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: Metadata{Lane: LaneEvent}}, nil), NewRealtimeSessionState("player-1"), 0)
	if fallback.PacketFamily != PacketFamilyEventBatch || fallback.Lane != LaneEvent || fallback.Kind != RealtimeLaneCandidateKindEventBatch {
		t.Fatalf("unexpected fallback diagnostics: %#v", fallback)
	}
	if fallback.Sequence != 0 || fallback.SnapshotID != "" || fallback.BaselineID != "" || fallback.ChunkCount != 0 || fallback.IsFinalChunk {
		t.Fatalf("expected zero-value metadata fallback, got %#v", fallback)
	}
}