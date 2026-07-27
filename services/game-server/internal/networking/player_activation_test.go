package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
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
	session.bindRoom(room)

	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected single-player start to succeed, got %v", err)
	}
	defer func() {
		if gameInstance := room.GameInstance(); gameInstance != nil {
			gameInstance.Stop()
		}
	}()

	activateRoomPlayers(room)

	if session.sessionContext().GamePlayerID != "player-1" {
		t.Fatalf("expected current game player id player-1, got %q", session.sessionContext().GamePlayerID)
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

func TestActivateRoomPlayersRollsBackWhenRoomActivationFails(t *testing.T) {
	room := rooms.NewRoom("room", rooms.RoomStateLobby, nil)
	member := room.AddMember(rooms.NewRoomMember("session-1"))
	originalPlayerID := member.PlayerID

	session := &webSocketSession{sessionID: "session-1", outbound: make(chan []byte, 1)}
	attachRoomSession(room, session.sessionID, session)
	t.Cleanup(func() { detachRoomSession(room, session.sessionID) })
	session.bindRoom(room)

	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("expected single-player start to succeed, got %v", err)
	}
	t.Cleanup(func() {
		if gameInstance := room.GameInstance(); gameInstance != nil {
			gameInstance.Stop()
		}
	})

	originalActivateMemberPlayerCall := activateMemberPlayerCall
	activateMemberPlayerCall = func(*rooms.Room, rooms.GameplayContext, string, string) bool { return false }
	t.Cleanup(func() { activateMemberPlayerCall = originalActivateMemberPlayerCall })

	activateRoomPlayers(room)
	if got := session.sessionContext().GamePlayerID; got != "" {
		t.Fatalf("expected empty game player ID after rollback, got %q", got)
	}
	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session-1 player lookup to succeed")
	}
	if playerID != originalPlayerID {
		t.Fatalf("expected original player ID %q, got %q", originalPlayerID, playerID)
	}
	if got := room.ActivePlayerCount(); got != 0 {
		t.Fatalf("expected active player count 0, got %d", got)
	}
	if facts := room.GameInstance().PlayerMatchFacts(); len(facts) != 0 {
		t.Fatalf("expected rolled-back game player facts to be empty, got %d", len(facts))
	}
	if decision := room.GameInstance().MatchDecision(); decision.IsOver {
		t.Fatalf("expected rolled-back provisional activation not to complete match, got %+v", decision)
	}
}

func TestActivateRoomPlayersRestoresBotsAcrossConsecutiveMatches(t *testing.T) {
	room := rooms.NewRoom("room", rooms.RoomStateLobby, nil)
	owner := room.AddMember(rooms.NewRoomMember("session-owner"))
	firstBot, roomErr := room.AddBotForOwnerSession(owner.SessionID)
	if roomErr != nil {
		t.Fatalf("add first bot: %v", roomErr)
	}
	secondBot, roomErr := room.AddBotForOwnerSession(owner.SessionID)
	if roomErr != nil {
		t.Fatalf("add second bot: %v", roomErr)
	}
	owner.SetReady(true)

	session := &webSocketSession{sessionID: owner.SessionID, outbound: make(chan []byte, 1)}
	attachRoomSession(room, session.sessionID, session)
	t.Cleanup(func() {
		detachRoomSession(room, session.sessionID)
		if gameInstance := room.GameInstance(); gameInstance != nil {
			gameInstance.Stop()
		}
	})
	session.bindRoom(room)

	if err := room.StartGameForSession(owner.SessionID, game.New); err != nil {
		t.Fatalf("start first game: %v", err)
	}
	activateRoomPlayers(room)
	if facts := room.GameInstance().PlayerMatchFacts(); len(facts) != 3 {
		t.Fatalf("first-match facts = %#v, want owner and two bots", facts)
	}

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("mark first game over: %v", err)
	}
	if err := room.ResetToLobbyForSession(owner.SessionID); err != nil {
		t.Fatalf("return to lobby: %v", err)
	}
	deactivateRoomPlayers(room)

	for _, botSessionID := range []string{firstBot.SessionID, secondBot.SessionID} {
		playerID, ok := room.PlayerIDForSession(botSessionID)
		if !ok || playerID == "" || playerID[:1] != "P" {
			t.Fatalf("expected restored lobby bot ID for %q, got %q ok=%v", botSessionID, playerID, ok)
		}
	}

	owner.SetReady(true)
	if err := room.StartGameForSession(owner.SessionID, game.New); err != nil {
		t.Fatalf("start second game: %v", err)
	}
	activateRoomPlayers(room)

	facts := room.GameInstance().PlayerMatchFacts()
	if len(facts) != 3 {
		t.Fatalf("second-match facts = %#v, want owner and two bots", facts)
	}
	factIDs := make(map[string]bool, len(facts))
	for _, fact := range facts {
		factIDs[fact.GamePlayerID] = true
	}
	for _, botSessionID := range []string{firstBot.SessionID, secondBot.SessionID} {
		playerID, ok := room.PlayerIDForSession(botSessionID)
		if !ok || !factIDs[playerID] {
			t.Fatalf("bot %q resolved to %q but was absent from second match facts %#v", botSessionID, playerID, facts)
		}
	}
}

