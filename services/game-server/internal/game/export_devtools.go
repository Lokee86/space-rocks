package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

type DevtoolsStatus struct {
	Invincible    bool
	InfiniteLives bool
	WorldFrozen   bool
	PlayerFrozen  bool
}

func (game *Game) DevtoolsStatusFor(playerID string) DevtoolsStatus {
	status := DevtoolsStatus{
		WorldFrozen: game.worldSimulationOptions.IsWorldFrozen(),
	}

	if session, ok := game.playerSessions[playerID]; ok {
		status.Invincible = session.DamageOptions.Invincible
		status.InfiniteLives = session.LifeOptions.InfiniteLives
		status.PlayerFrozen = session.Suspension.DevFrozen
	}

	if player, ok := game.state.Players[playerID]; ok {
		status.Invincible = player.DamageOptions.Invincible
	}

	return status
}

func (game *Game) DevtoolsWorldFrozen() bool {
	return game.worldSimulationOptions.IsWorldFrozen()
}

func (game *Game) DevtoolsSetWorldFrozen(enabled bool) {
	game.worldSimulationOptions.SetFreezeWorld(enabled)
}

func (game *Game) DevtoolsPlayerInvincible(playerID string) (bool, bool) {
	found := false
	invincible := false

	if session, ok := game.playerSessions[playerID]; ok {
		invincible = session.DamageOptions.Invincible
		found = true
	}

	if player, ok := game.state.Players[playerID]; ok {
		invincible = player.DamageOptions.Invincible
		found = true
	}

	return invincible, found
}

func (game *Game) DevtoolsSetPlayerInvincible(playerID string, enabled bool) bool {
	found := false

	if session, ok := game.playerSessions[playerID]; ok {
		session.DamageOptions.Invincible = enabled
		found = true
	}

	if player, ok := game.state.Players[playerID]; ok {
		player.DamageOptions.Invincible = enabled
		found = true
	}

	return found
}

func (game *Game) DevtoolsInfiniteLives(playerID string) (bool, bool) {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false, false
	}
	return session.LifeOptions.InfiniteLives, true
}

func (game *Game) DevtoolsSetInfiniteLives(playerID string, enabled bool) bool {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	session.LifeOptions.InfiniteLives = enabled
	return true
}

func (game *Game) DevtoolsPlayerFrozen(playerID string) (bool, bool) {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false, false
	}
	return session.Suspension.DevFrozen, true
}

func (game *Game) DevtoolsSetPlayerFrozen(playerID string, enabled bool) bool {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	session.Suspension.SetDevFrozen(enabled)
	if enabled {
		if player, ok := game.state.Players[playerID]; ok {
			player.ClearInput()
		}
	}
	return true
}

func (game *Game) DevtoolsKillPlayer(sourcePlayerID string, targetPlayerID string) bool {
	targetPlayer, ok := game.state.Players[targetPlayerID]
	if !ok || targetPlayer == nil {
		return true
	}
	damageRequest := damage.DamageRequest{
		TargetEntityID:   targetPlayerID,
		TargetEntityType: damage.EntityTypePlayer,
		SourceEntityID:   sourcePlayerID,
		SourceEntityType: damage.EntityTypePlayer,
		CurrentHealth:    targetPlayer.Health,
		Amount:           targetPlayer.Health,
		Type:             damage.DamageTypeDebug,
	}
	damageResult := damage.Resolve(damageRequest)
	targetPlayer.Health = damageResult.RemainingHealth
	if damageResult.Fatal {
		game.applyFatalPlayerDamage(targetPlayerID, targetPlayer)
	}
	return true
}

func (game *Game) DevtoolsRandomUnitVector() physics.Vector2 {
	return game.spawner.RandomUnitVector()
}

func (game *Game) DevtoolsNextBulletID() string {
	return game.spawner.NextBulletID()
}

func (game *Game) DevtoolsAddBullet(bullet *entities.Bullet) bool {
	if bullet == nil {
		return false
	}
	game.state.Projectiles[bullet.ID] = bullet
	return true
}

func (game *Game) DevtoolsRandomAsteroidSpeed() float64 {
	return game.spawner.RandomAsteroidSpeed()
}

func (game *Game) DevtoolsApplyAsteroidSpawnPlan(plan spawning.AsteroidSpawnPlan) *entities.Asteroid {
	return game.applyAsteroidSpawn(plan)
}

func (game *Game) DevtoolsEnsurePlayerSession(playerID string, spawnPosition physics.Vector2) bool {
	return game.ensureDebugPlayerSession(playerID, spawnPosition) != nil
}

func (game *Game) DevtoolsSpawnPlayerShip(playerID string, spawnPosition physics.Vector2) bool {
	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return false
	}
	return game.applyDebugPlayerShip(playerID, session, spawnPosition)
}

func (game *Game) DevtoolsPlayerIDOccupied(playerID string) bool {
	return game.isDebugGameplayPlayerIDOccupied(playerID)
}

func (game *Game) DevtoolsReservePlayerID(playerID string) bool {
	return game.reserveDebugGameplayPlayerID(playerID)
}
