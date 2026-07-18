package matchresults

import "github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"

type PresentationPlayer struct {
	GamePlayerID string
	Outcome      Outcome
	Placement    int
	Score        int
	ShipDeaths   int
}

type PresentationSummary struct {
	MatchID        string
	ModeID         string
	TerminalStatus TerminalStatus
	EndReason      string
	Players        []PresentationPlayer
	Objectives     []ObjectiveResolution
}

type ProgressionParticipant struct {
	PlayerRef PlayerRef
	Outcome   Outcome
	Score     int
}

type AchievementFact struct {
	Kind        string
	PlayerRef   PlayerRef
	ObjectiveID string
	Value       float64
}

type DispatchSlices struct {
	Persistence      playerdata.MatchResultSummary
	Presentation     PresentationSummary
	Progression      []ProgressionParticipant
	AchievementFacts []AchievementFact
}

type MatchSummaryDispatcher struct{}

func (MatchSummaryDispatcher) Dispatch(summary MatchSummary) DispatchSlices {
	mode := playerdata.MatchModeMultiplayer
	if summary.Session == SessionSinglePlayer {
		mode = playerdata.MatchModeSinglePlayer
	}
	persistencePlayers := make([]playerdata.PlayerMatchSummary, 0, len(summary.Participants))
	presentationPlayers := make([]PresentationPlayer, 0, len(summary.Participants))
	progression := make([]ProgressionParticipant, 0, len(summary.Participants))
	achievementFacts := []AchievementFact{{Kind: "match_completed"}}
	for _, participant := range summary.Participants {
		persistencePlayers = append(persistencePlayers, playerdata.PlayerMatchSummary{
			GamePlayerID:   participant.PlayerRef.GamePlayerID,
			AccountID:      participant.PlayerRef.AccountID,
			LocalProfileID: participant.PlayerRef.LocalProfileID,
			IsBot:          participant.PlayerRef.IsBot,
			Score:          participant.FinalScore,
			ShipDeaths:     participant.ShipDeaths,
			Won:            participant.Outcome == OutcomeWon,
		})
		presentationPlayers = append(presentationPlayers, PresentationPlayer{
			GamePlayerID: participant.PlayerRef.GamePlayerID, Outcome: participant.Outcome,
			Placement: participant.Placement, Score: participant.FinalScore, ShipDeaths: participant.ShipDeaths,
		})
		progression = append(progression, ProgressionParticipant{
			PlayerRef: participant.PlayerRef, Outcome: participant.Outcome, Score: participant.FinalScore,
		})
		if participant.Outcome == OutcomeWon {
			achievementFacts = append(achievementFacts, AchievementFact{Kind: "match_won", PlayerRef: participant.PlayerRef})
		}
		achievementFacts = append(achievementFacts, AchievementFact{Kind: "score_finalized", PlayerRef: participant.PlayerRef, Value: float64(participant.FinalScore)})
	}
	visibleObjectives := make([]ObjectiveResolution, 0, len(summary.Objectives))
	for _, objective := range summary.Objectives {
		if objective.Discovered {
			visibleObjectives = append(visibleObjectives, objective)
		}
		if objective.Status == "completed" {
			achievementFacts = append(achievementFacts, AchievementFact{Kind: "objective_completed", ObjectiveID: objective.InstanceID})
		}
	}
	return DispatchSlices{
		Persistence:      playerdata.MatchResultSummary{TraceID: summary.TraceID, MatchID: summary.MatchID, Mode: mode, Players: persistencePlayers},
		Presentation:     PresentationSummary{MatchID: summary.MatchID, ModeID: summary.ModeID, TerminalStatus: summary.TerminalStatus, EndReason: summary.EndReason, Players: presentationPlayers, Objectives: visibleObjectives},
		Progression:      progression,
		AchievementFacts: achievementFacts,
	}
}
