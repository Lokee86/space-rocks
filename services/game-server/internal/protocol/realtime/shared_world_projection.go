package realtime

import (
	"fmt"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

const sharedWorldProjectionCacheKey = "realtime.world.wire.v1"

func sharedWorldWireProjection(gameInstance *game.Game, snapshot game.GameplayPresentationSnapshot) (WorldWireFullPacket, error) {
	value, err := gameInstance.PresentationDerived(snapshot.Generation, sharedWorldProjectionCacheKey, func() (any, error) {
		packet := BuildWorldFullPacket(snapshot, 0)
		quantized, err := quantizeWorldFullPacket(packet)
		if err != nil {
			return nil, fmt.Errorf("quantize shared world projection: %w", err)
		}
		return quantized, nil
	})
	if err != nil {
		return WorldWireFullPacket{}, err
	}
	projection, ok := value.(WorldWireFullPacket)
	if !ok {
		return WorldWireFullPacket{}, fmt.Errorf("shared world projection has unexpected type %T", value)
	}
	return projection, nil
}

func receiverWorldWireProjection(shared WorldWireFullPacket, snapshot game.GameplayPresentationSnapshot, sequence int) WorldWireFullPacket {
	metadata := shared.Metadata
	metadata.Sequence = sequence
	metadata.BaselineID = FullBaselineID(LaneWorld, sequence)
	metadata.SnapshotID = FullBaselineID(LaneWorld, sequence)
	metadata.ServerSentMsec = snapshot.ServerSentMsec
	metadata = metadata.WithChunk(0, 1)

	projection := WorldWireFullPacket{
		Type:     PacketFamilyWorldFull,
		Metadata: metadata,
	}

	if len(snapshot.Players) == len(shared.Ships) {
		projection.Ships = shared.Ships
	} else {
		projection.Ships = make([]WorldShipWireRecord, 0, len(snapshot.Players))
		for _, record := range shared.Ships {
			if _, ok := snapshot.Players[record.ID]; ok {
				projection.Ships = append(projection.Ships, record)
			}
		}
	}

	if len(snapshot.Bullets) == len(shared.Bullets) {
		projection.Bullets = shared.Bullets
	} else {
		projection.Bullets = make([]WorldBulletWireRecord, 0, len(snapshot.Bullets))
		for _, record := range shared.Bullets {
			if _, ok := snapshot.Bullets[record.ID]; ok {
				projection.Bullets = append(projection.Bullets, record)
			}
		}
	}

	if len(snapshot.Asteroids) == len(shared.Asteroids) {
		projection.Asteroids = shared.Asteroids
	} else {
		projection.Asteroids = make([]WorldAsteroidWireRecord, 0, len(snapshot.Asteroids))
		for _, record := range shared.Asteroids {
			if _, ok := snapshot.Asteroids[record.ID]; ok {
				projection.Asteroids = append(projection.Asteroids, record)
			}
		}
	}

	if len(snapshot.Pickups) == len(shared.Pickups) {
		projection.Pickups = shared.Pickups
	} else {
		projection.Pickups = make([]WorldPickupWireRecord, 0, len(snapshot.Pickups))
		for _, record := range shared.Pickups {
			if _, ok := snapshot.Pickups[record.ID]; ok {
				projection.Pickups = append(projection.Pickups, record)
			}
		}
	}

	return projection
}
