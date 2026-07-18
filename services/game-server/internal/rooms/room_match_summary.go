package rooms

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/matchresults"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

func buildEndOfMatchInput(capture gameOverCapture, state game.FinalMatchState, endReason string) matchresults.BuildInput {
	session := matchresults.SessionMultiplayer
	if !capture.Joinable {
		session = matchresults.SessionSinglePlayer
	}
	modeID := state.ModeID
	if modeID == "" {
		modeID = "arcade_survival"
	}
	structure := state.TeamStructure
	if structure == "" {
		structure = teams.StructureFFA
	}
	participants := make([]matchresults.ParticipantFact, 0, len(state.Players))
	for _, fact := range state.Players {
		identity := capture.ParticipantIdentities[fact.GamePlayerID]
		disposition := matchresults.DispositionParticipated
		if _, active := capture.Members[fact.GamePlayerID]; !active {
			disposition = matchresults.DispositionDeparted
		}
		participants = append(participants, matchresults.ParticipantFact{
			PlayerRef: matchresults.PlayerRef{
				GamePlayerID: fact.GamePlayerID, AccountID: identity.AccountID, LocalProfileID: identity.LocalProfileID, IsBot: identity.IsBot,
			},
			TeamID: fact.TeamID, Score: fact.Score, ShipDeaths: fact.ShipDeaths, Disposition: disposition,
		})
	}
	objectiveResolutions := make([]matchresults.ObjectiveResolution, 0, len(state.Objectives))
	for _, snapshot := range state.Objectives {
		objectiveResolutions = append(objectiveResolutions, matchresults.ObjectiveResolution{
			DefinitionID: string(snapshot.DefinitionID), InstanceID: string(snapshot.InstanceID), Scope: string(snapshot.Scope),
			OwnerID: snapshot.OwnerID, Status: string(snapshot.Status), Progress: snapshot.Progress, Target: snapshot.Target,
			FailureReason: snapshot.FailureReason, Discovered: snapshot.Discovered,
		})
	}
	return matchresults.BuildInput{
		MatchID: capture.MatchID, TraceID: capture.TraceID, ModeID: modeID, Session: session,
		TeamStructure: structure, LockedDecision: state.Decision, EndReason: endReason,
		Participants: participants, Objectives: objectiveResolutions,
	}
}

// buildMatchResultSummary preserves the legacy helper through the new summary/dispatcher path.
func buildMatchResultSummary(capture gameOverCapture, facts []game.PlayerMatchFact) playerdata.MatchResultSummary {
	input := buildEndOfMatchInput(capture, game.FinalMatchState{ModeID: "arcade_survival", TeamStructure: teams.StructureFFA, Players: facts}, "simulation_complete")
	flow := matchresults.NewEndOfMatchFlow()
	summary, _, err := flow.Run(input)
	if err != nil {
		return playerdata.MatchResultSummary{TraceID: capture.TraceID, MatchID: capture.MatchID}
	}
	return (matchresults.MatchSummaryDispatcher{}).Dispatch(summary).Persistence
}
