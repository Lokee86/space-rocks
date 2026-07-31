package rules

type PlayerSnapshot struct {
	ID                string
	HasActiveShip     bool
	HasRemainingLives bool
}

type MatchSnapshot struct {
	Players         []PlayerSnapshot
	HadParticipants bool
}

type PlayerParticipationStatus string

const (
	PlayerActive         PlayerParticipationStatus = "active"
	PlayerPendingRespawn PlayerParticipationStatus = "pending_respawn"
	PlayerEliminated     PlayerParticipationStatus = "eliminated"
)

type TerminalStatus string

const (
	TerminalCompleted                  TerminalStatus = "completed"
	TerminalFailed                     TerminalStatus = "failed"
	TerminalCancelled                  TerminalStatus = "cancelled"
	TerminalInvalid                    TerminalStatus = "invalid"
	TerminalAdministrativelyTerminated TerminalStatus = "administratively_terminated"
)

type Outcome string

const (
	OutcomeWon       Outcome = "won"
	OutcomeLost      Outcome = "lost"
	OutcomeDraw      Outcome = "draw"
	OutcomeCompleted Outcome = "completed"
	OutcomeFailed    Outcome = "failed"
	OutcomeAborted   Outcome = "aborted"
)

type PlayerDecision struct {
	ID             string
	Status         PlayerParticipationStatus
	Outcome        Outcome
	Placement      int
	CompletionTime float64
	TargetValue    float64
}

type MatchDecision struct {
	IsOver           bool
	TerminalStatus   TerminalStatus
	EndReason        string
	Players          []PlayerDecision
	WinningPlayerIDs []string
}

func EvaluateMatch(snapshot MatchSnapshot) MatchDecision {
	players := make([]PlayerDecision, 0, len(snapshot.Players))
	if len(snapshot.Players) == 0 {
		decision := MatchDecision{IsOver: snapshot.HadParticipants, Players: players}
		if decision.IsOver {
			decision.TerminalStatus = TerminalCompleted
			decision.EndReason = "no_active_participants"
		}
		return decision
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
	decision := MatchDecision{IsOver: isOver, Players: players}
	if decision.IsOver {
		decision.TerminalStatus = TerminalCompleted
		decision.EndReason = "no_active_participants"
	}
	return decision
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
