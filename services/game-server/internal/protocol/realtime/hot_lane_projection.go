package realtime

type ShipHotLaneProjection struct {
	Records []WorldShipWireRecord
}

type AsteroidHotLaneProjection struct {
	Records []WorldAsteroidWireRecord
}

type BulletHotLaneProjection struct {
	Records []WorldBulletWireRecord
}

func projectShipHotLane(packet WorldWireFullPacket) ShipHotLaneProjection {
	records := make([]WorldShipWireRecord, 0, len(packet.Ships))
	for _, record := range packet.Ships {
		records = append(records, WorldShipWireRecord{
			ID:        record.ID,
			X:         record.X,
			Y:         record.Y,
			Rotation:  record.Rotation,
			Thrusting: record.Thrusting,
		})
	}
	return ShipHotLaneProjection{Records: records}
}

func projectAsteroidHotLane(packet WorldWireFullPacket) AsteroidHotLaneProjection {
	records := make([]WorldAsteroidWireRecord, 0, len(packet.Asteroids))
	for _, record := range packet.Asteroids {
		records = append(records, WorldAsteroidWireRecord{
			ID: record.ID,
			X:  record.X,
			Y:  record.Y,
		})
	}
	return AsteroidHotLaneProjection{Records: records}
}

func projectBulletHotLane(packet WorldWireFullPacket) BulletHotLaneProjection {
	records := make([]WorldBulletWireRecord, 0, len(packet.Bullets))
	for _, record := range packet.Bullets {
		records = append(records, WorldBulletWireRecord{
			ID:       record.ID,
			X:        record.X,
			Y:        record.Y,
			Rotation: record.Rotation,
		})
	}
	return BulletHotLaneProjection{Records: records}
}

func shipHotLaneUpdates(previous ShipHotLaneProjection, current ShipHotLaneProjection) []map[string]any {
	delta := CompareLaneRecordFields(previous.Records, current.Records,
		func(record WorldShipWireRecord) string { return record.ID },
		"id",
	)
	return delta.Updates
}

func asteroidHotLaneUpdates(previous AsteroidHotLaneProjection, current AsteroidHotLaneProjection) []map[string]any {
	delta := CompareLaneRecordFields(previous.Records, current.Records,
		func(record WorldAsteroidWireRecord) string { return record.ID },
		"id",
	)
	return delta.Updates
}

func bulletHotLaneUpdates(previous BulletHotLaneProjection, current BulletHotLaneProjection) []map[string]any {
	delta := CompareLaneRecordFields(previous.Records, current.Records,
		func(record WorldBulletWireRecord) string { return record.ID },
		"id",
	)
	return delta.Updates
}

func previousShipHotLaneProjection(state RealtimeSessionState, fallback WorldWireFullPacket) ShipHotLaneProjection {
	if projection, ok := state.BaselineProjection(LaneShips); ok {
		if typed, ok := projection.(ShipHotLaneProjection); ok {
			return typed
		}
	}
	return projectShipHotLane(fallback)
}

func previousAsteroidHotLaneProjection(state RealtimeSessionState, fallback WorldWireFullPacket) AsteroidHotLaneProjection {
	if projection, ok := state.BaselineProjection(LaneAsteroids); ok {
		if typed, ok := projection.(AsteroidHotLaneProjection); ok {
			return typed
		}
	}
	return projectAsteroidHotLane(fallback)
}

func previousBulletHotLaneProjection(state RealtimeSessionState, fallback WorldWireFullPacket) BulletHotLaneProjection {
	if projection, ok := state.BaselineProjection(LaneBullets); ok {
		if typed, ok := projection.(BulletHotLaneProjection); ok {
			return typed
		}
	}
	return projectBulletHotLane(fallback)
}

func seedHotLaneProjections(state *RealtimeSessionState, world WorldWireFullPacket) {
	state.StoreBaselineProjection(LaneShips, projectShipHotLane(world))
	state.StoreBaselineProjection(LaneAsteroids, projectAsteroidHotLane(world))
	state.StoreBaselineProjection(LaneBullets, projectBulletHotLane(world))
}

func syncHotLaneProjectionMembership(state *RealtimeSessionState, world WorldWireFullPacket) {
	state.StoreBaselineProjection(LaneShips, mergeShipHotLaneMembership(state, world))
	state.StoreBaselineProjection(LaneAsteroids, mergeAsteroidHotLaneMembership(state, world))
	state.StoreBaselineProjection(LaneBullets, mergeBulletHotLaneMembership(state, world))
}

func mergeShipHotLaneMembership(state *RealtimeSessionState, world WorldWireFullPacket) ShipHotLaneProjection {
	current := projectShipHotLane(world)
	previous, ok := state.BaselineProjection(LaneShips)
	if !ok {
		return current
	}
	typed, ok := previous.(ShipHotLaneProjection)
	if !ok {
		return current
	}
	previousByID := make(map[string]WorldShipWireRecord, len(typed.Records))
	for _, record := range typed.Records {
		previousByID[record.ID] = record
	}
	for index, record := range current.Records {
		if existing, ok := previousByID[record.ID]; ok {
			current.Records[index] = existing
		}
	}
	return current
}

func mergeAsteroidHotLaneMembership(state *RealtimeSessionState, world WorldWireFullPacket) AsteroidHotLaneProjection {
	current := projectAsteroidHotLane(world)
	previous, ok := state.BaselineProjection(LaneAsteroids)
	if !ok {
		return current
	}
	typed, ok := previous.(AsteroidHotLaneProjection)
	if !ok {
		return current
	}
	previousByID := make(map[string]WorldAsteroidWireRecord, len(typed.Records))
	for _, record := range typed.Records {
		previousByID[record.ID] = record
	}
	for index, record := range current.Records {
		if existing, ok := previousByID[record.ID]; ok {
			current.Records[index] = existing
		}
	}
	return current
}

func mergeBulletHotLaneMembership(state *RealtimeSessionState, world WorldWireFullPacket) BulletHotLaneProjection {
	current := projectBulletHotLane(world)
	previous, ok := state.BaselineProjection(LaneBullets)
	if !ok {
		return current
	}
	typed, ok := previous.(BulletHotLaneProjection)
	if !ok {
		return current
	}
	previousByID := make(map[string]WorldBulletWireRecord, len(typed.Records))
	for _, record := range typed.Records {
		previousByID[record.ID] = record
	}
	for index, record := range current.Records {
		if existing, ok := previousByID[record.ID]; ok {
			current.Records[index] = existing
		}
	}
	return current
}
