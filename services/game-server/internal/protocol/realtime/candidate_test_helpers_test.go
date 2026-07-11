package realtime

func testCandidate(lane Lane, kind RealtimeLaneCandidateKind) RealtimeLaneCandidate {
	metadata := Metadata{Lane: lane, Sequence: 1, SnapshotKind: SnapshotKind("test"), IsFinalChunk: true}
	switch lane {
	case LaneWorld:
		if kind == RealtimeLaneCandidateKindFull {
			return mustRealtimeLaneCandidate(WorldWireFullPacket{Type: PacketFamilyWorldFull, Metadata: metadata}, nil)
		}
		return mustRealtimeLaneCandidate(WorldWireDeltaPacket{Type: PacketFamilyWorldDelta, Metadata: metadata}, nil)
	case LaneOverlay:
		if kind == RealtimeLaneCandidateKindFull {
			return mustRealtimeLaneCandidate(OverlayWireFullPacket{Type: PacketFamilyOverlayFull, Metadata: metadata}, nil)
		}
		return mustRealtimeLaneCandidate(OverlayWireLaneDelta{Metadata: metadata}, nil)
	case LaneSession:
		if kind == RealtimeLaneCandidateKindFull {
			return mustRealtimeLaneCandidate(SessionWireFullPacket{Type: PacketFamilySessionFull, Metadata: metadata}, nil)
		}
		return mustRealtimeLaneCandidate(SessionWireLaneDelta{Metadata: metadata}, nil)
	case LaneAsteroids, LaneAsteroidsLifecycle:
		family := PacketFamilyAsteroidDelta
		if lane == LaneAsteroidsLifecycle {
			family = PacketFamilyAsteroidsLifecycle
		}
		return mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: family, Metadata: metadata}, nil)
	case LaneBullets, LaneBulletsLifecycle:
		family := PacketFamilyBulletDelta
		if lane == LaneBulletsLifecycle {
			family = PacketFamilyBulletsLifecycle
		}
		return mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: family, Metadata: metadata}, nil)
	case LaneEvent:
		return mustRealtimeLaneCandidate(EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: metadata}, nil)
	default:
		return RealtimeLaneCandidate{}
	}
}
