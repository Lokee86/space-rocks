package realtime

import (
	"fmt"
	"time"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

type LaneCandidateBuildDurations struct {
	StateAdvance      time.Duration
	WorldHotLifecycle time.Duration
	PlayerLocator     time.Duration
	Overlay           time.Duration
	Session           time.Duration
	Event             time.Duration
	CandidateFinalize time.Duration
}

func AssembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) (RealtimeLanePlan, error) {
	return assembleRealtimeLaneCandidates(snapshot, state, nil)
}

func assembleRealtimeLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, sessionState *RealtimeSessionState) (RealtimeLanePlan, error) {
	plan, _, err := assembleRealtimeLaneCandidatesMeasured(snapshot, state, sessionState, nil)
	return plan, err
}

func assembleRealtimeLaneCandidatesMeasured(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, sessionState *RealtimeSessionState, sharedWorld *WorldWireFullPacket) (RealtimeLanePlan, LaneCandidateBuildDurations, error) {
	candidates := make([]RealtimeLaneCandidate, 0, 4)
	durations := LaneCandidateBuildDurations{}

	started := time.Now()
	worldCandidates, err := buildWorldLaneCandidates(snapshot, state, sessionState, sharedWorld)
	durations.WorldHotLifecycle = time.Since(started)
	if err != nil {
		return RealtimeLanePlan{}, durations, err
	}
	candidates = append(candidates, worldCandidates...)

	started = time.Now()
	candidates = append(candidates, buildPlayerLocatorCandidate(snapshot, state)...)
	durations.PlayerLocator = time.Since(started)

	started = time.Now()
	overlayCandidates, err := buildOverlayLaneCandidates(snapshot, state)
	durations.Overlay = time.Since(started)
	if err != nil {
		return RealtimeLanePlan{}, durations, err
	}
	candidates = append(candidates, overlayCandidates...)

	started = time.Now()
	sessionCandidates, err := buildSessionLaneCandidates(snapshot, state)
	durations.Session = time.Since(started)
	if err != nil {
		return RealtimeLanePlan{}, durations, err
	}
	candidates = append(candidates, sessionCandidates...)

	started = time.Now()
	candidates = append(candidates, buildEventLaneCandidates(snapshot, state)...)
	durations.Event = time.Since(started)

	started = time.Now()
	for index := range candidates {
		candidates[index].MatchID = state.MatchID
	}
	durations.CandidateFinalize = time.Since(started)

	return RealtimeLanePlan{Candidates: candidates}, durations, nil
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
