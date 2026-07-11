package realtime

import (
	"fmt"
	"testing"
)

func makeBulletUpdate(id string, x int, y int) map[string]any {
	return map[string]any{
		"id": id,
		"x":  x,
		"y":  y,
	}
}

func makeBulletUpdateWithRotation(id string, x int, y int, rotation int) map[string]any {
	update := makeBulletUpdate(id, x, y)
	update["rotation"] = rotation
	return update
}

func makeAsteroidUpdate(id string, x int, y int) map[string]any {
	return map[string]any{
		"id": id,
		"x":  x,
		"y":  y,
	}
}

func encodedCandidateBytes(t *testing.T, candidate RealtimeLaneCandidate) int {
	t.Helper()
	encoded, recordedBytes := mustEncodeLanePacketUnchecked(t, candidate)
	if recordedBytes <= 0 || len(encoded) == 0 {
		t.Fatalf("expected encoded candidate bytes for %#v, got %d and %d", candidate, len(encoded), recordedBytes)
	}
	return recordedBytes
}

func assertConservativeEncodedBytes(t *testing.T, label string, candidate RealtimeLaneCandidate, estimated int) {
	t.Helper()
	_, actual := mustEncodeLanePacketUnchecked(t, candidate)
	if estimated < actual {
		t.Fatalf("%s estimated bytes = %d, actual bytes = %d", label, estimated, actual)
	}
}

func TestExpandHotLaneCandidateChunksLeavesSmallAsteroidDeltaAsOneFinalChunk(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{
		Type: PacketFamilyAsteroidDelta,
		Metadata: Metadata{
			Lane:         LaneAsteroids,
			Sequence:     21,
			SnapshotKind: SnapshotKind("delta"),
		},
		AsteroidUpdates: []map[string]any{
			makeAsteroidUpdate("asteroid-000001", 1, 2),
		},
	}, nil)

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %d", len(chunks))
	}

	packet, ok := chunks[0].Payload.(AsteroidWireDeltaPacket)
	if !ok {
		t.Fatalf("expected AsteroidWireDeltaPacket, got %#v", chunks[0].Payload)
	}
	if packet.Metadata.ChunkIndex != 0 {
		t.Fatalf("chunk index = %d, want 0", packet.Metadata.ChunkIndex)
	}
	if packet.Metadata.ChunkCount != 1 {
		t.Fatalf("chunk count = %d, want 1", packet.Metadata.ChunkCount)
	}
	if !packet.Metadata.IsFinalChunk {
		t.Fatal("expected small asteroid delta chunk to be final")
	}
}

func TestExpandHotLaneCandidateChunksSplitsOversizedAsteroidDelta(t *testing.T) {
	updates := make([]map[string]any, 0, 300)
	for i := 1; i <= 300; i++ {
		updates = append(updates, makeAsteroidUpdate(fmt.Sprintf("asteroid-%06d", i), i, i+1))
	}

	candidate := mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{
		Type: PacketFamilyAsteroidDelta,
		Metadata: Metadata{
			Lane:         LaneAsteroids,
			Sequence:     22,
			SnapshotKind: SnapshotKind("delta"),
		},
		AsteroidUpdates: updates,
	}, nil)

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) <= 1 {
		t.Fatalf("expected oversized asteroid delta to split, got %d chunk(s)", len(chunks))
	}

	totalUpdates := 0
	for index, chunk := range chunks {
		if chunk.Lane() != LaneAsteroids {
			t.Fatalf("chunk %d lane = %q, want asteroids", index, chunk.Lane())
		}
		if chunk.Kind() != RealtimeLaneCandidateKindDelta {
			t.Fatalf("chunk %d kind = %q, want delta", index, chunk.Kind())
		}

		packet, ok := chunk.Payload.(AsteroidWireDeltaPacket)
		if !ok {
			t.Fatalf("chunk %d delta type = %#v, want AsteroidWireDeltaPacket", index, chunk.Payload)
		}
		if packet.Metadata.Sequence != 22 {
			t.Fatalf("chunk %d sequence = %d, want 22", index, packet.Metadata.Sequence)
		}
		if packet.Metadata.ChunkIndex != index {
			t.Fatalf("chunk %d chunk index = %d, want %d", index, packet.Metadata.ChunkIndex, index)
		}
		if packet.Metadata.ChunkCount != len(chunks) {
			t.Fatalf("chunk %d chunk count = %d, want %d", index, packet.Metadata.ChunkCount, len(chunks))
		}
		if packet.Metadata.IsFinalChunk != (index == len(chunks)-1) {
			t.Fatalf("chunk %d final flag = %t, want %t", index, packet.Metadata.IsFinalChunk, index == len(chunks)-1)
		}

		updateCount := len(packet.AsteroidUpdates)
		totalUpdates += updateCount
		encodedBytes := encodedCandidateBytes(t, chunk)
		if updateCount != 1 && encodedBytes > HardCapBytes {
			t.Fatalf("chunk %d encoded bytes = %d, want <= %d", index, encodedBytes, HardCapBytes)
		}
	}

	if totalUpdates != len(updates) {
		t.Fatalf("total updates across chunks = %d, want %d", totalUpdates, len(updates))
	}
}

