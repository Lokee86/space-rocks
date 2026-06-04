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
	_, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected room to have member")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected member count 1, got %d", count)
	}

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session to resolve before removal")
	}
	room.RemoveMember(playerID)
	_, ok = room.PlayerIDForSession("session-1")
	if ok {
		t.Fatal("expected room member to be removed")
	}
}

func TestRoomSetReadyInLobbyBySession(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")

	setReadyInLobbyBySession(t, room, "session-1", true)

	members := room.MembersSnapshot()
	if len(members) != 1 {
		t.Fatalf("expected 1 member snapshot, got %d", len(members))
	}
	if !members[0].Ready {
		t.Fatal("expected member to be ready")
	}

	if _, ok := room.PlayerIDForSession("missing"); ok {
		t.Fatal("expected missing session not to resolve to a player")
	}
}

func TestRoomValidateStartAllowsOneReadyMember(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")
	setReadyInLobbyBySession(t, room, "session-1", true)

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session to resolve before start validation")
	}
	if roomErr := room.ValidateStart(playerID); roomErr != nil {
		t.Fatalf("expected ready solo member to start, got %s", roomErr.Code)
	}
}

func TestRoomValidateStartRequiresRequesterInRoom(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")
	setReadyInLobbyBySession(t, room, "session-1", true)

	roomErr := room.ValidateStart("missing-player")
	if roomErr == nil {
		t.Fatal("expected missing requester to be rejected")
	}
	if roomErr.Code != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomErr.Code)
	}
}

func TestRoomValidateStartRequiresLobbyState(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")
	setReadyInLobbyBySession(t, room, "session-1", true)
	room.State = rooms.RoomStateGameOver

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session to resolve before start validation")
	}
	roomErr := room.ValidateStart(playerID)
	if roomErr == nil {
		t.Fatal("expected non-lobby room to be rejected")
	}
	if roomErr.Code != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomErr.Code)
	}
}

func TestRoomValidateStartRejectsDoubleStartStates(t *testing.T) {
	for _, state := range []rooms.RoomState{rooms.RoomStateStarting, rooms.RoomStateInGame} {
		room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
		room.AddMemberSessionID("session-1")
		setReadyInLobbyBySession(t, room, "session-1", true)
		room.State = state

		playerID, ok := room.PlayerIDForSession("session-1")
		if !ok {
			t.Fatal("expected session to resolve before start validation")
		}
		roomErr := room.ValidateStart(playerID)
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
	room.AddMemberSessionID("session-1")
	room.AddMemberSessionID("session-2")
	setReadyInLobbyBySession(t, room, "session-1", true)

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session to resolve before start validation")
	}
	roomErr := room.ValidateStart(playerID)
	if roomErr == nil {
		t.Fatal("expected unready connected member to block start")
	}
	if roomErr.Code != rooms.RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotReady, roomErr.Code)
	}
}

func TestRoomValidateStartIgnoresDisconnectedUnreadyMembers(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")
	disconnected := room.AddMemberSessionID("session-2")
	disconnected.MarkDisconnected()
	setReadyInLobbyBySession(t, room, "session-1", true)

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session to resolve before start validation")
	}
	if roomErr := room.ValidateStart(playerID); roomErr != nil {
		t.Fatalf("expected disconnected unready member not to block start, got %s", roomErr.Code)
	}
}

func TestRoomStartGameForMemberAllowsOwnerWhenConnectedMembersReady(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")
	room.AddMemberSessionID("session-2")
	setReadyInLobbyBySession(t, room, "session-1", true)
	setReadyInLobbyBySession(t, room, "session-2", true)

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected owner session to resolve before start")
	}
	if roomErr := room.StartGameForMember(playerID, game.New); roomErr != nil {
		t.Fatalf("expected owner start to succeed, got %s", roomErr.Code)
	}
	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.GameInstance() == nil {
		t.Fatal("expected game to be created")
	}
}

func TestRoomStartGameForMemberRejectsNonOwner(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")
	room.AddMemberSessionID("session-2")
	setReadyInLobbyBySession(t, room, "session-1", true)
	setReadyInLobbyBySession(t, room, "session-2", true)

	nonOwnerID, ok := room.PlayerIDForSession("session-2")
	if !ok {
		t.Fatal("expected non-owner session to resolve before start")
	}
	roomErr := room.StartGameForMember(nonOwnerID, game.New)
	if roomErr == nil {
		t.Fatal("expected non-owner to be rejected")
	}
	if roomErr.Code != rooms.RoomErrorNotRoomOwner {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotRoomOwner, roomErr.Code)
	}
}

func TestRoomStartGameForMemberRejectsConnectedUnreadyMember(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")
	room.AddMemberSessionID("session-2")
	setReadyInLobbyBySession(t, room, "session-1", true)

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected owner session to resolve before start")
	}
	roomErr := room.StartGameForMember(playerID, game.New)
	if roomErr == nil {
		t.Fatal("expected unready connected member to block start")
	}
	if roomErr.Code != rooms.RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotReady, roomErr.Code)
	}
}

func TestRoomStartGameForMemberIgnoresDisconnectedUnreadyMember(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")
	disconnected := room.AddMemberSessionID("session-2")
	disconnected.MarkDisconnected()
	setReadyInLobbyBySession(t, room, "session-1", true)

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected owner session to resolve before start")
	}
	if roomErr := room.StartGameForMember(playerID, game.New); roomErr != nil {
		t.Fatalf("expected disconnected unready member not to block start, got %s", roomErr.Code)
	}
}

