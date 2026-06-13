package rooms

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/playerdata"
)

func TestGuestSinglePlayerResolvedMatchSummary(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.SetJoinable(false)
	member := room.AddMember(NewRoomMember("session-owner"))

	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected single-player start to succeed, got %v", err)
	}

	gameInstance := room.GameInstance()
	playerID := gameInstance.AddPlayer()
	remapLifecycleTickTestPlayerID(t, gameInstance, playerID, member.PlayerID)
	pruneLifecycleTickTestPlayers(t, gameInstance, member.PlayerID)
	gameInstance.SetPlayerScore(member.PlayerID, 175)
	markLifecycleTickTestGameOver(t, gameInstance)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to be stored")
	}
	if summary.Mode != playerdata.MatchModeSinglePlayer {
		t.Fatalf("expected mode %q, got %q", playerdata.MatchModeSinglePlayer, summary.Mode)
	}
	foundGuestPlayer := false
	for _, player := range summary.Players {
		if player.GamePlayerID != member.PlayerID {
			continue
		}

		foundGuestPlayer = true
		if player.Score != 175 {
			t.Fatalf("expected score 175, got %d", player.Score)
		}
		if player.Won {
			t.Fatal("expected single-player summary winner flag to be false")
		}
		if player.AccountID != "" {
			t.Fatalf("expected empty AccountID, got %q", player.AccountID)
		}
		if player.LocalProfileID != "" {
			t.Fatalf("expected empty LocalProfileID, got %q", player.LocalProfileID)
		}
	}
	if !foundGuestPlayer {
		t.Fatalf("expected guest player summary for %q", member.PlayerID)
	}
	for _, player := range summary.Players {
		if player.AccountID != "" || player.LocalProfileID != "" {
			t.Fatalf("expected guest identities to stay empty, got %+v", player)
		}
		if player.Won {
			t.Fatalf("expected guest winner flags to be false, got %+v", player)
		}
	}

	room.GameInstance().Stop()
}

func TestSinglePlayerResolvedMatchSummaryCopiesLocalProfileID(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.SetJoinable(false)
	member := room.AddMember(NewRoomMember("session-owner"))

	localProfileID := "local-profile-1"
	if !room.SetMemberLocalProfileIDForSession("session-owner", localProfileID) {
		t.Fatal("expected SetMemberLocalProfileIDForSession to succeed")
	}

	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected single-player start to succeed, got %v", err)
	}

	gameInstance := room.GameInstance()
	playerID := gameInstance.AddPlayer()
	remapLifecycleTickTestPlayerID(t, gameInstance, playerID, member.PlayerID)
	pruneLifecycleTickTestPlayers(t, gameInstance, member.PlayerID)
	gameInstance.SetPlayerScore(member.PlayerID, 175)
	markLifecycleTickTestGameOver(t, gameInstance)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to be stored")
	}
	foundPlayer := false
	for _, player := range summary.Players {
		if player.GamePlayerID != member.PlayerID {
			continue
		}

		foundPlayer = true
		if player.LocalProfileID != localProfileID {
			t.Fatalf("expected LocalProfileID %q, got %q", localProfileID, player.LocalProfileID)
		}
		if player.AccountID != "" {
			t.Fatalf("expected empty AccountID, got %q", player.AccountID)
		}
	}
	if !foundPlayer {
		t.Fatalf("expected summary for member %q", member.PlayerID)
	}

	room.GameInstance().Stop()
}

func TestMultiplayerResolvedMatchSummarySelectsUniqueWinner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	if err := room.StartGameForMember(owner.PlayerID, func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected multiplayer start to succeed, got %v", err)
	}

	gameInstance := room.GameInstance()
	player1 := gameInstance.AddPlayer()
	player2 := gameInstance.AddPlayer()
	remapLifecycleTickTestPlayerID(t, gameInstance, player1, owner.PlayerID)
	remapLifecycleTickTestPlayerID(t, gameInstance, player2, peer.PlayerID)
	gameInstance.SetPlayerScore(owner.PlayerID, 120)
	gameInstance.SetPlayerScore(peer.PlayerID, 250)
	markLifecycleTickTestGameOverForAllPlayers(t, gameInstance)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to be stored")
	}
	if summary.Mode != playerdata.MatchModeMultiplayer {
		t.Fatalf("expected mode %q, got %q", playerdata.MatchModeMultiplayer, summary.Mode)
	}
	if len(summary.Players) != 2 {
		t.Fatalf("expected 2 player summaries, got %d", len(summary.Players))
	}

	playersByID := map[string]playerdata.PlayerMatchSummary{}
	for _, player := range summary.Players {
		playersByID[player.GamePlayerID] = player
	}

	playerSummary1, ok := playersByID[owner.PlayerID]
	if !ok {
		t.Fatalf("expected summary for %q", owner.PlayerID)
	}
	if playerSummary1.Score != 120 {
		t.Fatalf("expected score 120 for %q, got %d", owner.PlayerID, playerSummary1.Score)
	}
	if playerSummary1.Won {
		t.Fatalf("expected %q to lose", owner.PlayerID)
	}

	playerSummary2, ok := playersByID[peer.PlayerID]
	if !ok {
		t.Fatalf("expected summary for %q", peer.PlayerID)
	}
	if playerSummary2.Score != 250 {
		t.Fatalf("expected score 250 for %q, got %d", peer.PlayerID, playerSummary2.Score)
	}
	if !playerSummary2.Won {
		t.Fatalf("expected %q to win", peer.PlayerID)
	}

	room.GameInstance().Stop()
}

