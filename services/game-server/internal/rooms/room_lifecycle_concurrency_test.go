package rooms

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

func newLifecycleTestRoom(t *testing.T) *Room {
	t.Helper()
	room := NewRoom("room", RoomStateLobby, nil)
	member := room.AddMember(NewRoomMember("session"))
	member.SetReady(true)
	room.Joinable = false
	return room
}

func newStartedLifecycleTestRoom(t *testing.T) *Room {
	t.Helper()
	room := newLifecycleTestRoom(t)
	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}
	gameInstance := room.GameInstance()
	t.Cleanup(func() {
		if gameInstance != nil {
			gameInstance.Stop()
		}
	})
	return room
}

func awaitSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for operation")
	}
}

func TestLifecycleFactoryCanReadRoomWhileStarting(t *testing.T) {
	room := newLifecycleTestRoom(t)
	completed := make(chan struct{})
	var observedState RoomState
	var startErr *RoomDomainError

	go func() {
		startErr = room.StartSinglePlayerGame(func() *game.Game {
			observedState = room.CurrentState()
			return game.New()
		})
		close(completed)
	}()
	awaitSignal(t, completed)
	if startErr != nil {
		t.Fatalf("start game: %v", startErr)
	}
	if observedState != RoomStateStarting {
		t.Fatalf("factory observed state %q", observedState)
	}
}

func TestLifecycleStartCanReadRoomWithoutLock(t *testing.T) {
	room := newLifecycleTestRoom(t)
	originalStartGameCall := startGameCall
	startGameCall = func(gameInstance *game.Game) {
		if room.GameplayContext().State != RoomStateStarting {
			t.Error("start callback did not observe Starting state")
		}
		originalStartGameCall(gameInstance)
	}
	t.Cleanup(func() { startGameCall = originalStartGameCall })

	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}
}

func TestLifecycleStopCanReadRoomWithoutLock(t *testing.T) {
	room := newStartedLifecycleTestRoom(t)
	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("mark game over: %v", err)
	}
	originalStopGameCall := stopGameCall
	var stopCalls atomic.Int32
	stopGameCall = func(gameInstance *game.Game) {
		stopCalls.Add(1)
		if room.CurrentState() != RoomStateLobby {
			t.Error("stop callback did not observe Lobby state")
		}
		originalStopGameCall(gameInstance)
	}
	t.Cleanup(func() { stopGameCall = originalStopGameCall })

	if err := room.ResetToLobby("Player-1"); err != nil {
		t.Fatalf("reset room: %v", err)
	}
	if stopCalls.Load() != 1 {
		t.Fatalf("stop calls = %d, want 1", stopCalls.Load())
	}
}

func TestLifecycleDecisionCanReadRoomWithoutLock(t *testing.T) {
	room := newStartedLifecycleTestRoom(t)
	originalDecisionCall := matchDecisionCall
	originalFinalStateCall := finalMatchStateCall
	matchDecisionCall = func(*game.Game) bool {
		if room.GameplayContext().State != RoomStateInGame {
			t.Error("decision callback observed unexpected state")
		}
		return true
	}
	finalMatchStateCall = func(*game.Game) (game.FinalMatchState, bool) {
		return lifecycleFinalState(nil), true
	}
	t.Cleanup(func() {
		matchDecisionCall = originalDecisionCall
		finalMatchStateCall = originalFinalStateCall
	})

	if !room.MarkGameOverIfComplete() {
		t.Fatal("expected game-over transition")
	}
}

func TestLifecyclePlayerFactsCanReadRoomWithoutLock(t *testing.T) {
	room := newStartedLifecycleTestRoom(t)
	originalFactsCall := playerMatchFactsCall
	playerMatchFactsCall = func(*game.Game) []game.PlayerMatchFact {
		if room.CurrentState() != RoomStateInGame {
			t.Error("facts callback observed unexpected state")
		}
		return nil
	}
	t.Cleanup(func() { playerMatchFactsCall = originalFactsCall })

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("mark game over: %v", err)
	}
}

func TestConcurrentStartHasOneWinner(t *testing.T) {
	room := newLifecycleTestRoom(t)
	var factoryCalls atomic.Int32
	var successfulStarts atomic.Int32
	var waitGroup sync.WaitGroup
	startGate := make(chan struct{})

	for i := 0; i < 8; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-startGate
			err := room.StartSinglePlayerGame(func() *game.Game {
				factoryCalls.Add(1)
				return game.New()
			})
			if err == nil {
				successfulStarts.Add(1)
			}
		}()
	}
	close(startGate)
	waitGroup.Wait()

	if successfulStarts.Load() != 1 || factoryCalls.Load() != 1 {
		t.Fatalf("successful starts = %d, factory calls = %d", successfulStarts.Load(), factoryCalls.Load())
	}
	if room.CurrentState() != RoomStateInGame || room.CurrentMatchID() != "room-match-1" {
		t.Fatalf("state = %q, match = %q", room.CurrentState(), room.CurrentMatchID())
	}
}

