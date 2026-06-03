package rooms

import "github.com/Lokee86/space-rocks/server/internal/rooms/roomrules"

func (room *Room) ValidateStart(playerID string) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.validateStartLocked(playerID)
}

func (room *Room) validateStartLocked(playerID string) *RoomDomainError {
	if roomErr := room.validateStartPreconditionsLocked(); roomErr != nil {
		return roomErr
	}

	memberSnapshot := room.membership.membersSnapshot()
	members := make([]roomrules.StartMember, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		members = append(members, roomrules.StartMember{
			PlayerID:  member.PlayerID,
			Ready:     member.Ready,
			Connected: member.Connected,
		})
	}

	decision := roomrules.DecideStart(roomrules.StartInput{
		State:              string(room.State),
		OwnerID:            room.membership.ownerIDValue(),
		RequestingPlayerID: playerID,
		Members:            members,
	})
	if roomErr := roomDomainErrorFromDecision(decision); roomErr != nil {
		return roomErr
	}

	return nil
}

func (room *Room) validateStartPreconditionsLocked() *RoomDomainError {
	switch room.State {
	case RoomStateLobby:
	case RoomStateStarting, RoomStateInGame:
		return &RoomDomainError{Code: RoomErrorRoomInGame, Message: "Room is already in game."}
	default:
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Game can only be started from the lobby."}
	}

	if room.membership.memberCount() == 0 {
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}

	return nil
}

func (room *Room) SetReadyInLobby(playerID string, ready bool) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateLobby {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Ready state can only be changed in the lobby."}
	}

	member, ok := room.membership.memberByPlayerID(playerID)
	if !ok {
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}

	member.SetReady(ready)
	return nil
}
