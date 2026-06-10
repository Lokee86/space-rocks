package playerdata

// BuildMatchResultSummary builds a match summary with resolved winners.
func BuildMatchResultSummary(matchID string, mode MatchMode, players []PlayerMatchSummary) MatchResultSummary {
	return MatchResultSummary{
		MatchID: matchID,
		Mode:    mode,
		Players: ResolveWinners(mode, players),
	}
}
