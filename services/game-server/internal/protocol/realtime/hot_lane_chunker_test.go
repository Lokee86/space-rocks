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
	encoded, recordedBytes := encodeLanePacketUnchecked(candidate)
	if recordedBytes <= 0 || len(encoded) == 0 {
		t.Fatalf("expected encoded candidate bytes for %#v, got %d and %d", candidate, len(encoded), recordedBytes)
	}
	return recordedBytes
}

func assertConservativeEncodedBytes(t *testing.T, label string, candidate RealtimeLaneCandidate, estimated int) {
	t.Helper()
	_, actual := encodeLanePacketUnchecked(candidate)
	if estimated < actual {
		t.Fatalf("%s estimated bytes = %d, actual bytes = %d", label, estimated, actual)
	}
}

func TestExpandHotLaneCandidateChunksLeavesSmallAsteroidDeltaAsOneFinalChunk(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneAsteroids,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: AsteroidWireDeltaPacket{
			Type: PacketFamilyAsteroidDelta,
			Metadata: Metadata{
				Lane:         LaneAsteroids,
				Sequence:     21,
				SnapshotKind: SnapshotKind("delta"),
			},
			AsteroidUpdates: []map[string]any{
				makeAsteroidUpdate("asteroid-000001", 1, 2),
			},
		},
	}

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %d", len(chunks))
	}

	packet, ok := chunks[0].Delta.(AsteroidWireDeltaPacket)
	if !ok {
		t.Fatalf("expected AsteroidWireDeltaPacket, got %#v", chunks[0].Delta)
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

	candidate := RealtimeLaneCandidate{
		Lane: LaneAsteroids,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: AsteroidWireDeltaPacket{
			Type: PacketFamilyAsteroidDelta,
			Metadata: Metadata{
				Lane:         LaneAsteroids,
				Sequence:     22,
				SnapshotKind: SnapshotKind("delta"),
			},
			AsteroidUpdates: updates,
		},
	}

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) <= 1 {
		t.Fatalf("expected oversized asteroid delta to split, got %d chunk(s)", len(chunks))
	}

	totalUpdates := 0
	for index, chunk := range chunks {
		if chunk.Lane != LaneAsteroids {
			t.Fatalf("chunk %d lane = %q, want asteroids", index, chunk.Lane)
		}
		if chunk.Kind != RealtimeLaneCandidateKindDelta {
			t.Fatalf("chunk %d kind = %q, want delta", index, chunk.Kind)
		}

		packet, ok := chunk.Delta.(AsteroidWireDeltaPacket)
		if !ok {
			t.Fatalf("chunk %d delta type = %#v, want AsteroidWireDeltaPacket", index, chunk.Delta)
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
	candidate := RealtimeLaneCandidate{
		Lane: LaneBullets,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: BulletWireDeltaPacket{
			Type: PacketFamilyBulletDelta,
			Metadata: Metadata{
				Lane:         LaneBullets,
				Sequence:     11,
				SnapshotKind: SnapshotKind("delta"),
			},
			BulletUpdates: []map[string]any{
				makeBulletUpdate("bullet-000001", 1, 2),
			},
		},
	}

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %d", len(chunks))
	}

	packet, ok := chunks[0].Delta.(BulletWireDeltaPacket)
	if !ok {
		t.Fatalf("expected BulletWireDeltaPacket, got %#v", chunks[0].Delta)
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

func TestExpandHotLaneCandidateChunksSplitsOversizedBulletDelta(t *testing.T) {
	updates := make([]map[string]any, 0, 240)
	for i := 1; i <= 240; i++ {
		updates = append(updates, makeBulletUpdate(fmt.Sprintf("bullet-%06d", i), i, i+1))
	}

	candidate := RealtimeLaneCandidate{
		Lane: LaneBullets,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: BulletWireDeltaPacket{
			Type: PacketFamilyBulletDelta,
			Metadata: Metadata{
				Lane:         LaneBullets,
				Sequence:     12,
				SnapshotKind: SnapshotKind("delta"),
			},
			BulletUpdates: updates,
		},
	}

	chunks := ExpandHotLaneCandidateChunks([]RealtimeLaneCandidate{candidate})
	if len(chunks) <= 1 {
		t.Fatalf("expected oversized bullet delta to split, got %d chunk(s)", len(chunks))
	}

	totalUpdates := 0
	for index, chunk := range chunks {
		if chunk.Lane != LaneBullets {
			t.Fatalf("chunk %d lane = %q, want bullets", index, chunk.Lane)
		}
		if chunk.Kind != RealtimeLaneCandidateKindDelta {
			t.Fatalf("chunk %d kind = %q, want delta", index, chunk.Kind)
		}

		packet, ok := chunk.Delta.(BulletWireDeltaPacket)
		if !ok {
			t.Fatalf("chunk %d delta type = %#v, want BulletWireDeltaPacket", index, chunk.Delta)
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

	packed := compactWirePackBulletMovementUpdate(map[string]any{
		"i": "bullet-rot",
		"x": 7,
		"y": 8,
		"r": 9,
	})
	tuple, ok := packed.([]any)
	if !ok {
		t.Fatalf("expected compact tuple, got %#v", packed)
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
			candidate := RealtimeLaneCandidate{
				Lane:  LaneBullets,
				Kind:  RealtimeLaneCandidateKindDelta,
				Delta: packet,
			}
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
			candidate := RealtimeLaneCandidate{
				Lane:  LaneAsteroids,
				Kind:  RealtimeLaneCandidateKindDelta,
				Delta: packet,
			}
			estimated := estimateAsteroidDeltaPacketBytes(packet, tc.updates)
			assertConservativeEncodedBytes(t, tc.name, candidate, estimated)
		})
	}
}
