package rooms

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/playerdata"
)

type fakeMatchResultReporter struct {
	calls       int
	lastSummary playerdata.MatchResultSummary
	err         error
}

func (r *fakeMatchResultReporter) ReportMatchResult(summary playerdata.MatchResultSummary) error {
	r.calls++
	r.lastSummary = summary
	return r.err
}

func TestTickRoomGameOverLifecycleTransitionsFinishedGameAndBroadcasts(t *testing.T) {
	finishedGame := game.New()
	markLifecycleTickTestGameOver(t, finishedGame)
	room := NewRoom("room", RoomStateInGame, finishedGame)
	broadcasts := 0

	if !TickRoomGameOverLifecycle(room, func(broadcastRoom *Room) {
		broadcasts++
		if broadcastRoom != room {
			t.Fatal("expected transitioned room to be broadcast")
		}
	}) {
		t.Fatal("expected finished room lifecycle tick to transition")
	}

	if room.State != RoomStateGameOver {
		t.Fatalf("expected room state %q, got %q", RoomStateGameOver, room.State)
	}
	if broadcasts != 1 {
		t.Fatalf("expected 1 broadcast, got %d", broadcasts)
	}
}

func TestTickRoomGameOverLifecycleDoesNotBroadcastWithoutTransition(t *testing.T) {
	activeGame := game.New()
	activeGame.AddPlayer()
	room := NewRoom("room", RoomStateInGame, activeGame)
	broadcasts := 0

	if TickRoomGameOverLifecycle(room, func(*Room) {
		broadcasts++
	}) {
		t.Fatal("expected active room lifecycle tick not to transition")
	}

	if room.State != RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", RoomStateInGame, room.State)
	}
	if broadcasts != 0 {
		t.Fatalf("expected no broadcast, got %d", broadcasts)
	}
}

func TestReportResolvedMatchResultOnceReturnsFalseForNilRoom(t *testing.T) {
	if ReportResolvedMatchResultOnce(nil, &fakeMatchResultReporter{}) {
		t.Fatal("expected nil room to return false")
	}
}

func TestReportResolvedMatchResultOnceReturnsFalseWithoutSummary(t *testing.T) {
	room := NewRoom("room", RoomStateGameOver, nil)

	if ReportResolvedMatchResultOnce(room, &fakeMatchResultReporter{}) {
		t.Fatal("expected room without resolved summary to return false")
	}
}

func TestReportResolvedMatchResultOnceReportsAndMarksOnce(t *testing.T) {
	room := NewRoom("room", RoomStateGameOver, nil)
	room.match.SetResolvedSummary(playerdata.MatchResultSummary{
		MatchID: "room-match-1",
		Players: []playerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
				Score:        42,
			},
		},
	})
	reporter := &fakeMatchResultReporter{}

	if !ReportResolvedMatchResultOnce(room, reporter) {
		t.Fatal("expected successful report to return true")
	}
	if reporter.calls != 1 {
		t.Fatalf("expected reporter to be called once, got %d", reporter.calls)
	}
	if !room.MatchResultReported() {
		t.Fatal("expected room to be marked as reported")
	}
}

func TestReportResolvedMatchResultOnceReturnsFalseAfterSuccess(t *testing.T) {
	room := NewRoom("room", RoomStateGameOver, nil)
	room.match.SetResolvedSummary(playerdata.MatchResultSummary{
		MatchID: "room-match-1",
		Players: []playerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
			},
		},
	})
	reporter := &fakeMatchResultReporter{}

	if !ReportResolvedMatchResultOnce(room, reporter) {
		t.Fatal("expected first report to succeed")
	}
	if ReportResolvedMatchResultOnce(room, reporter) {
		t.Fatal("expected second report attempt to return false")
	}
	if reporter.calls != 1 {
		t.Fatalf("expected reporter to be called once, got %d", reporter.calls)
	}
}

func TestReportResolvedMatchResultOnceReturnsFalseOnReporterError(t *testing.T) {
	room := NewRoom("room", RoomStateGameOver, nil)
	room.match.SetResolvedSummary(playerdata.MatchResultSummary{
		MatchID: "room-match-1",
		Players: []playerdata.PlayerMatchSummary{
			{
				GamePlayerID: "Player-1",
			},
		},
	})
	reporter := &fakeMatchResultReporter{err: errors.New("report failed")}

	if ReportResolvedMatchResultOnce(room, reporter) {
		t.Fatal("expected reporter error to return false")
	}
	if reporter.calls != 1 {
		t.Fatalf("expected reporter to be called once, got %d", reporter.calls)
	}
	if room.MatchResultReported() {
		t.Fatal("expected room to remain unreported after reporter error")
	}
}

func TestRoomGameOverLifecycleReportsMatchResultOnce(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))

	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected room start to succeed, got %v", err)
	}

	gameInstance := room.GameInstance()
	markLifecycleTickTestGameOver(t, gameInstance)

	broadcasts := 0
	if !TickRoomGameOverLifecycle(room, func(broadcastRoom *Room) {
		broadcasts++
		if broadcastRoom != room {
			t.Fatal("expected transitioned room to be broadcast")
		}
	}) {
		t.Fatal("expected room game-over lifecycle to advance")
	}
	if broadcasts != 1 {
		t.Fatalf("expected 1 broadcast, got %d", broadcasts)
	}

	reporter := &fakeMatchResultReporter{}
	if !ReportResolvedMatchResultOnce(room, reporter) {
		t.Fatal("expected first report attempt to succeed")
	}
	if reporter.calls != 1 {
		t.Fatalf("expected reporter to be called once, got %d", reporter.calls)
	}
	if reporter.lastSummary.MatchID == "" {
		t.Fatal("expected reporter to capture match summary")
	}
	if len(reporter.lastSummary.Players) != 1 {
		t.Fatalf("expected 1 player summary, got %d", len(reporter.lastSummary.Players))
	}
	if !room.MatchResultReported() {
		t.Fatal("expected room to be marked as reported")
	}

	if ReportResolvedMatchResultOnce(room, reporter) {
		t.Fatal("expected second report attempt to return false")
	}
	if reporter.calls != 1 {
		t.Fatalf("expected reporter to still be called once, got %d", reporter.calls)
	}

	room.GameInstance().Stop()
}

func markLifecycleTickTestGameOver(t *testing.T, gameInstance *game.Game) {
	t.Helper()

	playerID := gameInstance.AddPlayer()
	value := reflect.ValueOf(gameInstance).Elem()
	session := exportLifecycleTickTestValue(value.FieldByName("playerSessions")).
		MapIndex(reflect.ValueOf(playerID))
	exportLifecycleTickTestValue(session.Elem().FieldByName("Lives")).SetInt(0)
	players := exportLifecycleTickTestValue(value.FieldByName("entities").FieldByName("Players"))
	players.SetMapIndex(reflect.ValueOf(playerID), reflect.Value{})
}

func exportLifecycleTickTestValue(value reflect.Value) reflect.Value {
	if value.CanSet() {
		return value
	}

	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem()
}
