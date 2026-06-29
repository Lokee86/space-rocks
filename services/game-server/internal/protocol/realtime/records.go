package realtime

import "github.com/Lokee86/space-rocks/server/internal/game"

type WorldShipRecord struct {
	ID         string
	ShipType   string
	X          float64
	Y          float64
	Rotation   float64
	Health     int
	Shields    int
	Thrusting  bool
	TargetKind string
	TargetID   string
}

type WorldBulletRecord struct {
	ID             string
	OwnerID        string
	X              float64
	Y              float64
	Rotation       float64
	WeaponID       string
	ProjectileType string
}

type WorldAsteroidRecord struct {
	ID      string
	X       float64
	Y       float64
	Size    int
	Health  int
	Scale   float64
	Variant int
}

type WorldPickupRecord struct {
	ID              string
	Type            string
	PickupClass     string
	X               float64
	Y               float64
	Health          int
	AgeSeconds      float64
	LifespanSeconds float64
}

type OverlayReceiverRecord struct {
	SelfID                   string
	Lives                    int
	Score                    int
	RespawnCooldown          float64
	PrimaryWeaponID          string
	PrimaryAmmoPolicy        string
	PrimaryCooldownRemaining  float64
	PrimaryAmmoRemaining     int
	SecondaryWeaponID        string
	SecondaryAmmoPolicy      string
	SecondaryCooldownRemaining float64
	SecondaryAmmoRemaining   int
}

type SessionPlayerRecord struct {
	ID                  string
	ShipType            string
	Score               int
	Lives               int
	RespawnCooldown     float64
	PrimaryWeaponID     string
	PrimaryAmmoPolicy   string
	SecondaryWeaponID   string
	SecondaryAmmoPolicy string
	SpawnX              float64
	SpawnY              float64
}

type SessionSnapshotRecord struct {
	PlayerSessions map[string]SessionPlayerRecord
	PlayerLifecycle map[string]string
	TotalAsteroids int
}

type EventRecord struct {
	EventID string
	Event   game.EventState
}

type EventBatchRecord struct {
	BatchID string
	Sequence int
	Events   []EventRecord
}
