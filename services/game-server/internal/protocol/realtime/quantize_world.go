package realtime

import (
	"fmt"

	"github.com/Lokee86/space-rocks/server/internal/protocol/realtime/quantize"
)

func quantizeWorldFullPacket(packet WorldFullPacket) (WorldWireFullPacket, error) {
	quantized := WorldWireFullPacket{
		Type: packet.Type,
		Metadata: packet.Metadata,
		Ships: make([]WorldShipWireRecord, 0, len(packet.Ships)),
		Bullets: make([]WorldBulletWireRecord, 0, len(packet.Bullets)),
		Asteroids: make([]WorldAsteroidWireRecord, 0, len(packet.Asteroids)),
		Pickups: make([]WorldPickupWireRecord, 0, len(packet.Pickups)),
	}

	var err error
	for _, ship := range packet.Ships {
		wireShip := WorldShipWireRecord{
			ID:         ship.ID,
			ShipType:   ship.ShipType,
			Health:     ship.Health,
			Shields:    ship.Shields,
			Thrusting:  ship.Thrusting,
			TargetKind: ship.TargetKind,
			TargetID:   ship.TargetID,
		}
		wireShip.X, err = quantizeTypedFloat("world.ships.x", ship.X)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wireShip.Y, err = quantizeTypedFloat("world.ships.y", ship.Y)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wireShip.Rotation, err = quantizeTypedFloat("world.ships.rotation", ship.Rotation)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		quantized.Ships = append(quantized.Ships, wireShip)
	}

	for _, bullet := range packet.Bullets {
		wireBullet := WorldBulletWireRecord{
			ID:             bullet.ID,
			OwnerID:        bullet.OwnerID,
			WeaponID:       bullet.WeaponID,
			ProjectileType: bullet.ProjectileType,
		}
		wireBullet.X, err = quantizeTypedFloat("world.bullets.x", bullet.X)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wireBullet.Y, err = quantizeTypedFloat("world.bullets.y", bullet.Y)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wireBullet.Rotation, err = quantizeTypedFloat("world.bullets.rotation", bullet.Rotation)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		quantized.Bullets = append(quantized.Bullets, wireBullet)
	}

	for _, asteroid := range packet.Asteroids {
		wireAsteroid := WorldAsteroidWireRecord{
			ID:      asteroid.ID,
			Size:    asteroid.Size,
			Health:  asteroid.Health,
			Variant: asteroid.Variant,
		}
		wireAsteroid.X, err = quantizeTypedFloat("world.asteroids.x", asteroid.X)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wireAsteroid.Y, err = quantizeTypedFloat("world.asteroids.y", asteroid.Y)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wireAsteroid.Scale, err = quantizeTypedFloat("world.asteroids.scale", asteroid.Scale)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		quantized.Asteroids = append(quantized.Asteroids, wireAsteroid)
	}

	for _, pickup := range packet.Pickups {
		wirePickup := WorldPickupWireRecord{
			ID:              pickup.ID,
			Type:            pickup.Type,
			PickupClass:     pickup.PickupClass,
			Health:          pickup.Health,
			AgeSeconds:      0,
			LifespanSeconds: 0,
		}
		wirePickup.X, err = quantizeTypedFloat("world.pickups.x", pickup.X)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wirePickup.Y, err = quantizeTypedFloat("world.pickups.y", pickup.Y)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wirePickup.AgeSeconds, err = quantizeTypedFloat("world.pickups.age_seconds", pickup.AgeSeconds)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		wirePickup.LifespanSeconds, err = quantizeTypedFloat("world.pickups.lifespan_seconds", pickup.LifespanSeconds)
		if err != nil {
			return WorldWireFullPacket{}, err
		}
		quantized.Pickups = append(quantized.Pickups, wirePickup)
	}

	return quantized, nil
}

func quantizeTypedFloat(fieldPath string, value float64) (int64, error) {
	policy, ok := quantize.LookupPolicy(fieldPath)
	if !ok {
		return 0, fmt.Errorf("quantize %s: missing policy", fieldPath)
	}
	encoded, err := quantize.EncodeFloat(policy, value)
	if err != nil {
		return 0, err
	}
	return encoded, nil
}