func TestExpandHotLaneCandidateChunksLeavesSmallBulletDeltaAsOneFinalChunk(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(BulletWireDeltaPacket{
		Type: PacketFamilyBulletDelta,
		Metadata: Metadata{
			Lane:         LaneBullets,
			Sequence:     11,
			SnapshotKind: SnapshotKind("delta"),
		},
		BulletUpdates: []map[string]any{
			makeBulletUpdate("bullet-000001", 1, 2),
		},
	}, nil)

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %d", len(chunks))
	}

	packet, ok := chunks[0].Payload.(BulletWireDeltaPacket)
	if !ok {
		t.Fatalf("expected BulletWireDeltaPacket, got %#v", chunks[0].Payload)
	}
	if packet.Metadata.ChunkIndex != 0 {
		t.Fatalf("chunk index = %d, want 0", packet.Metadata.ChunkIndex)
	}
	if packet.Metadata.ChunkCount != 1 {
		t.Fatalf("chunk count = %d, want 1", packet.Metadata.ChunkCount)
	}
	if !packet.Metadata.IsFinalChunk {
		t.Fatal("expected small bullet delta chunk to be final")
	}
}

func TestExpandHotLaneCandidateChunksLeavesAsteroidLifecycleUntouched(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{
		Type:            PacketFamilyAsteroidsLifecycle,
		Metadata:        Metadata{Lane: LaneAsteroidsLifecycle, Sequence: 23, SnapshotKind: SnapshotKind("delta")},
		AsteroidUpdates: []map[string]any{makeAsteroidUpdate("asteroid-lifecycle-1", 1, 2)},
	}, nil)

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %d", len(chunks))
	}
	if chunks[0].Lane() != LaneAsteroidsLifecycle {
		t.Fatalf("lane = %q, want %q", chunks[0].Lane(), LaneAsteroidsLifecycle)
	}
	if chunks[0].Kind() != RealtimeLaneCandidateKindDelta {
		t.Fatalf("kind = %q, want delta", chunks[0].Kind())
	}
	packet, ok := chunks[0].Payload.(AsteroidWireDeltaPacket)
	if !ok {
		t.Fatalf("expected AsteroidWireDeltaPacket, got %#v", chunks[0].Payload)
	}
	if packet.Metadata.Lane != LaneAsteroidsLifecycle || packet.Metadata.ChunkIndex != 0 || packet.Metadata.ChunkCount != 0 || packet.Metadata.IsFinalChunk != false {
		t.Fatalf("asteroid lifecycle packet metadata = %#v, want index=0 count=0 final=false", packet.Metadata)
	}
}

