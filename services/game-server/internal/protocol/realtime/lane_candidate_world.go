package realtime

import (
	"fmt"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func buildWorldLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, sessionState *RealtimeSessionState, sharedWorld *WorldWireFullPacket) ([]RealtimeLaneCandidate, error) {
	candidates := make([]RealtimeLaneCandidate, 0, 7)

	worldState, worldSynced := state.LaneState(LaneWorld)
	worldReady := state.LaneBaselineReady(LaneWorld)
	worldSequence := NextLaneSequence(worldState, worldSynced)
	var currentWorld WorldWireFullPacket
	if sharedWorld != nil {
		currentWorld = receiverWorldWireProjection(*sharedWorld, snapshot, worldSequence)
	} else {
		worldFull := BuildWorldFullPacket(snapshot, worldSequence)
		var err error
		currentWorld, err = quantizeWorldFullPacket(worldFull)
		if err != nil {
			return nil, fmt.Errorf("quantize world full packet: %w", err)
		}
	}

	storedWorld, worldHasProjection := state.BaselineProjection(LaneWorld)
	worldCanUseProjection := worldReady && worldSynced && worldState.IsFinalChunk && worldState.BaselineID != "" && worldHasProjection
	if !worldCanUseProjection {
		return append(candidates, mustRealtimeLaneCandidate(currentWorld, currentWorld)), nil
	}
	previousWorld, ok := storedWorld.(WorldWireFullPacket)
	if !ok {
		return append(candidates, mustRealtimeLaneCandidate(currentWorld, currentWorld)), nil
	}

	if WorldWirePayloadChanged(previousWorld, currentWorld) {
		candidates = append(candidates, buildReliableWorldCandidates(previousWorld, currentWorld, state)...)
	}

	hotCandidates, cohortState := buildIndependentHotLaneCandidates(previousWorld, currentWorld, state)
	candidates = append(candidates, hotCandidates...)
	if sessionState != nil {
		sessionState.HotLaneCohorts = cohortState
	}
	return candidates, nil
}

func buildReliableWorldCandidates(previousWorld WorldWireFullPacket, currentWorld WorldWireFullPacket, state RealtimeSessionState) []RealtimeLaneCandidate {
	worldDelta := BuildWorldWireDeltaPacket(previousWorld, currentWorld)
	_, shipReliableUpdates := partitionShipUpdates(worldDelta.Ships.Updates)
	worldDelta.Ships.Updates = shipReliableUpdates
	worldDelta.Asteroids.Updates = nil
	worldDelta.Bullets.Updates = nil

	shipLifecyclePresent := len(worldDelta.Ships.Creates) > 0 || len(worldDelta.Ships.Updates) > 0 || len(worldDelta.Ships.Deletes) > 0
	asteroidLifecyclePresent := len(worldDelta.Asteroids.Creates) > 0 || len(worldDelta.Asteroids.Deletes) > 0
	bulletLifecyclePresent := len(worldDelta.Bullets.Creates) > 0 || len(worldDelta.Bullets.Deletes) > 0
	projectionAdvanceRequired := WorldWireDeltaHasChanges(worldDelta)
	if !projectionAdvanceRequired {
		return nil
	}

	candidates := make([]RealtimeLaneCandidate, 0, 4)
	if shipLifecyclePresent {
		if candidate, ok := buildShipLifecycleCandidate(worldDelta, state); ok {
			candidates = append(candidates, candidate)
			worldDelta.Ships.Creates = nil
			worldDelta.Ships.Updates = nil
			worldDelta.Ships.Deletes = nil
		}
	}
	if asteroidLifecyclePresent {
		if candidate, ok := buildAsteroidLifecycleCandidate(worldDelta, state); ok {
			candidates = append(candidates, candidate)
			worldDelta.Asteroids.Creates = nil
			worldDelta.Asteroids.Deletes = nil
		}
	}
	if bulletLifecyclePresent {
		if candidate, ok := buildBulletLifecycleCandidate(worldDelta, state); ok {
			candidates = append(candidates, candidate)
			worldDelta.Bullets.Creates = nil
			worldDelta.Bullets.Deletes = nil
		}
	}

	chainedWorldProjection := currentWorld
	chainedWorldProjection.Metadata = worldDelta.Metadata
	candidates = append(candidates, mustRealtimeLaneCandidate(worldDelta, chainedWorldProjection))
	return candidates
}

