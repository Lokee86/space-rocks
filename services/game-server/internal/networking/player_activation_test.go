package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestActivateRoomPlayersRebindsMemberPlayerIDAndPreservesAccountID(t *testing.T) {
	room := rooms.NewRoom("room", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")

	accountID := "11111111-2222-3333-4444-555555555555"
	if !room.SetMemberAccountIDForSession("session-1", accountID) {
		t.Fatal("expected SetMemberAccountIDForSession to succeed")
	}

	session := &webSocketSession{
		sessionID: "session-1",
		outbound:  make(chan []byte, 1),
		identity:  NewAuthenticatedAccountIdentity(123, accountID, "Ada"),
	}
	attachRoomSession(room, session.sessionID, session)

	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected single-player start to succeed, got %v", err)
	}
	defer func() {
		if gameInstance := room.GameInstance(); gameInstance != nil {
			gameInstance.Stop()
		}
	}()

	activateRoomPlayers(room)

	if session.currentGamePlayerID != "player-1" {
		t.Fatalf("expected current game player id player-1, got %q", session.currentGamePlayerID)
	}
	if playerID, ok := room.PlayerIDForSession("session-1"); !ok || playerID != "player-1" {
		t.Fatalf("expected session-1 to rebind to player-1, got %q ok=%v", playerID, ok)
	}
	if ownerID := room.OwnerID(); ownerID != "player-1" {
		t.Fatalf("expected owner id player-1 after activation, got %q", ownerID)
	}
	if count := room.ActivePlayerCount(); count != 1 {
		t.Fatalf("expected active player count 1, got %d", count)
	}

	gameInstance := room.GameInstance()
	gameInstance.SetPlayerScore("player-1", 120)
	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to be stored")
	}
	if len(summary.Players) != 1 {
		t.Fatalf("expected 1 player summary, got %d", len(summary.Players))
	}
	player := summary.Players[0]
	if player.GamePlayerID != "player-1" {
		t.Fatalf("expected GamePlayerID player-1, got %q", player.GamePlayerID)
	}
	if player.AccountID != accountID {
		t.Fatalf("expected AccountID %q, got %q", accountID, player.AccountID)
	}
}
