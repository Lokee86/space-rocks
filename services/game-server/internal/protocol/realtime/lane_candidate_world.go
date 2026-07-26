package realtime

import (
	"fmt"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func buildWorldLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, sessionState *RealtimeSessionState) ([]RealtimeLaneCandidate, error) {
	candidates := make([]RealtimeLaneCandidate, 0, 7)

	worldState, worldSynced := state.LaneState(LaneWorld)
	worldReady := state.LaneBaselineReady(LaneWorld)
	worldSequence := NextLaneSequence(worldState, worldSynced)
	worldFull := BuildWorldFullPacket(snapshot, worldSequence)
	quantizedWorldFull, err := quantizeWorldFullPacket(worldFull)
	if err != nil {
		return nil, fmt.Errorf("quantize world full packet: %w", err)
	}
	worldProjection, worldHasProjection := state.BaselineProjection(LaneWorld)
	worldCanUseProjection := worldReady && worldSynced && worldState.IsFinalChunk && worldState.BaselineID != "" && worldHasProjection
	if !worldCanUseProjection {
		candidates = append(candidates, mustRealtimeLaneCandidate(quantizedWorldFull, quantizedWorldFull))
	} else {
		previousWorldFull, ok := worldProjection.(WorldWireFullPacket)
		if !ok {
			candidates = append(candidates, mustRealtimeLaneCandidate(quantizedWorldFull, quantizedWorldFull))
		} else if !WorldWirePayloadChanged(previousWorldFull, quantizedWorldFull) {
			// No world candidate when the projection is unchanged.
		} else {
			worldDelta := BuildWorldWireDeltaPacket(previousWorldFull, quantizedWorldFull)
			split := SplitWorldHotUpdates(worldDelta, state.HotLaneCohorts, DefaultHotLaneOffloadPolicy())

			shipLifecyclePresent := len(split.WorldDelta.Ships.Creates) > 0 || len(split.WorldDelta.Ships.Updates) > 0 || len(split.WorldDelta.Ships.Deletes) > 0
			bulletLifecyclePresent := len(split.WorldDelta.Bullets.Creates) > 0 || len(split.WorldDelta.Bullets.Deletes) > 0
			asteroidLifecyclePresent := len(split.WorldDelta.Asteroids.Creates) > 0 || len(split.WorldDelta.Asteroids.Deletes) > 0
			lifecycleCandidates := make([]RealtimeLaneCandidate, 0, 3)
			if shipLifecyclePresent {
				if shipCandidate, ok := buildShipLifecycleCandidate(split.WorldDelta, state); ok {
					lifecycleCandidates = append(lifecycleCandidates, shipCandidate)
					split.WorldDelta.Ships.Creates = nil
					split.WorldDelta.Ships.Updates = nil
					split.WorldDelta.Ships.Deletes = nil
				}
			}
			if bulletLifecyclePresent {
				if bulletCandidate, ok := buildBulletLifecycleCandidate(split.WorldDelta, state); ok {
					lifecycleCandidates = append(lifecycleCandidates, bulletCandidate)
					split.WorldDelta.Bullets.Creates = nil
					split.WorldDelta.Bullets.Deletes = nil
				}
			}
			if asteroidLifecyclePresent {
				if asteroidCandidate, ok := buildAsteroidLifecycleCandidate(split.WorldDelta, state); ok {
					lifecycleCandidates = append(lifecycleCandidates, asteroidCandidate)
					split.WorldDelta.Asteroids.Creates = nil
					split.WorldDelta.Asteroids.Deletes = nil
				}
			}

			shipHotPresent := split.ShipDelta != nil && len(split.ShipDelta.ShipUpdates) > 0
			asteroidHotPresent := split.AsteroidDelta != nil && len(split.AsteroidDelta.AsteroidUpdates) > 0
			bulletHotPresent := split.BulletDelta != nil && len(split.BulletDelta.BulletUpdates) > 0
			worldDeltaHasChanges := WorldWireDeltaHasChanges(split.WorldDelta)
			projectionAdvanceRequired := worldDeltaHasChanges || shipLifecyclePresent || asteroidLifecyclePresent || bulletLifecyclePresent

			shipState, shipSynced := state.LaneState(LaneShips)
			shipSequence := NextLaneSequence(shipState, shipSynced)
			asteroidState, asteroidSynced := state.LaneState(LaneAsteroids)
			asteroidSequence := NextLaneSequence(asteroidState, asteroidSynced)
			bulletState, bulletSynced := state.LaneState(LaneBullets)
			bulletSequence := NextLaneSequence(bulletState, bulletSynced)

			if split.ShipDelta != nil {
				metadata := split.ShipDelta.Metadata
				metadata.Lane = LaneShips
				metadata.Sequence = shipSequence
				metadata.SnapshotID = DeltaSnapshotID(LaneShips, shipSequence)
				metadata.SnapshotKind = SnapshotKind("delta")
				metadata.ServerSentMsec = split.WorldDelta.Metadata.ServerSentMsec
				metadata.MatchID = state.MatchID
				metadata = metadata.WithChunk(0, 1)
				split.ShipDelta.Metadata = metadata
				split.ShipMode = hotLaneModeForShipChunkCount(shipWireDeltaChunkCount(*split.ShipDelta))
				split.CohortState.ShipMode = split.ShipMode
			}
			if split.AsteroidDelta != nil {
				metadata := split.AsteroidDelta.Metadata
				metadata.Lane = LaneAsteroids
				metadata.Sequence = asteroidSequence
				metadata.SnapshotID = DeltaSnapshotID(LaneAsteroids, asteroidSequence)
				metadata.SnapshotKind = SnapshotKind("delta")
				metadata.ServerSentMsec = split.WorldDelta.Metadata.ServerSentMsec
				metadata.MatchID = state.MatchID
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
				metadata.MatchID = state.MatchID
				metadata = metadata.WithChunk(0, 1)
				split.BulletDelta.Metadata = metadata
				split.BulletMode = hotLaneModeForBulletChunkCount(bulletWireDeltaChunkCount(*split.BulletDelta))
				split.CohortState.BulletMode = split.BulletMode
			}
			if split.AsteroidDelta != nil {
				if asteroidWireDeltaRequiresChunking(*split.AsteroidDelta) {
					split.AsteroidMode = HotLaneModeFullOwned30Hz
				} else {
					split.AsteroidMode = HotLaneModeFullOwned60Hz
				}
				split.CohortState.AsteroidMode = split.AsteroidMode
			}
			if sessionState != nil {
				sessionState.HotLaneCohorts = split.CohortState
			}

			shipHotAllowed := shipHotPresent && (worldDeltaHasChanges || shipLifecyclePresent || hotPacketCadenceAllows(split.CohortState.ShipMode, state.HotLaneTick))
			asteroidHotAllowed := asteroidHotPresent && (worldDeltaHasChanges || asteroidLifecyclePresent || hotPacketCadenceAllows(split.CohortState.AsteroidMode, state.HotLaneTick))
			bulletHotAllowed := bulletHotPresent && (worldDeltaHasChanges || bulletLifecyclePresent || hotPacketCadenceAllows(split.CohortState.BulletMode, state.HotLaneTick))
			allPresentHotAllowed := (!shipHotPresent || shipHotAllowed) && (!asteroidHotPresent || asteroidHotAllowed) && (!bulletHotPresent || bulletHotAllowed)
			anyHotAllowed := shipHotAllowed || asteroidHotAllowed || bulletHotAllowed
			if allPresentHotAllowed && (projectionAdvanceRequired || anyHotAllowed) {
				candidates = append(candidates, lifecycleCandidates...)
				chainedWorldProjection := quantizedWorldFull
				chainedWorldProjection.Metadata = split.WorldDelta.Metadata
				candidates = append(candidates, mustRealtimeLaneCandidate(split.WorldDelta, chainedWorldProjection))
			}
			if shipHotAllowed {
				candidates = append(candidates, mustRealtimeLaneCandidate(*split.ShipDelta, nil))
			}
			if asteroidHotAllowed {
				candidates = append(candidates, mustRealtimeLaneCandidate(*split.AsteroidDelta, nil))
			}
			if bulletHotAllowed {
				candidates = append(candidates, mustRealtimeLaneCandidate(*split.BulletDelta, nil))
			}
		}
	}

	return candidates, nil
}
