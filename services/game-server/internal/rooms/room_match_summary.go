package rooms

import (
	"github.com/Lokee86/space-rocks/server/internal/playerdata"
)

func (room *Room) buildMatchResultSummaryLocked() (playerdata.MatchResultSummary, bool) {
	if room.match.Game() == nil {
		return playerdata.MatchResultSummary{}, false
	}

	mode := playerdata.MatchModeMultiplayer
	if !room.Joinable {
		mode = playerdata.MatchModeSinglePlayer
	}

	facts := room.match.Game().PlayerMatchFacts()
	players := make([]playerdata.PlayerMatchSummary, 0, len(facts))
	for _, fact := range facts {
		summary := playerdata.PlayerMatchSummary{
			GamePlayerID: fact.GamePlayerID,
			Score:        fact.Score,
			ShipDeaths:   fact.ShipDeaths,
		}

		if member, ok := room.membership.memberByPlayerID(fact.GamePlayerID); ok && member != nil {
			summary.AccountID = member.AccountID
			summary.LocalProfileID = member.LocalProfileID
		}

		players = append(players, summary)
	}

	return playerdata.BuildMatchResultSummary(room.match.CurrentMatchID(), mode, players), true
}
