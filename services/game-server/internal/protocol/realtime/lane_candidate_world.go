package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
)

func buildWorldLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, sessionState *RealtimeSessionState) []RealtimeLaneCandidate {
	candidates := make([]RealtimeLaneCandidate, 0, 4)

	worldState, worldSynced := state.LaneState(LaneWorld)
	worldReady := state.LaneBaselineReady(LaneWorld)
	worldSequence := NextLaneSequence(worldState, worldSynced)
	worldFull := BuildWorldFullPacket(snapshot, worldSequence)
	quantizedWorldFull, err := quantizeWorldFullPacket(worldFull)
	if err != nil {
		return candidates
	}
	worldProjection, worldHasProjection := state.BaselineProjection(LaneWorld)
	worldCanUseProjection := worldReady && worldSynced && worldState.IsFinalChunk && worldState.BaselineID != "" && worldHasProjection
	if !worldCanUseProjection {
		candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull, Full: quantizedWorldFull, Projection: quantizedWorldFull})
	} else {
		previousWorldFull, ok := worldProjection.(WorldWireFullPacket)
		if !ok {
			candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull, Full: quantizedWorldFull, Projection: quantizedWorldFull})
		} else if !WorldWirePayloadChanged(previousWorldFull, quantizedWorldFull) {
			// No world candidate when the projection is unchanged.
		} else {
			worldDelta := BuildWorldWireDeltaPacket(previousWorldFull, quantizedWorldFull)
			split := SplitWorldHotUpdates(worldDelta, state.HotLaneCohorts, DefaultHotLaneOffloadPolicy())
			if sessionState != nil {
				sessionState.HotLaneCohorts = split.CohortState
			}
			if len(split.WorldDelta.Bullets.Creates) > 0 || len(split.WorldDelta.Bullets.Deletes) > 0 {
				if bulletCandidate, ok := buildBulletLifecycleCandidate(split.WorldDelta, state); ok {
					candidates = append(candidates, bulletCandidate)
					split.WorldDelta.Bullets.Creates = nil
					split.WorldDelta.Bullets.Deletes = nil
				}
			}
			if len(split.WorldDelta.Asteroids.Creates) > 0 || len(split.WorldDelta.Asteroids.Deletes) > 0 {
				if asteroidCandidate, ok := buildAsteroidLifecycleCandidate(split.WorldDelta, state); ok {
					candidates = append(candidates, asteroidCandidate)
					split.WorldDelta.Asteroids.Creates = nil
					split.WorldDelta.Asteroids.Deletes = nil
				}
			}

			asteroidHotPresent := split.AsteroidDelta != nil && len(split.AsteroidDelta.AsteroidUpdates) > 0
			bulletHotPresent := split.BulletDelta != nil && len(split.BulletDelta.BulletUpdates) > 0
			worldDeltaHasChanges := WorldWireDeltaHasChanges(split.WorldDelta)

			asteroidHotAllowed := asteroidHotPresent
			bulletHotAllowed := bulletHotPresent
			asteroidState, asteroidSynced := state.LaneState(LaneAsteroids)
			asteroidSequence := NextLaneSequence(asteroidState, asteroidSynced)
			bulletState, bulletSynced := state.LaneState(LaneBullets)
			bulletSequence := NextLaneSequence(bulletState, bulletSynced)
			if split.AsteroidDelta != nil {
				metadata := split.AsteroidDelta.Metadata
				metadata.Lane = LaneAsteroids
				metadata.Sequence = asteroidSequence
				metadata.SnapshotID = DeltaSnapshotID(LaneAsteroids, asteroidSequence)
				metadata.SnapshotKind = SnapshotKind("delta")
				metadata.ServerSentMsec = split.WorldDelta.Metadata.ServerSentMsec
				metadata = metadata.WithChunk(0, 1)
				split.AsteroidDelta.Metadata = metadata
			}
			if split.BulletDelta != nil {
				metadata := split.BulletDelta.Metadata
				metadata.Lane = LaneBullets
				metadata.Sequence = bulletSequence
				metadata.SnapshotID = DeltaSnapshotID(LaneBullets, bulletSequence)
				metadata.SnapshotKind = SnapshotKind("delta")
				metadata.ServerSentMsec = split.WorldDelta.Metadata.ServerSentMsec
				metadata = metadata.WithChunk(0, 1)
				split.BulletDelta.Metadata = metadata
			}
			if worldDeltaHasChanges || asteroidHotAllowed || bulletHotAllowed {
				chainedWorldProjection := quantizedWorldFull
				chainedWorldProjection.Metadata = split.WorldDelta.Metadata
				candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindDelta, Delta: split.WorldDelta, Projection: chainedWorldProjection})
			}
			if asteroidHotAllowed {
				candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneAsteroids, Kind: RealtimeLaneCandidateKindDelta, Delta: *split.AsteroidDelta})
			}
			if bulletHotAllowed {
				candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneBullets, Kind: RealtimeLaneCandidateKindDelta, Delta: *split.BulletDelta})
			}
		}
	}

	return candidates
}