func buildIndependentHotLaneCandidates(previousWorld WorldWireFullPacket, currentWorld WorldWireFullPacket, state RealtimeSessionState) ([]RealtimeLaneCandidate, HotLaneCohortState) {
	cohortState := state.HotLaneCohorts
	cohortState.EnsureInitialized()
	candidates := make([]RealtimeLaneCandidate, 0, 3)

	currentShips := projectShipHotLane(currentWorld)
	shipActive := shipHotProjectionIDs(currentShips)
	assignHotRoutesWithOverride(cohortState.ShipRoutes, sortedHotLaneUpdateIDs(shipActive), HotUpdateRouteShips)
	cohortState.RemoveMissingShips(shipActive)
	shipUpdates := shipHotLaneUpdates(previousShipHotLaneProjection(state, previousWorld), currentShips)
	if len(shipUpdates) > 0 {
		packet := ShipWireDeltaPacket{
			Type:        PacketFamilyShipDelta,
			Metadata:    nextHotLaneMetadata(currentWorld.Metadata, state, LaneShips),
			ShipUpdates: shipUpdates,
		}
		cohortState.ShipMode = hotLaneModeForChunkCount(shipWireDeltaChunkCount(packet))
		if hotPacketCadenceAllows(cohortState.ShipMode, state.HotLaneTick) {
			candidates = append(candidates, mustRealtimeLaneCandidate(packet, currentShips))
		}
	} else {
		cohortState.ShipMode = HotLaneModeInline
	}

	currentAsteroids := projectAsteroidHotLane(currentWorld)
	asteroidActive := asteroidHotProjectionIDs(currentAsteroids)
	assignHotRoutesWithOverride(cohortState.AsteroidRoutes, sortedHotLaneUpdateIDs(asteroidActive), HotUpdateRouteAsteroids)
	cohortState.RemoveMissingAsteroids(asteroidActive)
	asteroidUpdates := asteroidHotLaneUpdates(previousAsteroidHotLaneProjection(state, previousWorld), currentAsteroids)
	if len(asteroidUpdates) > 0 {
		packet := AsteroidWireDeltaPacket{
			Type:            PacketFamilyAsteroidDelta,
			Metadata:        nextHotLaneMetadata(currentWorld.Metadata, state, LaneAsteroids),
			AsteroidUpdates: asteroidUpdates,
		}
		cohortState.AsteroidMode = hotLaneModeForChunkCount(asteroidWireDeltaChunkCount(packet))
		if hotPacketCadenceAllows(cohortState.AsteroidMode, state.HotLaneTick) {
			candidates = append(candidates, mustRealtimeLaneCandidate(packet, currentAsteroids))
		}
	} else {
		cohortState.AsteroidMode = HotLaneModeInline
	}

	currentBullets := projectBulletHotLane(currentWorld)
	bulletActive := bulletHotProjectionIDs(currentBullets)
	assignHotRoutesWithOverride(cohortState.BulletRoutes, sortedHotLaneUpdateIDs(bulletActive), HotUpdateRouteBullets)
	cohortState.RemoveMissingBullets(bulletActive)
	bulletUpdates := bulletHotLaneUpdates(previousBulletHotLaneProjection(state, previousWorld), currentBullets)
	if len(bulletUpdates) > 0 {
		packet := BulletWireDeltaPacket{
			Type:          PacketFamilyBulletDelta,
			Metadata:      nextHotLaneMetadata(currentWorld.Metadata, state, LaneBullets),
			BulletUpdates: bulletUpdates,
		}
		cohortState.BulletMode = hotLaneModeForChunkCount(bulletWireDeltaChunkCount(packet))
		if hotPacketCadenceAllows(cohortState.BulletMode, state.HotLaneTick) {
			candidates = append(candidates, mustRealtimeLaneCandidate(packet, currentBullets))
		}
	} else {
		cohortState.BulletMode = HotLaneModeInline
	}

	return candidates, cohortState
}

func nextHotLaneMetadata(world Metadata, state RealtimeSessionState, lane Lane) Metadata {
	laneState, synced := state.LaneState(lane)
	sequence := NextLaneSequence(laneState, synced)
	if worldState, ok := state.LaneState(LaneWorld); ok && worldState.BaselineID != "" {
		world.BaselineID = worldState.BaselineID
	}
	world.MatchID = state.MatchID
	world.Lane = lane
	world.Sequence = sequence
	world.SnapshotID = DeltaSnapshotID(lane, sequence)
	world.SnapshotKind = SnapshotKind("delta")
	world = world.WithChunk(0, 1)
	return world
}

func shipHotProjectionIDs(projection ShipHotLaneProjection) map[string]bool {
	ids := make(map[string]bool, len(projection.Records))
	for _, record := range projection.Records {
		ids[record.ID] = true
	}
	return ids
}

func asteroidHotProjectionIDs(projection AsteroidHotLaneProjection) map[string]bool {
	ids := make(map[string]bool, len(projection.Records))
	for _, record := range projection.Records {
		ids[record.ID] = true
	}
	return ids
}

func bulletHotProjectionIDs(projection BulletHotLaneProjection) map[string]bool {
	ids := make(map[string]bool, len(projection.Records))
	for _, record := range projection.Records {
		ids[record.ID] = true
	}
	return ids
}