func TestExpandHotLaneCandidateChunksLeavesBulletLifecycleUntouched(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(BulletWireDeltaPacket{
		Type:          PacketFamilyBulletsLifecycle,
		Metadata:      Metadata{Lane: LaneBulletsLifecycle, Sequence: 24, SnapshotKind: SnapshotKind("delta")},
		BulletUpdates: []map[string]any{makeBulletUpdate("bullet-lifecycle-1", 1, 2)},
	}, nil)

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %d", len(chunks))
	}
	if chunks[0].Lane() != LaneBulletsLifecycle {
		t.Fatalf("lane = %q, want %q", chunks[0].Lane(), LaneBulletsLifecycle)
	}
	if chunks[0].Kind() != RealtimeLaneCandidateKindDelta {
		t.Fatalf("kind = %q, want delta", chunks[0].Kind())
	}
	packet, ok := chunks[0].Payload.(BulletWireDeltaPacket)
	if !ok {
		t.Fatalf("expected BulletWireDeltaPacket, got %#v", chunks[0].Payload)
	}
	if packet.Metadata.Lane != LaneBulletsLifecycle || packet.Metadata.ChunkIndex != 0 || packet.Metadata.ChunkCount != 0 || packet.Metadata.IsFinalChunk != false {
		t.Fatalf("bullet lifecycle packet metadata = %#v, want index=0 count=0 final=false", packet.Metadata)
	}
}

func TestExpandHotLaneCandidateChunksTreatsLifecycleLanesAsUntouchedAndHotLanesAsFinalChunks(t *testing.T) {
	tests := []struct {
		name           string
		candidate      RealtimeLaneCandidate
		wantChunkIndex int
		wantChunkCount int
		wantFinalChunk bool
	}{
		{
			name: "asteroid lifecycle",
			candidate: mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{
				Type:            PacketFamilyAsteroidsLifecycle,
				Metadata:        Metadata{Lane: LaneAsteroidsLifecycle, Sequence: 23, SnapshotKind: SnapshotKind("delta")},
				AsteroidUpdates: []map[string]any{makeAsteroidUpdate("asteroid-lifecycle-1", 1, 2)},
			}, nil),
			wantChunkIndex: 0,
			wantChunkCount: 0,
			wantFinalChunk: false,
		},
		{
			name: "bullet lifecycle",
			candidate: mustRealtimeLaneCandidate(BulletWireDeltaPacket{
				Type:          PacketFamilyBulletsLifecycle,
				Metadata:      Metadata{Lane: LaneBulletsLifecycle, Sequence: 24, SnapshotKind: SnapshotKind("delta")},
				BulletUpdates: []map[string]any{makeBulletUpdate("bullet-lifecycle-1", 1, 2)},
			}, nil),
			wantChunkIndex: 0,
			wantChunkCount: 0,
			wantFinalChunk: false,
		},
		{
			name: "asteroid hot lane",
			candidate: mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{
				Type:            PacketFamilyAsteroidDelta,
				Metadata:        Metadata{Lane: LaneAsteroids, Sequence: 25, SnapshotKind: SnapshotKind("delta")},
				AsteroidUpdates: []map[string]any{makeAsteroidUpdate("asteroid-hot-1", 1, 2)},
			}, nil),
			wantChunkIndex: 0,
			wantChunkCount: 1,
			wantFinalChunk: true,
		},
		{
			name: "bullet hot lane",
			candidate: mustRealtimeLaneCandidate(BulletWireDeltaPacket{
				Type:          PacketFamilyBulletDelta,
				Metadata:      Metadata{Lane: LaneBullets, Sequence: 26, SnapshotKind: SnapshotKind("delta")},
				BulletUpdates: []map[string]any{makeBulletUpdate("bullet-hot-1", 1, 2)},
			}, nil),
			wantChunkIndex: 0,
			wantChunkCount: 1,
			wantFinalChunk: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{tc.candidate})
			if len(chunks) != 1 {
				t.Fatalf("expected one chunk, got %d", len(chunks))
			}
			if chunks[0].Lane() != tc.candidate.Lane() {
				t.Fatalf("lane = %q, want %q", chunks[0].Lane(), tc.candidate.Lane())
			}
			if chunks[0].Kind() != tc.candidate.Kind() {
				t.Fatalf("kind = %q, want %q", chunks[0].Kind(), tc.candidate.Kind())
			}
			if err := ValidateRealtimeLanePayload(chunks[0].Payload); err != nil {
				t.Fatalf("replacement payload validation failed: %v", err)
			}
			if chunks[0].PacketFamily() != tc.candidate.PacketFamily() {
				t.Fatalf("family = %q, want %q", chunks[0].PacketFamily(), tc.candidate.PacketFamily())
			}
			switch packet := chunks[0].Payload.(type) {
			case AsteroidWireDeltaPacket:
				metadata, ok := chunks[0].Metadata()
				if !ok || metadata != packet.Metadata {
					t.Fatalf("asteroid metadata mismatch: candidate=%#v payload=%#v", metadata, packet.Metadata)
				}
				if packet.Metadata.ChunkIndex != tc.wantChunkIndex || packet.Metadata.ChunkCount != tc.wantChunkCount || packet.Metadata.IsFinalChunk != tc.wantFinalChunk {
					t.Fatalf("asteroid packet metadata = %#v, want index=%d count=%d final=%t", packet.Metadata, tc.wantChunkIndex, tc.wantChunkCount, tc.wantFinalChunk)
				}
			case BulletWireDeltaPacket:
				metadata, ok := chunks[0].Metadata()
				if !ok || metadata != packet.Metadata {
					t.Fatalf("bullet metadata mismatch: candidate=%#v payload=%#v", metadata, packet.Metadata)
				}
				if packet.Metadata.ChunkIndex != tc.wantChunkIndex || packet.Metadata.ChunkCount != tc.wantChunkCount || packet.Metadata.IsFinalChunk != tc.wantFinalChunk {
					t.Fatalf("bullet packet metadata = %#v, want index=%d count=%d final=%t", packet.Metadata, tc.wantChunkIndex, tc.wantChunkCount, tc.wantFinalChunk)
				}
			default:
				t.Fatalf("unexpected delta type %#v", chunks[0].Payload)
			}
		})
	}
}

