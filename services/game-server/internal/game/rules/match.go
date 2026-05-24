package rules

type PlayerSnapshot struct {
	ID                string
	HasActiveShip     bool
	HasRemainingLives bool
}

type MatchSnapshot struct {
	Players []PlayerSnapshot
}

type MatchDecision struct {
	IsOver bool
}

func EvaluateMatch(snapshot MatchSnapshot) MatchDecision {
	if len(snapshot.Players) == 0 {
		return MatchDecision{IsOver: false}
	}
	for _, player := range snapshot.Players {
		if player.HasRemainingLives {
			return MatchDecision{IsOver: false}
		}
	}
	for _, player := range snapshot.Players {
		if player.HasActiveShip {
			return MatchDecision{IsOver: false}
		}
	}
	return MatchDecision{IsOver: true}
}
