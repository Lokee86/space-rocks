package realtime

func buildShipLifecycleCandidate(worldDelta WorldWireDeltaPacket, state RealtimeSessionState) (RealtimeLaneCandidate, bool) {
	if len(worldDelta.Ships.Creates) == 0 && len(worldDelta.Ships.Updates) == 0 && len(worldDelta.Ships.Deletes) == 0 {
		return RealtimeLaneCandidate{}, false
	}

	lifecycleState, lifecycleSynced := state.LaneState(LaneShipsLifecycle)
	lifecycleSequence := NextLaneSequence(lifecycleState, lifecycleSynced)
	metadata := worldDelta.Metadata
	metadata.Lane = LaneShipsLifecycle
	metadata.Sequence = lifecycleSequence
	metadata.SnapshotID = DeltaSnapshotID(LaneShipsLifecycle, lifecycleSequence)
	metadata.SnapshotKind = SnapshotKind("delta")
	metadata = metadata.WithChunk(0, 1)

	return mustRealtimeLaneCandidate(ShipWireDeltaPacket{
		Type:        PacketFamilyShipsLifecycle,
		Metadata:    metadata,
		ShipCreates: worldDelta.Ships.Creates,
		ShipUpdates: worldDelta.Ships.Updates,
		ShipDeletes: worldDelta.Ships.Deletes,
	}, nil), true
}

func buildBulletLifecycleCandidate(worldDelta WorldWireDeltaPacket, state RealtimeSessionState) (RealtimeLaneCandidate, bool) {
	if len(worldDelta.Bullets.Creates) == 0 && len(worldDelta.Bullets.Deletes) == 0 {
		return RealtimeLaneCandidate{}, false
	}

	lifecycleState, lifecycleSynced := state.LaneState(LaneBulletsLifecycle)
	lifecycleSequence := NextLaneSequence(lifecycleState, lifecycleSynced)
	metadata := worldDelta.Metadata
	metadata.Lane = LaneBulletsLifecycle
	metadata.Sequence = lifecycleSequence
	metadata.SnapshotID = DeltaSnapshotID(LaneBulletsLifecycle, lifecycleSequence)
	metadata.SnapshotKind = SnapshotKind("delta")
	metadata = metadata.WithChunk(0, 1)

	return mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: PacketFamilyBulletsLifecycle, Metadata: metadata, BulletCreates: worldDelta.Bullets.Creates, BulletDeletes: worldDelta.Bullets.Deletes}, nil), true
}

func buildAsteroidLifecycleCandidate(worldDelta WorldWireDeltaPacket, state RealtimeSessionState) (RealtimeLaneCandidate, bool) {
	if len(worldDelta.Asteroids.Creates) == 0 && len(worldDelta.Asteroids.Deletes) == 0 {
		return RealtimeLaneCandidate{}, false
	}

	lifecycleState, lifecycleSynced := state.LaneState(LaneAsteroidsLifecycle)
	lifecycleSequence := NextLaneSequence(lifecycleState, lifecycleSynced)
	metadata := worldDelta.Metadata
	metadata.Lane = LaneAsteroidsLifecycle
	metadata.Sequence = lifecycleSequence
	metadata.SnapshotID = DeltaSnapshotID(LaneAsteroidsLifecycle, lifecycleSequence)
	metadata.SnapshotKind = SnapshotKind("delta")
	metadata = metadata.WithChunk(0, 1)

	return mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidsLifecycle, Metadata: metadata, AsteroidCreates: worldDelta.Asteroids.Creates, AsteroidDeletes: worldDelta.Asteroids.Deletes}, nil), true
}
