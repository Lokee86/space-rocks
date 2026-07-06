package realtime

import "sort"

type HotLaneSplitResult struct {
	WorldDelta        WorldWireDeltaPacket
	AsteroidDelta     *AsteroidWireDeltaPacket
	BulletDelta       *BulletWireDeltaPacket
	CohortState       HotLaneCohortState
	WorldHotCount     int
	AsteroidHotCount  int
	BulletHotCount    int
	AsteroidOffloaded int
	BulletOffloaded   int
	AsteroidMode      HotLaneMode
	BulletMode        HotLaneMode
}

type hotLaneSortedUpdateID struct {
	ID string
}

type hotLaneSortedUpdateIDs []hotLaneSortedUpdateID

func SplitWorldHotUpdates(worldDelta WorldWireDeltaPacket, cohortState HotLaneCohortState, policy HotLaneOffloadPolicy) HotLaneSplitResult {
	cohortState.EnsureInitialized()

	asteroidActive := activeUpdateIDsFromWireRecords(worldDelta.Asteroids.Updates)
	bulletActive := activeUpdateIDsFromWireRecords(worldDelta.Bullets.Updates)
	asteroidIDs := sortedHotLaneUpdateIDs(asteroidActive)
	bulletIDs := sortedHotLaneUpdateIDs(bulletActive)

	result := HotLaneSplitResult{
		WorldDelta:       worldDelta,
		CohortState:      cohortState,
		AsteroidMode:     HotLaneModeInline,
		BulletMode:       HotLaneModeInline,
		AsteroidHotCount: len(asteroidIDs),
		BulletHotCount:   len(bulletIDs),
	}
	result.WorldHotCount = result.AsteroidHotCount + result.BulletHotCount

	if len(asteroidIDs) > 0 {
		assignHotRoutesWithOverride(result.CohortState.AsteroidRoutes, asteroidIDs, HotUpdateRouteAsteroids)
		result.WorldDelta.Asteroids.Updates = nil
		result.AsteroidMode = hotLaneModeForFullOwnedCount(len(asteroidIDs), policy.AsteroidHotLaneEntityBudget)
		result.CohortState.AsteroidMode = result.AsteroidMode
		result.AsteroidOffloaded = len(asteroidIDs)
		result.AsteroidDelta = &AsteroidWireDeltaPacket{
			Type:            PacketFamilyAsteroidDelta,
			Sequence:        worldDelta.Metadata.Sequence,
			ServerSentMsec:  worldDelta.Metadata.ServerSentMsec,
			AsteroidUpdates: worldDelta.Asteroids.Updates,
		}
	}
	if len(bulletIDs) > 0 {
		assignHotRoutesWithOverride(result.CohortState.BulletRoutes, bulletIDs, HotUpdateRouteBullets)
		result.WorldDelta.Bullets.Updates = nil
		result.BulletMode = hotLaneModeForFullOwnedCount(len(bulletIDs), policy.BulletHotLaneEntityBudget)
		result.CohortState.BulletMode = result.BulletMode
		result.BulletOffloaded = len(bulletIDs)
		result.BulletDelta = &BulletWireDeltaPacket{
			Type:           PacketFamilyBulletDelta,
			Sequence:       worldDelta.Metadata.Sequence,
			ServerSentMsec: worldDelta.Metadata.ServerSentMsec,
			BulletUpdates:   worldDelta.Bullets.Updates,
		}
	}
	result.CohortState.RemoveMissingAsteroids(asteroidActive)
	result.CohortState.RemoveMissingBullets(bulletActive)
	return result
}

func hotLaneModeForFullOwnedCount(count int, budget int) HotLaneMode {
	if count > budget*3 {
		return HotLaneModeNeedsChunking
	}
	if count > budget*2 {
		return HotLaneModeFullOwned20Hz
	}
	return HotLaneModeFullOwned30Hz
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
