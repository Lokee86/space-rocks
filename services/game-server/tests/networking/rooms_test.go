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

func TestRoomManagerUsesDefaultRoomForBlankID(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	defaultRoom := manager.DefaultRoom()
	blankRoom := manager.GetOrCreate("")
	spaceRoom := manager.GetOrCreate("   ")

	if blankRoom != defaultRoom {
		t.Fatal("expected blank room id to use default room")
	}
	if spaceRoom != defaultRoom {
		t.Fatal("expected whitespace room id to use default room")
	}
}

func TestRoomManagerCreatesSeparateRoomsByID(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	first := manager.GetOrCreate("abc")
	again := manager.GetOrCreate("abc")
	second := manager.GetOrCreate("xyz")

	if first != again {
		t.Fatal("expected same room id to return same room")
	}
	if first == second {
		t.Fatal("expected different room ids to return different rooms")
	}
	if first.Game == second.Game {
		t.Fatal("expected different rooms to own different games")
	}
}

func TestCompatibilityRoomsStartInGame(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	room := manager.GetOrCreate("abc")

	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected compatibility room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.Game == nil {
		t.Fatal("expected compatibility room to create a game")
	}
}

func TestRoomInitializesMemberStorage(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	room := manager.GetOrCreate("abc")

	if members := room.MembersSnapshot(); len(members) != 0 {
		t.Fatalf("expected no initial room members, got %d", len(members))
	}
}

func TestRoomAddAndRemoveMember(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	room := manager.GetOrCreate("abc")
	member := room.AddMemberID("session-1")

	if member.SessionID != "session-1" {
		t.Fatalf("expected member session id %q, got %q", "session-1", member.SessionID)
	}
	if member.Ready {
		t.Fatal("expected new room member to start not ready")
	}
	if !member.Connected {
		t.Fatal("expected new room member to start connected")
	}
	if !room.HasMember("session-1") {
		t.Fatal("expected member to be stored by session id")
	}

	room.RemoveMember("session-1")
	if room.HasMember("session-1") {
		t.Fatal("expected member to be removed")
	}
}

func TestRoomMemberCountAndFullCapacity(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	room := manager.GetOrCreate("abc")
	if count := room.MemberCount(); count != 0 {
		t.Fatalf("expected empty room member count 0, got %d", count)
	}
	if room.IsFull() {
		t.Fatal("expected empty room not to be full")
	}

	for index := 0; index < rooms.MaxPlayersPerRoom-1; index++ {
		room.AddMemberID(string(rune('a' + index)))
	}
	if count := room.MemberCount(); count != rooms.MaxPlayersPerRoom-1 {
		t.Fatalf("expected member count %d, got %d", rooms.MaxPlayersPerRoom-1, count)
	}
	if room.IsFull() {
		t.Fatal("expected room below capacity not to be full")
	}

	room.AddMemberID("last")
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

	room := manager.GetOrCreate("abc")
	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL)+"/?room_id=abc", nil)
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

	defaultRoom := manager.DefaultRoom()
	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeInput}); err != nil {
		t.Fatalf("write session-only packet: %v", err)
	}
	time.Sleep(30 * time.Millisecond)
	if count := defaultRoom.MemberCount(); count != 0 {
		t.Fatalf("expected session-only websocket not to add default room member, got %d", count)
	}
	if count := manager.RoomCount(); count != 1 {
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
	if snapshot.LocalMemberID == "" {
		t.Fatal("expected local member id")
	}
	if len(snapshot.Members) != 1 {
		t.Fatalf("expected snapshot with 1 member, got %d", len(snapshot.Members))
	}
	if snapshot.Members[0].MemberID != snapshot.LocalMemberID {
		t.Fatalf("expected local member in snapshot, got member %q local %q", snapshot.Members[0].MemberID, snapshot.LocalMemberID)
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
	if snapshot.LocalMemberID == "" {
		t.Fatal("expected local member id")
	}
	if len(snapshot.Members) != 1 {
		t.Fatalf("expected single-player snapshot with 1 member, got %d", len(snapshot.Members))
	}
	if snapshot.Members[0].MemberID != snapshot.LocalMemberID {
		t.Fatalf("expected local member in snapshot, got member %q local %q", snapshot.Members[0].MemberID, snapshot.LocalMemberID)
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

	manager.GetOrCreate("ABCDEF")
	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:     servergame.PacketTypeJoinRoomRequest,
		RoomCode: "ABCDEF",
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
		room.AddMemberID(string(rune('a' + index)))
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
	if leaveSnapshot.Members[0].MemberID != createdSnapshot.LocalMemberID {
		t.Fatalf("expected remaining member %q, got %q", createdSnapshot.LocalMemberID, leaveSnapshot.Members[0].MemberID)
	}
	if leaveSnapshot.Members[0].MemberID == joinedJoinerSnapshot.LocalMemberID {
		t.Fatalf("expected leaving member %q to be excluded from leave broadcast", joinedJoinerSnapshot.LocalMemberID)
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
	if readySnapshot.Members[0].MemberID != createdSnapshot.LocalMemberID {
		t.Fatalf("expected ready member %q, got %q", createdSnapshot.LocalMemberID, readySnapshot.Members[0].MemberID)
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
	room.SetMemberReady(createdSnapshot.LocalMemberID, true)
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

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	room.Game = servergame.New()
	room.State = rooms.RoomStateGameOver
	room.SetMemberReady(createdSnapshot.LocalMemberID, true)

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
	members := room.MembersSnapshot()
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Ready {
		t.Fatal("expected return to lobby to clear ready state")
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
	if disconnectSnapshot.Members[0].MemberID != createdSnapshot.LocalMemberID {
		t.Fatalf("expected remaining member %q, got %q", createdSnapshot.LocalMemberID, disconnectSnapshot.Members[0].MemberID)
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

func TestRoomManagerCleansUpEmptyRoomAfterGracePeriod(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	room, leave := manager.Join("abc")
	if room == nil {
		t.Fatal("expected room")
	}
	leave()

	if !waitUntil(200*time.Millisecond, func() bool {
		return manager.GetOrCreate("abc") != room
	}) {
		t.Fatal("expected empty room to be cleaned up after grace period")
	}
}

func TestRoomManagerDoesNotCleanUpActiveRoom(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(10 * time.Millisecond)
	defer manager.StopAll()

	room, leave := manager.Join("abc")
	defer leave()

	time.Sleep(30 * time.Millisecond)
	if manager.GetOrCreate("abc") != room {
		t.Fatal("expected active room to stay alive")
	}
}

func TestRoomManagerCancelsCleanupWhenRoomIsRejoined(t *testing.T) {
	manager := networking.NewRoomManagerWithCleanupDelay(30 * time.Millisecond)
	defer manager.StopAll()

	room, leave := manager.Join("abc")
	leave()

	rejoined, leaveRejoined := manager.Join("abc")
	defer leaveRejoined()
	if rejoined != room {
		t.Fatal("expected reconnect during grace period to reuse room")
	}

	time.Sleep(60 * time.Millisecond)
	if manager.GetOrCreate("abc") != room {
		t.Fatal("expected rejoined room to survive canceled cleanup")
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
