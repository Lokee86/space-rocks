package realtime

import "fmt"

func expandHardCapCandidate(candidate RealtimeLaneCandidate) ([]RealtimeLaneCandidate, error) {
	if candidate.Lane() == LaneWorld && candidate.Kind() == RealtimeLaneCandidateKindFull {
		packet, ok := candidate.Payload.(WorldWireFullPacket)
		if !ok {
			return []RealtimeLaneCandidate{candidate}, nil
		}
		return chunkWorldFullCandidate(candidate, packet)
	}

	switch candidate.Lane() {
	case LaneShipsLifecycle:
		packet, ok := candidate.Payload.(ShipWireDeltaPacket)
		if !ok {
			return []RealtimeLaneCandidate{candidate}, nil
		}
		return chunkShipLifecycleCandidate(candidate, packet)
	case LaneAsteroidsLifecycle:
		packet, ok := candidate.Payload.(AsteroidWireDeltaPacket)
		if !ok {
			return []RealtimeLaneCandidate{candidate}, nil
		}
		return chunkAsteroidLifecycleCandidate(candidate, packet)
	case LaneBulletsLifecycle:
		packet, ok := candidate.Payload.(BulletWireDeltaPacket)
		if !ok {
			return []RealtimeLaneCandidate{candidate}, nil
		}
		return chunkBulletLifecycleCandidate(candidate, packet)
	default:
		return []RealtimeLaneCandidate{candidate}, nil
	}
}

func chunkWorldFullCandidate(base RealtimeLaneCandidate, source WorldWireFullPacket) ([]RealtimeLaneCandidate, error) {
	source.Metadata = source.Metadata.WithChunk(0, 1)
	fits, err := candidatePayloadFitsHardCap(base, source)
	if err != nil {
		return nil, fmt.Errorf("measure world_full packet: %w", err)
	}
	if fits {
		return normalizeWorldFullChunks(base, []WorldWireFullPacket{source})
	}

	total := worldFullRecordCount(source)
	chunks := make([]WorldWireFullPacket, 0, 2)
	for start := 0; start < total; {
		end, err := largestHardCapRangeEnd(start, total, func(end int) (int, error) {
			trial := worldFullRecordRange(source, start, end)
			trial.Metadata = trial.Metadata.WithChunk(0, conservativeChunkCount(total))
			return candidateEncodedSize(base, trial)
		})
		if err != nil {
			return nil, fmt.Errorf("measure world_full chunk: %w", err)
		}
		if end == start {
			return nil, fmt.Errorf("world_full record cannot fit within hard cap of %d bytes", HardCapBytes)
		}
		chunks = append(chunks, worldFullRecordRange(source, start, end))
		start = end
	}
	return normalizeWorldFullChunks(base, chunks)
}

func emptyWorldFullLike(source WorldWireFullPacket) WorldWireFullPacket {
	source.Ships = nil
	source.Bullets = nil
	source.Asteroids = nil
	source.Pickups = nil
	return source
}

func worldFullRecordCount(packet WorldWireFullPacket) int {
	return len(packet.Ships) + len(packet.Bullets) + len(packet.Asteroids) + len(packet.Pickups)
}

func worldFullRecordRange(source WorldWireFullPacket, start, end int) WorldWireFullPacket {
	packet := emptyWorldFullLike(source)
	offset := 0
	packet.Ships = recordRange(source.Ships, offset, start, end)
	offset += len(source.Ships)
	packet.Bullets = recordRange(source.Bullets, offset, start, end)
	offset += len(source.Bullets)
	packet.Asteroids = recordRange(source.Asteroids, offset, start, end)
	offset += len(source.Asteroids)
	packet.Pickups = recordRange(source.Pickups, offset, start, end)
	return packet
}

func normalizeWorldFullChunks(base RealtimeLaneCandidate, chunks []WorldWireFullPacket) ([]RealtimeLaneCandidate, error) {
	if len(chunks) == 0 {
		chunks = []WorldWireFullPacket{emptyWorldFullLike(base.Payload.(WorldWireFullPacket))}
	}
	result := make([]RealtimeLaneCandidate, 0, len(chunks))
	for index, packet := range chunks {
		packet.Metadata = packet.Metadata.WithChunk(index, len(chunks))
		candidate := base
		candidate.Payload = packet
		encodedBytes, err := candidateEncodedSize(candidate, packet)
		if err != nil {
			return nil, fmt.Errorf("measure normalized world_full chunk %d: %w", index, err)
		}
		if encodedBytes > HardCapBytes {
			return nil, fmt.Errorf("normalized world_full chunk %d is %d bytes, exceeds hard cap of %d", index, encodedBytes, HardCapBytes)
		}
		result = append(result, candidate)
	}
	return result, nil
}

