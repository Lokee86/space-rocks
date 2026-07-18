package matchresults

import "fmt"

type EndOfMatchFlow struct {
	summary *MatchSummary
}

func NewEndOfMatchFlow() *EndOfMatchFlow { return &EndOfMatchFlow{} }

func (flow *EndOfMatchFlow) Run(input BuildInput) (MatchSummary, bool, error) {
	if flow == nil {
		return MatchSummary{}, false, fmt.Errorf("end-of-match flow is required")
	}
	if flow.summary != nil {
		return cloneSummary(*flow.summary), false, nil
	}
	decision, err := ResolveDecision(input)
	if err != nil {
		return MatchSummary{}, false, err
	}
	summary := MatchSummary{
		MatchID: input.MatchID, TraceID: input.TraceID, ModeID: input.ModeID, Session: input.Session,
		TerminalStatus: decision.TerminalStatus, EndReason: decision.EndReason,
		Participants: decision.Participants, Teams: decision.Teams,
		WinningPlayerRefs: decision.WinningPlayerRefs, WinningTeamRefs: decision.WinningTeamRefs,
		Objectives: cloneObjectives(input.Objectives), Missions: cloneMissions(input.Missions),
		Challenges: cloneChallenges(input.Challenges),
	}
	flow.summary = &summary
	return cloneSummary(summary), true, nil
}

func (flow *EndOfMatchFlow) Summary() (MatchSummary, bool) {
	if flow == nil || flow.summary == nil {
		return MatchSummary{}, false
	}
	return cloneSummary(*flow.summary), true
}
