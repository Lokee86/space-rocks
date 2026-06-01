package networkingtests

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/networking"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

func TestRoomInitializesMemberStorage(t *testing.T) {
	room := rooms.NewRoom("abc", rooms.RoomStateLobby, nil)

	if members := room.MembersSnapshot(); len(members) != 0 {
		t.Fatalf("expected no initial room members, got %d", len(members))
	}
}

func TestRoomAddAndRemoveMember(t *testing.T) {
	room := rooms.NewRoom("abc", rooms.RoomStateLobby, nil)
	member := room.AddMemberSessionID("session-1")

	if member.SessionID != "session-1" {
		t.Fatalf("expected member session id %q, got %q", "session-1", member.SessionID)
	}
	if member.Ready {
		t.Fatal("expected new room member to start not ready")
	}
	if !member.Connected {
		t.Fatal("expected new room member to start connected")
	}
	if _, ok := room.PlayerIDForSession("session-1"); !ok {
		t.Fatal("expected member to be stored by session id")
	}

	room.RemoveMember(member.PlayerID)
	if _, ok := room.PlayerIDForSession("session-1"); ok {
		t.Fatal("expected member to be removed")
	}
}

func TestRoomMemberCountAndFullCapacity(t *testing.T) {
	room := rooms.NewRoom("abc", rooms.RoomStateLobby, nil)
	if count := room.MemberCount(); count != 0 {
		t.Fatalf("expected empty room member count 0, got %d", count)
	}
	if room.IsFull() {
		t.Fatal("expected empty room not to be full")
	}

	for index := 0; index < rooms.MaxPlayersPerRoom-1; index++ {
		room.AddMemberSessionID(string(rune('a' + index)))
	}
	if count := room.MemberCount(); count != rooms.MaxPlayersPerRoom-1 {
		t.Fatalf("expected member count %d, got %d", rooms.MaxPlayersPerRoom-1, count)
	}
	if room.IsFull() {
		t.Fatal("expected room below capacity not to be full")
	}

	room.AddMemberSessionID("last")
	if count := room.MemberCount(); count != rooms.MaxPlayersPerRoom {
		t.Fatalf("expected full room member count %d, got %d", rooms.MaxPlayersPerRoom, count)
	}
	if !room.IsFull() {
		t.Fatal("expected room at capacity to be full")
	}
}

func TestWebSocketRoomIDQueryDoesNotJoinOrSpawn(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL)+"/?room_id="+room.ID, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	time.Sleep(30 * time.Millisecond)
	if count := room.MemberCount(); count != 0 {
		t.Fatalf("expected room_id query websocket not to add room member, got %d", count)
	}
	if room.ActivePlayers != 0 {
		t.Fatalf("expected room_id query websocket not to spawn active player, got %d", room.ActivePlayers)
	}
}

func TestWebSocketWithoutRoomIDDoesNotJoinOrSpawn(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeInput}); err != nil {
		t.Fatalf("write session-only packet: %v", err)
	}
	time.Sleep(30 * time.Millisecond)
	if count := manager.RoomCount(); count != 0 {
		t.Fatalf("expected session-only websocket not to create rooms, got %d rooms", count)
	}
}

func TestCreateRoomRequestCreatesLobbyRoomWithoutGame(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	if err := conn.ReadJSON(&snapshot); err != nil {
		t.Fatalf("read room snapshot: %v", err)
	}
	if snapshot.Type != servergame.PacketTypeRoomSnapshot {
		t.Fatalf("expected room snapshot packet, got %q", snapshot.Type)
	}
	if snapshot.RoomCode == "" {
		t.Fatal("expected generated room code")
	}
	if snapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateLobby, snapshot.RoomState)
	}
	if snapshot.LocalPlayerID == "" {
		t.Fatal("expected local player id")
	}
	if len(snapshot.Members) != 1 {
		t.Fatalf("expected snapshot with 1 member, got %d", len(snapshot.Members))
	}
	if snapshot.Members[0].PlayerID != snapshot.LocalPlayerID {
		t.Fatalf("expected local player in snapshot, got member %q local %q", snapshot.Members[0].PlayerID, snapshot.LocalPlayerID)
	}
	if snapshot.Members[0].Ready {
		t.Fatal("expected new room member not to be ready")
	}
	if snapshot.MaxPlayers != rooms.MaxPlayersPerRoom {
		t.Fatalf("expected max players %d, got %d", rooms.MaxPlayersPerRoom, snapshot.MaxPlayers)
	}

	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected generated room %q to exist", snapshot.RoomCode)
	}
	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected created room state %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected lobby room not to create game simulation")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected created room member count 1, got %d", count)
	}
}

