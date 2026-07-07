package realtime

import "testing"

func TestLaneCandidateClassificationTreatsLifecycleLanesAsRequiredCritical(t *testing.T) {
	tests := []struct {
		name         string
		candidate    RealtimeLaneCandidate
		wantDelivery DeliveryClass
		wantPriority  Priority
	}{
		{
			name: "asteroid lifecycle",
			candidate: RealtimeLaneCandidate{
				Lane: LaneAsteroidsLifecycle,
				Kind: RealtimeLaneCandidateKindDelta,
				Delta: AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidDelta, Metadata: Metadata{Lane: LaneAsteroidsLifecycle, Sequence: 1}},
			},
			wantDelivery: DeliveryClassRequired,
			wantPriority:  PriorityCritical,
		},
		{
			name: "bullet lifecycle",
			candidate: RealtimeLaneCandidate{
				Lane: LaneBulletsLifecycle,
				Kind: RealtimeLaneCandidateKindDelta,
				Delta: BulletWireDeltaPacket{Type: PacketFamilyBulletDelta, Metadata: Metadata{Lane: LaneBulletsLifecycle, Sequence: 1}},
			},
			wantDelivery: DeliveryClassRequired,
			wantPriority:  PriorityCritical,
		},
		{
			name: "asteroid hot lane",
			candidate: RealtimeLaneCandidate{
				Lane: LaneAsteroids,
				Kind: RealtimeLaneCandidateKindDelta,
				Delta: AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidDelta, Metadata: Metadata{Lane: LaneAsteroids, Sequence: 1}},
			},
			wantDelivery: DeliveryClassHotSupersedable,
			wantPriority:  PriorityHigh,
		},
		{
			name: "bullet hot lane",
			candidate: RealtimeLaneCandidate{
				Lane: LaneBullets,
				Kind: RealtimeLaneCandidateKindDelta,
				Delta: BulletWireDeltaPacket{Type: PacketFamilyBulletDelta, Metadata: Metadata{Lane: LaneBullets, Sequence: 1}},
			},
			wantDelivery: DeliveryClassHotSupersedable,
			wantPriority:  PriorityHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deliveryClassForCandidate(tt.candidate); got != tt.wantDelivery {
				t.Fatalf("delivery class = %q, want %q", got, tt.wantDelivery)
			}
			if got := priorityForCandidate(tt.candidate); got != tt.wantPriority {
				t.Fatalf("priority = %q, want %q", got, tt.wantPriority)
			}
		})
	}
}
