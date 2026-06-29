package realtime

import (
	"sort"

	game "github.com/Lokee86/space-rocks/server/internal/game"
)

type WorldLaneProjection struct {
	Ships     []WorldShipRecord
	Bullets   []WorldBulletRecord
	Asteroids []WorldAsteroidRecord
	Pickups   []WorldPickupRecord
}

func ProjectWorldLane(snapshot game.GameplayPresentationSnapshot) WorldLaneProjection {
	shipKeys := make([]string, 0, len(snapshot.Players))
	for id := range snapshot.Players {
		shipKeys = append(shipKeys, id)
	}
	sort.Strings(shipKeys)

	ships := make([]WorldShipRecord, 0, len(shipKeys))
	for _, id := range shipKeys {
		player := snapshot.Players[id]
		ships = append(ships, WorldShipRecord{
			ID:         id,
			ShipType:   player.ShipType,
			X:          player.X,
			Y:          player.Y,
			Rotation:   player.Rotation,
			Health:     player.Health,
			Shields:    player.Shields,
			Thrusting:  player.Thrusting,
			TargetKind: player.TargetKind,
			TargetID:   player.TargetID,
		})
	}

	bulletKeys := make([]string, 0, len(snapshot.Bullets))
	for id := range snapshot.Bullets {
		bulletKeys = append(bulletKeys, id)
	}
	sort.Strings(bulletKeys)

	bullets := make([]WorldBulletRecord, 0, len(bulletKeys))
	for _, id := range bulletKeys {
		bullet := snapshot.Bullets[id]
		bullets = append(bullets, WorldBulletRecord{
			ID:             id,
			OwnerID:        bullet.OwnerID,
			X:              bullet.X,
			Y:              bullet.Y,
			Rotation:       bullet.Rotation,
			WeaponID:       bullet.WeaponID,
			ProjectileType: bullet.ProjectileType,
		})
	}

	asteroidKeys := make([]string, 0, len(snapshot.Asteroids))
	for id := range snapshot.Asteroids {
		asteroidKeys = append(asteroidKeys, id)
	}
	sort.Strings(asteroidKeys)

	asteroids := make([]WorldAsteroidRecord, 0, len(asteroidKeys))
	for _, id := range asteroidKeys {
		asteroid := snapshot.Asteroids[id]
		asteroids = append(asteroids, WorldAsteroidRecord{
			ID:      id,
			X:       asteroid.X,
			Y:       asteroid.Y,
			Size:    asteroid.Size,
			Health:  asteroid.Health,
			Scale:   asteroid.Scale,
			Variant: asteroid.Variant,
		})
	}

	pickupKeys := make([]string, 0, len(snapshot.Pickups))
	for id := range snapshot.Pickups {
		pickupKeys = append(pickupKeys, id)
	}
	sort.Strings(pickupKeys)

	pickups := make([]WorldPickupRecord, 0, len(pickupKeys))
	for _, id := range pickupKeys {
		pickup := snapshot.Pickups[id]
		pickups = append(pickups, WorldPickupRecord{
			ID:              id,
			Type:            pickup.Type,
			PickupClass:     pickup.PickupClass,
			X:               pickup.X,
			Y:               pickup.Y,
			Health:          pickup.Health,
			AgeSeconds:      pickup.AgeSeconds,
			LifespanSeconds: pickup.LifespanSeconds,
		})
	}

	return WorldLaneProjection{
		Ships:     ships,
		Bullets:   bullets,
		Asteroids: asteroids,
		Pickups:   pickups,
	}
}

func BuildWorldFullPacket(snapshot game.GameplayPresentationSnapshot, sequence int) WorldFullPacket {
	projection := ProjectWorldLane(snapshot)
	return WorldFullPacket{
		Type: PacketFamilyWorldFull,
		Metadata: Metadata{
			Lane:           LaneWorld,
			Sequence:       sequence,
			BaselineID:     snapshot.SelfID,
			SnapshotID:     snapshot.SelfID,
			ServerSentMsec: snapshot.ServerSentMsec,
			SnapshotKind:   SnapshotKind("full"),
			ChunkIndex:     0,
			ChunkCount:     1,
			IsFinalChunk:   true,
		},
		Ships:     projection.Ships,
		Bullets:   projection.Bullets,
		Asteroids: projection.Asteroids,
		Pickups:   projection.Pickups,
	}
}

