package networkingtests

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/authclient"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/networking"
	realtimemode "github.com/Lokee86/space-rocks/server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

func newAuthenticatedRoomTestServer(t *testing.T, manager *rooms.RoomManager) *httptest.Server {
	t.Helper()

	verifier := &fakeTokenVerifier{
		result: authclient.VerifyResult{
			Valid: true,
			Identity: authclient.Identity{
				UserID:      123,
				AccountID:   "11111111-2222-3333-4444-555555555555",
				DisplayName: "Ada",
			},
		},
	}

	return httptest.NewServer(networking.WebSocketHandlerWithAuth(manager, verifier))
}

func dialAuthenticatedRoomWebSocket(t *testing.T, serverURL string) *websocket.Conn {
	t.Helper()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(serverURL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeAuthenticateRequest, Token: "submitted-token"}); err != nil {
		t.Fatalf("write authenticate request: %v", err)
	}

	var authResult struct {
		Type          string `json:"type"`
		Authenticated bool   `json:"authenticated"`
		UserID        int64  `json:"user_id"`
		DisplayName   string `json:"display_name"`
		ErrorCode     string `json:"error_code"`
	}
	readJSON(t, conn, &authResult)
	if !authResult.Authenticated {
		t.Fatal("expected authenticated websocket")
	}

	return conn
}

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
	if room.ActivePlayerCount() != 0 {
		t.Fatalf("expected room_id query websocket not to spawn active player, got %d", room.ActivePlayerCount())
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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
	if room.GameInstance() != nil {
		t.Fatal("expected lobby room not to create game simulation")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected created room member count 1, got %d", count)
	}
}

func TestCreateRoomRequestRejectsSessionAlreadyInRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

func TestStartSinglePlayerRequestCreatesInGameRoomAndBootstrapsLane(t *testing.T) {
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
	if room.GameInstance() == nil {
		t.Fatal("expected single-player room to create game simulation")
	}
	if room.IsJoinable() {
		t.Fatal("expected single-player room not to be joinable")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected single-player room member count 1, got %d", count)
	}
	if room.ActivePlayerCount() != 1 {
		t.Fatalf("expected single-player room active players 1, got %d", room.ActivePlayerCount())
	}

	readLaneBootstrapPackets(t, conn)
}

func TestStartSinglePlayerRequestRejectsSessionAlreadyInRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

