package realtime

import (
	"fmt"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func AssembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) (RealtimeLanePlan, error) {
	return assembleRealtimeLaneCandidates(snapshot, state, nil)
}

func assembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, sessionState *RealtimeSessionState) (RealtimeLanePlan, error) {
	candidates := make([]RealtimeLaneCandidate, 0, 4)

	worldCandidates, err := buildWorldLaneCandidates(snapshot, state, sessionState)
	if err != nil {
		return RealtimeLanePlan{}, err
	}
	candidates = append(candidates, worldCandidates...)

	overlayCandidates, err := buildOverlayLaneCandidates(snapshot, state)
	if err != nil {
		return RealtimeLanePlan{}, err
	}
	candidates = append(candidates, overlayCandidates...)

	sessionCandidates, err := buildSessionLaneCandidates(snapshot, state)
	if err != nil {
		return RealtimeLanePlan{}, err
	}
	candidates = append(candidates, sessionCandidates...)
	candidates = append(candidates, buildEventLaneCandidates(snapshot, state)...)
	for index := range candidates {
		candidates[index].MatchID = state.MatchID
	}

	return RealtimeLanePlan{Candidates: candidates}, nil
}

func prepareRealtimeSendPlan(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) (RealtimeSendPrepared, error) {
	state.AdvanceHotLaneTick()
	candidatePlan, err := assembleRealtimeLaneCandidates(snapshot, state, &state)
	if err != nil {
		return RealtimeSendPrepared{}, fmt.Errorf("assemble realtime lane candidates: %w", err)
	}
	candidatePlan.Candidates, err = ExpandRealtimeCandidateChunks(candidatePlan.Candidates)
	if err != nil {
		return RealtimeSendPrepared{}, err
	}

	records := make([]ScheduleRecord, 0, len(candidatePlan.Candidates))
	for i, candidate := range candidatePlan.Candidates {
		records = append(records, scheduleRecordForCandidate(i, candidate))
	}

	return RealtimeSendPrepared{
		CandidatePlan: candidatePlan,
		Records:       records,
		SendPlan:      SelectSendPlan(records),
		SessionState:  state,
	}, nil
}
