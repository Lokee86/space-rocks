package gametests

import (
	"encoding/json"
	"testing"

	servergame "github.com/Lokee86/space-rocks/server/internal/game"
)

func TestGeneratedLobbyRequestPacketFields(t *testing.T) {
	clientPacket := servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: "ABCD",
		Ready:    true,
	}
	if clientPacket.RoomCode != "ABCD" {
		t.Fatalf("expected client packet room code field, got %q", clientPacket.RoomCode)
	}
	if !clientPacket.Ready {
		t.Fatal("expected client packet ready field")
	}

	joinRequest := servergame.JoinRoomRequest{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: "ABCD",
	}
	if joinRequest.RoomCode != "ABCD" {
		t.Fatalf("expected join request room code field, got %q", joinRequest.RoomCode)
	}

	setReadyRequest := servergame.SetReadyRequest{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}
	if !setReadyRequest.Ready {
		t.Fatal("expected set ready request ready field")
	}

	requestTypes := []string{
		servergame.CreateRoomRequest{Type: servergame.PacketTypeCreateRoomRequest}.Type,
		servergame.LeaveRoomRequest{Type: servergame.PacketTypeLeaveRoomRequest}.Type,
		servergame.StartGameRequest{Type: servergame.PacketTypeStartGameRequest}.Type,
		servergame.ReturnToLobbyRequest{Type: servergame.PacketTypeReturnToLobbyRequest}.Type,
	}
	for _, packetType := range requestTypes {
		if packetType == "" {
			t.Fatal("expected generated request packet type field")
		}
	}
}

func TestGeneratedLobbyPacketFields(t *testing.T) {
	snapshot := servergame.RoomSnapshot{
		Type:          servergame.PacketTypeRoomSnapshot,
		RoomCode:      "TEST",
		RoomState:     "Lobby",
		LocalPlayerID: "player-1",
		OwnerID:       "player-1",
		Members: []servergame.RoomMemberState{
			{PlayerID: "player-1", Ready: true, Connected: true},
			{PlayerID: "player-2", Ready: false, Connected: true},
		},
		MaxPlayers: 8,
	}
	if snapshot.RoomState != "Lobby" {
		t.Fatalf("expected room snapshot packet field, got %q", snapshot.RoomState)
	}
	if len(snapshot.Members) != 2 {
		t.Fatalf("expected room snapshot members, got %d", len(snapshot.Members))
	}
	if !snapshot.Members[0].Ready || snapshot.Members[1].Ready {
		t.Fatalf("expected generated member ready states, got %#v", snapshot.Members)
	}
	if snapshot.Members[0].PlayerID != "player-1" || snapshot.Members[1].PlayerID != "player-2" {
		t.Fatalf("expected generated member player IDs, got %#v", snapshot.Members)
	}
	if snapshot.LocalPlayerID != "player-1" || snapshot.OwnerID != "player-1" {
		t.Fatalf("expected local/owner player ID fields, got local=%q owner=%q", snapshot.LocalPlayerID, snapshot.OwnerID)
	}
	if snapshot.MaxPlayers != 8 {
		t.Fatalf("expected max players field, got %d", snapshot.MaxPlayers)
	}

	rawSnapshot, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal room snapshot: %v", err)
	}
	if !jsonContainsFields(t, rawSnapshot, "room_state", "members", "max_players") {
		t.Fatal("expected room snapshot JSON fields")
	}
	if jsonContainsFields(t, rawSnapshot, "member_id", "local_member_id") {
		t.Fatal("expected room snapshot JSON not to include removed member fields")
	}
	if !jsonSnapshotMembersContainReady(t, rawSnapshot) {
		t.Fatal("expected room snapshot member JSON to include ready states")
	}

	stateChanged := servergame.RoomStateChanged{
		Type:      servergame.PacketTypeRoomStateChanged,
		RoomCode:  "TEST",
		RoomState: "InGame",
	}
	if stateChanged.RoomState != "InGame" {
		t.Fatalf("expected room state changed field, got %q", stateChanged.RoomState)
	}

	roomError := servergame.RoomError{
		Type:      servergame.PacketTypeRoomError,
		ErrorCode: "room_full",
		Message:   "Room is full.",
	}
	if roomError.ErrorCode != "room_full" || roomError.Message == "" {
		t.Fatalf("expected room error code/message fields, got %#v", roomError)
	}
}

func jsonContainsFields(t *testing.T, raw []byte, fields ...string) bool {
	t.Helper()

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal generated packet JSON: %v", err)
	}

	for _, field := range fields {
		if _, ok := data[field]; !ok {
			return false
		}
	}

	return true
}

func jsonSnapshotMembersContainReady(t *testing.T, raw []byte) bool {
	t.Helper()

	var data struct {
		Members []map[string]any `json:"members"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal generated room snapshot JSON: %v", err)
	}
	if len(data.Members) == 0 {
		return false
	}
	for _, member := range data.Members {
		if _, ok := member["ready"]; !ok {
			return false
		}
	}

	return true
}