func TestMultiplayerResolvedMatchSummaryClearsTiedWinners(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	if err := room.StartGameForMember(owner.PlayerID, func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected multiplayer start to succeed, got %v", err)
	}

	gameInstance := room.GameInstance()
	player1 := gameInstance.AddPlayer()
	player2 := gameInstance.AddPlayer()
	remapLifecycleTickTestPlayerID(t, gameInstance, player1, owner.PlayerID)
	remapLifecycleTickTestPlayerID(t, gameInstance, player2, peer.PlayerID)
	gameInstance.SetPlayerScore(owner.PlayerID, 250)
	gameInstance.SetPlayerScore(peer.PlayerID, 250)
	markLifecycleTickTestGameOverForAllPlayers(t, gameInstance)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to be stored")
	}
	if summary.Mode != playerdata.MatchModeMultiplayer {
		t.Fatalf("expected mode %q, got %q", playerdata.MatchModeMultiplayer, summary.Mode)
	}
	if len(summary.Players) != 2 {
		t.Fatalf("expected 2 player summaries, got %d", len(summary.Players))
	}
	for _, player := range summary.Players {
		if player.Won {
			t.Fatalf("expected no winners for tied high score, got %+v", player)
		}
	}

	room.GameInstance().Stop()
}

func TestMultiplayerResolvedMatchSummaryCopiesAccountIDFromRoomMember(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	accountID := "11111111-2222-3333-4444-555555555555"
	if !room.SetMemberAccountIDForSession("session-owner", accountID) {
		t.Fatal("expected SetMemberAccountIDForSession to succeed")
	}

	if err := room.StartGameForMember(owner.PlayerID, func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected multiplayer start to succeed, got %v", err)
	}

	gameInstance := room.GameInstance()
	player1 := gameInstance.AddPlayer()
	player2 := gameInstance.AddPlayer()
	remapLifecycleTickTestPlayerID(t, gameInstance, player1, owner.PlayerID)
	remapLifecycleTickTestPlayerID(t, gameInstance, player2, peer.PlayerID)
	gameInstance.SetPlayerScore(owner.PlayerID, 120)
	gameInstance.SetPlayerScore(peer.PlayerID, 250)
	markLifecycleTickTestGameOverForAllPlayers(t, gameInstance)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to be stored")
	}

	found := false
	for _, player := range summary.Players {
		if player.GamePlayerID != owner.PlayerID {
			continue
		}

		found = true
		if player.AccountID != accountID {
			t.Fatalf("expected AccountID %q, got %q", accountID, player.AccountID)
		}
		if player.LocalProfileID != "" {
			t.Fatalf("expected empty LocalProfileID, got %q", player.LocalProfileID)
		}
	}
	if !found {
		t.Fatalf("expected summary for room member %q", owner.PlayerID)
	}

	room.GameInstance().Stop()
}

func TestMultiplayerResolvedMatchSummaryCopiesAccountIDAfterPlayerIDRekey(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))
	accountID := "11111111-2222-3333-4444-555555555555"

	if !room.SetMemberAccountIDForSession("session-owner", accountID) {
		t.Fatal("expected SetMemberAccountIDForSession to succeed")
	}
	if !room.SetMemberPlayerIDForSession("session-owner", "player-1") {
		t.Fatal("expected SetMemberPlayerIDForSession to succeed")
	}
	if ownerID := room.OwnerID(); ownerID != "player-1" {
		t.Fatalf("expected OwnerID player-1, got %q", ownerID)
	}
	if err := room.SetReadyInLobby("player-1", true); err != nil {
		t.Fatalf("expected ready update to succeed, got %v", err)
	}

	if err := room.StartGameForMember("player-1", func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected multiplayer start to succeed, got %v", err)
	}
	defer func() {
		if gameInstance := room.GameInstance(); gameInstance != nil {
			gameInstance.Stop()
		}
	}()

	gameInstance := room.GameInstance()
	playerID := gameInstance.AddPlayer()
	if playerID != "player-1" {
		t.Fatalf("expected authoritative game player id player-1, got %q", playerID)
	}
	gameInstance.SetPlayerScore("player-1", 120)
	markLifecycleTickTestGameOverForAllPlayers(t, gameInstance)

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

