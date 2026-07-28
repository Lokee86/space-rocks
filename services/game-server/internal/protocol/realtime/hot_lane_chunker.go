package realtime

func ExpandRealtimeCandidateChunks(candidates []RealtimeLaneCandidate) ([]RealtimeLaneCandidate, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	expanded := make([]RealtimeLaneCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		chunks, err := expandRealtimeCandidate(candidate)
		if err != nil {
			return nil, err
		}
		expanded = append(expanded, chunks...)
	}
	return expanded, nil
}

func expandRealtimeCandidate(candidate RealtimeLaneCandidate) ([]RealtimeLaneCandidate, error) {
	hardCapChunks, err := expandHardCapCandidate(candidate)
	if err != nil {
		return nil, err
	}
	if candidate.Kind() == RealtimeLaneCandidateKindFull && candidate.Lane() == LaneWorld {
		return hardCapChunks, nil
	}
	lane, kind := candidate.Lane(), candidate.Kind()
	if kind != RealtimeLaneCandidateKindDelta {
		return hardCapChunks, nil
	}

	switch lane {
	case LaneShipsLifecycle, LaneBulletsLifecycle, LaneAsteroidsLifecycle:
		return hardCapChunks, nil
	case LaneShips:
		packet, ok := shipWireDeltaPacketFromCandidate(candidate)
		if !ok {
			return []RealtimeLaneCandidate{candidate}, nil
		}
		packet.Metadata.MatchID = candidate.MatchID

		if shipWireDeltaChunkCount(packet) == 1 {
			return []RealtimeLaneCandidate{normalizedShipWireDeltaCandidate(candidate, packet, packet.ShipUpdates, 0, 1)}, nil
		}

		chunkUpdates := greedyShipWireDeltaChunks(packet)
		if len(chunkUpdates) == 0 {
			return []RealtimeLaneCandidate{candidate}, nil
		}

		chunks := make([]RealtimeLaneCandidate, 0, len(chunkUpdates))
		chunkCount := len(chunkUpdates)
		for index, updates := range chunkUpdates {
			chunks = append(chunks, normalizedShipWireDeltaCandidate(candidate, packet, updates, index, chunkCount))
		}
		return chunks, nil
	case LaneBullets:
		packet, ok := bulletWireDeltaPacketFromCandidate(candidate)
		if !ok {
			return []RealtimeLaneCandidate{candidate}, nil
		}
		packet.Metadata.MatchID = candidate.MatchID

		if bulletWireDeltaChunkCount(packet) == 1 {
			return []RealtimeLaneCandidate{normalizedBulletWireDeltaCandidate(candidate, packet, packet.BulletUpdates, 0, 1)}, nil
		}

		chunkUpdates := greedyBulletWireDeltaChunks(packet)
		if len(chunkUpdates) == 0 {
			return []RealtimeLaneCandidate{candidate}, nil
		}

		chunks := make([]RealtimeLaneCandidate, 0, len(chunkUpdates))
		chunkCount := len(chunkUpdates)
		for index, updates := range chunkUpdates {
			chunks = append(chunks, normalizedBulletWireDeltaCandidate(candidate, packet, updates, index, chunkCount))
		}
		return chunks, nil
	case LaneAsteroids:
		packet, ok := asteroidWireDeltaPacketFromCandidate(candidate)
		if !ok {
			return []RealtimeLaneCandidate{candidate}, nil
		}
		packet.Metadata.MatchID = candidate.MatchID

		if !asteroidWireDeltaRequiresChunking(packet) {
			return []RealtimeLaneCandidate{normalizedAsteroidWireDeltaCandidate(candidate, packet, packet.AsteroidUpdates, 0, 1)}, nil
		}

		chunkUpdates := greedyAsteroidWireDeltaChunks(packet)
		if len(chunkUpdates) == 0 {
			return []RealtimeLaneCandidate{candidate}, nil
		}

		chunks := make([]RealtimeLaneCandidate, 0, len(chunkUpdates))
		chunkCount := len(chunkUpdates)
		for index, updates := range chunkUpdates {
			chunks = append(chunks, normalizedAsteroidWireDeltaCandidate(candidate, packet, updates, index, chunkCount))
		}
		return chunks, nil
	default:
		return []RealtimeLaneCandidate{candidate}, nil
	}
}

func shipWireDeltaChunkCount(packet ShipWireDeltaPacket) int {
	if estimateShipDeltaPacketBytes(packet, packet.ShipUpdates) <= HardCapBytes {
		return 1
	}
	return len(greedyShipWireDeltaChunks(packet))
}

func hotLaneModeForChunkCount(chunkCount int) HotLaneMode {
	switch {
	case chunkCount <= 1:
		return HotLaneModeFullOwned60Hz
	case chunkCount == 2:
		return HotLaneModeFullOwned30Hz
	case chunkCount == 3:
		return HotLaneModeFullOwned20Hz
	default:
		return HotLaneModeFullOwned15Hz
	}
}

func hotLaneModeForShipChunkCount(chunkCount int) HotLaneMode {
	return hotLaneModeForChunkCount(chunkCount)
}

