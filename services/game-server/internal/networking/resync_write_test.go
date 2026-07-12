package networking

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	"github.com/gorilla/websocket"
)

func resyncWriteSession(t *testing.T) (*webSocketSession, string) {
	t.Helper()
	room := rooms.NewRoom("room-1", rooms.RoomStateLobby, nil)
	room.SetJoinable(false)
	room.AddMember(rooms.NewRoomMember("session-owner"))
	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("start single-player game: %v", err)
	}
	t.Cleanup(func() {
		if room.GameInstance() != nil {
			room.GameInstance().Stop()
		}
	})
	matchID := room.CurrentMatchID()
	return &webSocketSession{
		room:                room,
		currentRoomID:       room.ID,
		currentGamePlayerID: "player-1",
		realtimeState:       realtime.NewRealtimeSessionState("player-1", matchID),
	}, matchID
}

func readyResyncWriteState(session *webSocketSession) {
	session.realtimeState.UpdateLane(realtime.LaneWorld, realtime.Metadata{Lane: realtime.LaneWorld, Sequence: 7, BaselineID: "baseline-7", SnapshotID: "snapshot-7", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	session.realtimeState.MarkBaselineReady(realtime.LaneWorld)
	session.realtimeState.StoreBaselineProjection(realtime.LaneWorld, "projection")
}

func TestWriteResyncRequiredAndApplyWritesBeforeInvalidating(t *testing.T) {
	session, matchID := resyncWriteSession(t)
	readyResyncWriteState(session)
	request := queuedResyncRequest{Request: realtime.ResyncRequest{MatchID: matchID, Lane: realtime.LaneWorld, BaselineID: "baseline-7", Sequence: 7, Reason: "missing_baseline"}, RoomID: session.room.ID, ReceiverID: "player-1", MatchID: matchID}
	called := false
	original := writeResyncMessage
	t.Cleanup(func() { writeResyncMessage = original })
	writeResyncMessage = func(_ *websocket.Conn, message []byte, _ func(error)) bool {
		called = true
		var payload map[string]any
		if err := json.Unmarshal(message, &payload); err != nil {
			t.Fatal(err)
		}
		if payload["type"] != realtime.PacketFamilyResyncRequired || payload["match_id"] != matchID || payload["lane"] != "world" || payload["baseline_id"] != "baseline-7" || payload["reason"] != "missing_baseline" || payload["sequence"] != float64(7) {
			t.Fatalf("unexpected payload: %#v", payload)
		}
		if !session.realtimeState.LaneBaselineReady(realtime.LaneWorld) {
			t.Fatal("state invalidated before write")
		}
		if projection, ok := session.realtimeState.BaselineProjection(realtime.LaneWorld); !ok || projection != "projection" {
			t.Fatalf("projection unavailable during write: %#v, %t", projection, ok)
		}
		return true
	}
	if !writeResyncRequiredAndApply(session, request, "remote") || !called {
		t.Fatal("expected successful write")
	}
	if session.realtimeState.LaneBaselineReady(realtime.LaneWorld) {
		t.Fatal("expected baseline invalidated")
	}
	if projection, ok := session.realtimeState.BaselineProjection(realtime.LaneWorld); ok || projection != nil {
		t.Fatalf("expected projection cleared: %#v, %t", projection, ok)
	}
	lane, _ := session.realtimeState.LaneState(realtime.LaneWorld)
	if lane.Sequence != 7 || lane.BaselineID != "baseline-7" {
		t.Fatalf("metadata changed: %#v", lane)
	}
}

func TestWriteResyncRequiredAndApplyFailedWritePreservesState(t *testing.T) {
	session, matchID := resyncWriteSession(t)
	readyResyncWriteState(session)
	original := writeResyncMessage
	t.Cleanup(func() { writeResyncMessage = original })
	writeResyncMessage = func(_ *websocket.Conn, _ []byte, _ func(error)) bool { return false }
	request := queuedResyncRequest{Request: realtime.ResyncRequest{MatchID: matchID, Lane: realtime.LaneWorld, BaselineID: "baseline-7", Sequence: 7, Reason: "missing_baseline"}, RoomID: session.room.ID, ReceiverID: "player-1", MatchID: matchID}
	if writeResyncRequiredAndApply(session, request, "remote") {
		t.Fatal("expected failed write")
	}
	if !session.realtimeState.LaneBaselineReady(realtime.LaneWorld) {
		t.Fatal("readiness changed after failed write")
	}
	if projection, ok := session.realtimeState.BaselineProjection(realtime.LaneWorld); !ok || projection != "projection" {
		t.Fatalf("projection changed after failed write: %#v, %t", projection, ok)
	}
}

func TestWriteResyncRequiredAndApplyIgnoresStaleOrInvalidRequests(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*webSocketSession, *queuedResyncRequest)
	}{
		{"unsupported lane", func(_ *webSocketSession, r *queuedResyncRequest) { r.Request.Lane = realtime.LaneControl }},
		{"missing request match", func(_ *webSocketSession, r *queuedResyncRequest) { r.Request.MatchID = "" }},
		{"missing queued match", func(_ *webSocketSession, r *queuedResyncRequest) { r.MatchID = "" }},
		{"request queued mismatch", func(_ *webSocketSession, r *queuedResyncRequest) { r.Request.MatchID = "other-match" }},
		{"nil session", func(_ *webSocketSession, _ *queuedResyncRequest) {}},
		{"nil room", func(s *webSocketSession, _ *queuedResyncRequest) { s.room = nil }},
		{"nil game", func(s *webSocketSession, _ *queuedResyncRequest) {
			s.room = rooms.NewRoom("room-1", rooms.RoomStateInGame, nil)
		}},
		{"empty player", func(s *webSocketSession, _ *queuedResyncRequest) { s.currentGamePlayerID = "" }},
		{"stale room", func(_ *webSocketSession, r *queuedResyncRequest) { r.RoomID = "old-room" }},
		{"stale receiver", func(_ *webSocketSession, r *queuedResyncRequest) { r.ReceiverID = "old-player" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			session, matchID := resyncWriteSession(t)
			readyResyncWriteState(session)
			request := queuedResyncRequest{Request: realtime.ResyncRequest{MatchID: matchID, Lane: realtime.LaneWorld, BaselineID: "baseline-7", Sequence: 7, Reason: "missing_baseline"}, RoomID: session.room.ID, ReceiverID: "player-1", MatchID: matchID}
			called := false
			original := writeResyncMessage
			t.Cleanup(func() { writeResyncMessage = original })
			writeResyncMessage = func(_ *websocket.Conn, _ []byte, _ func(error)) bool { called = true; return true }
			if tc.name == "nil session" {
				tc.mutate(nil, &request)
				if writeResyncRequiredAndApply(nil, request, "remote") == false {
					t.Fatal("nil context should be ignored")
				}
			} else {
				tc.mutate(session, &request)
				writeResyncRequiredAndApply(session, request, "remote")
			}
			if called {
				t.Fatal("writer called for invalid request")
			}
			if !session.realtimeState.LaneBaselineReady(realtime.LaneWorld) {
				t.Fatal("state mutated")
			}
		})
	}
}

