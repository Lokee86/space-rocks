package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	entities "github.com/Lokee86/space-rocks/server/internal/game/entities"
)

type statePacketWithDebugStatus struct {
	Type            string                            `json:"type"`
	SelfID          string                            `json:"self_id"`
	Lives           int                               `json:"lives"`
	ServerSentMsec  int                               `json:"server_sent_msec"`
	DebugStatus     DebugStatus                       `json:"debug_status"`
	DebugStatuses   map[string]DebugStatus            `json:"debug_statuses"`
	Players         map[string]entities.ShipState     `json:"players"`
	PlayerLifecycle map[string]string                 `json:"player_lifecycle"`
	Bullets         map[string]entities.BulletState   `json:"bullets"`
	Asteroids       map[string]entities.AsteroidState `json:"asteroids"`
	Events          []game.EventState                 `json:"events"`
}

func WrapStatePacket(state game.StatePacket, status DebugStatus, statuses map[string]DebugStatus) any {
	return statePacketWithDebugStatus{
		Type:            state.Type,
		SelfID:          state.SelfID,
		Lives:           state.Lives,
		ServerSentMsec:  state.ServerSentMsec,
		DebugStatus:     status,
		DebugStatuses:   statuses,
		Players:         state.Players,
		PlayerLifecycle: state.PlayerLifecycle,
		Bullets:         state.Bullets,
		Asteroids:       state.Asteroids,
		Events:          state.Events,
	}
}
