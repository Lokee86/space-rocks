package lives

import "fmt"

func ValidatePolicy(policy Policy) error {
	return policy.Validate()
}

func ValidateRestorationPolicy(restoration RestorationPolicy) error {
	return restoration.Validate()
}

func ValidateTeamPoolPolicy(teamPool TeamPoolPolicy) error {
	return teamPool.Validate()
}

func (policy Policy) Validate() error {
	if policy.Model != LifeModelFinitePerPlayer && policy.Model != LifeModelSharedTeamPool && policy.Model != LifeModelInfinite {
		return fmt.Errorf("unknown life model %q", policy.Model)
	}
	if policy.StartingLives < 0 {
		return fmt.Errorf("starting lives must be nonnegative")
	}
	if policy.RespawnDelay < 0 {
		return fmt.Errorf("respawn delay must be nonnegative")
	}
	if policy.RespawnTrigger != RespawnTriggerManual {
		return fmt.Errorf("respawn trigger %q is not implemented", policy.RespawnTrigger)
	}
	if policy.SpawnProfileID == "" {
		return fmt.Errorf("spawn profile ID is required")
	}
	if policy.SpawnProfileID != DefaultSpawnProfileID {
		return fmt.Errorf("unsupported spawn profile ID %q", policy.SpawnProfileID)
	}
	if err := policy.Restoration.Validate(); err != nil {
		return err
	}

	switch policy.Model {
	case LifeModelFinitePerPlayer:
		if policy.StartingLives == 0 {
			return fmt.Errorf("finite per-player lives require starting lives")
		}
		if policy.TeamPool != nil {
			return fmt.Errorf("finite per-player lives do not accept a team pool")
		}
	case LifeModelSharedTeamPool:
		if policy.TeamPool == nil {
			return fmt.Errorf("shared team-pool lives require a team-pool policy")
		}
		if policy.StartingLives != 0 {
			return fmt.Errorf("shared team-pool lives keep starting lives on the team-pool policy")
		}
		if err := policy.TeamPool.Validate(); err != nil {
			return err
		}
	case LifeModelInfinite:
		if policy.StartingLives != 0 {
			return fmt.Errorf("infinite lives do not accept starting lives")
		}
		if policy.TeamPool != nil {
			return fmt.Errorf("infinite lives do not accept a team pool")
		}
	}
	return nil
}

func (restoration RestorationPolicy) Validate() error {
	if restoration.Health != RestorationNone && restoration.Health != RestorationFull {
		return fmt.Errorf("unknown health restoration mode %q", restoration.Health)
	}
	if restoration.Shields != RestorationNone && restoration.Shields != RestorationFull {
		return fmt.Errorf("unknown shields restoration mode %q", restoration.Shields)
	}
	if restoration.ShortCooldownThreshold < 0 {
		return fmt.Errorf("short cooldown threshold must be nonnegative")
	}
	if restoration.TemporaryEffects != TemporaryEffectsRemove && restoration.TemporaryEffects != TemporaryEffectsPersist {
		return fmt.Errorf("unknown temporary-effects policy %q", restoration.TemporaryEffects)
	}
	if restoration.Loadout != LoadoutPersist && restoration.Loadout != LoadoutReset {
		return fmt.Errorf("unknown loadout persistence %q", restoration.Loadout)
	}
	return nil
}

func (teamPool TeamPoolPolicy) Validate() error {
	if teamPool.PoolID == "" {
		return fmt.Errorf("team-pool ID is required")
	}
	if teamPool.StartingLives < 0 {
		return fmt.Errorf("team-pool starting lives must be nonnegative")
	}
	if teamPool.StartingLives == 0 {
		return fmt.Errorf("team-pool lives require starting lives")
	}
	return nil
}
