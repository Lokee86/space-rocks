package realtime

import (
	game "github.com/Lokee86/space-rocks/server/internal/game"
)

type OverlayLaneProjection struct {
	Receiver OverlayReceiverRecord
}

func ProjectOverlayLane(snapshot game.GameplayPresentationSnapshot, receiverPlayerID string) OverlayLaneProjection {
	receiverSession := snapshot.PlayerSessions[receiverPlayerID]
	receiverShip := snapshot.Players[receiverPlayerID]

	return OverlayLaneProjection{
		Receiver: OverlayReceiverRecord{
			SelfID:                     snapshot.SelfID,
			Lives:                      snapshot.Lives,
			Score:                      receiverSession.Score,
			RespawnCooldown:            receiverSession.RespawnCooldown,
			PrimaryWeaponID:            receiverShip.PrimaryWeaponID,
			PrimaryAmmoPolicy:          receiverShip.PrimaryAmmoPolicy,
			PrimaryCooldownRemaining:   receiverShip.PrimaryCooldownRemaining,
			PrimaryAmmoRemaining:       receiverShip.PrimaryAmmoRemaining,
			SecondaryWeaponID:          receiverShip.SecondaryWeaponID,
			SecondaryAmmoPolicy:        receiverShip.SecondaryAmmoPolicy,
			SecondaryCooldownRemaining: receiverShip.SecondaryCooldownRemaining,
			SecondaryAmmoRemaining:     receiverShip.SecondaryAmmoRemaining,
		},
	}
}

func BuildOverlayFullPacket(snapshot game.GameplayPresentationSnapshot, receiverPlayerID string, sequence int) OverlayFullPacket {
	projection := ProjectOverlayLane(snapshot, receiverPlayerID)
	return OverlayFullPacket{
		Type: PacketFamilyOverlayFull,
		Metadata: Metadata{
			Lane:           LaneOverlay,
			Sequence:       sequence,
			BaselineID:     FullBaselineID(LaneOverlay, sequence),
			SnapshotID:     FullBaselineID(LaneOverlay, sequence),
			ServerSentMsec: snapshot.ServerSentMsec,
			SnapshotKind:   SnapshotKind("full"),
			ChunkIndex:     0,
			ChunkCount:     1,
			IsFinalChunk:   true,
		},
		Receiver: projection.Receiver,
	}
}


