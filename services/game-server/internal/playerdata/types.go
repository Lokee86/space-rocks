package playerdata

// Stats is the logical V1.1 player stats contract.
type Stats struct {
	TotalScore  int
	HighScore   int
	ShipDeaths  int
	GamesPlayed int
	// Wins is account/multiplayer-only for V1.1.
	Wins int
}

// PlayerMatchSummary is the logical V1.1 per-player match summary contract.
type PlayerMatchSummary struct {
	GamePlayerID   string
	AccountID      string
	LocalProfileID string
	Score          int
	ShipDeaths     int
	Won            bool
}

// MatchResultSummary is the logical V1.1 match summary contract.
type MatchResultSummary struct {
	MatchID string
	Mode    string
	Players []PlayerMatchSummary
}
