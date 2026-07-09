package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/rules"
	runtimepkg "github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

type StatusTarget interface {
	TargetPlayerIDs() []string
	WorldFrozen() bool
	AsteroidsFrozen() bool
	BulletsFrozen() bool
	SpawningFrozen() bool
	CollisionsFrozen() bool
	PlayerInvincible(playerID string) (bool, bool)
	InfiniteLives(playerID string) (bool, bool)
	PlayerFrozen(playerID string) (bool, bool)
	MatchDecision() rules.MatchDecision
}

type ToggleTarget interface {
	WorldFrozen() bool
	SetWorldFrozen(enabled bool)
	ToggleFreezeWorld() bool
	ToggleFreezeAsteroids() bool
	ToggleFreezeBullets() bool
	ToggleFreezeSpawning() bool
	ToggleFreezeCollisions() bool
	PlayerInvincible(playerID string) (bool, bool)
	SetPlayerInvincible(playerID string, enabled bool) bool
	InfiniteLives(playerID string) (bool, bool)
	SetInfiniteLives(playerID string, enabled bool) bool
	PlayerFrozen(playerID string) (bool, bool)
	SetPlayerFrozen(playerID string, enabled bool) bool
	ApplyPlayerDefeat(sourcePlayerID string, targetPlayerID string) bool
}

type PlayerSpawnTarget interface {
	EnsurePlayerSession(playerID string, spawnPosition physics.Vector2) bool
	SpawnPlayerShip(playerID string, spawnPosition physics.Vector2, cameraConfig runtimepkg.ClientConfig) bool
	PlayerIDOccupied(playerID string) bool
	ReservePlayerID(playerID string) bool
}

type SpawnTarget interface {
	RandomUnitVector() physics.Vector2
	NextBulletID() string
	AddBullet(bullet *runtimepkg.Bullet) bool
	SpawnBullet(ownerID string, position physics.Vector2, direction physics.Vector2) (*runtimepkg.Bullet, bool)
	SpawnPickup(pickupType pickups.PickupType, position physics.Vector2) (*pickups.Pickup, bool, error)
	RandomAsteroidSpeed() float64
	ApplyAsteroidSpawnPlan(plan spawning.AsteroidSpawnPlan) *runtimepkg.Asteroid
}

type RespawnTarget interface {
	SafeRespawnPosition(playerID string) (physics.Vector2, bool)
	ForceRespawnPlayer(playerID string, position physics.Vector2, cameraConfig runtimepkg.ClientConfig) bool
}

type CounterTarget interface {
	SetPlayerScore(playerID string, score int) bool
	AddPlayerScore(playerID string, amount int) bool
	SetPlayerLives(playerID string, lives int) bool
	AddPlayerLives(playerID string, amount int) bool
}

type ClearTarget interface {
	ClearBullets() int
	ClearAsteroids() int
}

type StreamTarget interface {
	ObserverKey() any
	BulletsCanMove() bool
	SpawnDebugBullet(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool
	RegisterSimulationStepObserver(observer func(float64))
}

type CollisionTelemetryTarget interface {
	CollisionBodiesByKind() map[string][]physics.CollisionBody
}

type Target interface {
	StatusTarget
	ToggleTarget
	PlayerSpawnTarget
	SpawnTarget
	RespawnTarget
	CounterTarget
	ClearTarget
	StreamTarget
	CollisionTelemetryTarget
}