func TestCreateRoomRequestRejectsSessionAlreadyInRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write first create room request: %v", err)
	}
	var snapshot servergame.RoomSnapshot
	if err := conn.ReadJSON(&snapshot); err != nil {
		t.Fatalf("read first room snapshot: %v", err)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write second create room request: %v", err)
	}
	var roomError servergame.RoomError
	if err := conn.ReadJSON(&roomError); err != nil {
		t.Fatalf("read room error: %v", err)
	}
	if roomError.Type != servergame.PacketTypeRoomError {
		t.Fatalf("expected room error packet, got %q", roomError.Type)
	}
	if roomError.ErrorCode != rooms.RoomErrorAlreadyInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorAlreadyInRoom, roomError.ErrorCode)
	}
}

func TestStartSinglePlayerRequestCreatesInGameRoomAndStartsState(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartSinglePlayerRequest}); err != nil {
		t.Fatalf("write start single-player request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)
	if snapshot.Type != servergame.PacketTypeRoomSnapshot {
		t.Fatalf("expected room snapshot packet, got %q", snapshot.Type)
	}
	if snapshot.RoomState != string(rooms.RoomStateInGame) {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, snapshot.RoomState)
	}
	if snapshot.LocalPlayerID == "" {
		t.Fatal("expected local player id")
	}
	if len(snapshot.Members) != 1 {
		t.Fatalf("expected single-player snapshot with 1 member, got %d", len(snapshot.Members))
	}
	if snapshot.Members[0].PlayerID != snapshot.LocalPlayerID {
		t.Fatalf("expected local player in snapshot, got member %q local %q", snapshot.Members[0].PlayerID, snapshot.LocalPlayerID)
	}

	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected single-player room %q to exist", snapshot.RoomCode)
	}
	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.Game == nil {
		t.Fatal("expected single-player room to create game simulation")
	}
	if room.IsJoinable() {
		t.Fatal("expected single-player room not to be joinable")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected single-player room member count 1, got %d", count)
	}
	if room.ActivePlayers != 1 {
		t.Fatalf("expected single-player room active players 1, got %d", room.ActivePlayers)
	}

	var state servergame.StatePacket
	readJSON(t, conn, &state)
	if state.Type != servergame.PacketTypeState {
		t.Fatalf("expected state packet, got %q", state.Type)
	}
	if state.SelfID == "" {
		t.Fatal("expected state packet self id")
	}
	if len(state.Players) != 1 {
		t.Fatalf("expected state packet with 1 player, got %d", len(state.Players))
	}
	if _, ok := state.Players[state.SelfID]; !ok {
		t.Fatalf("expected state packet players to include self id %q", state.SelfID)
	}
}

func TestStartSinglePlayerRequestRejectsSessionAlreadyInRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartSinglePlayerRequest}); err != nil {
		t.Fatalf("write start single-player request: %v", err)
	}
	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.Type != servergame.PacketTypeRoomError {
		t.Fatalf("expected room error packet, got %q", roomError.Type)
	}
	if roomError.ErrorCode != rooms.RoomErrorAlreadyInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorAlreadyInRoom, roomError.ErrorCode)
	}
}

func TestSetTargetPlayerRequestUpdatesCanonicalTargetInState(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartSinglePlayerRequest}); err != nil {
		t.Fatalf("write start single-player request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)
	var initialState servergame.StatePacket
	readJSON(t, conn, &initialState)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:           servergame.PacketTypeSetTargetPlayerRequest,
		TargetPlayerID: initialState.SelfID,
	}); err != nil {
		t.Fatalf("write set target player request: %v", err)
	}

	var updatedState servergame.StatePacket
	readJSON(t, conn, &updatedState)
	selfState, ok := updatedState.Players[updatedState.SelfID]
	if !ok {
		t.Fatalf("expected updated state to include self player %q", updatedState.SelfID)
	}
	if selfState.TargetPlayerID != updatedState.SelfID {
		t.Fatalf("expected target_player_id %q, got %q", updatedState.SelfID, selfState.TargetPlayerID)
	}
}

