package realtime

import (
	"testing"
)

func TestScheduleRecordForHotAsteroidAndBulletDeltaCandidatesUsesHotSupersedableHighPriority(t *testing.T) {
	tests := []struct {
		name       string
		candidate  RealtimeLaneCandidate
		wantLane   Lane
		wantFamily string
	}{
		{
			name:       "asteroid",
			candidate:  mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidDelta, Metadata: Metadata{Lane: LaneAsteroids, Sequence: 1}}, nil),
			wantLane:   LaneAsteroids,
			wantFamily: PacketFamilyAsteroidDelta,
		},
		{
			name:       "bullet",
			candidate:  mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: PacketFamilyBulletDelta, Metadata: Metadata{Lane: LaneBullets, Sequence: 1}}, nil),
			wantLane:   LaneBullets,
			wantFamily: PacketFamilyBulletDelta,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := scheduleRecordForCandidate(0, tt.candidate)
			if record.Lane != tt.wantLane {
				t.Fatalf("record lane = %q, want %q", record.Lane, tt.wantLane)
			}
			if record.PacketFamily != tt.wantFamily {
				t.Fatalf("record packet family = %q, want %q", record.PacketFamily, tt.wantFamily)
			}
			if record.DeliveryClass != DeliveryClassHotSupersedable {
				t.Fatalf("record delivery class = %q, want hot supersedable", record.DeliveryClass)
			}
			if record.Priority != PriorityHigh {
				t.Fatalf("record priority = %q, want high", record.Priority)
			}
		})
	}
}

func TestScheduleRecordForAsteroidLifecycleDeltaCandidateUsesRequiredDelivery(t *testing.T) {
	record := scheduleRecordForCandidate(0, mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidsLifecycle, Metadata: Metadata{Lane: LaneAsteroidsLifecycle, Sequence: 1}}, nil))

	if record.Lane != LaneAsteroidsLifecycle {
		t.Fatalf("record lane = %q, want %q", record.Lane, LaneAsteroidsLifecycle)
	}
	if record.PacketFamily != PacketFamilyAsteroidsLifecycle {
		t.Fatalf("record packet family = %q, want %q", record.PacketFamily, PacketFamilyAsteroidsLifecycle)
	}
	if record.DeliveryClass != DeliveryClassRequired {
		t.Fatalf("record delivery class = %q, want required", record.DeliveryClass)
	}
	if record.Priority != PriorityCritical {
		t.Fatalf("record priority = %q, want critical", record.Priority)
	}
}

func TestScheduleRecordForBulletLifecycleDeltaCandidateUsesRequiredDelivery(t *testing.T) {
	record := scheduleRecordForCandidate(0, mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: PacketFamilyBulletsLifecycle, Metadata: Metadata{Lane: LaneBulletsLifecycle, Sequence: 1}}, nil))

	if record.Lane != LaneBulletsLifecycle {
		t.Fatalf("record lane = %q, want %q", record.Lane, LaneBulletsLifecycle)
	}
	if record.PacketFamily != PacketFamilyBulletsLifecycle {
		t.Fatalf("record packet family = %q, want %q", record.PacketFamily, PacketFamilyBulletsLifecycle)
	}
	if record.DeliveryClass != DeliveryClassRequired {
		t.Fatalf("record delivery class = %q, want required", record.DeliveryClass)
	}
	if record.Priority != PriorityCritical {
		t.Fatalf("record priority = %q, want critical", record.Priority)
	}
}

func TestScheduleRecordForLifecycleDeltaCandidateStaysSingleChunk(t *testing.T) {
	tests := []struct {
		name      string
		candidate RealtimeLaneCandidate
	}{
		{
			name:      "asteroid lifecycle",
			candidate: mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidsLifecycle, Metadata: Metadata{Lane: LaneAsteroidsLifecycle, Sequence: 1, ChunkIndex: 0, ChunkCount: 1, IsFinalChunk: true}}, nil),
		},
		{
			name:      "bullet lifecycle",
			candidate: mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: PacketFamilyBulletsLifecycle, Metadata: Metadata{Lane: LaneBulletsLifecycle, Sequence: 1, ChunkIndex: 0, ChunkCount: 1, IsFinalChunk: true}}, nil),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			record := scheduleRecordForCandidate(0, tc.candidate)
			if record.DeliveryClass != DeliveryClassRequired || record.Priority != PriorityCritical {
				t.Fatalf("record = %#v, want required critical lifecycle traffic", record)
			}
			if record.ChunkCount != 1 || !record.IsFinalChunk {
				t.Fatalf("record = %#v, want single final chunk", record)
			}
		})
	}
}