func TestConcurrentMarkGameOverIfCompleteTransitionsOnce(t *testing.T) {
	room := newStartedLifecycleTestRoom(t)
	originalDecisionCall := matchDecisionCall
	originalFactsCall := playerMatchFactsCall
	originalFinalStateCall := finalMatchStateCall
	facts := []game.PlayerMatchFact{{GamePlayerID: "player-1", Score: 4}}
	matchDecisionCall = func(*game.Game) bool { return true }
	playerMatchFactsCall = func(*game.Game) []game.PlayerMatchFact { return facts }
	finalMatchStateCall = func(*game.Game) (game.FinalMatchState, bool) {
		return lifecycleFinalState(facts), true
	}
	t.Cleanup(func() {
		matchDecisionCall = originalDecisionCall
		playerMatchFactsCall = originalFactsCall
		finalMatchStateCall = originalFinalStateCall
	})

	var successfulTransitions atomic.Int32
	var waitGroup sync.WaitGroup
	for i := 0; i < 8; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if room.MarkGameOverIfComplete() {
				successfulTransitions.Add(1)
			}
		}()
	}
	waitGroup.Wait()

	if successfulTransitions.Load() != 1 || room.CurrentState() != RoomStateGameOver {
		t.Fatalf("transitions = %d, state = %q", successfulTransitions.Load(), room.CurrentState())
	}
	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("resolved summary missing")
	}
	stableSummary, ok := room.ResolvedMatchSummary()
	if !ok || stableSummary.MatchID != summary.MatchID || len(stableSummary.Players) != len(summary.Players) || stableSummary.Players[0] != summary.Players[0] {
		t.Fatal("resolved summary changed after concurrent completion")
	}
}

func TestMarkGameOverIfCompletePreservesExistingSummary(t *testing.T) {
	room := newStartedLifecycleTestRoom(t)
	presetSummary := playerdata.MatchResultSummary{
		MatchID: room.CurrentMatchID(),
		Players: []playerdata.PlayerMatchSummary{{GamePlayerID: "preset", Score: 9, ShipDeaths: 2}},
	}
	room.mu.Lock()
	room.match.SetResolvedSummary(presetSummary)
	room.mu.Unlock()
	originalDecisionCall := matchDecisionCall
	originalFinalStateCall := finalMatchStateCall
	matchDecisionCall = func(*game.Game) bool { return true }
	finalMatchStateCall = func(*game.Game) (game.FinalMatchState, bool) {
		return lifecycleFinalState(nil), true
	}
	t.Cleanup(func() {
		matchDecisionCall = originalDecisionCall
		finalMatchStateCall = originalFinalStateCall
	})

	if !room.MarkGameOverIfComplete() {
		t.Fatal("expected transition")
	}
	actualSummary, ok := room.ResolvedMatchSummary()
	if !ok || actualSummary.MatchID != presetSummary.MatchID || len(actualSummary.Players) != len(presetSummary.Players) || actualSummary.Players[0] != presetSummary.Players[0] {
		t.Fatalf("preset summary was not preserved: %+v", actualSummary)
	}
}

func TestMarkGameOverNilGameTransitions(t *testing.T) {
	room := newLifecycleTestRoom(t)
	room.mu.Lock()
	room.State = RoomStateInGame
	room.mu.Unlock()

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("mark game over: %v", err)
	}
	if room.CurrentState() != RoomStateGameOver {
		t.Fatalf("state = %q", room.CurrentState())
	}
	if _, ok := room.ResolvedMatchSummary(); ok {
		t.Fatal("nil-game transition created a resolved summary")
	}
}

func lifecycleFinalState(players []game.PlayerMatchFact) game.FinalMatchState {
	decision := rules.MatchDecision{
		IsOver:         true,
		TerminalStatus: rules.TerminalCompleted,
		EndReason:      "simulation_complete",
		Players:        make([]rules.PlayerDecision, 0, len(players)),
	}
	for _, player := range players {
		decision.Players = append(decision.Players, rules.PlayerDecision{
			ID: player.GamePlayerID, Outcome: rules.OutcomeCompleted,
		})
	}
	return game.FinalMatchState{Decision: decision, Players: players}
}

type blockingMatchResultReporter struct {
	callCount atomic.Int32
	entered   chan struct{}
	release   chan struct{}
	err       error
	mutate    bool
}

