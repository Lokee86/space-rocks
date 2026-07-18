package rooms

import (
	"reflect"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	runtimepkg "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
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

func TestMultiplayerResolvedMatchSummaryIncludesRemovedPlayer(t *testing.T) {
	manager := NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	peerAccountID := "11111111-2222-3333-4444-555555555555"
	if !room.SetMemberAccountIDForSession(peer.SessionID, peerAccountID) {
		t.Fatal("expected peer account identity to be stored")
	}
	if err := room.StartGameForMember(owner.PlayerID, game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}

	gameInstance := room.GameInstance()
	ownerPlayerID := gameInstance.AddPlayer()
	peerPlayerID := gameInstance.AddPlayer()
	context := room.GameplayContext()
	if !room.ActivateMemberPlayer(context, owner.SessionID, ownerPlayerID) ||
		!room.ActivateMemberPlayer(context, peer.SessionID, peerPlayerID) {
		t.Fatal("activate players")
	}

	gameInstance.SetPlayerScore(ownerPlayerID, 125)
	gameInstance.SetPlayerScore(peerPlayerID, 275)
	setLifecycleTickTestShipDeaths(t, gameInstance, peerPlayerID, 2)

	leaveResult, leaveErr := manager.LeaveMember(room.ID, peer.SessionID, "")
	if leaveErr != nil {
		t.Fatalf("leave peer: %v", leaveErr)
	}
	if !leaveResult.PlayerRemoved || leaveResult.PlayerID != peerPlayerID {
		t.Fatalf("leave result = %+v, want peer removed from active play", leaveResult)
	}

	markLifecycleTickTestGameOverForAllPlayers(t, gameInstance)
	if !room.MarkGameOverIfComplete() {
		t.Fatal("removed player blocked room game-over transition")
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary")
	}
	if len(summary.Players) != 2 {
		t.Fatalf("summary players = %+v, want both participants", summary.Players)
	}

	playersByID := make(map[string]playerdata.PlayerMatchSummary, len(summary.Players))
	for _, player := range summary.Players {
		playersByID[player.GamePlayerID] = player
	}
	removedPlayer, ok := playersByID[peerPlayerID]
	if !ok {
		t.Fatalf("missing removed player %q from summary", peerPlayerID)
	}
	if removedPlayer.Score != 275 || removedPlayer.ShipDeaths != 2 {
		t.Fatalf("removed player summary = %+v, want score 275 and two deaths", removedPlayer)
	}
	if removedPlayer.AccountID != peerAccountID {
		t.Fatalf("removed player AccountID = %q, want %q", removedPlayer.AccountID, peerAccountID)
	}
	if activePlayer, ok := playersByID[ownerPlayerID]; !ok || activePlayer.Score != 125 {
		t.Fatalf("active player summary = %+v, want score 125", activePlayer)
	}
}

func TestRoomGameOverLifecycleCompletesAfterAllPlayersLeave(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(time.Hour)
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	ownerAccountID := "11111111-2222-3333-4444-555555555555"
	peerLocalProfileID := "local-profile-peer"
	if !room.SetMemberAccountIDForSession(owner.SessionID, ownerAccountID) {
		t.Fatal("store owner account identity")
	}
	if !room.SetMemberLocalProfileIDForSession(peer.SessionID, peerLocalProfileID) {
		t.Fatal("store peer local-profile identity")
	}
	if err := room.StartGameForMember(owner.PlayerID, game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}

	gameInstance := room.GameInstance()
	ownerPlayerID := gameInstance.AddPlayer()
	peerPlayerID := gameInstance.AddPlayer()
	context := room.GameplayContext()
	if !room.ActivateMemberPlayer(context, owner.SessionID, ownerPlayerID) ||
		!room.ActivateMemberPlayer(context, peer.SessionID, peerPlayerID) {
		t.Fatal("activate players")
	}
	gameInstance.SetPlayerScore(ownerPlayerID, 140)
	gameInstance.SetPlayerScore(peerPlayerID, 260)
	setLifecycleTickTestShipDeaths(t, gameInstance, ownerPlayerID, 1)
	setLifecycleTickTestShipDeaths(t, gameInstance, peerPlayerID, 2)

	for _, sessionID := range []string{owner.SessionID, peer.SessionID} {
		leaveResult, leaveErr := manager.LeaveMember(room.ID, sessionID, "")
		if leaveErr != nil {
			t.Fatalf("leave %q: %v", sessionID, leaveErr)
		}
		if !leaveResult.PlayerRemoved {
			t.Fatalf("leave result for %q = %+v, want active player removed", sessionID, leaveResult)
		}
	}
	if population := room.Population(); population.Members != 0 || population.ActivePlayers != 0 {
		t.Fatalf("population after leaves = %+v, want empty room", population)
	}

	broadcasts := 0
	if !TickRoomGameOverLifecycle(room, func(broadcastRoom *Room) {
		broadcasts++
		if broadcastRoom != room {
			t.Fatal("broadcasted unexpected room")
		}
	}) {
		t.Fatal("normal lifecycle did not complete match after all players left")
	}
	if room.State != RoomStateGameOver || broadcasts != 1 {
		t.Fatalf("room state/broadcasts = %q/%d, want game over/1", room.State, broadcasts)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary")
	}
	if len(summary.Players) != 2 {
		t.Fatalf("summary players = %+v, want both historical participants", summary.Players)
	}
	playersByID := make(map[string]playerdata.PlayerMatchSummary, len(summary.Players))
	for _, player := range summary.Players {
		playersByID[player.GamePlayerID] = player
	}
	if player := playersByID[ownerPlayerID]; player.AccountID != ownerAccountID || player.Score != 140 || player.ShipDeaths != 1 {
		t.Fatalf("owner summary = %+v, want retained account, score, and deaths", player)
	}
	if player := playersByID[peerPlayerID]; player.LocalProfileID != peerLocalProfileID || player.Score != 260 || player.ShipDeaths != 2 {
		t.Fatalf("peer summary = %+v, want retained profile, score, and deaths", player)
	}

	summaryType := reflect.TypeOf(playerdata.PlayerMatchSummary{})
	for _, fieldName := range []string{"Disconnected", "Departed", "Forfeited"} {
		if _, exists := summaryType.FieldByName(fieldName); exists {
			t.Fatalf("unexpected participant result label %q", fieldName)
		}
	}

	reporter := &fakeMatchResultReporter{}
	if !ReportResolvedMatchResultOnce(room, reporter) || reporter.calls != 1 {
		t.Fatalf("result reporting before cleanup failed: calls=%d", reporter.calls)
	}
	if _, stillManaged := manager.Find(room.ID); !stillManaged {
		t.Fatal("room cleaned up before resolved result could be reported")
	}
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
	if !sessions.MapIndex(reflect.ValueOf(oldPlayerID)).IsValid() {
		t.Fatalf("expected session %q to exist", oldPlayerID)
	}

	records := exportLifecycleTickTestValue(value.FieldByName("participantRecords"))
	record := records.MapIndex(reflect.ValueOf(oldPlayerID))
	if !record.IsValid() {
		t.Fatalf("expected participant record %q to exist", oldPlayerID)
	}
	control := game.NewControl(gameInstance)
	if !control.EnsurePlayerSession(newPlayerID, physics.Vector2{}) {
		t.Fatalf("expected test fixture to register session %q", newPlayerID)
	}
	position, ok := control.SafeRespawnPosition(newPlayerID)
	if !ok || !control.ForceRespawnPlayer(newPlayerID, position, runtimepkg.ClientConfig{}) {
		t.Fatalf("expected test fixture to activate player %q", newPlayerID)
	}
	gameInstance.RemovePlayer(oldPlayerID)

	records.SetMapIndex(reflect.ValueOf(oldPlayerID), reflect.Value{})
	records.SetMapIndex(reflect.ValueOf(newPlayerID), record)
	exportLifecycleTickTestValue(record.Elem().FieldByName("ID")).SetString(newPlayerID)
}

func setLifecycleTickTestShipDeaths(t *testing.T, gameInstance *game.Game, playerID string, shipDeaths int) {
	t.Helper()

	control := game.NewControl(gameInstance)
	for death := 0; death < shipDeaths; death++ {
		if !control.ApplyPlayerDefeat("test-fixture", playerID) {
			t.Fatalf("expected test fixture to apply death %d for %q", death+1, playerID)
		}
		if death == shipDeaths-1 {
			continue
		}

		position, ok := control.SafeRespawnPosition(playerID)
		if !ok {
			t.Fatalf("expected test fixture to find respawn position for %q", playerID)
		}
		if !control.ForceRespawnPlayer(playerID, position, runtimepkg.ClientConfig{}) {
			t.Fatalf("expected test fixture to respawn %q", playerID)
		}
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
	control := game.NewControl(gameInstance)
	for _, key := range sessions.MapKeys() {
		playerID := key.String()
		if !control.SetPlayerLives(playerID, 0) {
			t.Fatalf("expected test fixture to set lives for %q", playerID)
		}
		if !control.ApplyPlayerDefeat("test-fixture", playerID) {
			t.Fatalf("expected test fixture to apply defeat for %q", playerID)
		}
	}
}
