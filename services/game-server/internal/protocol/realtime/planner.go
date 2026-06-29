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
	Lane  Lane
	Kind  RealtimeLaneCandidateKind
	Full  any
	Delta any
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
	if !worldReady || !worldSynced || !worldState.IsFinalChunk || worldState.BaselineID == "" {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneWorld,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildWorldFullPacket(snapshot, worldState.Sequence),
		})
	} else {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneWorld,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildWorldFullPacket(snapshot, worldState.Sequence),
		})
	}

	overlayState, overlaySynced := state.LaneState(LaneOverlay)
	overlayReady := state.LaneBaselineReady(LaneOverlay)
	if !overlayReady || !overlaySynced || !overlayState.IsFinalChunk || overlayState.BaselineID == "" {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneOverlay,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildOverlayFullPacket(snapshot, state.ReceiverID, overlayState.Sequence),
		})
	} else {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneOverlay,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildOverlayFullPacket(snapshot, state.ReceiverID, overlayState.Sequence),
		})
	}

	sessionState, sessionSynced := state.LaneState(LaneSession)
	sessionReady := state.LaneBaselineReady(LaneSession)
	if !sessionReady || !sessionSynced || !sessionState.IsFinalChunk || sessionState.BaselineID == "" {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneSession,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildSessionFullPacket(snapshot, sessionState.Sequence),
		})
	} else {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneSession,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildSessionFullPacket(snapshot, sessionState.Sequence),
		})
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
	for _, candidate := range candidatePlan.Candidates {
		records = append(records, scheduleRecordForCandidate(candidate))
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

func scheduleRecordForCandidate(candidate RealtimeLaneCandidate) ScheduleRecord {
	packetFamily := packetFamilyForCandidate(candidate)
	record := ScheduleRecord{
		Lane:           candidate.Lane,
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
