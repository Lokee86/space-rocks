package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
)

type RealtimeLaneCandidateKind string

const (
	RealtimeLaneCandidateKindFull       RealtimeLaneCandidateKind = "full"
	RealtimeLaneCandidateKindDelta      RealtimeLaneCandidateKind = "delta"
	RealtimeLaneCandidateKindEventBatch RealtimeLaneCandidateKind = "event_batch"
)

type RealtimeLaneCandidate struct {
	Lane       Lane
	Kind       RealtimeLaneCandidateKind
	Full       any
	Projection any
	Delta      any
}

type RealtimeLanePlan struct {
	Candidates []RealtimeLaneCandidate
}

type RealtimeSendPrepared struct {
	CandidatePlan RealtimeLanePlan
	Records       []ScheduleRecord
	SendPlan      SendPlan
}

func AssembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) RealtimeLanePlan {
	candidates := make([]RealtimeLaneCandidate, 0, 4)

	worldState, worldSynced := state.LaneState(LaneWorld)
	worldReady := state.LaneBaselineReady(LaneWorld)
	worldSequence := NextLaneSequence(worldState, worldSynced)
	worldFull := BuildWorldFullPacket(snapshot, worldSequence)
	worldProjection, worldHasProjection := state.BaselineProjection(LaneWorld)
	worldCanUseProjection := worldReady && worldSynced && worldState.IsFinalChunk && worldState.BaselineID != "" && worldHasProjection
	if !worldCanUseProjection {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:       LaneWorld,
			Kind:       RealtimeLaneCandidateKindFull,
			Full:       worldFull,
			Projection: worldFull,
		})
	} else {
		previousWorldFull, ok := worldProjection.(WorldFullPacket)
		if !ok {
			candidates = append(candidates, RealtimeLaneCandidate{
				Lane:       LaneWorld,
				Kind:       RealtimeLaneCandidateKindFull,
				Full:       worldFull,
				Projection: worldFull,
			})
		} else if !ProjectionChanged(previousWorldFull, worldFull) {
			// No world candidate when the projection is unchanged.
		} else {
			worldDelta := BuildWorldDeltaPacket(previousWorldFull, worldFull)
			if WorldDeltaHasChanges(worldDelta) {
				candidates = append(candidates, RealtimeLaneCandidate{
					Lane:       LaneWorld,
					Kind:       RealtimeLaneCandidateKindDelta,
					Delta:      worldDelta,
					Projection: worldFull,
				})
			}
		}
	}

	overlayState, overlaySynced := state.LaneState(LaneOverlay)
	overlayReady := state.LaneBaselineReady(LaneOverlay)
	overlaySequence := NextLaneSequence(overlayState, overlaySynced)
	overlayFull := BuildOverlayFullPacket(snapshot, state.ReceiverID, overlaySequence)
	overlayProjection, overlayHasProjection := state.BaselineProjection(LaneOverlay)
	overlayCanUseProjection := overlayReady && overlaySynced && overlayState.IsFinalChunk && overlayState.BaselineID != "" && overlayHasProjection
	if !overlayCanUseProjection {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:       LaneOverlay,
			Kind:       RealtimeLaneCandidateKindFull,
			Full:       overlayFull,
			Projection: overlayFull,
		})
	} else {
		previousOverlayFull, ok := overlayProjection.(OverlayFullPacket)
		if !ok {
			candidates = append(candidates, RealtimeLaneCandidate{
				Lane:       LaneOverlay,
				Kind:       RealtimeLaneCandidateKindFull,
				Full:       overlayFull,
				Projection: overlayFull,
			})
		} else if !ProjectionChanged(previousOverlayFull, overlayFull) {
			// No overlay candidate when the projection is unchanged.
		} else {
			overlayDelta := BuildOverlayDeltaPacket(previousOverlayFull, overlayFull)
			if OverlayDeltaHasChanges(overlayDelta) {
				candidates = append(candidates, RealtimeLaneCandidate{
					Lane:       LaneOverlay,
					Kind:       RealtimeLaneCandidateKindDelta,
					Delta:      overlayDelta,
					Projection: overlayFull,
				})
			}
		}
	}

	sessionState, sessionSynced := state.LaneState(LaneSession)
	sessionReady := state.LaneBaselineReady(LaneSession)
	sessionSequence := NextLaneSequence(sessionState, sessionSynced)
	sessionFull := BuildSessionFullPacket(snapshot, sessionSequence)
	sessionProjection, sessionHasProjection := state.BaselineProjection(LaneSession)
	sessionCanUseProjection := sessionReady && sessionSynced && sessionState.IsFinalChunk && sessionState.BaselineID != "" && sessionHasProjection
	if !sessionCanUseProjection {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:       LaneSession,
			Kind:       RealtimeLaneCandidateKindFull,
			Full:       sessionFull,
			Projection: sessionFull,
		})
	} else {
		previousSessionFull, ok := sessionProjection.(SessionFullPacket)
		if !ok {
			candidates = append(candidates, RealtimeLaneCandidate{
				Lane:       LaneSession,
				Kind:       RealtimeLaneCandidateKindFull,
				Full:       sessionFull,
				Projection: sessionFull,
			})
		} else if !ProjectionChanged(previousSessionFull, sessionFull) {
			// No session candidate when the projection is unchanged.
		} else {
			sessionDelta := BuildSessionDeltaPacket(previousSessionFull, sessionFull)
			if SessionDeltaHasChanges(sessionDelta) {
				candidates = append(candidates, RealtimeLaneCandidate{
					Lane:       LaneSession,
					Kind:       RealtimeLaneCandidateKindDelta,
					Delta:      sessionDelta,
					Projection: sessionFull,
				})
			}
		}
	}

	if len(snapshot.PendingEvents) > 0 {
		eventState, _ := state.LaneState(LaneEvent)
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneEvent,
			Kind: RealtimeLaneCandidateKindEventBatch,
			Full: BuildEventBatchPacket(snapshot.PendingEvents, eventState.Sequence, snapshot.ServerSentMsec),
		})
	}

	return RealtimeLanePlan{Candidates: candidates}
}

func prepareRealtimeSendPlan(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) RealtimeSendPrepared {
	candidatePlan := AssembleRealtimeLaneCandidates(snapshot, state)

	records := make([]ScheduleRecord, 0, len(candidatePlan.Candidates))
	for i, candidate := range candidatePlan.Candidates {
		records = append(records, scheduleRecordForCandidate(i, candidate))
	}

	return RealtimeSendPrepared{
		CandidatePlan: candidatePlan,
		Records:       records,
		SendPlan:      SelectSendPlan(records),
	}
}

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
		case LaneWorld, LaneOverlay:
			return DeliveryClassHotSupersedable
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
		case LaneWorld, LaneOverlay:
			return PriorityHigh
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