func TestRoomStartGameForMemberRejectsStartingRoom(t *testing.T) {
	for _, state := range []rooms.RoomState{rooms.RoomStateStarting, rooms.RoomStateInGame} {
		room := rooms.NewRoom("TEST", state, nil)
		room.AddMemberSessionID("session-1")
		playerID, ok := room.PlayerIDForSession("session-1")
		if !ok {
			t.Fatal("expected session to resolve before start")
		}

		roomErr := room.StartGameForMember(playerID, game.New)
		if roomErr == nil {
			t.Fatalf("expected state %q to reject start", state)
		}
		if roomErr.Code != rooms.RoomErrorRoomInGame {
			t.Fatalf("expected error code %q for state %q, got %q", rooms.RoomErrorRoomInGame, state, roomErr.Code)
		}
	}
}

func TestRoomStartGameForMemberRejectsGameOverRoom(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateGameOver, nil)
	room.AddMemberSessionID("session-1")
	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session to resolve before start")
	}

	roomErr := room.StartGameForMember(playerID, game.New)
	if roomErr == nil {
		t.Fatal("expected game-over room to reject start")
	}
	if roomErr.Code != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomErr.Code)
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

func TestRoomStartSinglePlayerGameFromLobbySucceeds(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")

	if roomErr := room.StartSinglePlayerGame(game.New); roomErr != nil {
		t.Fatalf("expected single-player start to succeed, got %s", roomErr.Code)
	}
	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.GameInstance() == nil {
		t.Fatal("expected game instance to be created")
	}
}

func TestRoomStartSinglePlayerGameRequiresMember(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)

	roomErr := room.StartSinglePlayerGame(game.New)
	if roomErr == nil {
		t.Fatal("expected empty room to reject single-player start")
	}
	if roomErr.Code != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomErr.Code)
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
	if room.GameInstance() == nil {
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

func TestRoomMarkStartingRequiresLobby(t *testing.T) {
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
	}
}

func TestRoomMarkInGameRequiresStarting(t *testing.T) {
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
	if room.GameInstance() == nil {
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

func TestRoomMarkGameOverIfCompleteSkipsIncompleteRooms(t *testing.T) {
	var missingRoom *rooms.Room
	if missingRoom.MarkGameOverIfComplete() {
		t.Fatal("expected nil room not to transition")
	}

	lobbyRoom := rooms.NewRoom("TEST", rooms.RoomStateLobby, game.New())
	if lobbyRoom.MarkGameOverIfComplete() {
		t.Fatal("expected lobby room not to transition")
	}
	if lobbyRoom.State != rooms.RoomStateLobby {
		t.Fatalf("expected lobby room state to remain %q, got %q", rooms.RoomStateLobby, lobbyRoom.State)
	}

	inGameWithoutGame := rooms.NewRoom("TEST", rooms.RoomStateInGame, nil)
	if inGameWithoutGame.MarkGameOverIfComplete() {
		t.Fatal("expected in-game room without game not to transition")
	}
	if inGameWithoutGame.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state to remain %q, got %q", rooms.RoomStateInGame, inGameWithoutGame.State)
	}

	activeGame := game.New()
	activeGame.AddPlayer()
	activeRoom := rooms.NewRoom("TEST", rooms.RoomStateInGame, activeGame)
	if activeRoom.MarkGameOverIfComplete() {
		t.Fatal("expected active game not to transition")
	}
	if activeRoom.State != rooms.RoomStateInGame {
		t.Fatalf("expected active room state to remain %q, got %q", rooms.RoomStateInGame, activeRoom.State)
	}
}

func TestRoomMarkGameOverIfCompleteTransitionsFinishedGame(t *testing.T) {
	finishedGame := game.New()
	markGameOver(t, finishedGame)
	room := rooms.NewRoom("TEST", rooms.RoomStateInGame, finishedGame)

	if !room.MarkGameOverIfComplete() {
		t.Fatal("expected finished game to transition")
	}
	if room.State != rooms.RoomStateGameOver {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateGameOver, room.State)
	}
}

func TestRoomResetToLobbyClearsGameAndReadyStates(t *testing.T) {
	oldGame := game.New()
	room := rooms.NewRoom("TEST", rooms.RoomStateGameOver, oldGame)
	first := room.AddMemberSessionID("session-1")
	second := room.AddMemberSessionID("session-2")
	first.SetReady(true)
	second.SetReady(true)

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session to resolve before reset to lobby")
	}
	if roomErr := room.ResetToLobby(playerID); roomErr != nil {
		t.Fatalf("expected game-over room to reset to lobby, got %s", roomErr.Code)
	}
	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.GameInstance() != nil {
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
	room.AddMemberSessionID("session-1")

	roomErr := room.ResetToLobby("missing-player")
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
		member := room.AddMemberSessionID("session-1")
		member.SetReady(true)

		playerID, ok := room.PlayerIDForSession("session-1")
		if !ok {
			t.Fatal("expected session to resolve before reset to lobby")
		}
		roomErr := room.ResetToLobby(playerID)
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
		if room.GameInstance() == nil {
			t.Fatalf("expected rejected reset to preserve game for %q", state)
		}
	}
}

func TestRoomMembersSnapshotIsCopy(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberSessionID("session-1")

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
	memberRoom.AddMemberSessionID("session-1")

	if memberRoom.IsEmpty() {
		t.Fatal("expected room with member not to be empty")
	}
	if memberRoom.ShouldCleanup() {
		t.Fatal("expected room with member not to be cleanup-eligible")
	}

	activeRoom := rooms.NewRoom("TEST", rooms.RoomStateInGame, game.New())
	activeRoom.SetActivePlayerCount(1)

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
	players := exportRoomTestValue(value.FieldByName("entities").FieldByName("Players"))
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
