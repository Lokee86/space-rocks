package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
	"github.com/google/uuid"
)

func TestStartGameForMemberMovesLobbyRoomToInGame(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	newGame := func() *game.Game { return game.New() }

	if err := room.StartGameForMember("Player-1", newGame); err != nil {
		t.Fatalf("expected start to succeed, got %v", err)
	}
	if room.State != RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", RoomStateInGame, room.State)
	}
	if room.GameInstance() == nil {
		t.Fatal("expected game to be created")
	}
	if room.CurrentMatchID() == "" {
		t.Fatal("expected match ID to be set")
	}
	room.GameInstance().Stop()
}

func TestStartGameForMemberRejectsNonOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	err := room.StartGameForMember("Player-2", func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected non-owner start to fail")
	}
	if err.Code != RoomErrorNotRoomOwner {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotRoomOwner, err.Code)
	}
}

func TestStartGameForMemberRejectsUnreadyConnectedMember(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(false)

	err := room.StartGameForMember("Player-1", func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected unready connected member to block start")
	}
	if err.Code != RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotReady, err.Code)
	}
}

func TestStartGameForMemberRejectsNonLobbyRoom(t *testing.T) {
	room := NewRoom("room", RoomStateStarting, nil)
	room.AddMember(NewRoomMember("session-owner"))

	err := room.StartGameForMember("Player-1", func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected non-lobby start to fail")
	}
	if err.Code != RoomErrorRoomInGame {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomInGame, err.Code)
	}
}

func TestStartSinglePlayerGameMovesLobbyRoomWithMemberToInGame(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))

	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected single-player start to succeed, got %v", err)
	}
	if room.State != RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", RoomStateInGame, room.State)
	}
	if room.GameInstance() == nil {
		t.Fatal("expected game to be created")
	}
	if room.CurrentMatchID() == "" {
		t.Fatal("expected match ID to be set")
	}
	room.GameInstance().Stop()
}

func TestStartGameForMemberAdvancesMatchIDAfterReturnToLobby(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	newGame := func() *game.Game { return game.New() }

	if err := room.StartGameForMember("Player-1", newGame); err != nil {
		t.Fatalf("expected first start to succeed, got %v", err)
	}
	firstMatchID := room.CurrentMatchID()
	if firstMatchID == "" {
		t.Fatal("expected first match ID to be set")
	}
	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}
	if err := room.ResetToLobby("Player-1"); err != nil {
		t.Fatalf("expected reset to lobby to succeed, got %v", err)
	}

	owner.SetReady(true)
	peer.SetReady(true)

	if err := room.StartGameForMember("Player-1", newGame); err != nil {
		t.Fatalf("expected second start to succeed, got %v", err)
	}
	secondMatchID := room.CurrentMatchID()
	if secondMatchID == "" {
		t.Fatal("expected second match ID to be set")
	}
	if secondMatchID == firstMatchID {
		t.Fatalf("expected second match ID to differ from first match ID %q", firstMatchID)
	}
	room.GameInstance().Stop()
}

func TestStartGameForMemberResetsMatchResultReportedStateAfterNewMatch(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	newGame := func() *game.Game { return game.New() }

	if err := room.StartGameForMember("Player-1", newGame); err != nil {
		t.Fatalf("expected first start to succeed, got %v", err)
	}
	if room.MatchResultReported() {
		t.Fatal("expected new match to start with match result reported state false")
	}

	room.MarkMatchResultReported()
	if !room.MatchResultReported() {
		t.Fatal("expected match result reported state to be true after marking it")
	}

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}
	if err := room.ResetToLobby("Player-1"); err != nil {
		t.Fatalf("expected reset to lobby to succeed, got %v", err)
	}

	owner.SetReady(true)
	peer.SetReady(true)

	if err := room.StartGameForMember("Player-1", newGame); err != nil {
		t.Fatalf("expected second start to succeed, got %v", err)
	}
	if room.MatchResultReported() {
		t.Fatal("expected match result reported state to reset for the new match")
	}

	room.GameInstance().Stop()
}

func TestStartSinglePlayerGameRejectsRoomWithoutMembers(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	err := room.StartSinglePlayerGame(func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected single-player start without members to fail")
	}
	if err.Code != RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotInRoom, err.Code)
	}
}

func TestStartSinglePlayerGameRejectsNonLobbyRoom(t *testing.T) {
	room := NewRoom("room", RoomStateStarting, nil)
	room.AddMember(NewRoomMember("session-owner"))

	err := room.StartSinglePlayerGame(func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected single-player start from non-lobby room to fail")
	}
	if err.Code != RoomErrorRoomInGame {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomInGame, err.Code)
	}
}

func TestResetToLobbyOnlyWorksFromGameOver(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))

	err := room.ResetToLobby("Player-1")
	if err == nil {
		t.Fatal("expected reset from non-game-over state to fail")
	}
	if err.Code != RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", RoomErrorInvalidRoomState, err.Code)
	}
}

func TestResetToLobbyClearsReadyState(t *testing.T) {
	room := NewRoom("room", RoomStateGameOver, nil)
	member := room.AddMember(NewRoomMember("session-owner"))
	member.SetReady(true)

	if err := room.ResetToLobby("Player-1"); err != nil {
		t.Fatalf("expected reset to lobby to succeed, got %v", err)
	}
	if room.State != RoomStateLobby {
		t.Fatalf("expected room state %q, got %q", RoomStateLobby, room.State)
	}
	if member.Ready {
		t.Fatal("expected ready state to be cleared")
	}
}

