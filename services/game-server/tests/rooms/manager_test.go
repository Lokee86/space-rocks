package roomstests

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestRoomManagerUsesDefaultRoomForBlankID(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	blankRoom := manager.GetOrCreate("")
	spaceRoom := manager.GetOrCreate("   ")

	if spaceRoom != blankRoom {
		t.Fatal("expected blank and whitespace room ids to use the same default room")
	}
}

func TestRoomManagerCreatesAndFindsCompatibilityRooms(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	first := manager.GetOrCreate("abc")
	again := manager.GetOrCreate("abc")
	second := manager.GetOrCreate("xyz")

	if first != again {
		t.Fatal("expected same room for same id")
	}
	if first == second {
		t.Fatal("expected different room for different id")
	}
	if first.State != rooms.RoomStateInGame {
		t.Fatalf("expected compatibility room state %q, got %q", rooms.RoomStateInGame, first.State)
	}
	if first.Game == nil {
		t.Fatal("expected compatibility room to create a game")
	}

	found, ok := manager.Find("abc")
	if !ok {
		t.Fatal("expected room to be found")
	}
	if found != first {
		t.Fatal("expected found room to match created room")
	}
}

func TestRoomManagerCreateLobbyRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}

	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected lobby room state %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected lobby room not to create a game")
	}
	if !rooms.IsValidRoomCode(room.ID) {
		t.Fatalf("expected generated room code, got %q", room.ID)
	}

	found, ok := manager.Find(room.ID)
	if !ok {
		t.Fatal("expected created lobby room to be found")
	}
	if found != room {
		t.Fatal("expected found lobby room to match created room")
	}
}

func TestRoomManagerCreateSinglePlayerRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateSinglePlayerRoom("session-1")
	if err != nil {
		t.Fatalf("create single-player room: %v", err)
	}

	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected single-player room state %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected single-player room not to create a game")
	}
	if room.IsJoinable() {
		t.Fatal("expected single-player room not to be joinable")
	}
	if !rooms.IsValidRoomCode(room.ID) {
		t.Fatalf("expected generated room code, got %q", room.ID)
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected single-player room member count 1, got %d", count)
	}
	if !room.HasMember("session-1") {
		t.Fatal("expected requesting session to be the only room member")
	}

	found, ok := manager.Find(room.ID)
	if !ok {
		t.Fatal("expected created single-player room to be found")
	}
	if found != room {
		t.Fatal("expected found single-player room to match created room")
	}
}

func TestRoomManagerCreateStartedSinglePlayerRoomCreatesInGameRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, roomErr := manager.CreateStartedSinglePlayerRoom("session-1")
	if roomErr != nil {
		t.Fatalf("create started single-player room: %v", roomErr)
	}

	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.Game == nil {
		t.Fatal("expected started single-player room to create a game")
	}
	if room.IsJoinable() {
		t.Fatal("expected started single-player room not to be joinable")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected single-player room member count 1, got %d", count)
	}
	if !room.HasMember("session-1") {
		t.Fatal("expected requesting session to be the room member")
	}
}

func TestRoomManagerCreateSinglePlayerRoomRejectsJoinRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateSinglePlayerRoom("session-1")
	if err != nil {
		t.Fatalf("create single-player room: %v", err)
	}

	joinedRoom, roomErr := manager.JoinRoom("session-2", room.ID)
	if roomErr == nil {
		t.Fatal("expected invalid_room_state error")
	}
	if joinedRoom != nil {
		t.Fatal("expected no joined room")
	}
	if roomErr.Code != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomErr.Code)
	}
	if room.HasMember("session-2") {
		t.Fatal("expected rejected join not to add member")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected single-player room member count to remain 1, got %d", count)
	}
}

func TestRoomManagerReturnRoomToLobbyResetsGameAndReadiness(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	room.AddMemberID("session-1")
	room.AddMemberID("session-2")
	room.SetMemberReady("session-1", true)
	room.SetMemberReady("session-2", true)
	oldGame := game.New()
	room.Game = oldGame
	room.State = rooms.RoomStateGameOver

	returnedRoom, roomErr := manager.ReturnRoomToLobby(room.ID, "session-1")
	if roomErr != nil {
		t.Fatalf("return room to lobby: %v", roomErr)
	}
	if returnedRoom != room {
		t.Fatal("expected returned room to match lobby room")
	}
	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected return to lobby to clear game")
	}
	if !gameStopped(t, oldGame) {
		t.Fatal("expected return to lobby to stop old game")
	}

	for _, member := range room.MembersSnapshot() {
		if member.Ready {
			t.Fatalf("expected member %q ready state to reset", member.SessionID)
		}
	}
}

func TestRoomManagerJoinRoomAddsMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}

	joinedRoom, roomErr := manager.JoinRoom("session-1", room.ID)
	if roomErr != nil {
		t.Fatalf("join room: %v", roomErr)
	}
	if joinedRoom != room {
		t.Fatal("expected joined room to match lobby room")
	}
	if !room.HasMember("session-1") {
		t.Fatal("expected joined member to be added")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected member count 1, got %d", count)
	}
}