func TestResolvedMatchSummaryIsNotRebuiltAfterGameOver(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	if err := room.StartGameForMember(owner.PlayerID, func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected multiplayer start to succeed, got %v", err)
	}

	gameInstance := room.GameInstance()
	player1 := gameInstance.AddPlayer()
	player2 := gameInstance.AddPlayer()
	remapLifecycleTickTestPlayerID(t, gameInstance, player1, owner.PlayerID)
	remapLifecycleTickTestPlayerID(t, gameInstance, player2, peer.PlayerID)
	gameInstance.SetPlayerScore(owner.PlayerID, 100)
	gameInstance.SetPlayerScore(peer.PlayerID, 200)
	markLifecycleTickTestGameOverForAllPlayers(t, gameInstance)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	before, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to be stored")
	}

	gameInstance.SetPlayerScore(owner.PlayerID, 999)
	if err := room.MarkGameOver(); err == nil {
		t.Fatal("expected second game over transition to fail")
	}

	after, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to remain stored")
	}
	if after.MatchID != before.MatchID {
		t.Fatalf("expected MatchID to stay %q, got %q", before.MatchID, after.MatchID)
	}
	if after.Mode != before.Mode {
		t.Fatalf("expected Mode to stay %q, got %q", before.Mode, after.Mode)
	}
	if len(after.Players) != len(before.Players) {
		t.Fatalf("expected player count to stay %d, got %d", len(before.Players), len(after.Players))
	}
	for i := range before.Players {
		if after.Players[i] != before.Players[i] {
			t.Fatalf("expected summary player %d to stay %+v, got %+v", i, before.Players[i], after.Players[i])
		}
	}

	room.GameInstance().Stop()
}

func remapLifecycleTickTestPlayerID(t *testing.T, gameInstance *game.Game, oldPlayerID string, newPlayerID string) {
	t.Helper()

	value := reflect.ValueOf(gameInstance).Elem()
	sessions := exportLifecycleTickTestValue(value.FieldByName("playerSessions"))
	session := sessions.MapIndex(reflect.ValueOf(oldPlayerID))
	if !session.IsValid() {
		t.Fatalf("expected session %q to exist", oldPlayerID)
	}

	sessions.SetMapIndex(reflect.ValueOf(oldPlayerID), reflect.Value{})
	sessions.SetMapIndex(reflect.ValueOf(newPlayerID), session)
	exportLifecycleTickTestValue(session.Elem().FieldByName("ID")).SetString(newPlayerID)

	players := exportLifecycleTickTestValue(value.FieldByName("entities").FieldByName("Players"))
	player := players.MapIndex(reflect.ValueOf(oldPlayerID))
	if player.IsValid() {
		players.SetMapIndex(reflect.ValueOf(oldPlayerID), reflect.Value{})
		players.SetMapIndex(reflect.ValueOf(newPlayerID), player)
		exportLifecycleTickTestValue(player.Elem().FieldByName("ID")).SetString(newPlayerID)
	}
}

func pruneLifecycleTickTestPlayers(t *testing.T, gameInstance *game.Game, keepPlayerID string) {
	t.Helper()

	value := reflect.ValueOf(gameInstance).Elem()
	sessions := exportLifecycleTickTestValue(value.FieldByName("playerSessions"))
	for _, key := range sessions.MapKeys() {
		playerID := key.String()
		if playerID == keepPlayerID {
			continue
		}
		gameInstance.RemovePlayer(playerID)
	}
}

func markLifecycleTickTestGameOverForAllPlayers(t *testing.T, gameInstance *game.Game) {
	t.Helper()

	value := reflect.ValueOf(gameInstance).Elem()
	sessions := exportLifecycleTickTestValue(value.FieldByName("playerSessions"))
	for _, key := range sessions.MapKeys() {
		playerID := key.String()
		session := sessions.MapIndex(key)
		exportLifecycleTickTestValue(session.Elem().FieldByName("Lives")).SetInt(0)
		players := exportLifecycleTickTestValue(value.FieldByName("entities").FieldByName("Players"))
		players.SetMapIndex(reflect.ValueOf(playerID), reflect.Value{})
	}
}