func shipWireDeltaPacketFromCandidate(candidate RealtimeLaneCandidate) (ShipWireDeltaPacket, bool) {
	packet, ok := candidate.Payload.(ShipWireDeltaPacket)
	return packet, ok
}

func normalizedShipWireDeltaCandidate(candidate RealtimeLaneCandidate, packet ShipWireDeltaPacket, updates []map[string]any, chunkIndex int, chunkCount int) RealtimeLaneCandidate {
	packet.ShipUpdates = updates
	packet.Metadata = packet.Metadata.WithChunk(chunkIndex, chunkCount)
	candidate.Payload = packet
	return candidate
}

func greedyShipWireDeltaChunks(packet ShipWireDeltaPacket) [][]map[string]any {
	estimatedPacket := packet
	estimatedPacket.Metadata.ChunkCount = 2
	estimatedPacket.Metadata.IsFinalChunk = true
	return greedyHotLaneUpdateChunks(
		packet.ShipUpdates,
		estimateShipDeltaPacketBytes(estimatedPacket, nil),
		estimateCompactShipMovementUpdateBytes,
	)
}

func bulletWireDeltaChunkCount(packet BulletWireDeltaPacket) int {
	if estimateBulletDeltaPacketBytes(packet, packet.BulletUpdates) <= HardCapBytes {
		return 1
	}
	return len(greedyBulletWireDeltaChunks(packet))
}

func hotLaneModeForBulletChunkCount(chunkCount int) HotLaneMode {
	return hotLaneModeForChunkCount(chunkCount)
}

func bulletWireDeltaPacketFromCandidate(candidate RealtimeLaneCandidate) (BulletWireDeltaPacket, bool) {
	packet, ok := candidate.Payload.(BulletWireDeltaPacket)
	return packet, ok
}

func normalizedBulletWireDeltaCandidate(candidate RealtimeLaneCandidate, packet BulletWireDeltaPacket, updates []map[string]any, chunkIndex int, chunkCount int) RealtimeLaneCandidate {
	packet.BulletUpdates = updates
	packet.Metadata = packet.Metadata.WithChunk(chunkIndex, chunkCount)
	candidate.Payload = packet
	return candidate
}

func greedyBulletWireDeltaChunks(packet BulletWireDeltaPacket) [][]map[string]any {
	estimatedPacket := packet
	estimatedPacket.Metadata.ChunkCount = 2
	estimatedPacket.Metadata.IsFinalChunk = true
	return greedyHotLaneUpdateChunks(
		packet.BulletUpdates,
		estimateBulletDeltaPacketBytes(estimatedPacket, nil),
		estimateCompactBulletMovementUpdateBytes,
	)
}

func asteroidWireDeltaPacketFromCandidate(candidate RealtimeLaneCandidate) (AsteroidWireDeltaPacket, bool) {
	packet, ok := candidate.Payload.(AsteroidWireDeltaPacket)
	return packet, ok
}

func asteroidWireDeltaChunkCount(packet AsteroidWireDeltaPacket) int {
	if estimateAsteroidDeltaPacketBytes(packet, packet.AsteroidUpdates) <= HardCapBytes {
		return 1
	}
	return len(greedyAsteroidWireDeltaChunks(packet))
}

func asteroidWireDeltaRequiresChunking(packet AsteroidWireDeltaPacket) bool {
	return asteroidWireDeltaChunkCount(packet) > 1
}

func normalizedAsteroidWireDeltaCandidate(candidate RealtimeLaneCandidate, packet AsteroidWireDeltaPacket, updates []map[string]any, chunkIndex int, chunkCount int) RealtimeLaneCandidate {
	packet.AsteroidUpdates = updates
	packet.Metadata = packet.Metadata.WithChunk(chunkIndex, chunkCount)
	candidate.Payload = packet
	return candidate
}

func greedyAsteroidWireDeltaChunks(packet AsteroidWireDeltaPacket) [][]map[string]any {
	estimatedPacket := packet
	estimatedPacket.Metadata.ChunkCount = 2
	estimatedPacket.Metadata.IsFinalChunk = true
	return greedyHotLaneUpdateChunks(
		packet.AsteroidUpdates,
		estimateAsteroidDeltaPacketBytes(estimatedPacket, nil),
		estimateCompactAsteroidMovementUpdateBytes,
	)
}

func greedyHotLaneUpdateChunks(updates []map[string]any, emptyPacketBytes int, estimateUpdate func(map[string]any) int) [][]map[string]any {
	if len(updates) == 0 {
		return [][]map[string]any{{}}
	}

	chunks := make([][]map[string]any, 0, 1)
	chunkStart := 0
	chunkBytes := emptyPacketBytes
	chunkRecords := 0
	for index, update := range updates {
		additionalBytes := estimateUpdate(update)
		if chunkRecords > 0 {
			additionalBytes++
		}
		if chunkRecords > 0 && chunkBytes+additionalBytes > HardCapBytes {
			chunks = append(chunks, updates[chunkStart:index])
			chunkStart = index
			chunkBytes = emptyPacketBytes
			chunkRecords = 0
			additionalBytes = estimateUpdate(update)
		}
		chunkBytes += additionalBytes
		chunkRecords++
	}
	chunks = append(chunks, updates[chunkStart:])
	return chunks
}
