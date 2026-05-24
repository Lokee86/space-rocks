package rules

type PlayerSnapshot struct {
	ID                string
	HasActiveShip     bool
	HasRemainingLives bool
}

type MatchSnapshot struct {
	Players []PlayerSnapshot
}

type PlayerParticipationStatus string

const (
	PlayerActive         PlayerParticipationStatus = "active"
	PlayerPendingRespawn PlayerParticipationStatus = "pending_respawn"
	PlayerEliminated     PlayerParticipationStatus = "eliminated"
)

type PlayerDecision struct {
	ID     string
	Status PlayerParticipationStatus
}

type MatchDecision struct {
	IsOver  bool
	Players []PlayerDecision
}

func EvaluateMatch(snapshot MatchSnapshot) MatchDecision {
	players := make([]PlayerDecision, 0, len(snapshot.Players))
	if len(snapshot.Players) == 0 {
		return MatchDecision{IsOver: false, Players: players}
	}
	isOver := true
	for _, player := range snapshot.Players {
		decision := PlayerDecision{
			ID:     player.ID,
			Status: classifyPlayer(player),
		}
		players = append(players, decision)
		if decision.Status != PlayerEliminated {
			isOver = false
		}
	}
	return MatchDecision{IsOver: isOver, Players: players}
}

func classifyPlayer(player PlayerSnapshot) PlayerParticipationStatus {
	if player.HasActiveShip {
		return PlayerActive
	}
	if player.HasRemainingLives {
		return PlayerPendingRespawn
	}
	return PlayerEliminated
}
