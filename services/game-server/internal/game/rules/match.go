package rules

import playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"

type PlayerSnapshot struct {
	ID                string
	Status            playerstate.Status
	HasActiveShip     bool
	HasRemainingLives bool
}

type MatchSnapshot struct {
	Players         []PlayerSnapshot
	HadParticipants bool
}

type PlayerDecision struct {
	ID     string
	Status playerstate.Status
}

type MatchDecision struct {
	IsOver  bool
	Players []PlayerDecision
}

func EvaluateMatch(snapshot MatchSnapshot) MatchDecision {
	players := make([]PlayerDecision, 0, len(snapshot.Players))
	if len(snapshot.Players) == 0 {
		return MatchDecision{IsOver: snapshot.HadParticipants, Players: players}
	}
	isOver := true
	for _, player := range snapshot.Players {
		decision := PlayerDecision{
			ID:     player.ID,
			Status: classifyPlayer(player),
		}
		players = append(players, decision)
		if decision.Status != playerstate.StatusEliminated {
			isOver = false
		}
	}
	return MatchDecision{IsOver: isOver, Players: players}
}

func classifyPlayer(player PlayerSnapshot) playerstate.Status {
	return player.Status
}
