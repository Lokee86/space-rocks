package rooms

type RoomMember struct {
	SessionID string
	PlayerID  string
	Ready     bool
	Connected bool
}

func NewRoomMember(sessionID string) *RoomMember {
	return &RoomMember{
		SessionID: sessionID,
		Connected: true,
	}
}

func (member *RoomMember) SetReady(ready bool) {
	member.Ready = ready
}

func (member *RoomMember) MarkConnected() {
	member.Connected = true
}

func (member *RoomMember) MarkDisconnected() {
	member.Connected = false
}
