package realtime

func deliveryClassForCandidate(candidate RealtimeLaneCandidate) DeliveryClass {
	if candidate.PacketFamily() == PacketFamilyPlayerLocator {
		return DeliveryClassHotSupersedable
	}
	lane, kind := candidate.Lane(), candidate.Kind()
	switch kind {
	case RealtimeLaneCandidateKindEventBatch:
		return DeliveryClassEventOnce
	case RealtimeLaneCandidateKindDelta:
		switch lane {
		case LaneSession, LaneWorld:
			return DeliveryClassRequired
		case LaneOverlay, LaneShips, LaneAsteroids, LaneBullets:
			return DeliveryClassHotSupersedable
		case LaneShipsLifecycle, LaneAsteroidsLifecycle, LaneBulletsLifecycle:
			return DeliveryClassRequired
		}
	default:
		return DeliveryClassRequired
	}

	return DeliveryClassRequired
}

func priorityForCandidate(candidate RealtimeLaneCandidate) Priority {
	if candidate.PacketFamily() == PacketFamilyPlayerLocator {
		return PriorityMedium
	}
	lane, kind := candidate.Lane(), candidate.Kind()
	switch kind {
	case RealtimeLaneCandidateKindEventBatch, RealtimeLaneCandidateKindFull:
		return PriorityCritical
	case RealtimeLaneCandidateKindDelta:
		switch lane {
		case LaneWorld:
			return PriorityCritical
		case LaneSession:
			return PriorityMedium
		case LaneOverlay, LaneShips, LaneAsteroids, LaneBullets:
			return PriorityHigh
		case LaneShipsLifecycle, LaneAsteroidsLifecycle, LaneBulletsLifecycle:
			return PriorityCritical
		}
	}

	return PriorityCritical
}

func scheduleRecordForCandidate(candidateIndex int, candidate RealtimeLaneCandidate) ScheduleRecord {
	packetFamily := candidate.PacketFamily()
	record := ScheduleRecord{
		Lane:           candidate.Lane(),
		CandidateIndex: candidateIndex,
		PacketFamily:   packetFamily,
		RecordKind:     string(candidate.Kind()),
		Priority:       priorityForCandidate(candidate),
		DeliveryClass:  deliveryClassForCandidate(candidate),
		EstimatedBytes: EstimatePacketBytes(packetFamily, 1, 0),
		ChunkCount:     1,
		IsFinalChunk:   true,
		PayloadRef:     candidate.Payload,
	}
	if metadata, ok := candidate.Metadata(); ok {
		record.ChunkIndex = metadata.ChunkIndex
		record.ChunkCount = metadata.ChunkCount
		record.IsFinalChunk = metadata.IsFinalChunk
	}
	return record
}

func CandidateProjection(candidate RealtimeLaneCandidate) (any, bool) {
	if candidate.Kind() == RealtimeLaneCandidateKindEventBatch {
		return nil, false
	}
	if candidate.Projection == nil {
		return nil, false
	}
	return candidate.Projection, true
}
