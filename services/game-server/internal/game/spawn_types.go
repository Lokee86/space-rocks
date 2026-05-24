package game

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

type SpawnEntityType string

const (
	SpawnEntityTypeAsteroid SpawnEntityType = "asteroid"
	SpawnEntityTypePlayer   SpawnEntityType = "player"
)

type SpawnReason string

const (
	SpawnReasonTimedAsteroid    SpawnReason = "timed_asteroid"
	SpawnReasonAsteroidFragment SpawnReason = "asteroid_fragment"
	SpawnReasonInitialPlayer    SpawnReason = "initial_player"
	SpawnReasonPlayerRespawn    SpawnReason = "player_respawn"
)

type AsteroidSpawnPlan struct {
	EntityType SpawnEntityType
	Reason     SpawnReason
	Position   physics.Vector2
	Velocity   physics.Vector2
	Size       int
	Variant    int
}
