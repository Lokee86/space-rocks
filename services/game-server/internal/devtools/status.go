package devtools

import "github.com/Lokee86/space-rocks/server/internal/game"

func StatusFor(target *game.Game, playerID string) DebugStatus {
	status := target.DevtoolsStatusFor(playerID)
	return DebugStatus{
		Invincible:    status.Invincible,
		InfiniteLives: status.InfiniteLives,
		WorldFrozen:   status.WorldFrozen,
		PlayerFrozen:  status.PlayerFrozen,
	}
}
