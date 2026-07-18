package lives

import "github.com/Lokee86/space-rocks/services/game-server/internal/constants"

type Policy struct {
	Model          LifeModel
	StartingLives  int
	RespawnDelay   float64
	RespawnTrigger RespawnTrigger
	Restoration    RestorationPolicy
	SpawnProfileID string
	TeamPool       *TeamPoolPolicy
}

type DeathResult struct {
	RemainingLives  int
	RespawnCooldown float64
}

func NewBaselinePolicy() Policy {
	return Policy{
		Model:          LifeModelFinitePerPlayer,
		StartingLives:  constants.PlayerStartingLives,
		RespawnDelay:   constants.PlayerRespawnDelay,
		RespawnTrigger: RespawnTriggerManual,
		Restoration:    NewBaselineRestorationPolicy(),
		SpawnProfileID: DefaultSpawnProfileID,
	}
}

func (policy Policy) ApplyDeath(currentLives int, infiniteLives bool) DeathResult {
	remainingLives := currentLives
	if policy.Model == LifeModelInfinite && remainingLives <= 0 {
		remainingLives = InfiniteLives
	}
	if policy.Model != LifeModelInfinite && !infiniteLives && remainingLives > 0 {
		remainingLives--
	}

	respawnCooldown := 0.0
	if remainingLives > 0 || policy.Model == LifeModelInfinite {
		respawnCooldown = policy.RespawnDelay
	}

	return DeathResult{
		RemainingLives:  remainingLives,
		RespawnCooldown: respawnCooldown,
	}
}
