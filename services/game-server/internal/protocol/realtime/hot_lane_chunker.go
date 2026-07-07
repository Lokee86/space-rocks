package realtime

func ExpandHotLaneCandidateChunks(candidates []RealtimeLaneCandidate) []RealtimeLaneCandidate {
	if len(candidates) == 0 {
		return nil
	}

	expanded := make([]RealtimeLaneCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		chunks := expandHotLaneCandidateChunks(candidate)
		expanded = append(expanded, chunks...)
	}
	return expanded
}

func expandHotLaneCandidateChunks(candidate RealtimeLaneCandidate) []RealtimeLaneCandidate {
	if candidate.Kind != RealtimeLaneCandidateKindDelta {
		return []RealtimeLaneCandidate{candidate}
	}

	switch candidate.Lane {
	case LaneBullets:
		packet, ok := bulletWireDeltaPacketFromCandidate(candidate)
		if !ok {
			return []RealtimeLaneCandidate{candidate}
		}

		if estimateBulletDeltaPacketBytes(packet, packet.BulletUpdates) <= HardCapBytes {
			return []RealtimeLaneCandidate{normalizedBulletWireDeltaCandidate(candidate, packet, packet.BulletUpdates, 0, 1)}
		}

		chunkUpdates := greedyBulletWireDeltaChunks(packet)
		if len(chunkUpdates) == 0 {
			return []RealtimeLaneCandidate{candidate}
		}

		chunks := make([]RealtimeLaneCandidate, 0, len(chunkUpdates))
		chunkCount := len(chunkUpdates)
		for index, updates := range chunkUpdates {
			chunks = append(chunks, normalizedBulletWireDeltaCandidate(candidate, packet, updates, index, chunkCount))
		}
		return chunks
	case LaneAsteroids:
		packet, ok := asteroidWireDeltaPacketFromCandidate(candidate)
		if !ok {
			return []RealtimeLaneCandidate{candidate}
		}

		if estimateAsteroidDeltaPacketBytes(packet, packet.AsteroidUpdates) <= HardCapBytes {
			return []RealtimeLaneCandidate{normalizedAsteroidWireDeltaCandidate(candidate, packet, packet.AsteroidUpdates, 0, 1)}
		}

		chunkUpdates := greedyAsteroidWireDeltaChunks(packet)
		if len(chunkUpdates) == 0 {
			return []RealtimeLaneCandidate{candidate}
		}

		chunks := make([]RealtimeLaneCandidate, 0, len(chunkUpdates))
		chunkCount := len(chunkUpdates)
		for index, updates := range chunkUpdates {
			chunks = append(chunks, normalizedAsteroidWireDeltaCandidate(candidate, packet, updates, index, chunkCount))
		}
		return chunks
	default:
		return []RealtimeLaneCandidate{candidate}
	}
}

func bulletWireDeltaPacketFromCandidate(candidate RealtimeLaneCandidate) (BulletWireDeltaPacket, bool) {
	switch packet := candidate.Delta.(type) {
	case BulletWireDeltaPacket:
		return packet, true
	case *BulletWireDeltaPacket:
		if packet == nil {
			return BulletWireDeltaPacket{}, false
		}
		return *packet, true
	default:
		return BulletWireDeltaPacket{}, false
	}
}

func normalizedBulletWireDeltaCandidate(candidate RealtimeLaneCandidate, packet BulletWireDeltaPacket, updates []map[string]any, chunkIndex int, chunkCount int) RealtimeLaneCandidate {
	packet.BulletUpdates = updates
	packet.Metadata = packet.Metadata.WithChunk(chunkIndex, chunkCount)
	candidate.Delta = packet
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
	switch packet := candidate.Delta.(type) {
	case AsteroidWireDeltaPacket:
		return packet, true
	case *AsteroidWireDeltaPacket:
		if packet == nil {
			return AsteroidWireDeltaPacket{}, false
		}
		return *packet, true
	default:
		return AsteroidWireDeltaPacket{}, false
	}
}

func normalizedAsteroidWireDeltaCandidate(candidate RealtimeLaneCandidate, packet AsteroidWireDeltaPacket, updates []map[string]any, chunkIndex int, chunkCount int) RealtimeLaneCandidate {
	packet.AsteroidUpdates = updates
	packet.Metadata = packet.Metadata.WithChunk(chunkIndex, chunkCount)
	candidate.Delta = packet
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
