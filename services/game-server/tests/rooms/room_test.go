package roomstests

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestRoomMemberAccessors(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	member := rooms.NewRoomMember("session-1")

	stored := room.AddMember(member)
	if stored != member {
		t.Fatal("expected AddMember to return stored member")
	}
	if !room.HasMember("session-1") {
		t.Fatal("expected room to have member")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected member count 1, got %d", count)
	}

	room.RemoveMember("session-1")
	if room.HasMember("session-1") {
		t.Fatal("expected room member to be removed")
	}
}

func TestRoomSetMemberReady(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")

	if !room.SetMemberReady("session-1", true) {
		t.Fatal("expected SetMemberReady to find member")
	}

	members := room.MembersSnapshot()
	if len(members) != 1 {
		t.Fatalf("expected 1 member snapshot, got %d", len(members))
	}
	if !members[0].Ready {
		t.Fatal("expected member to be ready")
	}

	if room.SetMemberReady("missing", true) {
		t.Fatal("expected SetMemberReady to fail for missing member")
	}
}

func TestRoomValidateStartAllowsOneReadyMember(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")
	room.SetMemberReady("session-1", true)

	if roomErr := room.ValidateStart("session-1"); roomErr != nil {
		t.Fatalf("expected ready solo member to start, got %s", roomErr.Code)
	}
}

func TestRoomValidateStartRequiresRequesterInRoom(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")
	room.SetMemberReady("session-1", true)

	roomErr := room.ValidateStart("missing")
	if roomErr == nil {
		t.Fatal("expected missing requester to be rejected")
	}
	if roomErr.Code != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomErr.Code)
	}
}

func TestRoomValidateStartRequiresLobbyState(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateGameOver, nil)
	room.AddMemberID("session-1")
	room.SetMemberReady("session-1", true)

	roomErr := room.ValidateStart("session-1")
	if roomErr == nil {
		t.Fatal("expected non-lobby room to be rejected")
	}
	if roomErr.Code != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomErr.Code)
	}
}

func TestRoomValidateStartRejectsDoubleStartStates(t *testing.T) {
	for _, state := range []rooms.RoomState{rooms.RoomStateStarting, rooms.RoomStateInGame} {
		room := rooms.NewRoom("TEST", state, nil)
		room.AddMemberID("session-1")
		room.SetMemberReady("session-1", true)

		roomErr := room.ValidateStart("session-1")
		if roomErr == nil {
			t.Fatalf("expected state %q to reject double start", state)
		}
		if roomErr.Code != rooms.RoomErrorRoomInGame {
			t.Fatalf("expected error code %q for state %q, got %q", rooms.RoomErrorRoomInGame, state, roomErr.Code)
		}
	}
}

func TestRoomValidateStartRequiresAllConnectedMembersReady(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")
	room.AddMemberID("session-2")
	room.SetMemberReady("session-1", true)

	roomErr := room.ValidateStart("session-1")
	if roomErr == nil {
		t.Fatal("expected unready connected member to block start")
	}
	if roomErr.Code != rooms.RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotReady, roomErr.Code)
	}
}

func TestRoomValidateStartIgnoresDisconnectedUnreadyMembers(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")
	disconnected := room.AddMemberID("session-2")
	disconnected.MarkDisconnected()
	room.SetMemberReady("session-1", true)

	if roomErr := room.ValidateStart("session-1"); roomErr != nil {
		t.Fatalf("expected disconnected unready member not to block start, got %s", roomErr.Code)
	}
}

func TestRoomMarkStartingMovesLobbyToStarting(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)

	if roomErr := room.MarkStarting(); roomErr != nil {
		t.Fatalf("expected lobby room to mark starting, got %s", roomErr.Code)
	}
	if room.State != rooms.RoomStateStarting {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateStarting, room.State)
	}
}

func TestRoomMarkStartingRejectsInvalidState(t *testing.T) {
	for _, state := range []rooms.RoomState{
		rooms.RoomStateStarting,
		rooms.RoomStateInGame,
		rooms.RoomStateGameOver,
		rooms.RoomStateClosed,
	} {
		room := rooms.NewRoom("TEST", state, nil)

		roomErr := room.MarkStarting()
		if roomErr == nil {
			t.Fatalf("expected state %q to reject mark starting", state)
		}
		if roomErr.Code != rooms.RoomErrorInvalidRoomState {
			t.Fatalf("expected error code %q for state %q, got %q", rooms.RoomErrorInvalidRoomState, state, roomErr.Code)
		}
		if room.State != state {
			t.Fatalf("expected rejected room to stay %q, got %q", state, room.State)
		}
	}
}

func TestRoomMarkInGameMovesStartingToInGame(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateStarting, game.New())

	if roomErr := room.MarkInGame(); roomErr != nil {
		t.Fatalf("expected starting room to mark in-game, got %s", roomErr.Code)
	}
	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.Game == nil {
		t.Fatal("expected game instance to remain after in-game transition")
	}
}

