package rooms

import "time"

const (
	DefaultRoomID        = "default"
	RoomCleanupGraceTime = 30 * time.Second
	MaxPlayersPerRoom    = 8
	RoomCodeLength       = 6
	RoomCodeAlphabet     = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

type RoomState string

const (
	RoomStateLobby    RoomState = "Lobby"
	RoomStateStarting RoomState = "Starting"
	RoomStateInGame   RoomState = "InGame"
	RoomStateGameOver RoomState = "GameOver"
	RoomStateClosed   RoomState = "Closed"
)

const (
	RoomErrorRoomNotFound     = "room_not_found"
	RoomErrorRoomClosed       = "room_closed"
	RoomErrorRoomInGame       = "room_in_game"
	RoomErrorRoomFull         = "room_full"
	RoomErrorAlreadyInRoom    = "already_in_room"
	RoomErrorNotInRoom        = "not_in_room"
	RoomErrorInvalidRoomCode  = "invalid_room_code"
	RoomErrorNotReady         = "not_ready"
	RoomErrorInvalidRoomState = "invalid_room_state"
)
