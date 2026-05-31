package devtools

import "github.com/Lokee86/space-rocks/server/internal/game"

func StatusFor(target *game.Game, playerID string) DebugStatus {
	status := target.DevtoolsStatusFor(playerID)
	return DebugStatus{
		Invincible:       status.Invincible,
		InfiniteLives:    status.InfiniteLives,
		WorldFrozen:      status.WorldFrozen,
		AsteroidsFrozen:  status.AsteroidsFrozen,
		BulletsFrozen:    status.BulletsFrozen,
		SpawningFrozen:   status.SpawningFrozen,
		CollisionsFrozen: status.CollisionsFrozen,
		PlayerFrozen:     status.PlayerFrozen,
	}
}

func StatusesForAllPlayers(target *game.Game) map[string]DebugStatus {
	statuses := make(map[string]DebugStatus)
	for _, player := range target.MatchDecision().Players {
		statuses[player.ID] = StatusFor(target, player.ID)
	}
	return statuses
}