func TestSetTargetPlayerRequestInvalidTargetDoesNotOverwriteExistingTarget(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartSinglePlayerRequest}); err != nil {
		t.Fatalf("write start single-player request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)
	var initialState servergame.StatePacket
	readJSON(t, conn, &initialState)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:           servergame.PacketTypeSetTargetPlayerRequest,
		TargetPlayerID: initialState.SelfID,
	}); err != nil {
		t.Fatalf("write valid set target player request: %v", err)
	}
	var targetedState servergame.StatePacket
	readJSON(t, conn, &targetedState)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:           servergame.PacketTypeSetTargetPlayerRequest,
		TargetPlayerID: "player-missing",
	}); err != nil {
		t.Fatalf("write invalid set target player request: %v", err)
	}
	var afterInvalidState servergame.StatePacket
	readJSON(t, conn, &afterInvalidState)

	selfState, ok := afterInvalidState.Players[afterInvalidState.SelfID]
	if !ok {
		t.Fatalf("expected state to include self player %q", afterInvalidState.SelfID)
	}
	if selfState.TargetPlayerID != afterInvalidState.SelfID {
		t.Fatalf("expected invalid request to keep target_player_id %q, got %q", afterInvalidState.SelfID, selfState.TargetPlayerID)
	}
}

func TestSetTargetPlayerRequestEmptyTargetClearsTarget(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartSinglePlayerRequest}); err != nil {
		t.Fatalf("write start single-player request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)
	var initialState servergame.StatePacket
	readJSON(t, conn, &initialState)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:           servergame.PacketTypeSetTargetPlayerRequest,
		TargetPlayerID: initialState.SelfID,
	}); err != nil {
		t.Fatalf("write set target player request: %v", err)
	}
	var targetedState servergame.StatePacket
	readJSON(t, conn, &targetedState)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:           servergame.PacketTypeSetTargetPlayerRequest,
		TargetPlayerID: "",
	}); err != nil {
		t.Fatalf("write clear target player request: %v", err)
	}
	var clearedState servergame.StatePacket
	readJSON(t, conn, &clearedState)

	selfState, ok := clearedState.Players[clearedState.SelfID]
	if !ok {
		t.Fatalf("expected state to include self player %q", clearedState.SelfID)
	}
	if selfState.TargetPlayerID != "" {
		t.Fatalf("expected clear target request to produce empty target_player_id, got %q", selfState.TargetPlayerID)
	}
}

func TestSelectTargetAtPositionRequestRoutesToServerTargetSelection(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartSinglePlayerRequest}); err != nil {
		t.Fatalf("write start single-player request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)
	var initialState servergame.StatePacket
	readJSON(t, conn, &initialState)

	selfState, ok := initialState.Players[initialState.SelfID]
	if !ok {
		t.Fatalf("expected initial state to include self player %q", initialState.SelfID)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSelectTargetAtPositionRequest,
		X:          selfState.X,
		Y:          selfState.Y,
		TargetKind: "player",
		TargetID:   initialState.SelfID,
	}); err != nil {
		t.Fatalf("write select target at position request: %v", err)
	}

	var updatedState servergame.StatePacket
	readJSON(t, conn, &updatedState)
	updatedSelfState, ok := updatedState.Players[updatedState.SelfID]
	if !ok {
		t.Fatalf("expected updated state to include self player %q", updatedState.SelfID)
	}
	if updatedSelfState.TargetKind != "player" {
		t.Fatalf("expected target_kind %q, got %q", "player", updatedSelfState.TargetKind)
	}
	if updatedSelfState.TargetID != updatedState.SelfID {
		t.Fatalf("expected target_id %q, got %q", updatedState.SelfID, updatedSelfState.TargetID)
	}
}

