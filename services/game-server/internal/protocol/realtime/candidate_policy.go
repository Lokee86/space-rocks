package realtime

func packetFamilyForCandidate(candidate RealtimeLaneCandidate) string {
	switch candidate.Lane {
	case LaneWorld:
		switch candidate.Kind {
		case RealtimeLaneCandidateKindFull:
			return PacketFamilyWorldFull
		case RealtimeLaneCandidateKindDelta:
			return PacketFamilyWorldDelta
		}
	case LaneOverlay:
		switch candidate.Kind {
		case RealtimeLaneCandidateKindFull:
			return PacketFamilyOverlayFull
		case RealtimeLaneCandidateKindDelta:
			return PacketFamilyOverlayDelta
		}
	case LaneSession:
		switch candidate.Kind {
		case RealtimeLaneCandidateKindFull:
			return PacketFamilySessionFull
		case RealtimeLaneCandidateKindDelta:
			return PacketFamilySessionDelta
		}
	case LaneAsteroids:
		if candidate.Kind == RealtimeLaneCandidateKindDelta {
			return PacketFamilyAsteroidDelta
		}
	case LaneAsteroidsLifecycle:
		if candidate.Kind == RealtimeLaneCandidateKindDelta {
			return PacketFamilyAsteroidsLifecycle
		}
	case LaneBullets:
		if candidate.Kind == RealtimeLaneCandidateKindDelta {
			return PacketFamilyBulletDelta
		}
	case LaneBulletsLifecycle:
		if candidate.Kind == RealtimeLaneCandidateKindDelta {
			return PacketFamilyBulletsLifecycle
		}
	case LaneEvent:
		if candidate.Kind == RealtimeLaneCandidateKindEventBatch {
			return PacketFamilyEventBatch
		}
	}

	return ""
}

func deliveryClassForCandidate(candidate RealtimeLaneCandidate) DeliveryClass {
	switch candidate.Kind {
	case RealtimeLaneCandidateKindEventBatch:
		return DeliveryClassEventOnce
	case RealtimeLaneCandidateKindDelta:
		switch candidate.Lane {
		case LaneSession:
			return DeliveryClassDeferrable
		case LaneWorld, LaneOverlay, LaneAsteroids, LaneBullets:
			return DeliveryClassHotSupersedable
		case LaneAsteroidsLifecycle, LaneBulletsLifecycle:
			return DeliveryClassRequired
		}
	default:
		return DeliveryClassRequired
	}

	return DeliveryClassRequired
}

func priorityForCandidate(candidate RealtimeLaneCandidate) Priority {
	switch candidate.Kind {
	case RealtimeLaneCandidateKindEventBatch, RealtimeLaneCandidateKindFull:
		return PriorityCritical
	case RealtimeLaneCandidateKindDelta:
		switch candidate.Lane {
		case LaneSession:
			return PriorityMedium
		case LaneWorld, LaneOverlay, LaneAsteroids, LaneBullets:
			return PriorityHigh
		case LaneAsteroidsLifecycle, LaneBulletsLifecycle:
			return PriorityCritical
		}
	}

	return PriorityCritical
}

func scheduleRecordForCandidate(candidateIndex int, candidate RealtimeLaneCandidate) ScheduleRecord {
	packetFamily := packetFamilyForCandidate(candidate)
	record := ScheduleRecord{
		Lane:           candidate.Lane,
		CandidateIndex: candidateIndex,
		PacketFamily:   packetFamily,
		RecordKind:     string(candidate.Kind),
		Priority:       priorityForCandidate(candidate),
		DeliveryClass:  deliveryClassForCandidate(candidate),
		EstimatedBytes: EstimatePacketBytes(packetFamily, 1, 0),
		ChunkCount:     1,
		IsFinalChunk:   true,
	}

	switch candidate.Kind {
	case RealtimeLaneCandidateKindFull, RealtimeLaneCandidateKindEventBatch:
		record.PayloadRef = candidate.Full
	case RealtimeLaneCandidateKindDelta:
		record.PayloadRef = candidate.Delta
		switch packet := candidate.Delta.(type) {
		case BulletWireDeltaPacket:
			record.ChunkIndex = packet.Metadata.ChunkIndex
			record.ChunkCount = packet.Metadata.ChunkCount
			record.IsFinalChunk = packet.Metadata.IsFinalChunk
		case *BulletWireDeltaPacket:
			if packet != nil {
				record.ChunkIndex = packet.Metadata.ChunkIndex
				record.ChunkCount = packet.Metadata.ChunkCount
				record.IsFinalChunk = packet.Metadata.IsFinalChunk
			}
		case AsteroidWireDeltaPacket:
			record.ChunkIndex = packet.Metadata.ChunkIndex
			record.ChunkCount = packet.Metadata.ChunkCount
			record.IsFinalChunk = packet.Metadata.IsFinalChunk
		case *AsteroidWireDeltaPacket:
			if packet != nil {
				record.ChunkIndex = packet.Metadata.ChunkIndex
				record.ChunkCount = packet.Metadata.ChunkCount
				record.IsFinalChunk = packet.Metadata.IsFinalChunk
			}
		}
	}

	return record
}

func CandidateProjection(candidate RealtimeLaneCandidate) (any, bool) {
	if candidate.Kind == RealtimeLaneCandidateKindEventBatch {
		return nil, false
	}
	if candidate.Projection == nil {
		return nil, false
	}
	return candidate.Projection, true
}