func TestRoomManagerMultiplayerLobbyRoomRemainsJoinable(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}

	joinedRoom, roomErr := manager.JoinRoom("session-1", room.ID)
	if roomErr != nil {
		t.Fatalf("join lobby room: %v", roomErr)
	}
	if joinedRoom != room {
		t.Fatal("expected joined room to match lobby room")
	}
	if !room.HasMember("session-1") {
		t.Fatal("expected joined member to be added")
	}
}

func TestRoomManagerJoinRoomRejectsUnjoinableRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	room.SetJoinable(false)

	joinedRoom, roomErr := manager.JoinRoom("session-1", room.ID)
	if roomErr == nil {
		t.Fatal("expected invalid_room_state error")
	}
	if joinedRoom != nil {
		t.Fatal("expected no joined room")
	}
	if roomErr.Code != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomErr.Code)
	}
	if room.HasMember("session-1") {
		t.Fatal("expected rejected join not to add member")
	}
}

func TestRoomManagerJoinRoomBelowCapacitySucceeds(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	for index := 0; index < rooms.MaxPlayersPerRoom-1; index++ {
		room.AddMemberID(string(rune('a' + index)))
	}

	joinedRoom, roomErr := manager.JoinRoom("last", room.ID)
	if roomErr != nil {
		t.Fatalf("join below capacity: %v", roomErr)
	}
	if joinedRoom != room {
		t.Fatal("expected joined room to match lobby room")
	}
	if count := room.MemberCount(); count != rooms.MaxPlayersPerRoom {
		t.Fatalf("expected member count %d, got %d", rooms.MaxPlayersPerRoom, count)
	}
}

func TestRoomManagerJoinRoomAtCapacityFails(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	for index := 0; index < rooms.MaxPlayersPerRoom; index++ {
		room.AddMemberID(string(rune('a' + index)))
	}

	joinedRoom, roomErr := manager.JoinRoom("overflow", room.ID)
	if roomErr == nil {
		t.Fatal("expected room_full error")
	}
	if joinedRoom != nil {
		t.Fatal("expected no joined room")
	}
	if roomErr.Code != rooms.RoomErrorRoomFull {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomFull, roomErr.Code)
	}
	if count := room.MemberCount(); count != rooms.MaxPlayersPerRoom {
		t.Fatalf("expected failed join to preserve member count %d, got %d", rooms.MaxPlayersPerRoom, count)
	}
	if room.HasMember("overflow") {
		t.Fatal("expected failed join not to add member")
	}
}

func TestRoomManagerLeaveRoomRemovesMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	if _, roomErr := manager.JoinRoom("session-1", room.ID); roomErr != nil {
		t.Fatalf("join room: %v", roomErr)
	}

	result, roomErr := manager.LeaveRoom(room.ID, "session-1")
	if roomErr != nil {
		t.Fatalf("leave room: %v", roomErr)
	}
	if result.Room != room {
		t.Fatal("expected leave result room to match lobby room")
	}
	if result.RoomID != room.ID {
		t.Fatalf("expected leave result room id %q, got %q", room.ID, result.RoomID)
	}
	if result.MemberID != "session-1" {
		t.Fatalf("expected leave result member id %q, got %q", "session-1", result.MemberID)
	}
	if result.RemainingMembers != 0 {
		t.Fatalf("expected remaining members 0, got %d", result.RemainingMembers)
	}
	if room.HasMember("session-1") {
		t.Fatal("expected member to be removed")
	}
}

func TestRoomManagerLeaveMemberRemovesMemberAndReportsBroadcast(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	room.AddMemberID("session-1")
	room.AddMemberID("session-2")

	result, roomErr := manager.LeaveMember(room.ID, "session-1", "")
	if roomErr != nil {
		t.Fatalf("leave member: %v", roomErr)
	}
	if result.Room != room {
		t.Fatal("expected leave member result room to match lobby room")
	}
	if result.MemberID != "session-1" {
		t.Fatalf("expected member id %q, got %q", "session-1", result.MemberID)
	}
	if result.RemainingMembers != 1 {
		t.Fatalf("expected remaining members 1, got %d", result.RemainingMembers)
	}
	if result.PlayerRemoved {
		t.Fatal("expected no player to be removed")
	}
	if result.CleanupScheduled {
		t.Fatal("expected cleanup not to be scheduled while room still has a member")
	}
	if !result.ShouldBroadcastSnapshot {
		t.Fatal("expected leave member result to request snapshot broadcast")
	}
	if room.HasMember("session-1") {
		t.Fatal("expected leaving member to be removed")
	}
	if !room.HasMember("session-2") {
		t.Fatal("expected remaining member to stay in room")
	}
	if room.CleanupTimer != nil {
		t.Fatal("expected cleanup timer not to be scheduled")
	}
}

