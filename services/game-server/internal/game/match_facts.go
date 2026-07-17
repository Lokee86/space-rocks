package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"

// PlayerMatchFact is the game-owned match fact used to derive playerdata summaries.
type PlayerMatchFact struct {
	GamePlayerID string
	TeamID       teams.ID
	Score        int
	ShipDeaths   int
}

type participantRecord struct {
	ID         string
	TeamID     teams.ID
	Score      int
	ShipDeaths int
}
