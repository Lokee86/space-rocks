package rooms

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func newMembershipTestRoom(t *testing.T, memberCount int) (*Room, GameplayContext) {
	t.Helper()
	room := NewRoom("membership-room", RoomStateLobby, nil)
	for index := 0; index < memberCount; index++ {
		member := room.AddMember(NewRoomMember(fmt.Sprintf("session-%d", index+1)))
		member.SetReady(true)
	}
	room.Joinable = false
	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}
	gameplayContext := room.GameplayContext()
	t.Cleanup(func() {
		if gameplayContext.Game != nil {
			gameplayContext.Game.Stop()
		}
	})
	return room, gameplayContext
}

func TestActivateMemberPlayerRekeysMemberOwnerAndCountsOnce(t *testing.T) {
	room, context := newMembershipTestRoom(t, 1)
	if !room.ActivateMemberPlayer(context, "session-1", "player-1") {
		t.Fatal("activation failed")
	}
	if room.OwnerID() != "player-1" || room.ActivePlayerCount() != 1 {
		t.Fatalf("owner/count = %q/%d", room.OwnerID(), room.ActivePlayerCount())
	}
}

func TestActivateMemberPlayerRepeatedSameMatchIsIdempotent(t *testing.T) {
	room, context := newMembershipTestRoom(t, 1)
	if !room.ActivateMemberPlayer(context, "session-1", "player-1") || !room.ActivateMemberPlayer(context, "session-1", "player-1") {
		t.Fatal("activation failed")
	}
	if room.ActivePlayerCount() != 1 {
		t.Fatalf("count=%d", room.ActivePlayerCount())
	}
}

func TestActivateMemberPlayerSameIDCountsInNextMatch(t *testing.T) {
	room, context := newMembershipTestRoom(t, 1)
	if !room.ActivateMemberPlayer(context, "session-1", "player-1") {
		t.Fatal("first activation failed")
	}
	room.mu.Lock()
	room.match.BeginNextMatch(room.ID)
	room.mu.Unlock()
	if !room.ActivateMemberPlayer(room.GameplayContext(), "session-1", "player-1") {
		t.Fatal("second activation failed")
	}
	if room.ActivePlayerCount() != 1 {
		t.Fatalf("count=%d", room.ActivePlayerCount())
	}
}

func TestActivateMemberPlayerRejectsStaleGameplayContext(t *testing.T) {
	room, context := newMembershipTestRoom(t, 1)
	stale := context
	room.SetGameInstance(game.New())
	if room.ActivateMemberPlayer(stale, "session-1", "player-1") {
		t.Fatal("stale context accepted")
	}
	if room.ActivePlayerCount() != 0 {
		t.Fatal("count mutated")
	}
}

func TestActivateMemberPlayerRejectsPlayerIDCollision(t *testing.T) {
	room, context := newMembershipTestRoom(t, 2)
	if !room.ActivateMemberPlayer(context, "session-1", "player-1") || room.ActivateMemberPlayer(context, "session-2", "player-1") {
		t.Fatal("collision accepted")
	}
}

func TestActivateMemberPlayerConcurrentDistinctMembersHasExactCount(t *testing.T) {
	room, context := newMembershipTestRoom(t, 8)
	var wait sync.WaitGroup
	for index := 1; index <= 8; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			if !room.ActivateMemberPlayer(context, fmt.Sprintf("session-%d", index), fmt.Sprintf("player-%d", index)) {
				t.Errorf("activation %d failed", index)
			}
		}(index)
	}
	wait.Wait()
	if room.ActivePlayerCount() != 8 {
		t.Fatalf("count=%d", room.ActivePlayerCount())
	}
}

func TestRemoveMemberForSessionReturnsRemovedMemberAndReassignsOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMemberSessionID("one")
	room.AddMemberSessionID("two")
	removed, remaining, ok := room.RemoveMemberForSession("one")
	if !ok || removed.SessionID != "one" || remaining != 1 || room.OwnerID() != "Player-2" {
		t.Fatalf("removed=%+v remaining=%d ok=%v owner=%s", removed, remaining, ok, room.OwnerID())
	}
}

func TestDecrementActivePlayerCountNeverGoesNegativeConcurrently(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.SetActivePlayerCount(3)
	var wait sync.WaitGroup
	for index := 0; index < 16; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if room.DecrementActivePlayerCount() < 0 {
				t.Error("negative count")
			}
		}()
	}
	wait.Wait()
	if room.ActivePlayerCount() != 0 {
		t.Fatalf("count=%d", room.ActivePlayerCount())
	}
}

func TestDeactivateMemberPlayerIsIdempotent(t *testing.T) {
	room, context := newMembershipTestRoom(t, 1)
	if !room.ActivateMemberPlayer(context, "session-1", "player-1") {
		t.Fatal("activate failed")
	}
	if !room.DeactivateMemberPlayer("session-1") || room.DeactivateMemberPlayer("session-1") || room.ActivePlayerCount() != 0 {
		t.Fatal("deactivation not idempotent")
	}
}

func TestLeaveMemberUsesAtomicallyRemovedPlayerIdentity(t *testing.T) {
	manager := NewRoomManager()
	defer manager.StopAll()
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	room.AddMemberSessionID("session-1")
	room.AddMemberSessionID("session-2")
	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}
	gameInstance := room.GameInstance()
	playerOne, playerTwo := gameInstance.AddPlayer(), gameInstance.AddPlayer()
	if !room.ActivateMemberPlayer(room.GameplayContext(), "session-1", playerOne) || !room.ActivateMemberPlayer(room.GameplayContext(), "session-2", playerTwo) {
		t.Fatal("activate players")
	}
	result, leaveErr := manager.LeaveMember(room.ID, "session-1", playerTwo)
	if leaveErr != nil {
		t.Fatalf("leave: %v", leaveErr)
	}
	if result.PlayerID != playerOne || !result.PlayerRemoved || result.ActivePlayers != 1 {
		t.Fatalf("result=%+v", result)
	}
	facts := gameInstance.PlayerMatchFacts()
	if len(facts) != 1 || facts[0].GamePlayerID != playerTwo {
		t.Fatalf("facts=%+v", facts)
	}
	if mapped, ok := room.PlayerIDForSession("session-2"); !ok || mapped != playerTwo {
		t.Fatalf("mapping=%q ok=%v", mapped, ok)
	}
}
