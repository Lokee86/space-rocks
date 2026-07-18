package lives

import playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"

type RespawnFact struct {
	PlayerID        string
	Accepted        bool
	PreviousStatus  playerstate.Status
	ResultingStatus playerstate.Status
	RemainingLives  int
	RespawnDelay    float64
	ReasonCode      string
}

func (policy Policy) EvaluateRespawn(playerID string, status playerstate.Status, remainingLives int, respawnCooldown float64) RespawnFact {
	fact := RespawnFact{
		PlayerID:        playerID,
		PreviousStatus:  status,
		ResultingStatus: status,
		RemainingLives:  remainingLives,
		RespawnDelay:    respawnCooldown,
	}

	switch {
	case status == playerstate.StatusActive:
		fact.ReasonCode = "already_active"
	case status == playerstate.StatusEliminated || remainingLives <= 0 || respawnCooldown != 0:
		fact.ReasonCode = "respawn_cooldown_or_lives_exhausted"
	case status != playerstate.StatusPendingRespawn:
		fact.ReasonCode = "invalid_status"
	default:
		fact.Accepted = true
		fact.ResultingStatus = playerstate.StatusActive
		fact.ReasonCode = ""
	}

	return fact
}

func RejectedRespawn(playerID string, reasonCode string) RespawnFact {
	return RespawnFact{
		PlayerID:   playerID,
		ReasonCode: reasonCode,
	}
}
