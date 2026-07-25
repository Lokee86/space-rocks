package realtime

import "sort"

type HotLaneSplitResult struct {
	WorldDelta        WorldWireDeltaPacket
	ShipDelta         *ShipWireDeltaPacket
	AsteroidDelta     *AsteroidWireDeltaPacket
	BulletDelta       *BulletWireDeltaPacket
	CohortState       HotLaneCohortState
	WorldHotCount     int
	ShipHotCount      int
	AsteroidHotCount  int
	BulletHotCount    int
	ShipOffloaded     int
	AsteroidOffloaded int
	BulletOffloaded   int
	ShipMode          HotLaneMode
	AsteroidMode      HotLaneMode
	BulletMode        HotLaneMode
}

type hotLaneSortedUpdateID struct {
	ID string
}

type hotLaneSortedUpdateIDs []hotLaneSortedUpdateID

func SplitWorldHotUpdates(worldDelta WorldWireDeltaPacket, cohortState HotLaneCohortState, _ HotLaneOffloadPolicy) HotLaneSplitResult {
	cohortState.EnsureInitialized()

	shipActive := activeUpdateIDsFromWireRecords(worldDelta.Ships.Updates)
	asteroidActive := activeUpdateIDsFromWireRecords(worldDelta.Asteroids.Updates)
	bulletActive := activeUpdateIDsFromWireRecords(worldDelta.Bullets.Updates)
	shipIDs := sortedHotLaneUpdateIDs(shipActive)
	asteroidIDs := sortedHotLaneUpdateIDs(asteroidActive)
	bulletIDs := sortedHotLaneUpdateIDs(bulletActive)

	result := HotLaneSplitResult{
		WorldDelta:       worldDelta,
		CohortState:      cohortState,
		ShipMode:         HotLaneModeInline,
		AsteroidMode:     HotLaneModeInline,
		BulletMode:       HotLaneModeInline,
		ShipHotCount:     len(shipIDs),
		AsteroidHotCount: len(asteroidIDs),
		BulletHotCount:   len(bulletIDs),
	}
	result.WorldHotCount = result.ShipHotCount + result.AsteroidHotCount + result.BulletHotCount

	if len(shipIDs) > 0 {
		assignHotRoutesWithOverride(result.CohortState.ShipRoutes, shipIDs, HotUpdateRouteShips)
		result.WorldDelta.Ships.Updates = nil
		result.ShipMode = HotLaneModeFullOwned60Hz
		result.CohortState.ShipMode = result.ShipMode
		result.ShipOffloaded = len(shipIDs)
		metadata := worldDelta.Metadata
		metadata.Lane = LaneShips
		metadata.SnapshotKind = SnapshotKind("delta")
		metadata = metadata.WithChunk(0, 1)
		result.ShipDelta = &ShipWireDeltaPacket{
			Type:        PacketFamilyShipDelta,
			Metadata:    metadata,
			ShipUpdates: worldDelta.Ships.Updates,
		}
	}
	if len(asteroidIDs) > 0 {
		assignHotRoutesWithOverride(result.CohortState.AsteroidRoutes, asteroidIDs, HotUpdateRouteAsteroids)
		result.WorldDelta.Asteroids.Updates = nil
		result.AsteroidMode = HotLaneModeFullOwned60Hz
		result.CohortState.AsteroidMode = result.AsteroidMode
		result.AsteroidOffloaded = len(asteroidIDs)
		metadata := worldDelta.Metadata
		metadata.Lane = LaneAsteroids
		metadata.SnapshotKind = SnapshotKind("delta")
		metadata = metadata.WithChunk(0, 1)
		result.AsteroidDelta = &AsteroidWireDeltaPacket{
			Type:            PacketFamilyAsteroidDelta,
			Metadata:        metadata,
			AsteroidUpdates: worldDelta.Asteroids.Updates,
		}
	}
	if len(bulletIDs) > 0 {
		assignHotRoutesWithOverride(result.CohortState.BulletRoutes, bulletIDs, HotUpdateRouteBullets)
		result.WorldDelta.Bullets.Updates = nil
		result.BulletMode = HotLaneModeFullOwned60Hz
		result.CohortState.BulletMode = result.BulletMode
		result.BulletOffloaded = len(bulletIDs)
		metadata := worldDelta.Metadata
		metadata.Lane = LaneBullets
		metadata.SnapshotKind = SnapshotKind("delta")
		metadata = metadata.WithChunk(0, 1)
		result.BulletDelta = &BulletWireDeltaPacket{
			Type:          PacketFamilyBulletDelta,
			Metadata:      metadata,
			BulletUpdates: worldDelta.Bullets.Updates,
		}
	}
	result.CohortState.RemoveMissingShips(shipActive)
	result.CohortState.RemoveMissingAsteroids(asteroidActive)
	result.CohortState.RemoveMissingBullets(bulletActive)
	return result
}

func sortedHotLaneUpdateIDs(active map[string]bool) hotLaneSortedUpdateIDs {
	ids := make(hotLaneSortedUpdateIDs, 0, len(active))
	for id := range active {
		ids = append(ids, hotLaneSortedUpdateID{ID: id})
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].ID < ids[j].ID
	})
	return ids
}

func assignHotRoutesWithOverride(routes map[string]HotUpdateRoute, ids hotLaneSortedUpdateIDs, route HotUpdateRoute) {
	for _, item := range ids {
		if existing, exists := routes[item.ID]; exists && existing == route {
			continue
		}
		routes[item.ID] = route
	}
}

func activeUpdateIDsFromWireRecords(records []map[string]any) map[string]bool {
	active := make(map[string]bool, len(records))
	for _, record := range records {
		id := asString(record["id"])
		if id == "" {
			continue
		}
		active[id] = true
	}
	return active
}
