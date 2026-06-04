package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	runtime "github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

type statePacketWithDebugStatus struct {
	Type            string                            `json:"type"`
	SelfID          string                            `json:"self_id"`
	Lives           int                               `json:"lives"`
	ServerSentMsec  int                               `json:"server_sent_msec"`
	DebugStatus     DebugStatus                       `json:"debug_status"`
	DebugStatuses   map[string]DebugStatus            `json:"debug_statuses"`
	Players         map[string]runtime.ShipState     `json:"players"`
	PlayerLifecycle map[string]string                 `json:"player_lifecycle"`
	PlayerSessions  map[string]game.PlayerSessionState `json:"player_sessions"`
	Bullets         map[string]runtime.BulletState    `json:"bullets"`
	Asteroids       map[string]runtime.AsteroidState  `json:"asteroids"`
	TotalAsteroids  int                                `json:"total_asteroids"`
	Events          []game.EventState                  `json:"events"`
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
		PlayerSessions:  state.PlayerSessions,
		Bullets:         state.Bullets,
		Asteroids:       state.Asteroids,
		TotalAsteroids:  state.TotalAsteroids,
		Events:          state.Events,
	}
}
