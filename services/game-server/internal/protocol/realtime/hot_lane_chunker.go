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
	case LaneBulletsLifecycle, LaneAsteroidsLifecycle:
		return hardCapChunks, nil
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

func bulletWireDeltaChunkCount(packet BulletWireDeltaPacket) int {
	if estimateBulletDeltaPacketBytes(packet, packet.BulletUpdates) <= HardCapBytes {
		return 1
	}
	return len(greedyBulletWireDeltaChunks(packet))
}

func hotLaneModeForBulletChunkCount(chunkCount int) HotLaneMode {
	switch {
	case chunkCount <= 1:
		return HotLaneModeFullOwned60Hz
	case chunkCount == 2:
		return HotLaneModeFullOwned30Hz
	default:
		return HotLaneModeFullOwned20Hz
	}
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
	updates := packet.BulletUpdates
	if len(updates) == 0 {
		return [][]map[string]any{{}}
	}

	estimatedPacket := packet
	estimatedPacket.Metadata.ChunkCount = 2
	estimatedPacket.Metadata.IsFinalChunk = true

	chunks := make([][]map[string]any, 0, len(updates))
	chunkStart := 0
	for idx := 0; idx < len(updates); idx++ {
		if idx > chunkStart {
			trialUpdates := updates[chunkStart : idx+1]
			if estimateBulletDeltaPacketBytes(estimatedPacket, trialUpdates) > HardCapBytes {
				chunks = append(chunks, updates[chunkStart:idx])
				chunkStart = idx
			}
		}
	}

	chunks = append(chunks, updates[chunkStart:])
	return chunks
}

func asteroidWireDeltaPacketFromCandidate(candidate RealtimeLaneCandidate) (AsteroidWireDeltaPacket, bool) {
	packet, ok := candidate.Payload.(AsteroidWireDeltaPacket)
	return packet, ok
}

func asteroidWireDeltaRequiresChunking(packet AsteroidWireDeltaPacket) bool {
	return estimateAsteroidDeltaPacketBytes(packet, packet.AsteroidUpdates) > HardCapBytes
}

func normalizedAsteroidWireDeltaCandidate(candidate RealtimeLaneCandidate, packet AsteroidWireDeltaPacket, updates []map[string]any, chunkIndex int, chunkCount int) RealtimeLaneCandidate {
	packet.AsteroidUpdates = updates
	packet.Metadata = packet.Metadata.WithChunk(chunkIndex, chunkCount)
	candidate.Payload = packet
	return candidate
}

func greedyAsteroidWireDeltaChunks(packet AsteroidWireDeltaPacket) [][]map[string]any {
	updates := packet.AsteroidUpdates
	if len(updates) == 0 {
		return [][]map[string]any{{}}
	}

	estimatedPacket := packet
	estimatedPacket.Metadata.ChunkCount = 2
	estimatedPacket.Metadata.IsFinalChunk = true

	chunks := make([][]map[string]any, 0, len(updates))
	chunkStart := 0
	for idx := 0; idx < len(updates); idx++ {
		if idx > chunkStart {
			trialUpdates := updates[chunkStart : idx+1]
			if estimateAsteroidDeltaPacketBytes(estimatedPacket, trialUpdates) > HardCapBytes {
				chunks = append(chunks, updates[chunkStart:idx])
				chunkStart = idx
			}
		}
	}

	chunks = append(chunks, updates[chunkStart:])
	return chunks
}
