package realtime

import "testing"

func TestLaneCandidateClassificationTreatsLifecycleAndSessionStateAsRequired(t *testing.T) {
	tests := []struct {
		name         string
		candidate    RealtimeLaneCandidate
		wantDelivery DeliveryClass
		wantPriority Priority
	}{
		{
			name:         "asteroid lifecycle",
			candidate:    mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidsLifecycle, Metadata: Metadata{Lane: LaneAsteroidsLifecycle, Sequence: 1}}, nil),
			wantDelivery: DeliveryClassRequired,
			wantPriority: PriorityCritical,
		},
		{
			name:         "bullet lifecycle",
			candidate:    mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: PacketFamilyBulletsLifecycle, Metadata: Metadata{Lane: LaneBulletsLifecycle, Sequence: 1}}, nil),
			wantDelivery: DeliveryClassRequired,
			wantPriority: PriorityCritical,
		},
		{
			name:         "session delta",
			candidate:    mustRealtimeLaneCandidate(SessionWireLaneDelta{Metadata: Metadata{Lane: LaneSession, Sequence: 1}}, nil),
			wantDelivery: DeliveryClassRequired,
			wantPriority: PriorityMedium,
		},
		{
			name:         "asteroid hot lane",
			candidate:    mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidDelta, Metadata: Metadata{Lane: LaneAsteroids, Sequence: 1}}, nil),
			wantDelivery: DeliveryClassHotSupersedable,
			wantPriority: PriorityHigh,
		},
		{
			name:         "bullet hot lane",
			candidate:    mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: PacketFamilyBulletDelta, Metadata: Metadata{Lane: LaneBullets, Sequence: 1}}, nil),
			wantDelivery: DeliveryClassHotSupersedable,
			wantPriority: PriorityHigh,
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
