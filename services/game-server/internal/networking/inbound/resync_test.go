package inbound

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

type resyncGameplaySessionFake struct {
	room     *rooms.Room
	playerID string
	requests []realtime.ResyncRequest
}

func (f *resyncGameplaySessionFake) CurrentRoom() *rooms.Room    { return f.room }
func (f *resyncGameplaySessionFake) CurrentGamePlayerID() string { return f.playerID }
func (f *resyncGameplaySessionFake) EnqueuePlayerPauseState()    {}
func (f *resyncGameplaySessionFake) EnqueueResyncRequest(request realtime.ResyncRequest) bool {
	f.requests = append(f.requests, request)
	return true
}

func activeResyncRoom(t *testing.T) (*rooms.Room, string) {
	t.Helper()
	room := rooms.NewRoom("room-1", rooms.RoomStateLobby, nil)
	room.AddMember(rooms.NewRoomMember("session-owner"))
	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("start single-player game: %v", err)
	}
	t.Cleanup(func() { room.GameInstance().Stop() })
	return room, room.CurrentMatchID()
}

func TestHandleGameplayPacketQueuesResyncRequest(t *testing.T) {
	room, matchID := activeResyncRoom(t)
	session := &resyncGameplaySessionFake{room: room, playerID: "player-1"}
	packet := game.ClientPacket{Type: game.PacketTypeResyncRequest, MatchID: matchID, Lane: string(realtime.LaneWorld), BaselineID: "baseline-4", Sequence: 4, Reason: "wrong_baseline"}
	if !HandleGameplayPacket(session, packet) {
		t.Fatal("expected valid resync request")
	}
	if len(session.requests) != 1 {
		t.Fatalf("queued requests = %d, want 1", len(session.requests))
	}
	if got := session.requests[0]; got.MatchID != matchID || got.Lane != realtime.LaneWorld || got.BaselineID != "baseline-4" || got.Sequence != 4 || got.Reason != "wrong_baseline" {
		t.Fatalf("unexpected request: %#v", got)
	}
}

func TestHandleGameplayPacketRejectsInvalidResyncContext(t *testing.T) {
	tests := []struct {
		name     string
		room     *rooms.Room
		playerID string
		matchID  string
	}{
		{"invalid lane", nil, "player-1", ""},
		{"missing match", nil, "player-1", ""},
		{"nil room", nil, "player-1", ""},
		{"empty player", nil, "", ""},
		{"nil game", rooms.NewRoom("room-1", rooms.RoomStateInGame, nil), "player-1", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			room := tc.room
			matchID := tc.matchID
			if tc.name == "invalid lane" || tc.name == "empty player" {
				room, matchID = activeResyncRoom(t)
			}
			session := &resyncGameplaySessionFake{room: room, playerID: tc.playerID}
			lane := string(realtime.LaneWorld)
			if tc.name == "invalid lane" {
				lane = "unsupported"
			}
			if HandleGameplayPacket(session, game.ClientPacket{Type: game.PacketTypeResyncRequest, MatchID: matchID, Lane: lane}) {
				t.Fatal("expected rejection")
			}
			if len(session.requests) != 0 {
				t.Fatalf("queued requests = %d, want 0", len(session.requests))
			}
		})
	}
}