func TestClearTargetRequestClearsGenericTarget(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartSinglePlayerRequest}); err != nil {
		t.Fatalf("write start single-player request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)
	var initialState servergame.StatePacket
	readJSON(t, conn, &initialState)

	selfState, ok := initialState.Players[initialState.SelfID]
	if !ok {
		t.Fatalf("expected initial state to include self player %q", initialState.SelfID)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSelectTargetAtPositionRequest,
		X:          selfState.X,
		Y:          selfState.Y,
		TargetKind: "player",
		TargetID:   initialState.SelfID,
	}); err != nil {
		t.Fatalf("write select target at position request: %v", err)
	}
	var targetedState servergame.StatePacket
	readJSON(t, conn, &targetedState)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type: servergame.PacketTypeClearTargetRequest,
	}); err != nil {
		t.Fatalf("write clear target request: %v", err)
	}
	var clearedState servergame.StatePacket
	readJSON(t, conn, &clearedState)

	clearedSelfState, ok := clearedState.Players[clearedState.SelfID]
	if !ok {
		t.Fatalf("expected cleared state to include self player %q", clearedState.SelfID)
	}
	if clearedSelfState.TargetKind != "" {
		t.Fatalf("expected cleared target_kind to be empty, got %q", clearedSelfState.TargetKind)
	}
	if clearedSelfState.TargetID != "" {
		t.Fatalf("expected cleared target_id to be empty, got %q", clearedSelfState.TargetID)
	}
	if clearedSelfState.TargetPlayerID != "" {
		t.Fatalf("expected cleared target_player_id to be empty, got %q", clearedSelfState.TargetPlayerID)
	}
}

func TestJoinRoomRequestJoinsLobbyAndBroadcastsSnapshots(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	creator, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial creator websocket: %v", err)
	}
	defer creator.Close()
	if err := creator.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &createdSnapshot)

	joiner, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial joiner websocket: %v", err)
	}
	defer joiner.Close()
	if err := joiner.WriteJSON(servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: createdSnapshot.RoomCode,
	}); err != nil {
		t.Fatalf("write join room request: %v", err)
	}

	var creatorSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &creatorSnapshot)
	var joinerSnapshot servergame.RoomSnapshot
	readJSON(t, joiner, &joinerSnapshot)

	if creatorSnapshot.RoomCode != createdSnapshot.RoomCode {
		t.Fatalf("expected creator broadcast room code %q, got %q", createdSnapshot.RoomCode, creatorSnapshot.RoomCode)
	}
	if joinerSnapshot.RoomCode != createdSnapshot.RoomCode {
		t.Fatalf("expected joiner snapshot room code %q, got %q", createdSnapshot.RoomCode, joinerSnapshot.RoomCode)
	}
	if len(creatorSnapshot.Members) != 2 {
		t.Fatalf("expected creator broadcast with 2 members, got %d", len(creatorSnapshot.Members))
	}
	if len(joinerSnapshot.Members) != 2 {
		t.Fatalf("expected joiner snapshot with 2 members, got %d", len(joinerSnapshot.Members))
	}

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	if count := room.MemberCount(); count != 2 {
		t.Fatalf("expected room member count 2, got %d", count)
	}
}

func TestJoinRoomRequestRejectsNonexistentRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: "ZZZZZZ",
	}); err != nil {
		t.Fatalf("write join room request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorRoomNotFound {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomNotFound, roomError.ErrorCode)
	}
}

func TestJoinRoomRequestRejectsInGameRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	room, roomErr := manager.CreateStartedSinglePlayerRoom("session-owner")
	if roomErr != nil {
		t.Fatalf("create started single-player room: %v", roomErr)
	}
	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: room.ID,
	}); err != nil {
		t.Fatalf("write join room request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorRoomInGame {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomInGame, roomError.ErrorCode)
	}
}

func TestJoinRoomRequestRejectsFullRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	for index := 0; index < rooms.MaxPlayersPerRoom; index++ {
		room.AddMemberSessionID(string(rune('a' + index)))
	}

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: room.ID,
	}); err != nil {
		t.Fatalf("write join room request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorRoomFull {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomFull, roomError.ErrorCode)
	}
}

func TestLeaveRoomRequestRejectsSessionNotInRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeLeaveRoomRequest}); err != nil {
		t.Fatalf("write leave room request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomError.ErrorCode)
	}
}