func TestSetTargetPlayerRequestUpdatesCanonicalTarget(t *testing.T) {
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
	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", snapshot.RoomCode)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSetTargetPlayerRequest,
		TargetKind: "player",
		TargetID:   snapshot.LocalPlayerID,
	}); err != nil {
		t.Fatalf("write set target player request: %v", err)
	}

	waitForPlayerTarget(t, room, snapshot.LocalPlayerID, "player", snapshot.LocalPlayerID)
	targetKind, targetID, _, _ := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)
	if targetKind != "player" {
		t.Fatalf("expected target_kind %q, got %q", "player", targetKind)
	}
	if targetID != snapshot.LocalPlayerID {
		t.Fatalf("expected target_id %q, got %q", snapshot.LocalPlayerID, targetID)
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
	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", snapshot.RoomCode)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSetTargetPlayerRequest,
		TargetKind: "player",
		TargetID:   snapshot.LocalPlayerID,
	}); err != nil {
		t.Fatalf("write valid set target player request: %v", err)
	}
	waitForPlayerTarget(t, room, snapshot.LocalPlayerID, "player", snapshot.LocalPlayerID)
	validKind, validID, _, _ := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)
	if validKind != "player" {
		t.Fatalf("expected valid target_kind %q, got %q", "player", validKind)
	}
	if validID != snapshot.LocalPlayerID {
		t.Fatalf("expected valid target_id %q, got %q", snapshot.LocalPlayerID, validID)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSetTargetPlayerRequest,
		TargetKind: "player",
		TargetID:   "player-missing",
	}); err != nil {
		t.Fatalf("write invalid set target player request: %v", err)
	}
	waitForPlayerTarget(t, room, snapshot.LocalPlayerID, validKind, validID)
	invalidKind, invalidID, _, _ := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)
	if invalidKind != validKind {
		t.Fatalf("expected invalid request to keep target_kind %q, got %q", validKind, invalidKind)
	}
	if invalidID != validID {
		t.Fatalf("expected invalid request to keep target_id %q, got %q", validID, invalidID)
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
	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", snapshot.RoomCode)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSetTargetPlayerRequest,
		TargetKind: "player",
		TargetID:   snapshot.LocalPlayerID,
	}); err != nil {
		t.Fatalf("write set target player request: %v", err)
	}
	waitForPlayerTarget(t, room, snapshot.LocalPlayerID, "player", snapshot.LocalPlayerID)
	setKind, setID, _, _ := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)
	if setKind != "player" {
		t.Fatalf("expected target_kind %q, got %q", "player", setKind)
	}
	if setID != snapshot.LocalPlayerID {
		t.Fatalf("expected target_id %q, got %q", snapshot.LocalPlayerID, setID)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSetTargetPlayerRequest,
		TargetKind: "player",
		TargetID:   "",
	}); err != nil {
		t.Fatalf("write clear target player request: %v", err)
	}
	waitForPlayerTarget(t, room, snapshot.LocalPlayerID, "", "")
	clearedKind, clearedID, _, _ := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)
	if clearedKind != "" {
		t.Fatalf("expected clear target request to produce empty target_kind, got %q", clearedKind)
	}
	if clearedID != "" {
		t.Fatalf("expected clear target request to produce empty target_id, got %q", clearedID)
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
	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", snapshot.RoomCode)
	}

	_, _, x, y := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSelectTargetAtPositionRequest,
		X:          x,
		Y:          y,
		TargetKind: "player",
		TargetID:   snapshot.LocalPlayerID,
	}); err != nil {
		t.Fatalf("write select target at position request: %v", err)
	}

	targetKind, targetID, _, _ := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)
	if targetKind != "player" {
		t.Fatalf("expected target_kind %q, got %q", "player", targetKind)
	}
	if targetID != snapshot.LocalPlayerID {
		t.Fatalf("expected target_id %q, got %q", snapshot.LocalPlayerID, targetID)
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
	room, ok := manager.Find(snapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", snapshot.RoomCode)
	}

	_, _, x, y := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type:       servergame.PacketTypeSelectTargetAtPositionRequest,
		X:          x,
		Y:          y,
		TargetKind: "player",
		TargetID:   snapshot.LocalPlayerID,
	}); err != nil {
		t.Fatalf("write select target at position request: %v", err)
	}
	setKind, setID, _, _ := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)
	if setKind != "player" {
		t.Fatalf("expected target_kind %q, got %q", "player", setKind)
	}
	if setID != snapshot.LocalPlayerID {
		t.Fatalf("expected target_id %q, got %q", snapshot.LocalPlayerID, setID)
	}

	if err := conn.WriteJSON(servergame.ClientPacket{
		Type: servergame.PacketTypeClearTargetRequest,
	}); err != nil {
		t.Fatalf("write clear target request: %v", err)
	}
	clearedKind, clearedID, _, _ := playerPresentationForRoom(t, room, snapshot.LocalPlayerID)
	if clearedKind != "" {
		t.Fatalf("expected cleared target_kind to be empty, got %q", clearedKind)
	}
	if clearedID != "" {
		t.Fatalf("expected cleared target_id to be empty, got %q", clearedID)
	}
}

func TestJoinRoomRequestJoinsLobbyAndBroadcastsSnapshots(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	creator := dialAuthenticatedRoomWebSocket(t, server.URL)
	defer creator.Close()
	if err := creator.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &createdSnapshot)

	joiner := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	room, roomErr := manager.CreateStartedSinglePlayerRoom("session-owner")
	if roomErr != nil {
		t.Fatalf("create started single-player room: %v", roomErr)
	}
	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	creator := dialAuthenticatedRoomWebSocket(t, server.URL)
	defer creator.Close()
	if err := creator.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &createdSnapshot)

	joiner := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	if room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, room.State)
	}
	if room.GameInstance() == nil {
		t.Fatal("expected valid start request to create game")
	}
	if room.ActivePlayerCount() != 1 {
		t.Fatalf("expected 1 active game player, got %d", room.ActivePlayerCount())
	}
	if readySnapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected ready snapshot to remain lobby, got %q", readySnapshot.RoomState)
	}

	readLaneBootstrapPackets(t, conn)
}

