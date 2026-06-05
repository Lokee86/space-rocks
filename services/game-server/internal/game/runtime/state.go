package runtime

import (
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

const DefaultShipTypeID = "v_wing"

type Ship struct {
	ID                       string
	ShipTypeID               string
	Stats                    ShipStats
	X                        float64
	Y                        float64
	Rotation                 float64
	Velocity                 physics.Vector2
	Input                    InputState
	Config                   ClientConfig
	ShootCooldown            float64
	TargetKind               string
	TargetID                 string
	Health                   int
	Shields                  int
	DamageOptions            DamageOptions
	InvulnerabilityRemaining float64
	PendingDespawn           bool
	DespawnDelay             float64
}

type CameraView struct {
	X      float64
	Y      float64
	Config ClientConfig
}

type Bullet struct {
	ID             string
	OwnerID        string
	X              float64
	Y              float64
	Rotation       float64
	Velocity       physics.Vector2
	Life           float64
	Damage         int
	PendingDespawn bool
	DespawnDelay   float64
}

type Asteroid struct {
	ID             string
	X              float64
	Y              float64
	Velocity       physics.Vector2
	Size           int
	Variant        int
	Health         int
	CollisionDamage int
	PendingDespawn bool
	DespawnDelay   float64
}

type EntityStore struct {
	Players     map[string]*Ship
	Projectiles map[string]*Bullet
	Asteroids   map[string]*Asteroid
	Enemies     map[string]*Ship
	Pickups     map[string]*pickups.Pickup
}

func NewEntityStore() EntityStore {
	return EntityStore{
		Players:     make(map[string]*Ship),
		Projectiles: make(map[string]*Bullet),
		Asteroids:   make(map[string]*Asteroid),
		Enemies:     make(map[string]*Ship),
		Pickups:     make(map[string]*pickups.Pickup),
	}
}