func TestRoomMarkInGameRejectsInvalidState(t *testing.T) {
	for _, state := range []rooms.RoomState{
		rooms.RoomStateLobby,
		rooms.RoomStateInGame,
		rooms.RoomStateGameOver,
		rooms.RoomStateClosed,
	} {
		room := rooms.NewRoom("TEST", state, nil)

		roomErr := room.MarkInGame()
		if roomErr == nil {
			t.Fatalf("expected state %q to reject mark in-game", state)
		}
		if roomErr.Code != rooms.RoomErrorInvalidRoomState {
			t.Fatalf("expected error code %q for state %q, got %q", rooms.RoomErrorInvalidRoomState, state, roomErr.Code)
		}
		if room.State != state {
			t.Fatalf("expected rejected room to stay %q, got %q", state, room.State)
		}
	}
}

func TestRoomMarkGameOverMovesInGameToGameOver(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateInGame, game.New())

	if roomErr := room.MarkGameOver(); roomErr != nil {
		t.Fatalf("expected in-game room to mark game over, got %s", roomErr.Code)
	}
	if room.State != rooms.RoomStateGameOver {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateGameOver, room.State)
	}
	if room.Game == nil {
		t.Fatal("expected game instance to remain after game-over transition")
	}
}

func TestRoomMarkGameOverRejectsInvalidState(t *testing.T) {
	for _, state := range []rooms.RoomState{
		rooms.RoomStateLobby,
		rooms.RoomStateStarting,
		rooms.RoomStateGameOver,
		rooms.RoomStateClosed,
	} {
		room := rooms.NewRoom("TEST", state, nil)

		roomErr := room.MarkGameOver()
		if roomErr == nil {
			t.Fatalf("expected state %q to reject mark game over", state)
		}
		if roomErr.Code != rooms.RoomErrorInvalidRoomState {
			t.Fatalf("expected error code %q for state %q, got %q", rooms.RoomErrorInvalidRoomState, state, roomErr.Code)
		}
		if room.State != state {
			t.Fatalf("expected rejected room to stay %q, got %q", state, room.State)
		}
	}
}

func TestRoomResetToLobbyClearsGameAndReadyStates(t *testing.T) {
	oldGame := game.New()
	room := rooms.NewRoom("TEST", rooms.RoomStateGameOver, oldGame)
	first := room.AddMemberID("session-1")
	second := room.AddMemberID("session-2")
	first.SetReady(true)
	second.SetReady(true)

	if roomErr := room.ResetToLobby("session-1"); roomErr != nil {
		t.Fatalf("expected game-over room to reset to lobby, got %s", roomErr.Code)
	}
	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected old game instance to be cleared")
	}
	if !gameStopped(t, oldGame) {
		t.Fatal("expected old game instance to be stopped before being cleared")
	}

	members := room.MembersSnapshot()
	if len(members) != 2 {
		t.Fatalf("expected connected members to remain, got %d", len(members))
	}
	for _, member := range members {
		if member.Ready {
			t.Fatalf("expected member %q ready state to reset", member.SessionID)
		}
		if !member.Connected {
			t.Fatalf("expected member %q to remain connected", member.SessionID)
		}
	}
}

func TestRoomResetToLobbyRequiresRequesterMember(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateGameOver, game.New())
	room.AddMemberID("session-1")

	roomErr := room.ResetToLobby("missing")
	if roomErr == nil {
		t.Fatal("expected missing member to be rejected")
	}
	if roomErr.Code != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomErr.Code)
	}
	if room.State != rooms.RoomStateGameOver {
		t.Fatalf("expected room state to remain %q, got %q", rooms.RoomStateGameOver, room.State)
	}
}

func TestRoomResetToLobbyRequiresGameOverState(t *testing.T) {
	for _, state := range []rooms.RoomState{
		rooms.RoomStateLobby,
		rooms.RoomStateStarting,
		rooms.RoomStateInGame,
		rooms.RoomStateClosed,
	} {
		room := rooms.NewRoom("TEST", state, game.New())
		member := room.AddMemberID("session-1")
		member.SetReady(true)

		roomErr := room.ResetToLobby("session-1")
		if roomErr == nil {
			t.Fatalf("expected state %q to reject reset to lobby", state)
		}
		if roomErr.Code != rooms.RoomErrorInvalidRoomState {
			t.Fatalf("expected error code %q for state %q, got %q", rooms.RoomErrorInvalidRoomState, state, roomErr.Code)
		}
		if room.State != state {
			t.Fatalf("expected rejected room to stay %q, got %q", state, room.State)
		}
		if !room.MembersSnapshot()[0].Ready {
			t.Fatalf("expected rejected reset to preserve ready state for %q", state)
		}
		if room.Game == nil {
			t.Fatalf("expected rejected reset to preserve game for %q", state)
		}
	}
}