func TestLeaveRoomRequestRemovesMemberAndBroadcastsSnapshot(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	creator, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial creator websocket: %v", err)
	}
	defer creator.Close()
	if err := creator.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &createdSnapshot)

	joiner, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial joiner websocket: %v", err)
	}
	defer joiner.Close()
	if err := joiner.WriteJSON(servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: createdSnapshot.RoomCode,
	}); err != nil {
		t.Fatalf("write join room request: %v", err)
	}
	var joinedCreatorSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &joinedCreatorSnapshot)
	var joinedJoinerSnapshot servergame.RoomSnapshot
	readJSON(t, joiner, &joinedJoinerSnapshot)

	if err := joiner.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeLeaveRoomRequest}); err != nil {
		t.Fatalf("write leave room request: %v", err)
	}
	var leaveSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &leaveSnapshot)

	if leaveSnapshot.RoomCode != createdSnapshot.RoomCode {
		t.Fatalf("expected leave broadcast room code %q, got %q", createdSnapshot.RoomCode, leaveSnapshot.RoomCode)
	}
	if len(leaveSnapshot.Members) != 1 {
		t.Fatalf("expected leave broadcast with 1 member, got %d", len(leaveSnapshot.Members))
	}
	if leaveSnapshot.Members[0].PlayerID != createdSnapshot.LocalPlayerID {
		t.Fatalf("expected remaining member %q, got %q", createdSnapshot.LocalPlayerID, leaveSnapshot.Members[0].PlayerID)
	}
	if leaveSnapshot.Members[0].PlayerID == joinedJoinerSnapshot.LocalPlayerID {
		t.Fatalf("expected leaving member %q to be excluded from leave broadcast", joinedJoinerSnapshot.LocalPlayerID)
	}

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected room member count 1 after leave, got %d", count)
	}
}

func TestLeaveRoomRequestSchedulesEmptyRoomCleanup(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeLeaveRoomRequest}); err != nil {
		t.Fatalf("write leave room request: %v", err)
	}

	if !waitUntil(200*time.Millisecond, func() bool {
		_, ok := manager.Find(snapshot.RoomCode)
		return !ok
	}) {
		t.Fatalf("expected empty room %q to be cleaned up", snapshot.RoomCode)
	}
}

func TestJoinAfterEmptyRoomCleanupFails(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()
	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeLeaveRoomRequest}); err != nil {
		t.Fatalf("write leave room request: %v", err)
	}
	if !waitUntil(200*time.Millisecond, func() bool {
		_, ok := manager.Find(snapshot.RoomCode)
		return !ok
	}) {
		t.Fatalf("expected empty room %q to be cleaned up", snapshot.RoomCode)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: snapshot.RoomCode,
	}); err != nil {
		t.Fatalf("write join room request after cleanup: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorRoomNotFound {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomNotFound, roomError.ErrorCode)
	}
}

func TestLeaveRoomRequestClearsSessionRoomAssociation(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeLeaveRoomRequest}); err != nil {
		t.Fatalf("write first leave room request: %v", err)
	}
	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeLeaveRoomRequest}); err != nil {
		t.Fatalf("write second leave room request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomError.ErrorCode)
	}

	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to remain until cleanup", snapshot.RoomCode)
	}
	if count := room.MemberCount(); count != 0 {
		t.Fatalf("expected room member count 0 after leave, got %d", count)
	}
}

func TestSetReadyRequestUpdatesMemberAndBroadcastsSnapshot(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write set ready request: %v", err)
	}

	var readySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &readySnapshot)
	if readySnapshot.Type != servergame.PacketTypeRoomSnapshot {
		t.Fatalf("expected room snapshot packet, got %q", readySnapshot.Type)
	}
	if readySnapshot.RoomCode != createdSnapshot.RoomCode {
		t.Fatalf("expected ready snapshot room code %q, got %q", createdSnapshot.RoomCode, readySnapshot.RoomCode)
	}
	if len(readySnapshot.Members) != 1 {
		t.Fatalf("expected ready snapshot with 1 member, got %d", len(readySnapshot.Members))
	}
	if readySnapshot.Members[0].PlayerID != createdSnapshot.LocalPlayerID {
		t.Fatalf("expected ready member %q, got %q", createdSnapshot.LocalPlayerID, readySnapshot.Members[0].PlayerID)
	}
	if !readySnapshot.Members[0].Ready {
		t.Fatal("expected member to be ready")
	}

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	members := room.MembersSnapshot()
	if len(members) != 1 {
		t.Fatalf("expected room with 1 member, got %d", len(members))
	}
	if !members[0].Ready {
		t.Fatal("expected room member ready state to be true")
	}
}

func TestSetReadyRequestRejectsSessionNotInRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write set ready request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomError.ErrorCode)
	}
}

