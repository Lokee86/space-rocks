package rooms

import "github.com/Lokee86/space-rocks/server/internal/rooms/roomrules"

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

func (room *Room) ValidateJoin() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	decision := roomrules.DecideJoin(roomrules.JoinInput{
		State:       string(room.State),
		Joinable:    room.Joinable,
		MemberCount: room.membership.memberCount(),
		MaxMembers:  MaxPlayersPerRoom,
	})
	if roomErr := roomDomainErrorFromDecision(decision); roomErr != nil {
		return roomErr
	}

	return nil
}

func (room *Room) JoinMember(sessionID string) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	decision := roomrules.DecideJoin(roomrules.JoinInput{
		State:       string(room.State),
		Joinable:    room.Joinable,
		MemberCount: room.membership.memberCount(),
		MaxMembers:  MaxPlayersPerRoom,
	})
	if roomErr := roomDomainErrorFromDecision(decision); roomErr != nil {
		return roomErr
	}

	room.addMemberLocked(NewRoomMember(sessionID))
	return nil
}