func TestRoomMembersSnapshotIsCopy(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")

	members := room.MembersSnapshot()
	members[0].SetReady(true)

	freshMembers := room.MembersSnapshot()
	if freshMembers[0].Ready {
		t.Fatal("expected member snapshot mutation not to mutate room")
	}
}

func TestRoomShouldCleanupWhenEmptyAcrossLifecycleStates(t *testing.T) {
	for _, state := range []rooms.RoomState{
		rooms.RoomStateLobby,
		rooms.RoomStateInGame,
		rooms.RoomStateGameOver,
	} {
		room := rooms.NewRoom("TEST", state, nil)

		if !room.IsEmpty() {
			t.Fatalf("expected empty %q room to be empty", state)
		}
		if !room.ShouldCleanup() {
			t.Fatalf("expected empty %q room to be cleanup-eligible", state)
		}
	}
}

func TestRoomShouldCleanupRejectsNonEmptyRooms(t *testing.T) {
	memberRoom := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	memberRoom.AddMemberID("session-1")

	if memberRoom.IsEmpty() {
		t.Fatal("expected room with member not to be empty")
	}
	if memberRoom.ShouldCleanup() {
		t.Fatal("expected room with member not to be cleanup-eligible")
	}

	activeRoom := rooms.NewRoom("TEST", rooms.RoomStateInGame, game.New())
	activeRoom.ActivePlayers = 1

	if activeRoom.IsEmpty() {
		t.Fatal("expected room with active player not to be empty")
	}
	if activeRoom.ShouldCleanup() {
		t.Fatal("expected room with active player not to be cleanup-eligible")
	}
}

func TestNilRoomShouldNotCleanup(t *testing.T) {
	var room *rooms.Room

	if room.ShouldCleanup() {
		t.Fatal("expected nil room not to be cleanup-eligible")
	}
}

func TestRoomIsGameOverReturnsFalseForLobbyOrMissingGame(t *testing.T) {
	var missingRoom *rooms.Room
	if missingRoom.IsGameOver() {
		t.Fatal("expected nil room not to be game over")
	}

	lobbyRoom := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	if lobbyRoom.IsGameOver() {
		t.Fatal("expected lobby room not to be game over")
	}

	inGameWithoutGame := rooms.NewRoom("TEST", rooms.RoomStateInGame, nil)
	if inGameWithoutGame.IsGameOver() {
		t.Fatal("expected in-game room without game not to be game over")
	}
}

func TestRoomIsGameOverDelegatesToInGameGame(t *testing.T) {
	activeGame := game.New()
	activeGame.AddPlayer()
	activeRoom := rooms.NewRoom("TEST", rooms.RoomStateInGame, activeGame)
	if activeRoom.IsGameOver() {
		t.Fatal("expected in-game room to reflect active game as not game over")
	}

	finishedGame := game.New()
	markGameOver(t, finishedGame)
	finishedRoom := rooms.NewRoom("TEST", rooms.RoomStateInGame, finishedGame)
	if !finishedRoom.IsGameOver() {
		t.Fatal("expected in-game room to delegate finished game-over state")
	}
}

func TestRoomIsGameOverIgnoresFinishedGameOutsideInGameState(t *testing.T) {
	finishedGame := game.New()
	markGameOver(t, finishedGame)
	room := rooms.NewRoom("TEST", rooms.RoomStateGameOver, finishedGame)

	if room.IsGameOver() {
		t.Fatal("expected game-over room not to report game-over detection outside in-game state")
	}
	if room.State != rooms.RoomStateGameOver {
		t.Fatalf("expected room state to remain %q, got %q", rooms.RoomStateGameOver, room.State)
	}
}

func markGameOver(t *testing.T, gameInstance *game.Game) {
	t.Helper()

	playerID := gameInstance.AddPlayer()
	value := reflect.ValueOf(gameInstance).Elem()
	session := exportRoomTestValue(value.FieldByName("playerSessions")).
		MapIndex(reflect.ValueOf(playerID))
	exportRoomTestValue(session.Elem().FieldByName("Lives")).SetInt(0)
	players := exportRoomTestValue(value.FieldByName("state").FieldByName("Players"))
	players.SetMapIndex(reflect.ValueOf(playerID), reflect.Value{})
}

func exportRoomTestValue(value reflect.Value) reflect.Value {
	if value.CanSet() {
		return value
	}

	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem()
}

func gameStopped(t *testing.T, gameInstance *game.Game) bool {
	t.Helper()

	stopSimulation := exportRoomTestValue(
		reflect.ValueOf(gameInstance).Elem().FieldByName("stopSimulation"),
	).Interface().(chan struct{})
	select {
	case <-stopSimulation:
		return true
	default:
		return false
	}
}
