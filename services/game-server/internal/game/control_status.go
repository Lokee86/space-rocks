package game

import "sort"

func (target *Control) TargetPlayerIDs() []string {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	playerIDs := make(map[string]struct{}, len(target.game.playerSessions)+len(target.game.entities.Players))
	for playerID := range target.game.playerSessions {
		if playerID == "" {
			continue
		}
		playerIDs[playerID] = struct{}{}
	}
	for playerID := range target.game.entities.Players {
		if playerID == "" {
			continue
		}
		playerIDs[playerID] = struct{}{}
	}

	ids := make([]string, 0, len(playerIDs))
	for playerID := range playerIDs {
		ids = append(ids, playerID)
	}
	sort.Strings(ids)
	return ids
}

func (target *Control) WorldFrozen() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.worldSimulationOptions.IsWorldFrozen()
}

func (target *Control) AsteroidsFrozen() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return !target.game.worldSimulationOptions.AsteroidsCanMove()
}

func (target *Control) BulletsFrozen() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return !target.game.worldSimulationOptions.BulletsCanMove()
}

func (target *Control) SpawningFrozen() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return !target.game.worldSimulationOptions.CanSpawnAsteroids()
}

func (target *Control) CollisionsFrozen() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return !target.game.worldSimulationOptions.CanRunCollisions()
}
