package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
)

type RealtimeLaneCandidateKind string

const (
	RealtimeLaneCandidateKindFull      RealtimeLaneCandidateKind = "full"
	RealtimeLaneCandidateKindDelta     RealtimeLaneCandidateKind = "delta"
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

func AssembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) RealtimeLanePlan {
	candidates := make([]RealtimeLaneCandidate, 0, 4)

	worldState, worldSynced := state.LaneState(LaneWorld)
	if !worldSynced || !worldState.IsFinalChunk || worldState.BaselineID == "" {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneWorld,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildWorldFullPacket(snapshot, worldState.Sequence),
		})
	} else {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:  LaneWorld,
			Kind:  RealtimeLaneCandidateKindDelta,
			Delta: ProjectWorldLane(snapshot),
		})
	}

	overlayState, overlaySynced := state.LaneState(LaneOverlay)
	if !overlaySynced || !overlayState.IsFinalChunk || overlayState.BaselineID == "" {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneOverlay,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildOverlayFullPacket(snapshot, state.ReceiverID, overlayState.Sequence),
		})
	} else {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:  LaneOverlay,
			Kind:  RealtimeLaneCandidateKindDelta,
			Delta: ProjectOverlayLane(snapshot, state.ReceiverID),
		})
	}

	sessionState, sessionSynced := state.LaneState(LaneSession)
	if !sessionSynced || !sessionState.IsFinalChunk || sessionState.BaselineID == "" {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane: LaneSession,
			Kind: RealtimeLaneCandidateKindFull,
			Full: BuildSessionFullPacket(snapshot, sessionState.Sequence),
		})
	} else {
		candidates = append(candidates, RealtimeLaneCandidate{
			Lane:  LaneSession,
			Kind:  RealtimeLaneCandidateKindDelta,
			Delta: ProjectSessionLane(snapshot),
		})
	}

	eventState, _ := state.LaneState(LaneEvent)
	candidates = append(candidates, RealtimeLaneCandidate{
		Lane: LaneEvent,
		Kind: RealtimeLaneCandidateKindEventBatch,
		Full: BuildEventBatchPacket(snapshot.PendingEvents, eventState.Sequence, snapshot.ServerSentMsec),
	})

	return RealtimeLanePlan{Candidates: candidates}
}