func TestWriteResyncRequiredAndApplyRejectsStaleMatchAfterRoomAdvances(t *testing.T) {
	session, oldMatchID := resyncWriteSession(t)
	readyResyncWriteState(session)
	if err := session.room.MarkGameOver(); err != nil {
		t.Fatalf("mark game over: %v", err)
	}
	if err := session.room.ResetToLobby("Player-1"); err != nil {
		t.Fatalf("reset room: %v", err)
	}
	if err := session.room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("start next match: %v", err)
	}
	newMatchID := session.room.CurrentMatchID()
	if newMatchID == oldMatchID {
		t.Fatalf("expected match to advance from %q", oldMatchID)
	}
	session.realtimeState = realtime.NewRealtimeSessionState("player-1", newMatchID)
	readyResyncWriteState(session)
	request := queuedResyncRequest{Request: realtime.ResyncRequest{MatchID: oldMatchID, Lane: realtime.LaneWorld, BaselineID: "baseline-7", Sequence: 7, Reason: "missing_baseline"}, RoomID: session.room.ID, ReceiverID: "player-1", MatchID: oldMatchID}
	called := false
	original := writeResyncMessage
	t.Cleanup(func() { writeResyncMessage = original })
	writeResyncMessage = func(_ *websocket.Conn, _ []byte, _ func(error)) bool { called = true; return true }
	writeResyncRequiredAndApply(session, request, "remote")
	if called {
		t.Fatal("writer called for stale match")
	}
	if !session.realtimeState.LaneBaselineReady(realtime.LaneWorld) {
		t.Fatal("new match readiness mutated")
	}
	if projection, ok := session.realtimeState.BaselineProjection(realtime.LaneWorld); !ok || projection != "projection" {
		t.Fatalf("new match projection mutated: %#v, %t", projection, ok)
	}
}
