package game

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

type SpawnEntityType string

const (
	SpawnEntityTypePlayer SpawnEntityType = "player"
)

type SpawnReason string

const (
	SpawnReasonInitialPlayer SpawnReason = "initial_player"
	SpawnReasonPlayerRespawn SpawnReason = "player_respawn"
)

type PlayerSpawnPlan struct {
	EntityType SpawnEntityType
	Reason     SpawnReason
	PlayerID   string
	Position   physics.Vector2
}
