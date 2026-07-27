package game

import (
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestGameplayPresentationSnapshotSetsPublishedServerTimestamp(t *testing.T) {
	game := New()

	before := time.Now().UnixMilli()
	playerID := game.AddPlayer()
	after := time.Now().UnixMilli()
	snapshot := game.GameplayPresentationSnapshot(playerID)

	if snapshot.ServerSentMsec <= 0 {
		t.Fatalf("expected positive server timestamp, got %d", snapshot.ServerSentMsec)
	}
	if snapshot.ServerSentMsec < int(before) || snapshot.ServerSentMsec > int(after) {
		t.Fatalf("expected timestamp %d within [%d, %d]", snapshot.ServerSentMsec, before, after)
	}
}

func TestGameplayPresentationSnapshotIncludesAuthoritativeTeamID(t *testing.T) {
	game := New()
	playerID := game.AddPlayerWithTeam(teams.Team3)

	snapshot := game.GameplayPresentationSnapshot(playerID)
	if snapshot.Players[playerID].TeamID != string(teams.Team3) {
		t.Fatalf("team ID = %q, want %q", snapshot.Players[playerID].TeamID, teams.Team3)
	}
}

func TestGameplayPresentationSnapshotPublishesOnAddAndReusesFrame(t *testing.T) {
	game := New()
	first := game.AddPlayer()
	second := game.AddPlayer()

	game.mu.Lock()
	frame := game.presentationFrame
	generation := frame.generation
	timestamp := frame.serverSentMsec
	game.mu.Unlock()

	firstSnapshot := game.GameplayPresentationSnapshot(first)
	secondSnapshot := game.GameplayPresentationSnapshot(second)

	game.mu.Lock()
	defer game.mu.Unlock()
	if game.presentationFrame != frame {
		t.Fatal("receiver snapshots caused an unexpected publication")
	}
	if len(frame.players) != 2 {
		t.Fatalf("expected both players in published frame, got %d", len(frame.players))
	}
	if firstSnapshot.ServerSentMsec != timestamp || secondSnapshot.ServerSentMsec != timestamp {
		t.Fatal("receiver snapshots did not use the same published timestamp")
	}
	if generation == 0 || game.presentationFrame.generation != generation {
		t.Fatal("receiver snapshots changed the published generation")
	}
}

func TestGameplayPresentationFrameRemainsImmutableAfterStep(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.mu.Lock()
	oldFrame := game.presentationFrame
	oldState := oldFrame.players[playerID]
	player := game.entities.Players[playerID]
	player.X += 100
	game.mu.Unlock()

	game.Step(1.0 / 60.0)

	game.mu.Lock()
	newFrame := game.presentationFrame
	game.mu.Unlock()
	if oldFrame.players[playerID] != oldState {
		t.Fatal("old published frame changed after live state mutation")
	}
	if newFrame.generation <= oldFrame.generation {
		t.Fatal("expected a later published generation")
	}
	if newFrame.players[playerID] == oldState {
		t.Fatal("replacement frame did not contain changed player state")
	}
}

func TestGameplayPresentationFramePublishesOnRemove(t *testing.T) {
	game := New()
	removed := game.AddPlayer()
	remaining := game.AddPlayer()

	game.mu.Lock()
	oldGeneration := game.presentationFrame.generation
	game.mu.Unlock()
	game.RemovePlayer(removed)

	game.mu.Lock()
	frame := game.presentationFrame
	game.mu.Unlock()
	if frame.generation <= oldGeneration {
		t.Fatal("expected RemovePlayer to publish a newer frame")
	}
	if _, ok := frame.players[removed]; ok {
		t.Fatal("removed player remained in published players")
	}
	if _, ok := frame.playerSessions[removed]; ok {
		t.Fatal("removed player remained in published sessions")
	}
	if _, ok := frame.players[remaining]; !ok {
		t.Fatal("remaining player was removed from published players")
	}
}

func TestGameplayPresentationSnapshotPendingEventsAreReceiverSpecificCopies(t *testing.T) {
	game := New()
	first := game.AddPlayer()
	second := game.AddPlayer()

	game.mu.Lock()
	game.pendingPresentationEvents[first] = []PendingPresentationEvent{{EventID: "first"}}
	game.pendingPresentationEvents[second] = []PendingPresentationEvent{{EventID: "second"}}
	game.mu.Unlock()

	firstSnapshot := game.GameplayPresentationSnapshot(first)
	secondSnapshot := game.GameplayPresentationSnapshot(second)
	firstSnapshot.PendingEvents[0].EventID = "changed"
	firstSnapshot.PendingEvents = append(firstSnapshot.PendingEvents, PendingPresentationEvent{EventID: "extra"})

	game.mu.Lock()
	defer game.mu.Unlock()
	if firstSnapshot.PendingEvents[0].EventID == game.pendingPresentationEvents[first][0].EventID {
		t.Fatal("first snapshot shared its pending event backing data")
	}
	if len(game.pendingPresentationEvents[first]) != 1 || len(game.pendingPresentationEvents[second]) != 1 {
		t.Fatal("snapshot mutation changed authoritative queues")
	}
	if secondSnapshot.PendingEvents[0].EventID != "second" {
		t.Fatal("receiver pending events were not isolated")
	}
}

func TestGameplayPresentationSnapshotConcurrentReadsDuringPublication(t *testing.T) {
	game := New()
	playerIDs := []string{game.AddPlayer(), game.AddPlayer(), game.AddPlayer(), game.AddPlayer()}
	const readers = 4
	const snapshotsPerReader = 40
	const publications = 40

	var wg sync.WaitGroup
	wg.Add(readers)
	for reader := 0; reader < readers; reader++ {
		playerID := playerIDs[reader%len(playerIDs)]
		go func() {
			defer wg.Done()
			for i := 0; i < snapshotsPerReader; i++ {
				snapshot := game.GameplayPresentationSnapshot(playerID)
				if _, ok := snapshot.Players[playerID]; !ok {
					t.Errorf("snapshot omitted reader player %q", playerID)
				}
				if _, ok := snapshot.PlayerSessions[playerID]; !ok {
					t.Errorf("snapshot omitted reader session %q", playerID)
				}
			}
		}()
	}

	for i := 0; i < publications; i++ {
		game.Step(0)
	}
	wg.Wait()
}