func TestSetReadyRequestRejectsNonLobbyRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	room.State = rooms.RoomStateInGame

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write set ready request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomError.ErrorCode)
	}
}

func TestStartGameRequestRejectsSessionNotInRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write start game request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomError.ErrorCode)
	}
}

func TestStartGameRequestRejectsNotReadyRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write start game request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotReady, roomError.ErrorCode)
	}
}

func TestStartGameRequestRejectsDoubleStartState(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	playerID := createdSnapshot.LocalPlayerID
	if roomErr := room.SetReadyInLobby(playerID, true); roomErr != nil {
		t.Fatalf("expected ready update for member %q, got %s", createdSnapshot.LocalPlayerID, roomErr.Code)
	}
	room.State = rooms.RoomStateInGame

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write start game request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorRoomInGame {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomInGame, roomError.ErrorCode)
	}
}

func TestStartGameRequestCreatesGameAndMarksRoomStarting(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write set ready request: %v", err)
	}
	var readySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &readySnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write start game request: %v", err)
	}

	var inGameSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &inGameSnapshot)
	if inGameSnapshot.Type != servergame.PacketTypeRoomSnapshot {
		t.Fatalf("expected room snapshot packet, got %q", inGameSnapshot.Type)
	}
	if inGameSnapshot.RoomState != string(rooms.RoomStateInGame) {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, inGameSnapshot.RoomState)
	}

	var state servergame.StatePacket
	readJSON(t, conn, &state)
	if state.Type != servergame.PacketTypeState {
		t.Fatalf("expected gameplay state packet after start, got %q", state.Type)
	}
	if state.SelfID == "" {
		t.Fatal("expected started player self id")
	}
	if len(state.Players) != 1 {
		t.Fatalf("expected 1 active game player, got %d", len(state.Players))
	}

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.Game == nil {
		t.Fatal("expected valid start request to create game")
	}
	if room.ActivePlayers != 1 {
		t.Fatalf("expected 1 active game player, got %d", room.ActivePlayers)
	}
	if readySnapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected ready snapshot to remain lobby, got %q", readySnapshot.RoomState)
	}
}

func TestReturnToLobbyRequestResetsGameOverRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write first set ready request: %v", err)
	}
	var firstReadySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &firstReadySnapshot)
	if firstReadySnapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected first ready room state %q, got %q", rooms.RoomStateLobby, firstReadySnapshot.RoomState)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write first start game request: %v", err)
	}
	var firstInGameSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &firstInGameSnapshot)
	if firstInGameSnapshot.RoomState != string(rooms.RoomStateInGame) {
		t.Fatalf("expected first start room state %q, got %q", rooms.RoomStateInGame, firstInGameSnapshot.RoomState)
	}
	var firstState servergame.StatePacket
	readJSON(t, conn, &firstState)
	if len(firstState.Players) != 1 {
		t.Fatalf("expected first match to spawn 1 player, got %d", len(firstState.Players))
	}

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	if room.ActivePlayers != 1 {
		t.Fatalf("expected first match active players 1, got %d", room.ActivePlayers)
	}
	room.State = rooms.RoomStateGameOver

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeReturnToLobbyRequest}); err != nil {
		t.Fatalf("write return to lobby request: %v", err)
	}

	var lobbySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &lobbySnapshot)
	if lobbySnapshot.Type != servergame.PacketTypeRoomSnapshot {
		t.Fatalf("expected room snapshot packet, got %q", lobbySnapshot.Type)
	}
	if lobbySnapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected snapshot room state %q, got %q", rooms.RoomStateLobby, lobbySnapshot.RoomState)
	}
	if len(lobbySnapshot.Members) != 1 {
		t.Fatalf("expected snapshot with 1 member, got %d", len(lobbySnapshot.Members))
	}
	if lobbySnapshot.Members[0].Ready {
		t.Fatal("expected snapshot ready state to be cleared")
	}

	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected return to lobby to clear game")
	}
	if room.ActivePlayers != 0 {
		t.Fatalf("expected return to lobby to clear active players, got %d", room.ActivePlayers)
	}
	members := room.MembersSnapshot()
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Ready {
		t.Fatal("expected return to lobby to clear ready state")
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write second set ready request: %v", err)
	}
	var secondReadySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &secondReadySnapshot)
	if secondReadySnapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected second ready room state %q, got %q", rooms.RoomStateLobby, secondReadySnapshot.RoomState)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write second start game request: %v", err)
	}
	var secondInGameSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &secondInGameSnapshot)
	if secondInGameSnapshot.RoomState != string(rooms.RoomStateInGame) {
		t.Fatalf("expected second start room state %q, got %q", rooms.RoomStateInGame, secondInGameSnapshot.RoomState)
	}
	var secondState servergame.StatePacket
	readJSON(t, conn, &secondState)
	if secondState.SelfID == "" {
		t.Fatal("expected second match self id")
	}
	if len(secondState.Players) != 1 {
		t.Fatalf("expected second match to spawn 1 fresh player, got %d", len(secondState.Players))
	}
	if _, ok := secondState.Players[secondState.SelfID]; !ok {
		t.Fatalf("expected second match state to include self player %q", secondState.SelfID)
	}
	if room.ActivePlayers != 1 {
		t.Fatalf("expected second match active players 1, got %d", room.ActivePlayers)
	}
	if room.MemberCount() != 1 {
		t.Fatalf("expected room membership to remain intact, got %d", room.MemberCount())
	}
}