func chunkShipLifecycleCandidate(base RealtimeLaneCandidate, source ShipWireDeltaPacket) ([]RealtimeLaneCandidate, error) {
	source.Metadata = source.Metadata.WithChunk(0, 1)
	fits, err := candidatePayloadFitsHardCap(base, source)
	if err != nil {
		return nil, fmt.Errorf("measure ships.lifecycle packet: %w", err)
	}
	if fits {
		return normalizeShipLifecycleChunks(base, []ShipWireDeltaPacket{source})
	}

	total := len(source.ShipCreates) + len(source.ShipUpdates) + len(source.ShipDeletes)
	chunks := make([]ShipWireDeltaPacket, 0, 2)
	for start := 0; start < total; {
		end, err := largestHardCapRangeEnd(start, total, func(end int) (int, error) {
			trial := shipLifecycleRecordRange(source, start, end)
			trial.Metadata = trial.Metadata.WithChunk(0, conservativeChunkCount(total))
			return candidateEncodedSize(base, trial)
		})
		if err != nil {
			return nil, fmt.Errorf("measure ships.lifecycle chunk: %w", err)
		}
		if end == start {
			return nil, fmt.Errorf("ships.lifecycle record cannot fit within hard cap of %d bytes", HardCapBytes)
		}
		chunks = append(chunks, shipLifecycleRecordRange(source, start, end))
		start = end
	}
	return normalizeShipLifecycleChunks(base, chunks)
}

func emptyShipLifecycleLike(source ShipWireDeltaPacket) ShipWireDeltaPacket {
	source.ShipCreates = nil
	source.ShipUpdates = nil
	source.ShipDeletes = nil
	return source
}

func shipLifecycleRecordRange(source ShipWireDeltaPacket, start, end int) ShipWireDeltaPacket {
	packet := emptyShipLifecycleLike(source)
	offset := 0
	packet.ShipCreates = recordRange(source.ShipCreates, offset, start, end)
	offset += len(source.ShipCreates)
	packet.ShipUpdates = recordRange(source.ShipUpdates, offset, start, end)
	offset += len(source.ShipUpdates)
	packet.ShipDeletes = recordRange(source.ShipDeletes, offset, start, end)
	return packet
}

func normalizeShipLifecycleChunks(base RealtimeLaneCandidate, chunks []ShipWireDeltaPacket) ([]RealtimeLaneCandidate, error) {
	result := make([]RealtimeLaneCandidate, 0, len(chunks))
	for index, packet := range chunks {
		packet.Metadata = packet.Metadata.WithChunk(index, len(chunks))
		candidate := base
		candidate.Payload = packet
		encodedBytes, err := candidateEncodedSize(candidate, packet)
		if err != nil {
			return nil, fmt.Errorf("measure normalized ships.lifecycle chunk %d: %w", index, err)
		}
		if encodedBytes > HardCapBytes {
			return nil, fmt.Errorf("normalized ships.lifecycle chunk %d exceeds hard cap", index)
		}
		result = append(result, candidate)
	}
	return result, nil
}

func chunkAsteroidLifecycleCandidate(base RealtimeLaneCandidate, source AsteroidWireDeltaPacket) ([]RealtimeLaneCandidate, error) {
	source.Metadata = source.Metadata.WithChunk(0, 1)
	fits, err := candidatePayloadFitsHardCap(base, source)
	if err != nil {
		return nil, fmt.Errorf("measure asteroids.lifecycle packet: %w", err)
	}
	if fits {
		return normalizeAsteroidLifecycleChunks(base, []AsteroidWireDeltaPacket{source})
	}

	total := len(source.AsteroidCreates) + len(source.AsteroidDeletes)
	chunks := make([]AsteroidWireDeltaPacket, 0, 2)
	for start := 0; start < total; {
		end, err := largestHardCapRangeEnd(start, total, func(end int) (int, error) {
			trial := asteroidLifecycleRecordRange(source, start, end)
			trial.Metadata = trial.Metadata.WithChunk(0, conservativeChunkCount(total))
			return candidateEncodedSize(base, trial)
		})
		if err != nil {
			return nil, fmt.Errorf("measure asteroids.lifecycle chunk: %w", err)
		}
		if end == start {
			return nil, fmt.Errorf("asteroids.lifecycle record cannot fit within hard cap of %d bytes", HardCapBytes)
		}
		chunks = append(chunks, asteroidLifecycleRecordRange(source, start, end))
		start = end
	}
	return normalizeAsteroidLifecycleChunks(base, chunks)
}

func emptyAsteroidLifecycleLike(source AsteroidWireDeltaPacket) AsteroidWireDeltaPacket {
	source.AsteroidCreates = nil
	source.AsteroidDeletes = nil
	return source
}

func asteroidLifecycleRecordRange(source AsteroidWireDeltaPacket, start, end int) AsteroidWireDeltaPacket {
	packet := emptyAsteroidLifecycleLike(source)
	offset := 0
	packet.AsteroidCreates = recordRange(source.AsteroidCreates, offset, start, end)
	offset += len(source.AsteroidCreates)
	packet.AsteroidDeletes = recordRange(source.AsteroidDeletes, offset, start, end)
	return packet
}