func TestRoomManagerLeaveMemberRemovesPlayerAndSchedulesCleanupWhenEmpty(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	room.AddMemberID("session-1")
	room.Game = game.New()
	playerID := room.Game.AddPlayer()
	room.ActivePlayers = 1

	result, roomErr := manager.LeaveMember(room.ID, "session-1", playerID)
	if roomErr != nil {
		t.Fatalf("leave member: %v", roomErr)
	}
	if room.HasMember("session-1") {
		t.Fatal("expected leaving member to be removed")
	}
	if gameTracksPlayer(t, room.Game, playerID) {
		t.Fatal("expected leaving player to be removed from game")
	}
	if !result.PlayerRemoved {
		t.Fatal("expected leave member result to report player removal")
	}
	if result.ActivePlayers != 0 {
		t.Fatalf("expected active players 0, got %d", result.ActivePlayers)
	}
	if result.RemainingMembers != 0 {
		t.Fatalf("expected remaining members 0, got %d", result.RemainingMembers)
	}
	if result.ShouldBroadcastSnapshot {
		t.Fatal("expected empty room not to request snapshot broadcast")
	}
	if !result.CleanupScheduled {
		t.Fatal("expected empty room cleanup to be scheduled")
	}
	if room.CleanupTimer == nil {
		t.Fatal("expected cleanup timer to be scheduled")
	}
}

func TestRoomManagerSetReadyUpdatesLobbyMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	if _, roomErr := manager.JoinRoom("session-1", room.ID); roomErr != nil {
		t.Fatalf("join room: %v", roomErr)
	}

	updatedRoom, roomErr := manager.SetReady(room.ID, "session-1", true)
	if roomErr != nil {
		t.Fatalf("set ready: %v", roomErr)
	}
	if updatedRoom != room {
		t.Fatal("expected updated room to match lobby room")
	}

	members := room.MembersSnapshot()
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if !members[0].Ready {
		t.Fatal("expected member to be ready")
	}
}

func TestRoomManagerSetReadyRejectsMissingMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}

	updatedRoom, roomErr := manager.SetReady(room.ID, "missing", true)
	if roomErr == nil {
		t.Fatal("expected not_in_room error")
	}
	if updatedRoom != nil {
		t.Fatal("expected no updated room")
	}
	if roomErr.Code != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomErr.Code)
	}
}

func TestRoomManagerSetReadyRejectsNonLobbyRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room := manager.GetOrCreate("ABCDEF")
	room.AddMemberID("session-1")

	updatedRoom, roomErr := manager.SetReady(room.ID, "session-1", true)
	if roomErr == nil {
		t.Fatal("expected invalid_room_state error")
	}
	if updatedRoom != nil {
		t.Fatal("expected no updated room")
	}
	if roomErr.Code != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomErr.Code)
	}
}

func TestRoomManagerStartRoomGameRejectsMissingRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	startedRoom, roomErr := manager.StartRoomGame("MISSING", "session-1")
	if roomErr == nil {
		t.Fatal("expected room_not_found error")
	}
	if startedRoom != nil {
		t.Fatal("expected no started room")
	}
	if roomErr.Code != rooms.RoomErrorRoomNotFound {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomNotFound, roomErr.Code)
	}
}

func TestRoomManagerStartRoomGameRejectsNonMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	room.AddMemberID("session-1")
	room.SetMemberReady("session-1", true)

	startedRoom, roomErr := manager.StartRoomGame(room.ID, "missing")
	if roomErr == nil {
		t.Fatal("expected not_in_room error")
	}
	if startedRoom != nil {
		t.Fatal("expected no started room")
	}
	if roomErr.Code != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomErr.Code)
	}
	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected room state to remain %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected rejected start not to create a game")
	}
}

func TestRoomManagerStartRoomGameRejectsNotReadyMembers(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	room.AddMemberID("session-1")
	room.AddMemberID("session-2")
	room.SetMemberReady("session-1", true)

	startedRoom, roomErr := manager.StartRoomGame(room.ID, "session-1")
	if roomErr == nil {
		t.Fatal("expected not_ready error")
	}
	if startedRoom != nil {
		t.Fatal("expected no started room")
	}
	if roomErr.Code != rooms.RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotReady, roomErr.Code)
	}
	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected room state to remain %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected rejected start not to create a game")
	}
}

func TestRoomManagerStartRoomGameTransitionsLobbyToInGame(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	room.AddMemberID("session-1")
	room.AddMemberID("session-2")
	room.SetMemberReady("session-1", true)
	room.SetMemberReady("session-2", true)

	startedRoom, roomErr := manager.StartRoomGame(room.ID, "session-1")
	if roomErr != nil {
		t.Fatalf("start room game: %v", roomErr)
	}
	if startedRoom != room {
		t.Fatal("expected started room to match lobby room")
	}
	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.Game == nil {
		t.Fatal("expected started room to create a game")
	}
}

func gameTracksPlayer(t *testing.T, gameInstance *game.Game, playerID string) bool {
	t.Helper()

	playerSessions := exportRoomTestValue(
		reflect.ValueOf(gameInstance).Elem().FieldByName("playerSessions"),
	)
	return playerSessions.MapIndex(reflect.ValueOf(playerID)).IsValid()
}
