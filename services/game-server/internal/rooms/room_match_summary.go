package rooms

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

func buildMatchResultSummary(capture gameOverCapture, facts []game.PlayerMatchFact) playerdata.MatchResultSummary {
	mode := playerdata.MatchModeMultiplayer
	if !capture.Joinable {
		mode = playerdata.MatchModeSinglePlayer
	}
	players := make([]playerdata.PlayerMatchSummary, 0, len(facts))
	for _, fact := range facts {
		summary := playerdata.PlayerMatchSummary{GamePlayerID: fact.GamePlayerID, TeamID: string(fact.TeamID), Score: fact.Score, ShipDeaths: fact.ShipDeaths}
		if identity, ok := capture.ParticipantIdentities[fact.GamePlayerID]; ok {
			summary.AccountID = identity.AccountID
			summary.LocalProfileID = identity.LocalProfileID
			summary.IsBot = identity.IsBot
		}
		players = append(players, summary)
	}
	summary := playerdata.BuildMatchResultSummary(capture.MatchID, mode, players)
	if mode == playerdata.MatchModeMultiplayer && capture.Game != nil && capture.Game.ResolvedMatchRules().TeamScoreEnabled {
		winningPlayers := make(map[string]struct{})
		for _, playerID := range capture.Game.MatchDecision().WinningPlayerIDs {
			winningPlayers[playerID] = struct{}{}
		}
		for index := range summary.Players {
			_, summary.Players[index].Won = winningPlayers[summary.Players[index].GamePlayerID]
		}
	}
	summary.TraceID = capture.TraceID
	return summary
}