func TestActivateRoomPlayersCountsSameIdentityAcrossConsecutiveMatches(t *testing.T) {
	room := rooms.NewRoom("room", rooms.RoomStateLobby, nil)
	room.AddMember(rooms.NewRoomMember("session-1"))

	session := &webSocketSession{sessionID: "session-1", outbound: make(chan []byte, 1)}
	attachRoomSession(room, session.sessionID, session)
	t.Cleanup(func() {
		detachRoomSession(room, session.sessionID)
		if gameInstance := room.GameInstance(); gameInstance != nil {
			gameInstance.Stop()
		}
	})
	session.bindRoom(room)

	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("expected first single-player start to succeed, got %v", err)
	}
	firstMatchID := room.CurrentMatchID()
	if firstMatchID == "" {
		t.Fatal("expected first match ID to be nonempty")
	}
	activateRoomPlayers(room)
	if got := session.sessionContext().GamePlayerID; got != "player-1" {
		t.Fatalf("expected session game player ID player-1, got %q", got)
	}
	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok || playerID != "player-1" {
		t.Fatalf("expected room player ID player-1, got %q ok=%v", playerID, ok)
	}
	if got := room.ActivePlayerCount(); got != 1 {
		t.Fatalf("expected first active player count 1, got %d", got)
	}
	if facts := room.GameInstance().PlayerMatchFacts(); len(facts) != 1 {
		t.Fatalf("expected one first-match player fact, got %d", len(facts))
	}

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}
	if err := room.ResetToLobby("player-1"); err != nil {
		t.Fatalf("expected lobby reset to succeed, got %v", err)
	}
	deactivateRoomPlayers(room)
	if got := session.sessionContext().GamePlayerID; got != "" {
		t.Fatalf("expected empty session game player ID after deactivation, got %q", got)
	}
	if got := room.ActivePlayerCount(); got != 0 {
		t.Fatalf("expected active player count 0, got %d", got)
	}
	playerID, ok = room.PlayerIDForSession("session-1")
	if !ok || playerID != "Player-1" {
		t.Fatalf("expected lobby player ID Player-1 after reset, got %q ok=%v", playerID, ok)
	}

	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("expected second single-player start to succeed, got %v", err)
	}
	secondMatchID := room.CurrentMatchID()
	if secondMatchID == "" || secondMatchID == firstMatchID {
		t.Fatalf("expected distinct nonempty second match ID, got %q (first %q)", secondMatchID, firstMatchID)
	}
	activateRoomPlayers(room)
	if got := session.sessionContext().GamePlayerID; got != "player-1" {
		t.Fatalf("expected second session game player ID player-1, got %q", got)
	}
	playerID, ok = room.PlayerIDForSession("session-1")
	if !ok || playerID != "player-1" {
		t.Fatalf("expected second room player ID player-1, got %q ok=%v", playerID, ok)
	}
	if got := room.ActivePlayerCount(); got != 1 {
		t.Fatalf("expected second active player count 1, got %d", got)
	}
	facts := room.GameInstance().PlayerMatchFacts()
	if len(facts) != 1 || facts[0].GamePlayerID != "player-1" {
		t.Fatalf("expected one second-match fact for player-1, got %#v", facts)
	}
}
