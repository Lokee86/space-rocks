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
	chunks := make([]WorldWireFullPacket, 0, 1)
	current := emptyWorldFullLike(source)

	appendRecord := func(add func(*WorldWireFullPacket)) error {
		trial := current
		add(&trial)
		trial.Metadata = trial.Metadata.WithChunk(0, conservativeChunkCount(worldFullRecordCount(source)))
		encodedBytes, err := candidateEncodedSize(base, trial)
		if err != nil {
			return fmt.Errorf("measure world_full chunk: %w", err)
		}
		if encodedBytes > HardCapBytes {
			if worldFullRecordCount(current) == 0 {
				return fmt.Errorf("world_full record cannot fit within hard cap of %d bytes", HardCapBytes)
			}
			chunks = append(chunks, current)
			current = emptyWorldFullLike(source)
			add(&current)
			encodedBytes, err = candidateEncodedSize(base, current)
			if err != nil {
				return fmt.Errorf("measure world_full single-record chunk: %w", err)
			}
			if encodedBytes > HardCapBytes {
				return fmt.Errorf("world_full record cannot fit within hard cap of %d bytes", HardCapBytes)
			}
		} else {
			current = trial
		}
		return nil
	}

	for _, record := range source.Ships {
		record := record
		if err := appendRecord(func(packet *WorldWireFullPacket) { packet.Ships = append(packet.Ships, record) }); err != nil {
			return nil, err
		}
	}
	for _, record := range source.Bullets {
		record := record
		if err := appendRecord(func(packet *WorldWireFullPacket) { packet.Bullets = append(packet.Bullets, record) }); err != nil {
			return nil, err
		}
	}
	for _, record := range source.Asteroids {
		record := record
		if err := appendRecord(func(packet *WorldWireFullPacket) { packet.Asteroids = append(packet.Asteroids, record) }); err != nil {
			return nil, err
		}
	}
	for _, record := range source.Pickups {
		record := record
		if err := appendRecord(func(packet *WorldWireFullPacket) { packet.Pickups = append(packet.Pickups, record) }); err != nil {
			return nil, err
		}
	}

	if len(chunks) == 0 || worldFullRecordCount(current) > 0 {
		chunks = append(chunks, current)
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

func chunkAsteroidLifecycleCandidate(base RealtimeLaneCandidate, source AsteroidWireDeltaPacket) ([]RealtimeLaneCandidate, error) {
	chunks := make([]AsteroidWireDeltaPacket, 0, 1)
	current := emptyAsteroidLifecycleLike(source)
	appendRecord := func(add func(*AsteroidWireDeltaPacket)) error {
		trial := current
		add(&trial)
		trial.Metadata = trial.Metadata.WithChunk(0, conservativeChunkCount(len(source.AsteroidCreates)+len(source.AsteroidDeletes)))
		encodedBytes, err := candidateEncodedSize(base, trial)

		if err != nil {
			return fmt.Errorf("measure asteroids.lifecycle chunk: %w", err)
		}
		if encodedBytes > HardCapBytes {
			if len(current.AsteroidCreates)+len(current.AsteroidDeletes) == 0 {
				return fmt.Errorf("asteroids.lifecycle record cannot fit within hard cap of %d bytes", HardCapBytes)
			}
			chunks = append(chunks, current)
			current = emptyAsteroidLifecycleLike(source)
			add(&current)
			encodedBytes, err = candidateEncodedSize(base, current)
			if err != nil {
				return fmt.Errorf("measure asteroids.lifecycle single-record chunk: %w", err)
			}
			if encodedBytes > HardCapBytes {
				return fmt.Errorf("asteroids.lifecycle record cannot fit within hard cap of %d bytes", HardCapBytes)
			}
		} else {
			current = trial
		}
		return nil
	}
	for _, record := range source.AsteroidCreates {
		record := record
		if err := appendRecord(func(packet *AsteroidWireDeltaPacket) { packet.AsteroidCreates = append(packet.AsteroidCreates, record) }); err != nil {
			return nil, err
		}
	}
	for _, id := range source.AsteroidDeletes {
		id := id
		if err := appendRecord(func(packet *AsteroidWireDeltaPacket) { packet.AsteroidDeletes = append(packet.AsteroidDeletes, id) }); err != nil {
			return nil, err
		}
	}
	if len(chunks) == 0 || len(current.AsteroidCreates)+len(current.AsteroidDeletes) > 0 {
		chunks = append(chunks, current)
	}
	return normalizeAsteroidLifecycleChunks(base, chunks)
}

func emptyAsteroidLifecycleLike(source AsteroidWireDeltaPacket) AsteroidWireDeltaPacket {
	source.AsteroidCreates = nil
	source.AsteroidDeletes = nil
	return source
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
	chunks := make([]BulletWireDeltaPacket, 0, 1)
	current := source
	current.BulletCreates = nil
	current.BulletDeletes = nil
	appendRecord := func(add func(*BulletWireDeltaPacket)) error {
		trial := current
		add(&trial)
		trial.Metadata = trial.Metadata.WithChunk(0, conservativeChunkCount(len(source.BulletCreates)+len(source.BulletDeletes)))
		encodedBytes, err := candidateEncodedSize(base, trial)
		if err != nil {
			return fmt.Errorf("measure bullets.lifecycle chunk: %w", err)
		}
		if encodedBytes > HardCapBytes {
			if len(current.BulletCreates)+len(current.BulletDeletes) == 0 {
				return fmt.Errorf("bullets.lifecycle record cannot fit within hard cap of %d bytes", HardCapBytes)
			}
			chunks = append(chunks, current)
			current = source
			current.BulletCreates = nil
			current.BulletDeletes = nil
			add(&current)
			encodedBytes, err = candidateEncodedSize(base, current)
			if err != nil {
				return fmt.Errorf("measure bullets.lifecycle single-record chunk: %w", err)
			}
			if encodedBytes > HardCapBytes {
				return fmt.Errorf("bullets.lifecycle record cannot fit within hard cap of %d bytes", HardCapBytes)
			}
		} else {
			current = trial
		}
		return nil
	}
	for _, record := range source.BulletCreates {
		record := record
		if err := appendRecord(func(packet *BulletWireDeltaPacket) { packet.BulletCreates = append(packet.BulletCreates, record) }); err != nil {
			return nil, err
		}
	}
	for _, id := range source.BulletDeletes {
		id := id
		if err := appendRecord(func(packet *BulletWireDeltaPacket) { packet.BulletDeletes = append(packet.BulletDeletes, id) }); err != nil {
			return nil, err
		}
	}
	if len(chunks) == 0 || len(current.BulletCreates)+len(current.BulletDeletes) > 0 {
		chunks = append(chunks, current)
	}
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
