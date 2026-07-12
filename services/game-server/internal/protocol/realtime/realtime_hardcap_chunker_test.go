package realtime

import (
	"reflect"
	"strings"
	"testing"
)

func assertChunkSeries(t *testing.T, chunks []RealtimeLaneCandidate, original Metadata) {
	t.Helper()
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	for index, candidate := range chunks {
		metadata, ok := candidate.Metadata()
		if !ok {
			t.Fatalf("chunk %d has no metadata", index)
		}
		if metadata.MatchID != original.MatchID || metadata.Lane != original.Lane || metadata.Sequence != original.Sequence || metadata.BaselineID != original.BaselineID || metadata.SnapshotID != original.SnapshotID || metadata.ServerSentMsec != original.ServerSentMsec || metadata.SnapshotKind != original.SnapshotKind {
			t.Fatalf("chunk %d changed identity metadata: %#v", index, metadata)
		}
		if metadata.ChunkIndex != index || metadata.ChunkCount != len(chunks) || metadata.IsFinalChunk != (index == len(chunks)-1) {
			t.Fatalf("chunk %d metadata = %#v", index, metadata)
		}
		if bytes := encodedCandidateBytes(t, candidate); bytes > HardCapBytes {
			t.Fatalf("chunk %d encoded bytes = %d, want <= %d", index, bytes, HardCapBytes)
		}
	}
}

func TestExpandWorldFullHardCapSmallAndOversizedPreservesRecords(t *testing.T) {
	metadata := Metadata{MatchID: "", Lane: LaneWorld, Sequence: 7, BaselineID: "baseline-7", SnapshotID: "snapshot-7", ServerSentMsec: 99, SnapshotKind: SnapshotKind("full")}
	small := WorldWireFullPacket{Type: PacketFamilyWorldFull, Metadata: metadata, Asteroids: []WorldAsteroidWireRecord{{ID: "a-1", X: 1, Y: 2, Size: 3, Health: 4, Scale: 5, Variant: 6}}}
	smallChunks, err := ExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{mustRealtimeLaneCandidate(small, nil)})
	if err != nil {
		t.Fatal(err)
	}
	assertChunkSeries(t, smallChunks, metadata)
	if len(smallChunks) != 1 || !reflect.DeepEqual(smallChunks[0].Payload.(WorldWireFullPacket).Asteroids, small.Asteroids) {
		t.Fatalf("small records changed: %#v", smallChunks)
	}

	large := small
	large.Asteroids = make([]WorldAsteroidWireRecord, 0, 300)
	for i := 0; i < 300; i++ {
		large.Asteroids = append(large.Asteroids, WorldAsteroidWireRecord{ID: "asteroid-" + strings.Repeat("x", 8) + string(rune('a'+i%26)), X: int64(i), Y: int64(i + 1), Size: 2, Health: 3, Scale: 4, Variant: 5})
	}
	largeChunks, err := ExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{mustRealtimeLaneCandidate(large, nil)})
	if err != nil {
		t.Fatal(err)
	}
	assertChunkSeries(t, largeChunks, metadata)
	var got []WorldAsteroidWireRecord
	for _, chunk := range largeChunks {
		got = append(got, chunk.Payload.(WorldWireFullPacket).Asteroids...)
	}
	if !reflect.DeepEqual(got, large.Asteroids) {
		t.Fatal("world_full records were not preserved exactly")
	}
}

func TestExpandLifecycleHardCapSmallAndOversizedPreservesRecords(t *testing.T) {
	tests := []struct {
		name      string
		candidate RealtimeLaneCandidate
		count     func(RealtimeLaneCandidate) int
	}{
		{"asteroids", mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidsLifecycle, Metadata: Metadata{MatchID: "m", Lane: LaneAsteroidsLifecycle, Sequence: 8, BaselineID: "b", SnapshotID: "s", SnapshotKind: SnapshotKind("delta")}, AsteroidCreates: []WorldAsteroidWireRecord{{ID: "a", X: 1}}}, nil), func(c RealtimeLaneCandidate) int { return len(c.Payload.(AsteroidWireDeltaPacket).AsteroidCreates) }},
		{"bullets", mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: PacketFamilyBulletsLifecycle, Metadata: Metadata{MatchID: "m", Lane: LaneBulletsLifecycle, Sequence: 9, BaselineID: "b", SnapshotID: "s", SnapshotKind: SnapshotKind("delta")}, BulletCreates: []WorldBulletWireRecord{{ID: "b", X: 1}}}, nil), func(c RealtimeLaneCandidate) int { return len(c.Payload.(BulletWireDeltaPacket).BulletCreates) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metadata, _ := tc.candidate.Metadata()
			chunks, err := ExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{tc.candidate})
			if err != nil {
				t.Fatal(err)
			}
			assertChunkSeries(t, chunks, metadata)
			if len(chunks) != 1 || tc.count(chunks[0]) != 1 {
				t.Fatalf("small lifecycle packet was not preserved")
			}
		})
	}
}

func TestExpandRealtimeCandidateChunksRejectsIndividuallyOversizedRecords(t *testing.T) {
	metadata := Metadata{MatchID: "m", Lane: LaneAsteroidsLifecycle, Sequence: 1, SnapshotKind: SnapshotKind("delta")}
	candidate := mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidsLifecycle, Metadata: metadata, AsteroidCreates: []WorldAsteroidWireRecord{{ID: strings.Repeat("oversized", 2000)}}}, nil)
	if _, err := ExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{candidate}); err == nil {
		t.Fatal("expected individually oversized record error")
	}
}
