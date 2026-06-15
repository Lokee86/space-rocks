package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/playerdata"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

type recordingMatchResultReporter struct {
	room         *rooms.Room
	calls        int
	memberCounts []int
}

func (reporter *recordingMatchResultReporter) ReportMatchResult(summary playerdata.MatchResultSummary) error {
	reporter.calls++
	reporter.memberCounts = append(reporter.memberCounts, reporter.room.MemberCount())
	return nil
}

func TestWebSocketSessionReportsResolvedMatchBeforeRoomExit(t *testing.T) {
	t.Run("requested leave reports before removing member", func(t *testing.T) {
		session, room, reporter, cleanup := newWebSocketSessionRoomExitTestSetup(t)
		defer cleanup()

		session.leaveRequestedRoom()

		if reporter.calls != 1 {
			t.Fatalf("expected reporter to be called once, got %d", reporter.calls)
		}
		if len(reporter.memberCounts) != 1 || reporter.memberCounts[0] != 1 {
			t.Fatalf("expected reporter to observe 1 room member before leave, got %v", reporter.memberCounts)
		}
		if room.MemberCount() != 0 {
			t.Fatalf("expected room member to be removed after leave, got %d", room.MemberCount())
		}
		if session.room != nil || session.currentRoomID != "" || session.currentGamePlayerID != "" {
			t.Fatal("expected session room state to be cleared after leave")
		}
	})

	t.Run("disconnected leave skips already reported match", func(t *testing.T) {
		session, room, reporter, cleanup := newWebSocketSessionRoomExitTestSetup(t)
		defer cleanup()

		room.MarkMatchResultReported()
		session.leaveDisconnectedRoom()

		if reporter.calls != 0 {
			t.Fatalf("expected reporter to be skipped for already reported match, got %d calls", reporter.calls)
		}
		if room.MemberCount() != 0 {
			t.Fatalf("expected room member to be removed after disconnect, got %d", room.MemberCount())
		}
		if session.room != nil || session.currentRoomID != "" || session.currentGamePlayerID != "" {
			t.Fatal("expected session room state to be cleared after disconnect")
		}
	})
}

func newWebSocketSessionRoomExitTestSetup(t *testing.T) (*webSocketSession, *rooms.Room, *recordingMatchResultReporter, func()) {
	t.Helper()

	manager := rooms.NewRoomManager()
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	sessionID := "session-1"
	room.AddMemberSessionID(sessionID)
	playerID, ok := room.PlayerIDForSession(sessionID)
	if !ok {
		t.Fatal("expected room member to be created")
	}

	gameInstance := game.New()
	if !gameInstance.DevtoolsEnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected devtools player session to be created")
	}
	gameInstance.DevtoolsSetPlayerLives(playerID, 0)

	if err := room.StartSinglePlayerGame(func() *game.Game {
		return gameInstance
	}); err != nil {
		t.Fatalf("start game: %v", err)
	}
	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("mark game over: %v", err)
	}

	reporter := &recordingMatchResultReporter{room: room}
	session := &webSocketSession{
		sessionID:           sessionID,
		currentRoomID:       room.ID,
		currentGamePlayerID: playerID,
		room:                room,
		rooms:               manager,
		matchResultReporter: reporter,
	}

	cleanup := func() {
		gameInstance.Stop()
	}

	return session, room, reporter, cleanup
}