func TestExpandHotLaneCandidateChunksSplitsOversizedBulletDelta(t *testing.T) {
	updates := make([]map[string]any, 0, 240)
	for i := 1; i <= 240; i++ {
		updates = append(updates, makeBulletUpdate(fmt.Sprintf("bullet-%06d", i), i, i+1))
	}

	candidate := mustRealtimeLaneCandidate(BulletWireDeltaPacket{
		Type: PacketFamilyBulletDelta,
		Metadata: Metadata{
			Lane:         LaneBullets,
			Sequence:     12,
			SnapshotKind: SnapshotKind("delta"),
		},
		BulletUpdates: updates,
	}, nil)

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) <= 1 {
		t.Fatalf("expected oversized bullet delta to split, got %d chunk(s)", len(chunks))
	}

	totalUpdates := 0
	for index, chunk := range chunks {
		if chunk.Lane() != LaneBullets {
			t.Fatalf("chunk %d lane = %q, want bullets", index, chunk.Lane())
		}
		if chunk.Kind() != RealtimeLaneCandidateKindDelta {
			t.Fatalf("chunk %d kind = %q, want delta", index, chunk.Kind())
		}

		packet, ok := chunk.Payload.(BulletWireDeltaPacket)
		if !ok {
			t.Fatalf("chunk %d delta type = %#v, want BulletWireDeltaPacket", index, chunk.Payload)
		}
		if packet.Metadata.Sequence != 12 {
			t.Fatalf("chunk %d sequence = %d, want 12", index, packet.Metadata.Sequence)
		}
		if packet.Metadata.ChunkIndex != index {
			t.Fatalf("chunk %d chunk index = %d, want %d", index, packet.Metadata.ChunkIndex, index)
		}
		if packet.Metadata.ChunkCount != len(chunks) {
			t.Fatalf("chunk %d chunk count = %d, want %d", index, packet.Metadata.ChunkCount, len(chunks))
		}
		if packet.Metadata.IsFinalChunk != (index == len(chunks)-1) {
			t.Fatalf("chunk %d final flag = %t, want %t", index, packet.Metadata.IsFinalChunk, index == len(chunks)-1)
		}

		updateCount := len(packet.BulletUpdates)
		totalUpdates += updateCount
		encodedBytes := encodedCandidateBytes(t, chunk)
		if updateCount != 1 && encodedBytes > HardCapBytes {
			t.Fatalf("chunk %d encoded bytes = %d, want <= %d", index, encodedBytes, HardCapBytes)
		}
	}

	if totalUpdates != len(updates) {
		t.Fatalf("total updates across chunks = %d, want %d", totalUpdates, len(updates))
	}
}

