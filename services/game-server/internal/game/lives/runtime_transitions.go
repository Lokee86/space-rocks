package lives

import (
	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func (runtime *Runtime) ApplyDeath(input DeathInput) DeathFact {
	input = input.normalized()
	playerID := input.PlayerID
	participant, ok := runtime.participants[playerID]
	if !ok {
		return DeathFact{PlayerID: playerID, Input: input.clone(), ReasonCode: "session_missing"}
	}
	if participant.status != playerstate.StatusActive {
		return DeathFact{
			PlayerID:        playerID,
			PreviousStatus:  participant.status,
			ResultingStatus: participant.status,
			RemainingLives:  runtime.effectiveLives(participant),
			RespawnDelay:    participant.respawnCooldown,
			DeathCount:      participant.deathCount,
			ReasonCode:      "not_active",
			Input:           input.clone(),
		}
	}

	previousStatus := participant.status
	deathResult := runtime.consumeLife(participant)
	participant.deathCount++
	participant.respawnCooldown = 0

	if runtime.hasRespawnEligibility(participant) {
		participant.status = playerstate.StatusPendingRespawn
		participant.respawnCooldown = deathResult.RespawnCooldown
	} else {
		participant.status = playerstate.StatusEliminated
		participant.deathExhausted = true
	}

	fact := DeathFact{
		Accepted:        true,
		PlayerID:        playerID,
		PreviousStatus:  previousStatus,
		ResultingStatus: participant.status,
		RemainingLives:  runtime.effectiveLives(participant),
		RespawnDelay:    participant.respawnCooldown,
		DeathCount:      participant.deathCount,
		Input:           input.clone(),
	}
	participant.deathHistory = append(participant.deathHistory, fact.clone())
	return fact
}

func (runtime *Runtime) EvaluateRespawn(playerID string) RespawnFact {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return RejectedRespawn(playerID, "session_missing")
	}

	fact := RespawnFact{
		PlayerID:        playerID,
		PreviousStatus:  participant.status,
		ResultingStatus: participant.status,
		RemainingLives:  runtime.effectiveLives(participant),
		RespawnDelay:    participant.respawnCooldown,
	}
	if participant.status == playerstate.StatusActive {
		fact.ReasonCode = "already_active"
		return fact
	}
	if participant.status != playerstate.StatusPendingRespawn {
		fact.ReasonCode = "invalid_status"
		return fact
	}
	if !runtime.hasRespawnEligibility(participant) || participant.respawnCooldown != 0 {
		fact.ReasonCode = "respawn_cooldown_or_lives_exhausted"
		return fact
	}

	fact.Accepted = true
	fact.ResultingStatus = playerstate.StatusActive
	return fact
}

func (runtime *Runtime) CommitRespawn(playerID string) RespawnFact {
	fact := runtime.EvaluateRespawn(playerID)
	if !fact.Accepted {
		return fact
	}

	participant := runtime.participants[playerID]
	participant.status = playerstate.StatusActive
	participant.respawnCooldown = 0
	participant.respawnCount++
	participant.deathExhausted = false
	fact.ResultingStatus = participant.status
	fact.RespawnDelay = participant.respawnCooldown
	return fact
}

func (runtime *Runtime) RemoveParticipant(playerID string, reasonCode string) RemovalFact {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return RemovalFact{TransitionFact: TransitionFact{PlayerID: playerID, ReasonCode: "session_missing"}}
	}
	if participant.status == StatusRemoved {
		return RemovalFact{TransitionFact: TransitionFact{
			PlayerID:        playerID,
			PreviousStatus:  participant.status,
			ResultingStatus: participant.status,
			ReasonCode:      "already_removed",
		}}
	}

	previousStatus := participant.status
	participant.status = StatusRemoved
	participant.respawnCooldown = 0
	return RemovalFact{TransitionFact: TransitionFact{
		Accepted:        true,
		PlayerID:        playerID,
		PreviousStatus:  previousStatus,
		ResultingStatus: participant.status,
		ReasonCode:      reasonCode,
	}}
}

func (runtime *Runtime) ForceActivateForDevtools(playerID string) TransitionFact {
	participant, ok := runtime.participants[playerID]
	if !ok {
		return TransitionFact{PlayerID: playerID, ReasonCode: "session_missing"}
	}

	previousStatus := participant.status
	participant.status = playerstate.StatusActive
	participant.respawnCooldown = 0
	participant.deathExhausted = false
	return TransitionFact{
		Accepted:        true,
		PlayerID:        playerID,
		PreviousStatus:  previousStatus,
		ResultingStatus: participant.status,
		ReasonCode:      "devtool_force_active",
	}
}

func (runtime *Runtime) consumeLife(participant *participantRecord) DeathResult {
	if participant.infiniteOverride {
		return DeathResult{
			RemainingLives:  runtime.numericLives(participant),
			RespawnCooldown: runtime.policy.RespawnDelay,
		}
	}
	if runtime.policy.Model == LifeModelSharedTeamPool {
		pool := runtime.teamPools[participant.teamID]
		result := runtime.policy.ApplyDeath(pool.RemainingLives, false)
		pool.RemainingLives = result.RemainingLives
		if pool.RemainingLives == 0 {
			runtime.closePendingTeamRespawns(participant.teamID)
		}
		return result
	}
	result := runtime.policy.ApplyDeath(participant.remainingLives, runtime.policy.Model == LifeModelInfinite)
	participant.remainingLives = result.RemainingLives
	return result
}

func (runtime *Runtime) closePendingTeamRespawns(teamID teams.ID) {
	for _, participant := range runtime.participants {
		if participant.teamID != teamID || participant.status != playerstate.StatusPendingRespawn || participant.infiniteOverride {
			continue
		}
		participant.status = playerstate.StatusEliminated
		participant.respawnCooldown = 0
		participant.deathExhausted = true
	}
}

func (runtime *Runtime) effectiveLives(participant *participantRecord) int {
	if runtime.policy.Model == LifeModelInfinite || participant.infiniteOverride {
		return InfiniteLives
	}
	if runtime.policy.Model == LifeModelSharedTeamPool {
		pool := runtime.teamPools[participant.teamID]
		if pool == nil {
			return 0
		}
		return pool.RemainingLives
	}
	return participant.remainingLives
}

func (runtime *Runtime) hasRespawnEligibility(participant *participantRecord) bool {
	return runtime.policy.Model == LifeModelInfinite || participant.infiniteOverride || runtime.effectiveLives(participant) > 0
}
