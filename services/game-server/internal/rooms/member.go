package rooms

type RoomMember struct {
	MemberID       string
	SessionID      string
	PlayerID       string
	AccountID      string
	LocalProfileID string
	Ready          bool
	Connected      bool
	IsBot          bool
}

func NewRoomMember(sessionID string) *RoomMember {
	return &RoomMember{
		MemberID:  newMemberID(),
		SessionID: sessionID,
		Connected: true,
	}
}

func NewBotRoomMember() *RoomMember {
	member := &RoomMember{
		MemberID:  newMemberID(),
		Ready:     true,
		Connected: true,
		IsBot:     true,
	}
	member.SessionID = "bot:" + member.MemberID
	return member
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
