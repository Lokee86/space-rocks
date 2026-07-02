package realtime

import (
	"sort"

	game "github.com/Lokee86/space-rocks/server/internal/game"
)

type SessionLaneProjection struct {
	Players       []SessionPlayerRecord
	PlayerLifecycle []SessionLifecycleRecord
	TotalAsteroids int
}

type SessionLifecycleRecord struct {
	PlayerID string
	Status   string
}

func ProjectSessionLane(snapshot game.GameplayPresentationSnapshot) SessionLaneProjection {
	playerKeys := make([]string, 0, len(snapshot.PlayerSessions))
	for id := range snapshot.PlayerSessions {
		playerKeys = append(playerKeys, id)
	}
	sort.Strings(playerKeys)

	players := make([]SessionPlayerRecord, 0, len(playerKeys))
	for _, id := range playerKeys {
		player := snapshot.PlayerSessions[id]
		players = append(players, SessionPlayerRecord{
			ID:                  player.ID,
			ShipType:            player.ShipType,
			Score:               player.Score,
			Lives:               player.Lives,
			RespawnCooldown:     player.RespawnCooldown,
			PrimaryWeaponID:     player.PrimaryWeaponID,
			PrimaryAmmoPolicy:   player.PrimaryAmmoPolicy,
			SecondaryWeaponID:   player.SecondaryWeaponID,
			SecondaryAmmoPolicy: player.SecondaryAmmoPolicy,
			SpawnX:              player.SpawnX,
			SpawnY:              player.SpawnY,
		})
	}

	lifecycleKeys := make([]string, 0, len(snapshot.PlayerLifecycle))
	for id := range snapshot.PlayerLifecycle {
		lifecycleKeys = append(lifecycleKeys, id)
	}
	sort.Strings(lifecycleKeys)

	lifecycle := make([]SessionLifecycleRecord, 0, len(lifecycleKeys))
	for _, id := range lifecycleKeys {
		lifecycle = append(lifecycle, SessionLifecycleRecord{
			PlayerID: id,
			Status:   snapshot.PlayerLifecycle[id],
		})
	}

	return SessionLaneProjection{
		Players:         players,
		PlayerLifecycle: lifecycle,
		TotalAsteroids:  snapshot.TotalAsteroids,
	}
}

func BuildSessionFullPacket(snapshot game.GameplayPresentationSnapshot, sequence int) SessionFullPacket {
	projection := ProjectSessionLane(snapshot)
	return SessionFullPacket{
		Type: PacketFamilySessionFull,
		Metadata: Metadata{
			Lane:           LaneSession,
			Sequence:       sequence,
			BaselineID:     FullBaselineID(LaneSession, sequence),
			SnapshotID:     FullBaselineID(LaneSession, sequence),
			ServerSentMsec: snapshot.ServerSentMsec,
			SnapshotKind:   SnapshotKind("full"),
			ChunkIndex:     0,
			ChunkCount:     1,
			IsFinalChunk:   true,
		},
		Players:         projection.Players,
		PlayerLifecycle: projection.PlayerLifecycle,
		TotalAsteroids:  projection.TotalAsteroids,
	}
}


