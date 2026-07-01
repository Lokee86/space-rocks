package realtime

type OverlayReceiverWireRecord struct {
	SelfID                    string
	Lives                     int
	Score                     int
	RespawnCooldown           int64
	PrimaryWeaponID           string
	PrimaryAmmoPolicy         string
	PrimaryCooldownRemaining  int64
	PrimaryAmmoRemaining      int
	SecondaryWeaponID         string
	SecondaryAmmoPolicy       string
	SecondaryCooldownRemaining int64
	SecondaryAmmoRemaining    int
}

type WorldShipWireRecord struct {
	ID         string
	ShipType   string
	X          int64
	Y          int64
	Rotation   int64
	Health     int
	Shields    int
	Thrusting  bool
	TargetKind string
	TargetID   string
}

type WorldBulletWireRecord struct {
	ID             string
	OwnerID        string
	X              int64
	Y              int64
	Rotation       int64
	WeaponID       string
	ProjectileType string
}

type WorldAsteroidWireRecord struct {
	ID      string
	X       int64
	Y       int64
	Size    int
	Health  int
	Scale   int64
	Variant int
}

type WorldPickupWireRecord struct {
	ID              string
	Type            string
	PickupClass     string
	X               int64
	Y               int64
	Health          int
	AgeSeconds      int64
	LifespanSeconds int64
}

type OverlayWireFullPacket struct {
	Type     string
	Metadata Metadata
	Receiver OverlayReceiverWireRecord
}

type WorldWireFullPacket struct {
	Type      string
	Metadata  Metadata
	Ships     []WorldShipWireRecord
	Bullets   []WorldBulletWireRecord
	Asteroids []WorldAsteroidWireRecord
	Pickups   []WorldPickupWireRecord
}

type SessionPlayerWireRecord struct {
	ID                  string
	ShipType            string
	Score               int
	Lives               int
	RespawnCooldown     int64
	PrimaryWeaponID     string
	PrimaryAmmoPolicy   string
	SecondaryWeaponID   string
	SecondaryAmmoPolicy string
	SpawnX              int64
	SpawnY              int64
}

type SessionWireFullPacket struct {
	Type           string
	Metadata       Metadata
	Players        []SessionPlayerWireRecord
	PlayerLifecycle []SessionLifecycleRecord
	TotalAsteroids int
}