func TestReturnToLobbyRequestResetsGameOverRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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
	room, ok := manager.Find(createdSnapshot.RoomCode)
	if !ok {
		t.Fatalf("expected room %q to exist", createdSnapshot.RoomCode)
	}
	if room.ActivePlayerCount() != 1 {
		t.Fatalf("expected first match active players 1, got %d", room.ActivePlayerCount())
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
	if room.GameInstance() != nil {
		t.Fatal("expected return to lobby to clear game")
	}
	if room.ActivePlayerCount() != 0 {
		t.Fatalf("expected return to lobby to clear active players, got %d", room.ActivePlayerCount())
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
	if room.ActivePlayerCount() != 1 {
		t.Fatalf("expected second match active players 1, got %d", room.ActivePlayerCount())
	}
	if room.MemberCount() != 1 {
		t.Fatalf("expected room membership to remain intact, got %d", room.MemberCount())
	}
}

func TestReturnToLobbyAllowsFreshSecondMatch(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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
	if room.ActivePlayerCount() != 1 {
		t.Fatalf("expected second match active players 1, got %d", room.ActivePlayerCount())
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	creator := dialAuthenticatedRoomWebSocket(t, server.URL)
	defer creator.Close()
	if err := creator.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}
	var createdSnapshot servergame.RoomSnapshot
	readJSON(t, creator, &createdSnapshot)

	joiner := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

	server := newAuthenticatedRoomTestServer(t, manager)
	defer server.Close()

	conn := dialAuthenticatedRoomWebSocket(t, server.URL)
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

func readLaneBootstrapPackets(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	required := []string{
		string(realtimemode.PacketFamilyWorldFull),
		string(realtimemode.PacketFamilyOverlayFull),
		string(realtimemode.PacketFamilySessionFull),
	}
	seen := map[string]bool{}
	deadline := time.Now().Add(time.Second)

	for {
		if seen[required[0]] && seen[required[1]] && seen[required[2]] {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for lane bootstrap packets; required=%v seen=%v", required, seenPacketFamilies(seen))
		}

		if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
			t.Fatalf("set websocket read deadline: %v", err)
		}

		var envelope struct {
			Type string `json:"type"`
		}
		if err := conn.ReadJSON(&envelope); err != nil {
			if isWebSocketTimeout(err) {
				continue
			}
			t.Fatalf("read lane bootstrap packet envelope: %v", err)
		}

		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			t.Fatalf("clear websocket read deadline: %v", err)
		}

		if envelope.Type == "state" {
			t.Fatal("unexpected state packet during lane bootstrap")
		}

		seen[envelope.Type] = true
	}
}

func seenPacketFamilies(seen map[string]bool) []string {
	families := make([]string, 0, len(seen))
	for family := range seen {
		families = append(families, family)
	}
	return families
}

func isWebSocketTimeout(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "i/o timeout") || strings.Contains(message, "deadline exceeded")
}

func playerPresentationForRoom(t *testing.T, room *rooms.Room, playerID string) (string, string, float64, float64) {
	t.Helper()

	instance := room.GameInstance()
	if instance == nil {
		t.Fatal("expected room game instance")
	}

	snapshot := instance.GameplayPresentationSnapshot(playerID)
	player, ok := snapshot.Players[playerID]
	if !ok {
		t.Fatalf("expected presentation snapshot to include player %q", playerID)
	}

	return player.TargetKind, player.TargetID, player.X, player.Y
}

func waitForPlayerTarget(t *testing.T, room *rooms.Room, playerID, expectedKind, expectedID string) {
	t.Helper()

	if !waitUntil(200*time.Millisecond, func() bool {
		instance := room.GameInstance()
		if instance == nil {
			return false
		}

		snapshot := instance.GameplayPresentationSnapshot(playerID)
		player, ok := snapshot.Players[playerID]
		if !ok {
			return false
		}

		return player.TargetKind == expectedKind && player.TargetID == expectedID
	}) {
		instance := room.GameInstance()
		if instance == nil {
			t.Fatalf("expected room game instance when waiting for player %q target kind %q id %q", playerID, expectedKind, expectedID)
		}

		snapshot := instance.GameplayPresentationSnapshot(playerID)
		player, ok := snapshot.Players[playerID]
		if !ok {
			t.Fatalf("expected presentation snapshot to include player %q while waiting for target kind %q id %q", playerID, expectedKind, expectedID)
		}

		t.Fatalf("expected player %q target to reach kind %q id %q, got kind %q id %q", playerID, expectedKind, expectedID, player.TargetKind, player.TargetID)
	}
}

func readJSON(t *testing.T, conn *websocket.Conn, value any) {
	t.Helper()

	const maxSkippedPackets = 10

	for skipped := 0; ; skipped++ {
		if skipped >= maxSkippedPackets {
			t.Fatalf("read websocket JSON: exceeded %d skipped async devtools packets", maxSkippedPackets)
		}

		if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
			t.Fatalf("set websocket read deadline: %v", err)
		}

		var raw map[string]any
		if err := conn.ReadJSON(&raw); err != nil {
			t.Fatalf("read websocket JSON: %v", err)
		}

		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			t.Fatalf("clear websocket read deadline: %v", err)
		}

		packetType, _ := raw["type"].(string)
		if packetType == "debug_status" || packetType == "debug_shape_catalog" {
			continue
		}

		encoded, err := json.Marshal(raw)
		if err != nil {
			t.Fatalf("marshal websocket JSON: %v", err)
		}
		if err := json.Unmarshal(encoded, value); err != nil {
			t.Fatalf("unmarshal websocket JSON: %v", err)
		}
		return
	}
}