func (reporter *blockingMatchResultReporter) ReportMatchResult(summary playerdata.MatchResultSummary) error {
	reporter.callCount.Add(1)
	if reporter.entered != nil {
		reporter.entered <- struct{}{}
	}
	if reporter.release != nil {
		<-reporter.release
	}
	if reporter.mutate && len(summary.Players) > 0 {
		summary.Players[0].Score = 99
		summary.Players = append(summary.Players, playerdata.PlayerMatchSummary{GamePlayerID: "mutated"})
	}
	return reporter.err
}

func newReportTestRoom(t *testing.T) (*Room, playerdata.MatchResultSummary) {
	t.Helper()
	room := newStartedLifecycleTestRoom(t)
	summary := playerdata.MatchResultSummary{
		MatchID: room.CurrentMatchID(),
		Players: []playerdata.PlayerMatchSummary{{GamePlayerID: "player-1", Score: 1}},
	}
	room.mu.Lock()
	room.match.SetResolvedSummary(summary)
	room.mu.Unlock()
	return room, summary
}

func TestConcurrentReportClaimOnlyOneReporter(t *testing.T) {
	room, _ := newReportTestRoom(t)
	reporter := &blockingMatchResultReporter{entered: make(chan struct{}, 1), release: make(chan struct{})}
	firstResult := make(chan bool, 1)
	go func() { firstResult <- ReportResolvedMatchResultOnce(room, reporter) }()
	awaitSignal(t, reporter.entered)

	secondaryResults := make(chan bool, 8)
	var waitGroup sync.WaitGroup
	for i := 0; i < 8; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			secondaryResults <- ReportResolvedMatchResultOnce(room, reporter)
		}()
	}
	waitGroup.Wait()
	close(reporter.release)
	if !<-firstResult || reporter.callCount.Load() != 1 {
		t.Fatalf("first report/call count = false/%d", reporter.callCount.Load())
	}
	for i := 0; i < 8; i++ {
		if <-secondaryResults {
			t.Fatal("secondary report succeeded")
		}
	}
}

func TestReportFailureReleasesClaim(t *testing.T) {
	room, _ := newReportTestRoom(t)
	failingReporter := &blockingMatchResultReporter{err: errors.New("report failed")}
	if ReportResolvedMatchResultOnce(room, failingReporter) {
		t.Fatal("failed report returned success")
	}
	if failingReporter.callCount.Load() != 1 || room.MatchResultReported() {
		t.Fatalf("failure calls = %d, reported = %v", failingReporter.callCount.Load(), room.MatchResultReported())
	}
	successfulReporter := &blockingMatchResultReporter{}
	if !ReportResolvedMatchResultOnce(room, successfulReporter) {
		t.Fatal("retry report failed")
	}
	if successfulReporter.callCount.Load() != 1 || !room.MatchResultReported() {
		t.Fatalf("success calls = %d, reported = %v", successfulReporter.callCount.Load(), room.MatchResultReported())
	}
}

func TestReporterMutationDoesNotAliasSummary(t *testing.T) {
	room, originalSummary := newReportTestRoom(t)
	reporter := &blockingMatchResultReporter{mutate: true}
	if !ReportResolvedMatchResultOnce(room, reporter) {
		t.Fatal("report failed")
	}
	actualSummary, ok := room.ResolvedMatchSummary()
	if !ok || actualSummary.MatchID != originalSummary.MatchID || len(actualSummary.Players) != len(originalSummary.Players) || actualSummary.Players[0] != originalSummary.Players[0] {
		t.Fatal("reporter mutated room-owned summary")
	}
}

func TestOldReportCannotCompleteNewMatchClaim(t *testing.T) {
	room, _ := newReportTestRoom(t)
	oldReporter := &blockingMatchResultReporter{entered: make(chan struct{}, 1), release: make(chan struct{})}
	oldResult := make(chan bool, 1)
	go func() { oldResult <- ReportResolvedMatchResultOnce(room, oldReporter) }()
	awaitSignal(t, oldReporter.entered)

	room.mu.Lock()
	room.match.BeginNextMatch(room.ID)
	room.match.SetResolvedSummary(playerdata.MatchResultSummary{MatchID: room.match.CurrentMatchID()})
	room.mu.Unlock()

	newReporter := &blockingMatchResultReporter{entered: make(chan struct{}, 1), release: make(chan struct{})}
	newResult := make(chan bool, 1)
	go func() { newResult <- ReportResolvedMatchResultOnce(room, newReporter) }()
	awaitSignal(t, newReporter.entered)

	close(oldReporter.release)
	if <-oldResult {
		t.Fatal("old report completed successfully")
	}
	if ReportResolvedMatchResultOnce(room, newReporter) {
		t.Fatal("third report succeeded while new claim was blocked")
	}
	if newReporter.callCount.Load() != 1 {
		t.Fatalf("new reporter calls = %d", newReporter.callCount.Load())
	}
	close(newReporter.release)
	if !<-newResult || !room.MatchResultReported() {
		t.Fatal("new report did not complete successfully")
	}
}