func TestMarkGameOverStoresResolvedMatchSummary(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	member := room.AddMember(NewRoomMember("session-owner"))
	member.SetReady(true)

	if err := room.StartGameForMember("Player-1", func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected start to succeed, got %v", err)
	}

	traceID := room.match.CurrentTraceID()
	if uuid.Validate(traceID) != nil {
		t.Fatalf("match trace ID is not a UUID: %q", traceID)
	}

	gameInstance := room.GameInstance()
	playerID := gameInstance.AddPlayer()
	if playerID == "" {
		t.Fatal("expected a game player ID")
	}
	gameInstance.SetPlayerScore(playerID, 275)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to be stored")
	}
	if summary.TraceID != traceID {
		t.Fatalf("expected summary TraceID %q, got %q", traceID, summary.TraceID)
	}
	if summary.MatchID != room.CurrentMatchID() {
		t.Fatalf("expected MatchID %q, got %q", room.CurrentMatchID(), summary.MatchID)
	}
	if summary.Mode != playerdata.MatchModeMultiplayer {
		t.Fatalf("expected multiplayer mode, got %q", summary.Mode)
	}
	if len(summary.Players) != 1 {
		t.Fatalf("expected 1 player summary, got %d", len(summary.Players))
	}
	if summary.Players[0].GamePlayerID != playerID {
		t.Fatalf("expected GamePlayerID %q, got %q", playerID, summary.Players[0].GamePlayerID)
	}
	if summary.Players[0].Score != 275 {
		t.Fatalf("expected score 275, got %d", summary.Players[0].Score)
	}
	room.GameInstance().Stop()
}

func TestMarkGameOverKeepsExistingResolvedMatchSummary(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	member := room.AddMember(NewRoomMember("session-owner"))
	member.SetReady(true)

	if err := room.StartGameForMember("Player-1", func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected start to succeed, got %v", err)
	}

	preset := playerdata.MatchResultSummary{
		MatchID: "preset-match",
		Mode:    playerdata.MatchModeSinglePlayer,
	}
	room.match.SetResolvedSummary(preset)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	summary, ok := room.ResolvedMatchSummary()
	if !ok {
		t.Fatal("expected resolved match summary to remain stored")
	}
	if summary.MatchID != preset.MatchID {
		t.Fatalf("expected MatchID %q, got %q", preset.MatchID, summary.MatchID)
	}
	if summary.Mode != preset.Mode {
		t.Fatalf("expected Mode %q, got %q", preset.Mode, summary.Mode)
	}
	room.GameInstance().Stop()
}

func TestStartGameForSessionRejectsMissingSessionAfterPlayerIDReuse(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	if room.AddMember(NewRoomMember("session-old")) == nil {
		t.Fatal("expected old member to be added")
	}
	if _, _, removed := room.RemoveMemberForSession("session-old"); !removed {
		t.Fatal("expected old member removal")
	}
	replacement := room.AddMember(NewRoomMember("session-replacement"))
	if replacement == nil {
		t.Fatal("expected replacement member to be added")
	}
	if playerID, ok := room.PlayerIDForSession("session-replacement"); !ok || playerID != "Player-1" {
		t.Fatalf("expected replacement session to map to Player-1, got %q, %v", playerID, ok)
	}
	replacement.SetReady(true)
	factoryCalls := 0
	err := room.StartGameForSession("session-old", func() *game.Game { factoryCalls++; return game.New() })
	if err == nil || err.Code != RoomErrorNotInRoom {
		t.Fatalf("expected not_in_room error, got %v", err)
	}
	if factoryCalls != 0 {
		t.Fatalf("expected factory not to be called, got %d calls", factoryCalls)
	}
	if room.State != RoomStateLobby || room.GameInstance() != nil || room.CurrentMatchID() != "" {
		t.Fatal("expected unchanged lobby state")
	}
	if playerID, ok := room.PlayerIDForSession("session-replacement"); !ok || playerID != "Player-1" {
		t.Fatalf("expected replacement mapping to remain Player-1, got %q, %v", playerID, ok)
	}
}

func TestResetToLobbyForSessionRejectsMissingSessionAfterPlayerIDReuse(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	if room.AddMember(NewRoomMember("session-old")) == nil {
		t.Fatal("expected old member to be added")
	}
	if _, _, removed := room.RemoveMemberForSession("session-old"); !removed {
		t.Fatal("expected old member removal")
	}
	replacement := room.AddMember(NewRoomMember("session-replacement"))
	if replacement == nil {
		t.Fatal("expected replacement member to be added")
	}
	if playerID, ok := room.PlayerIDForSession("session-replacement"); !ok || playerID != "Player-1" {
		t.Fatalf("expected replacement session to map to Player-1, got %q, %v", playerID, ok)
	}
	replacement.SetReady(true)
	if err := room.StartGameForSession("session-replacement", game.New); err != nil {
		t.Fatalf("expected replacement start to succeed, got %v", err)
	}
	traceID := room.match.CurrentTraceID()
	if uuid.Validate(traceID) != nil {
		t.Fatalf("match trace ID is not a UUID: %q", traceID)
	}

	gameInstance := room.GameInstance()
	t.Cleanup(func() { gameInstance.Stop() })
	matchID := room.CurrentMatchID()
	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}
	err := room.ResetToLobbyForSession("session-old")
	if err == nil || err.Code != RoomErrorNotInRoom {
		t.Fatalf("expected not_in_room error, got %v", err)
	}
	if room.State != RoomStateGameOver || room.GameInstance() != gameInstance || room.CurrentMatchID() != matchID {
		t.Fatal("expected game-over state and game to remain unchanged")
	}
	if playerID, ok := room.PlayerIDForSession("session-replacement"); !ok || playerID != "Player-1" {
		t.Fatalf("expected replacement mapping to remain Player-1, got %q, %v", playerID, ok)
	}
}
