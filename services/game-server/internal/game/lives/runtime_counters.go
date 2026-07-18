package lives

import playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"

func (runtime *Runtime) EffectiveLives(playerID string) (int, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return 0, false
	}
	return runtime.effectiveLives(participant), true
}

// ProjectedLives preserves the numeric lives display for per-player policies
// while exposing the shared team pool for shared-pool policies.
func (runtime *Runtime) ProjectedLives(playerID string) (int, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return 0, false
	}
	if runtime.policy.Model == LifeModelSharedTeamPool {
		return runtime.effectiveLives(participant), true
	}
	return participant.remainingLives, true
}

func (runtime *Runtime) Status(playerID string) (playerstate.Status, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return "", false
	}
	return participant.status, true
}

func (runtime *Runtime) RespawnCooldown(playerID string) (float64, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return 0, false
	}
	return participant.respawnCooldown, true
}

func (runtime *Runtime) DeathCount(playerID string) (int, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return 0, false
	}
	return participant.deathCount, true
}

func (runtime *Runtime) RespawnCount(playerID string) (int, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return 0, false
	}
	return participant.respawnCount, true
}

func (runtime *Runtime) InfiniteOverride(playerID string) (bool, bool) {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return false, false
	}
	return participant.infiniteOverride, true
}

func (runtime *Runtime) SetInfiniteOverride(playerID string, enabled bool) bool {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return false
	}
	participant.infiniteOverride = enabled
	return true
}

func (runtime *Runtime) SetLives(playerID string, lives int) LifeMutation {
	return runtime.mutateLives(playerID, lives, "lives_set")
}

func (runtime *Runtime) AddLives(playerID string, amount int) LifeMutation {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return LifeMutation{PlayerID: playerID, ReasonCode: "session_missing"}
	}
	if runtime.policy.Model == LifeModelInfinite {
		return rejectedNumericLivesMutation(playerID)
	}

	return runtime.mutateLives(playerID, runtime.numericLives(participant)+amount, "lives_added")
}

func (runtime *Runtime) mutateLives(playerID string, lives int, reasonCode string) LifeMutation {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return LifeMutation{PlayerID: playerID, ReasonCode: "session_missing"}
	}
	if runtime.policy.Model == LifeModelInfinite {
		return rejectedNumericLivesMutation(playerID)
	}

	before := runtime.numericLives(participant)
	previousStatus := participant.status
	after := max(0, lives)
	if runtime.policy.Model == LifeModelSharedTeamPool {
		runtime.teamPools[participant.teamID].RemainingLives = after
	} else {
		participant.remainingLives = after
	}
	if after > before && participant.deathExhausted && participant.status == playerstate.StatusEliminated {
		participant.status = playerstate.StatusPendingRespawn
		participant.respawnCooldown = 0
		participant.deathExhausted = false
	}
	return LifeMutation{
		Accepted:        true,
		PlayerID:        playerID,
		PreviousLives:   before,
		CurrentLives:    after,
		Delta:           after - before,
		ReasonCode:      reasonCode,
		PreviousStatus:  previousStatus,
		ResultingStatus: participant.status,
	}
}

func (runtime *Runtime) numericLives(participant *participantRecord) int {
	if runtime.policy.Model == LifeModelSharedTeamPool {
		return runtime.teamPools[participant.teamID].RemainingLives
	}
	return participant.remainingLives
}

func rejectedNumericLivesMutation(playerID string) LifeMutation {
	return LifeMutation{
		PlayerID:   playerID,
		ReasonCode: "numeric_lives_not_applicable",
	}
}