func TestEstimateCompactBulletMovementUpdateBytesKeepsRotationConservative(t *testing.T) {
	update := makeBulletUpdateWithRotation("bullet-rot", 7, 8, 9)
	estimated := estimateCompactBulletMovementUpdateBytes(update)
	if estimated <= 0 {
		t.Fatalf("expected positive estimate, got %d", estimated)
	}

	packedPacket := CompactWirePacket(map[string]any{
		"type": "bullet_delta",
		"bullet_updates": []any{map[string]any{
			"id":       "bullet-rot",
			"x":        7,
			"y":        8,
			"rotation": 9,
		}},
	})
	updates, ok := packedPacket["bu"].([]any)
	if !ok || len(updates) != 1 {
		t.Fatalf("expected one compact bullet update, got %#v", packedPacket["bu"])
	}
	tuple, ok := updates[0].([]any)
	if !ok {
		t.Fatalf("expected compact tuple, got %#v", updates[0])
	}
	if len(tuple) != 4 {
		t.Fatalf("expected rotation tuple to remain four entries, got %#v", tuple)
	}
	if tuple[0] != "bullet-rot" || tuple[1] != 7 || tuple[2] != 8 || tuple[3] != 9 {
		t.Fatalf("unexpected compact tuple %#v", tuple)
	}
	encoded := estimateCompactJSONTupleBytes(tuple)
	if estimated < encoded {
		t.Fatalf("estimated bytes = %d, want conservative estimate >= compact tuple bytes %d", estimated, encoded)
	}
}

func TestEstimateBulletDeltaPacketBytesIsConservative(t *testing.T) {
	tests := []struct {
		name    string
		updates []map[string]any
	}{
		{
			name:    "id x y",
			updates: []map[string]any{makeBulletUpdate("bullet-000001", 1, 2)},
		},
		{
			name:    "id x",
			updates: []map[string]any{{"id": "bullet-000002", "x": 3}},
		},
		{
			name:    "id null y",
			updates: []map[string]any{{"id": "bullet-000003", "y": 4}},
		},
		{
			name:    "id x y rotation",
			updates: []map[string]any{makeBulletUpdateWithRotation("bullet-000004", 5, 6, 7)},
		},
		{
			name:    "zero values",
			updates: []map[string]any{makeBulletUpdate("bullet-000005", 0, 0)},
		},
		{
			name:    "negative values",
			updates: []map[string]any{makeBulletUpdate("bullet-000006", -8, -9)},
		},
	}

	for index, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			packet := BulletWireDeltaPacket{
				Type: PacketFamilyBulletDelta,
				Metadata: Metadata{
					Lane:         LaneBullets,
					Sequence:     30 + index,
					SnapshotKind: SnapshotKind("delta"),
				},
				BulletUpdates: tc.updates,
			}
			candidate := mustRealtimeLaneCandidate(packet, nil)
			estimated := estimateBulletDeltaPacketBytes(packet, tc.updates)
			assertConservativeEncodedBytes(t, tc.name, candidate, estimated)
		})
	}
}

func TestEstimateAsteroidDeltaPacketBytesIsConservative(t *testing.T) {
	tests := []struct {
		name    string
		updates []map[string]any
	}{
		{
			name:    "id x y",
			updates: []map[string]any{makeAsteroidUpdate("asteroid-000001", 1, 2)},
		},
		{
			name:    "id x",
			updates: []map[string]any{{"id": "asteroid-000002", "x": 3}},
		},
		{
			name:    "id null y",
			updates: []map[string]any{{"id": "asteroid-000003", "y": 4}},
		},
		{
			name:    "zero values",
			updates: []map[string]any{makeAsteroidUpdate("asteroid-000004", 0, 0)},
		},
		{
			name:    "negative values",
			updates: []map[string]any{makeAsteroidUpdate("asteroid-000005", -8, -9)},
		},
	}

	for index, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			packet := AsteroidWireDeltaPacket{
				Type: PacketFamilyAsteroidDelta,
				Metadata: Metadata{
					Lane:         LaneAsteroids,
					Sequence:     40 + index,
					SnapshotKind: SnapshotKind("delta"),
				},
				AsteroidUpdates: tc.updates,
			}
			candidate := mustRealtimeLaneCandidate(packet, nil)
			estimated := estimateAsteroidDeltaPacketBytes(packet, tc.updates)
			assertConservativeEncodedBytes(t, tc.name, candidate, estimated)
		})
	}
}
