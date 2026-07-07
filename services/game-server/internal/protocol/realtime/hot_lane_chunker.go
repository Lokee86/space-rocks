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

		if fitsBulletWireDeltaCandidate(candidate, packet.Metadata.WithChunk(0, 999), packet.BulletUpdates) {
			return []RealtimeLaneCandidate{normalizedBulletWireDeltaCandidate(candidate, packet, packet.BulletUpdates, 0, 1)}
		}

		chunkUpdates := greedyBulletWireDeltaChunks(candidate, packet)
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

		if fitsAsteroidWireDeltaCandidate(candidate, packet.Metadata.WithChunk(0, 999), packet.AsteroidUpdates) {
			return []RealtimeLaneCandidate{normalizedAsteroidWireDeltaCandidate(candidate, packet, packet.AsteroidUpdates, 0, 1)}
		}

		chunkUpdates := greedyAsteroidWireDeltaChunks(candidate, packet)
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

func fitsBulletWireDeltaCandidate(candidate RealtimeLaneCandidate, metadata Metadata, updates []map[string]any) bool {
	packet, ok := bulletWireDeltaPacketFromCandidate(candidate)
	if !ok {
		return false
	}
	packet.Metadata = metadata
	packet.BulletUpdates = updates
	trial := candidate
	trial.Delta = packet
	_, recordedBytes := encodeLanePacketUnchecked(trial)
	return recordedBytes > 0 && recordedBytes <= HardCapBytes
}

func greedyBulletWireDeltaChunks(candidate RealtimeLaneCandidate, packet BulletWireDeltaPacket) [][]map[string]any {
	updates := packet.BulletUpdates
	if len(updates) == 0 {
		return [][]map[string]any{{}}
	}

	chunks := make([][]map[string]any, 0, len(updates))
	for start := 0; start < len(updates); {
		current := make([]map[string]any, 0, len(updates)-start)
		end := start
		for end < len(updates) {
			trialUpdates := append(append([]map[string]any(nil), current...), updates[end])
			if fitsBulletWireDeltaCandidate(candidate, packet.Metadata.WithChunk(0, 999), trialUpdates) {
				current = trialUpdates
				end++
				continue
			}
			if len(current) > 0 {
				break
			}
			current = trialUpdates
			end++
			break
		}
		chunks = append(chunks, current)
		start = end
	}
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

func fitsAsteroidWireDeltaCandidate(candidate RealtimeLaneCandidate, metadata Metadata, updates []map[string]any) bool {
	packet, ok := asteroidWireDeltaPacketFromCandidate(candidate)
	if !ok {
		return false
	}
	packet.Metadata = metadata
	packet.AsteroidUpdates = updates
	trial := candidate
	trial.Delta = packet
	_, recordedBytes := encodeLanePacketUnchecked(trial)
	return recordedBytes > 0 && recordedBytes <= HardCapBytes
}

func greedyAsteroidWireDeltaChunks(candidate RealtimeLaneCandidate, packet AsteroidWireDeltaPacket) [][]map[string]any {
	updates := packet.AsteroidUpdates
	if len(updates) == 0 {
		return [][]map[string]any{{}}
	}

	chunks := make([][]map[string]any, 0, len(updates))
	for start := 0; start < len(updates); {
		current := make([]map[string]any, 0, len(updates)-start)
		end := start
		for end < len(updates) {
			trialUpdates := append(append([]map[string]any(nil), current...), updates[end])
			if fitsAsteroidWireDeltaCandidate(candidate, packet.Metadata.WithChunk(0, 999), trialUpdates) {
				current = trialUpdates
				end++
				continue
			}
			if len(current) > 0 {
				break
			}
			current = trialUpdates
			end++
			break
		}
		chunks = append(chunks, current)
		start = end
	}
	return chunks
}