func normalizeAsteroidLifecycleChunks(base RealtimeLaneCandidate, chunks []AsteroidWireDeltaPacket) ([]RealtimeLaneCandidate, error) {
	result := make([]RealtimeLaneCandidate, 0, len(chunks))
	for index, packet := range chunks {
		packet.Metadata = packet.Metadata.WithChunk(index, len(chunks))
		candidate := base
		candidate.Payload = packet
		encodedBytes, err := candidateEncodedSize(candidate, packet)
		if err != nil {
			return nil, fmt.Errorf("measure normalized asteroids.lifecycle chunk %d: %w", index, err)
		}
		if encodedBytes > HardCapBytes {
			return nil, fmt.Errorf("normalized asteroids.lifecycle chunk %d exceeds hard cap", index)
		}
		result = append(result, candidate)
	}
	return result, nil
}

func chunkBulletLifecycleCandidate(base RealtimeLaneCandidate, source BulletWireDeltaPacket) ([]RealtimeLaneCandidate, error) {
	source.Metadata = source.Metadata.WithChunk(0, 1)
	fits, err := candidatePayloadFitsHardCap(base, source)
	if err != nil {
		return nil, fmt.Errorf("measure bullets.lifecycle packet: %w", err)
	}
	if fits {
		return normalizeBulletLifecycleChunks(base, []BulletWireDeltaPacket{source})
	}

	total := len(source.BulletCreates) + len(source.BulletDeletes)
	chunks := make([]BulletWireDeltaPacket, 0, 2)
	for start := 0; start < total; {
		end, err := largestHardCapRangeEnd(start, total, func(end int) (int, error) {
			trial := bulletLifecycleRecordRange(source, start, end)
			trial.Metadata = trial.Metadata.WithChunk(0, conservativeChunkCount(total))
			return candidateEncodedSize(base, trial)
		})
		if err != nil {
			return nil, fmt.Errorf("measure bullets.lifecycle chunk: %w", err)
		}
		if end == start {
			return nil, fmt.Errorf("bullets.lifecycle record cannot fit within hard cap of %d bytes", HardCapBytes)
		}
		chunks = append(chunks, bulletLifecycleRecordRange(source, start, end))
		start = end
	}
	return normalizeBulletLifecycleChunks(base, chunks)
}

func emptyBulletLifecycleLike(source BulletWireDeltaPacket) BulletWireDeltaPacket {
	source.BulletCreates = nil
	source.BulletDeletes = nil
	return source
}

func bulletLifecycleRecordRange(source BulletWireDeltaPacket, start, end int) BulletWireDeltaPacket {
	packet := emptyBulletLifecycleLike(source)
	offset := 0
	packet.BulletCreates = recordRange(source.BulletCreates, offset, start, end)
	offset += len(source.BulletCreates)
	packet.BulletDeletes = recordRange(source.BulletDeletes, offset, start, end)
	return packet
}

func normalizeBulletLifecycleChunks(base RealtimeLaneCandidate, chunks []BulletWireDeltaPacket) ([]RealtimeLaneCandidate, error) {
	result := make([]RealtimeLaneCandidate, 0, len(chunks))
	for index, packet := range chunks {
		packet.Metadata = packet.Metadata.WithChunk(index, len(chunks))
		candidate := base
		candidate.Payload = packet
		encodedBytes, err := candidateEncodedSize(candidate, packet)
		if err != nil {
			return nil, fmt.Errorf("measure normalized bullets.lifecycle chunk %d: %w", index, err)
		}
		if encodedBytes > HardCapBytes {
			return nil, fmt.Errorf("normalized bullets.lifecycle chunk %d exceeds hard cap", index)
		}
		result = append(result, candidate)
	}
	return result, nil
}

func candidatePayloadFitsHardCap(base RealtimeLaneCandidate, payload RealtimeLanePayload) (bool, error) {
	encodedBytes, err := candidateEncodedSize(base, payload)
	if err != nil {
		return false, err
	}
	return encodedBytes <= HardCapBytes, nil
}

func largestHardCapRangeEnd(start, total int, encodedSize func(end int) (int, error)) (int, error) {
	best := start
	low := start + 1
	high := total
	for low <= high {
		middle := low + (high-low)/2
		bytes, err := encodedSize(middle)
		if err != nil {
			return start, err
		}
		if bytes <= HardCapBytes {
			best = middle
			low = middle + 1
		} else {
			high = middle - 1
		}
	}
	return best, nil
}

func recordRange[T any](records []T, offset, start, end int) []T {
	localStart := start - offset
	if localStart < 0 {
		localStart = 0
	}
	localEnd := end - offset
	if localEnd > len(records) {
		localEnd = len(records)
	}
	if localStart >= localEnd || localStart >= len(records) || localEnd <= 0 {
		return nil
	}
	return records[localStart:localEnd]
}

func conservativeChunkCount(recordCount int) int {
	if recordCount < 2 {
		return 2
	}
	return recordCount
}

func candidateEncodedSize(base RealtimeLaneCandidate, payload RealtimeLanePayload) (int, error) {
	candidate := base
	candidate.Payload = payload
	_, encodedBytes, err := encodeLanePacketUnchecked(candidate)
	return encodedBytes, err
}
