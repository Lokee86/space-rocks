package game

// PlayerMatchFact is the game-owned match fact used to derive playerdata summaries.
type PlayerMatchFact struct {
	GamePlayerID string
	Score        int
	ShipDeaths   int
}

type participantRecord struct {
	ID         string
	Score      int
	ShipDeaths int
}
