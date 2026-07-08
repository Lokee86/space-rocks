package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
)

func AssembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) RealtimeLanePlan {
	return assembleRealtimeLaneCandidates(snapshot, state, nil)
}

func assembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, sessionState *RealtimeSessionState) RealtimeLanePlan {
	candidates := make([]RealtimeLaneCandidate, 0, 4)

	candidates = append(candidates, buildWorldLaneCandidates(snapshot, state, sessionState)...)

	candidates = append(candidates, buildOverlayLaneCandidates(snapshot, state)...)
	candidates = append(candidates, buildSessionLaneCandidates(snapshot, state)...)
	candidates = append(candidates, buildEventLaneCandidates(snapshot, state)...)

	return RealtimeLanePlan{Candidates: candidates}
}

func prepareRealtimeSendPlan(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) RealtimeSendPrepared {
	candidatePlan := assembleRealtimeLaneCandidates(snapshot, state, &state)
	candidatePlan.Candidates = ExpandHotLaneCandidateChunks(candidatePlan.Candidates)

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