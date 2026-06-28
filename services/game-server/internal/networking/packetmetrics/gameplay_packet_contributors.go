package packetmetrics

import "github.com/Lokee86/space-rocks/server/internal/game"

type GameplayPacketContributors struct {
	RoomState      string
	Players        int
	PlayerSessions int
	PlayerLifecycle int
	Asteroids      int
	Bullets        int
	Pickups        int
	Enemies        int
	Events         int
	TotalAsteroids int
}

func BuildGameplayPacketContributors(roomState string, statePacket game.StatePacket) GameplayPacketContributors {
	return GameplayPacketContributors{
		RoomState:       roomState,
		Players:         len(statePacket.Players),
		PlayerSessions:   len(statePacket.PlayerSessions),
		PlayerLifecycle:  len(statePacket.PlayerLifecycle),
		Asteroids:       len(statePacket.Asteroids),
		Bullets:         len(statePacket.Bullets),
		Pickups:         len(statePacket.Pickups),
		Enemies:         0,
		Events:          len(statePacket.Events),
		TotalAsteroids:  statePacket.TotalAsteroids,
	}
}
