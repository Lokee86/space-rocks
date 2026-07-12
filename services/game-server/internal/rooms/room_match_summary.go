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
		summary := playerdata.PlayerMatchSummary{GamePlayerID: fact.GamePlayerID, Score: fact.Score, ShipDeaths: fact.ShipDeaths}
		if member, ok := capture.Members[fact.GamePlayerID]; ok {
			summary.AccountID = member.AccountID
			summary.LocalProfileID = member.LocalProfileID
		}
		players = append(players, summary)
	}
	return playerdata.BuildMatchResultSummary(capture.MatchID, mode, players)
}
