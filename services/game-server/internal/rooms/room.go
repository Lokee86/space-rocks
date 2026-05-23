package rooms

import (
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

type Room struct {
	ID             string
	State          RoomState
	Game           *game.Game
	Members        map[string]*RoomMember
	Joinable       bool
	ActivePlayers  int
	CleanupTimer   *time.Timer
	CleanupVersion int
	mu             sync.Mutex
}

func NewRoom(roomID string, state RoomState, gameInstance *game.Game) *Room {
	return &Room{
		ID:       roomID,
		State:    state,
		Game:     gameInstance,
		Members:  make(map[string]*RoomMember),
		Joinable: true,
	}
}

func (room *Room) AddMember(member *RoomMember) *RoomMember {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.Members[member.SessionID] = member

	return member
}

func (room *Room) AddMemberID(sessionID string) *RoomMember {
	return room.AddMember(NewRoomMember(sessionID))
}

func (room *Room) RemoveMember(sessionID string) {
	room.mu.Lock()
	defer room.mu.Unlock()

	delete(room.Members, sessionID)
}

func (room *Room) HasMember(sessionID string) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	_, ok := room.Members[sessionID]
	return ok
}

func (room *Room) MemberCount() int {
	room.mu.Lock()
	defer room.mu.Unlock()

	return len(room.Members)
}

func (room *Room) SetMemberReady(sessionID string, ready bool) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	member, ok := room.Members[sessionID]
	if !ok {
		return false
	}

	member.SetReady(ready)
	return true
}

func (room *Room) SetJoinable(joinable bool) {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.Joinable = joinable
}

func (room *Room) IsJoinable() bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.Joinable
}

func (room *Room) ValidateStart(memberID string) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if _, ok := room.Members[memberID]; !ok {
		return &RoomDomainError{
			Code:    RoomErrorNotInRoom,
			Message: "Member is not in the room.",
		}
	}

	switch room.State {
	case RoomStateLobby:
	case RoomStateStarting, RoomStateInGame:
		return &RoomDomainError{
			Code:    RoomErrorRoomInGame,
			Message: "Room is already in game.",
		}
	default:
		return &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Game can only be started from the lobby.",
		}
	}

	for _, member := range room.Members {
		if member.Connected && !member.Ready {
			return &RoomDomainError{
				Code:    RoomErrorNotReady,
				Message: "All connected members must be ready.",
			}
		}
	}

	return nil
}

func (room *Room) MarkStarting() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateLobby {
		return &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room can only start from the lobby.",
		}
	}

	room.State = RoomStateStarting
	return nil
}

func (room *Room) MarkGameOver() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateInGame {
		return &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room can only move to game over from in-game.",
		}
	}

	room.State = RoomStateGameOver
	return nil
}

func (room *Room) ResetToLobby(memberID string) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if _, ok := room.Members[memberID]; !ok {
		return &RoomDomainError{
			Code:    RoomErrorNotInRoom,
			Message: "Member is not in the room.",
		}
	}

	if room.State != RoomStateGameOver {
		return &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room can only return to lobby from game over.",
		}
	}

	for _, member := range room.Members {
		member.SetReady(false)
	}
	room.Game = nil
	room.State = RoomStateLobby
	return nil
}

func (room *Room) IsFull() bool {
	return room.MemberCount() >= MaxPlayersPerRoom
}

func (room *Room) IsEmpty() bool {
	return room.ActivePlayers == 0 && room.MemberCount() == 0
}

func (room *Room) ShouldCleanup() bool {
	return room != nil && room.IsEmpty()
}

func (room *Room) IsGameOver() bool {
	if room == nil || room.State != RoomStateInGame || room.Game == nil {
		return false
	}

	// TODO: Delegate to game.Game once it exposes match-over state for all
	// active room players. The current lives/respawn data needed for an exact
	// multiplayer room game-over check is private to the game package.
	return false
}

func (room *Room) MembersSnapshot() []RoomMember {
	room.mu.Lock()
	defer room.mu.Unlock()

	members := make([]RoomMember, 0, len(room.Members))
	for _, member := range room.Members {
		members = append(members, *member)
	}

	return members
}