func TestReturnToLobbyAllowsFreshSecondMatch(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write first ready request: %v", err)
	}
	var firstReadySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &firstReadySnapshot)
	if firstReadySnapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected first ready room state %q, got %q", rooms.RoomStateLobby, firstReadySnapshot.RoomState)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write first start request: %v", err)
	}
	var firstInGameSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &firstInGameSnapshot)
	if firstInGameSnapshot.RoomState != string(rooms.RoomStateInGame) {
		t.Fatalf("expected first start room state %q, got %q", rooms.RoomStateInGame, firstInGameSnapshot.RoomState)
	}
	var firstState servergame.StatePacket
	readJSON(t, conn, &firstState)
	if firstState.SelfID == "" {
		t.Fatal("expected first match self id")
	}
	if firstState.SelfID != "player-1" {
		t.Fatalf("expected first fresh game player id %q, got %q", "player-1", firstState.SelfID)
	}
	if _, ok := firstState.Players[firstState.SelfID]; !ok {
		t.Fatalf("expected first match state to include self player %q", firstState.SelfID)
	}

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	room.State = rooms.RoomStateGameOver

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeReturnToLobbyRequest}); err != nil {
		t.Fatalf("write return to lobby request: %v", err)
	}
	var lobbySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &lobbySnapshot)
	if lobbySnapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected return snapshot room state %q, got %q", rooms.RoomStateLobby, lobbySnapshot.RoomState)
	}
	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected room state %q after return, got %q", rooms.RoomStateLobby, room.State)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write second ready request: %v", err)
	}
	var secondReadySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &secondReadySnapshot)
	if secondReadySnapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected second ready room state %q, got %q", rooms.RoomStateLobby, secondReadySnapshot.RoomState)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write second start request: %v", err)
	}
	var secondInGameSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &secondInGameSnapshot)
	if secondInGameSnapshot.RoomState != string(rooms.RoomStateInGame) {
		t.Fatalf("expected second start room state %q, got %q", rooms.RoomStateInGame, secondInGameSnapshot.RoomState)
	}
	var secondState servergame.StatePacket
	readJSON(t, conn, &secondState)
	if secondState.SelfID == "" {
		t.Fatal("expected second match self id")
	}
	if secondState.SelfID != "player-1" {
		t.Fatalf("expected second fresh game player id %q, got %q", "player-1", secondState.SelfID)
	}
	if _, ok := secondState.Players[secondState.SelfID]; !ok {
		t.Fatalf("expected second match state to include self player %q", secondState.SelfID)
	}
	if len(secondState.Players) != 1 {
		t.Fatalf("expected second match to create 1 active player, got %d", len(secondState.Players))
	}
	if room.ActivePlayers != 1 {
		t.Fatalf("expected second match active players 1, got %d", room.ActivePlayers)
	}
}

func TestReturnToLobbyRequestRejectsSessionNotInRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeReturnToLobbyRequest}); err != nil {
		t.Fatalf("write return to lobby request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomError.ErrorCode)
	}
}

func TestReturnToLobbyRequestRejectsNonGameOverRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeReturnToLobbyRequest}); err != nil {
		t.Fatalf("write return to lobby request: %v", err)
	}

	var roomError servergame.RoomError
	readJSON(t, conn, &roomError)
	if roomError.ErrorCode != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomError.ErrorCode)
	}
}

