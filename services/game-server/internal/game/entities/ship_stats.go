package entities

import "github.com/Lokee86/space-rocks/server/internal/constants"

const (
	DefaultShipRotationSpeedModifier     = 1.0
	DefaultShipThrustForceModifier       = 1.0
	DefaultShipMaxSpeedModifier          = 1.0
	DefaultShipDampingModifier           = 1.0
	DefaultShipBulletCooldownModifier    = 1.0
	DefaultShipBulletSpeedModifier       = 1.0
	DefaultShipBulletLifetimeModifier    = 1.0
	DefaultShipBulletSpawnOffsetModifier = 1.0
	DefaultShipCollisionShapeID          = "v_wing"
)

type ShipStats struct {
	RotationSpeed     float64
	ThrustForce       float64
	MaxSpeed          float64
	Damping           float64
	BulletCooldown    float64
	BulletSpeed       float64
	BulletLifetime    float64
	BulletSpawnOffset float64
	CollisionShapeID  string
}

type ShipStatModifiers struct {
	RotationSpeed     float64
	ThrustForce       float64
	MaxSpeed          float64
	Damping           float64
	BulletCooldown    float64
	BulletSpeed       float64
	BulletLifetime    float64
	BulletSpawnOffset float64
	CollisionShapeID  string
}

func DefaultShipStatModifiers() ShipStatModifiers {
	return ShipStatModifiers{
		RotationSpeed:     DefaultShipRotationSpeedModifier,
		ThrustForce:       DefaultShipThrustForceModifier,
		MaxSpeed:          DefaultShipMaxSpeedModifier,
		Damping:           DefaultShipDampingModifier,
		BulletCooldown:    DefaultShipBulletCooldownModifier,
		BulletSpeed:       DefaultShipBulletSpeedModifier,
		BulletLifetime:    DefaultShipBulletLifetimeModifier,
		BulletSpawnOffset: DefaultShipBulletSpawnOffsetModifier,
		CollisionShapeID:  DefaultShipCollisionShapeID,
	}
}

func ResolveShipStatModifiers(shipTypeID string) ShipStatModifiers {
	switch shipTypeID {
	case DefaultShipTypeID:
		return DefaultShipStatModifiers()
	default:
		return DefaultShipStatModifiers()
	}
}

func DefaultShipStats() ShipStats {
	return resolveShipStats(DefaultShipStatModifiers())
}

func ResolveShipStats(shipTypeID string) ShipStats {
	return resolveShipStats(ResolveShipStatModifiers(shipTypeID))
}

func resolveShipStats(modifiers ShipStatModifiers) ShipStats {
	return ShipStats{
		RotationSpeed:     constants.PlayerRotationSpeed * modifiers.RotationSpeed,
		ThrustForce:       constants.PlayerThrustForce * modifiers.ThrustForce,
		MaxSpeed:          constants.PlayerMaxSpeed * modifiers.MaxSpeed,
		Damping:           constants.PlayerDamping * modifiers.Damping,
		BulletCooldown:    constants.BulletCooldown * modifiers.BulletCooldown,
		BulletSpeed:       constants.BulletSpeed * modifiers.BulletSpeed,
		BulletLifetime:    constants.BulletLifetime * modifiers.BulletLifetime,
		BulletSpawnOffset: constants.BulletSpawnOffset * modifiers.BulletSpawnOffset,
		CollisionShapeID:  modifiers.CollisionShapeID,
	}
}