func TestDisconnectLeavesLobbyAndBroadcastsSnapshot(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	creator, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial creator websocket: %v", err)
	}
	defer creator.Close()
	if err := creator.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &createdSnapshot)

	joiner, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial joiner websocket: %v", err)
	}
	if err := joiner.WriteJSON(servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: createdSnapshot.RoomCode,
	}); err != nil {
		t.Fatalf("write join room request: %v", err)
	}
	var joinedCreatorSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &joinedCreatorSnapshot)
	var joinedJoinerSnapshot servergame.RoomSnapshot
	readJSON(t, joiner, &joinedJoinerSnapshot)

	if err := joiner.Close(); err != nil {
		t.Fatalf("close joiner websocket: %v", err)
	}
	var disconnectSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &disconnectSnapshot)

	if len(disconnectSnapshot.Members) != 1 {
		t.Fatalf("expected disconnect broadcast with 1 member, got %d", len(disconnectSnapshot.Members))
	}
	if disconnectSnapshot.Members[0].PlayerID != createdSnapshot.LocalPlayerID {
		t.Fatalf("expected remaining member %q, got %q", createdSnapshot.LocalPlayerID, disconnectSnapshot.Members[0].PlayerID)
	}
	if !waitUntil(100*time.Millisecond, func() bool {
		room, ok := manager.Find(createdSnapshot.RoomCode)
		return ok && room.MemberCount() == 1
	}) {
		t.Fatalf("expected non-empty room %q to survive disconnect cleanup", createdSnapshot.RoomCode)
	}
}

func TestDisconnectCleansEmptyLobbyRoom(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)

	if err := conn.Close(); err != nil {
		t.Fatalf("close websocket: %v", err)
	}
	if !waitUntil(200*time.Millisecond, func() bool {
		_, ok := manager.Find(snapshot.RoomCode)
		return !ok
	}) {
		t.Fatalf("expected empty lobby room %q to be cleaned up after disconnect", snapshot.RoomCode)
	}
}

func TestDisconnectCleansEmptyGameOverRoom(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)

	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", snapshot.RoomCode)
	}
	room.State = rooms.RoomStateGameOver

	if err := conn.Close(); err != nil {
		t.Fatalf("close websocket: %v", err)
	}
	if !waitUntil(200*time.Millisecond, func() bool {
		_, ok := manager.Find(snapshot.RoomCode)
		return !ok
	}) {
		t.Fatalf("expected empty game-over room %q to be cleaned up after disconnect", snapshot.RoomCode)
	}
}

func TestDisconnectCleansEmptyInGameRoom(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &createdSnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:  servergame.PacketTypeSetReadyRequest,
		Ready: true,
	}); err != nil {
		t.Fatalf("write set ready request: %v", err)
	}
	var readySnapshot servergame.RoomSnapshot
	readJSON(t, conn, &readySnapshot)

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartGameRequest}); err != nil {
		t.Fatalf("write start game request: %v", err)
	}
	var inGameSnapshot servergame.RoomSnapshot
	readJSON(t, conn, &inGameSnapshot)
	if inGameSnapshot.RoomState != string(rooms.RoomStateInGame) {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, inGameSnapshot.RoomState)
	}
	if readySnapshot.RoomCode != createdSnapshot.RoomCode {
		t.Fatalf("expected ready snapshot room %q, got %q", createdSnapshot.RoomCode, readySnapshot.RoomCode)
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("close websocket: %v", err)
	}
	if !waitUntil(200*time.Millisecond, func() bool {
		_, ok := manager.Find(createdSnapshot.RoomCode)
		return !ok
	}) {
		t.Fatalf("expected empty in-game room %q to be cleaned up after disconnect", createdSnapshot.RoomCode)
	}
}

func waitUntil(timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(time.Millisecond)
	}

	return condition()
}

func webSocketURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http")
}

func readJSON(t *testing.T, conn *websocket.Conn, value any) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		t.Fatalf("set websocket read deadline: %v", err)
	}
	if err := conn.ReadJSON(value); err != nil {
		t.Fatalf("read websocket JSON: %v", err)
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		t.Fatalf("clear websocket read deadline: %v", err)
	}
}
