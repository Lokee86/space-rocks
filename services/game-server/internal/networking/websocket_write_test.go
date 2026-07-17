package networking

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

func newActiveRoomForWriterTest(t *testing.T, gameInstance *game.Game) (*rooms.Room, string) {
	t.Helper()
	room := rooms.NewRoom("room-1", rooms.RoomStateLobby, nil)
	room.AddMember(rooms.NewRoomMember("session-owner"))
	if err := room.StartSinglePlayerGame(func() *game.Game { return gameInstance }); err != nil {
		t.Fatalf("expected active room to start, got %v", err)
	}
	matchID := room.CurrentMatchID()
	if matchID == "" {
		t.Fatal("expected active room match ID")
	}
	t.Cleanup(func() {
		if room.GameInstance() != nil {
			room.GameInstance().Stop()
		}
	})
	return room, matchID
}

func newReadyGameplayWebRTCTransportForTests() (*WebRTCTransport, map[string]*fakeWebRTCDataChannel) {
	world := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	overlay := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	sessionChannel := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	event := &fakeWebRTCDataChannel{readyState: webrtc.DataChannelStateOpen}
	channels := map[string]*fakeWebRTCDataChannel{
		webRTCGameplayChannelLaneWorld: world,
		"overlay":                      overlay,
		"session":                      sessionChannel,
		"event":                        event,
	}
	transport := &WebRTCTransport{
		channels: map[string]webRTCDataChannel{
			webRTCGameplayChannelLaneWorld: world,
			"overlay":                      overlay,
			"session":                      sessionChannel,
			"event":                        event,
		},
		ready: true,
	}
	return transport, channels
}

func TestMaybeWriteDebugShapeCatalogSendsOnlyOnceForSameRoom(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	originalBuilder := buildDebugShapeCatalogResponse
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return true
	}
	buildDebugShapeCatalogResponse = func(room *rooms.Room, roomID string) ([]byte, bool) {
		return []byte(`{"type":"debug_shape_catalog","shapes":{}}`), true
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
		buildDebugShapeCatalogResponse = originalBuilder
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	room := rooms.NewRoom("room-1", rooms.RoomStateInGame, game.New())
	session := &webSocketSession{
		conn:    serverConn,
		context: SessionContext{Room: room, RoomID: "room-1", GamePlayerID: "player-1"},
		rooms:   rooms.NewRoomManager(),
	}

	if !maybeWriteDebugShapeCatalog(session, session.sessionContext(), "127.0.0.1:1234") {
		t.Fatal("expected first debug shape catalog write to succeed")
	}
	assertDebugShapeCatalogPacket(t, clientConn)

	if maybeWriteDebugShapeCatalog(session, session.sessionContext(), "127.0.0.1:1234") {
		// no-op send still returns true; verify no duplicate packet instead.
	}
	assertNoMessageWithin(t, clientConn)
}

func TestMaybeWriteDebugShapeCatalogSendsAgainForNewRoomAfterReset(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	originalBuilder := buildDebugShapeCatalogResponse
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return true
	}
	buildDebugShapeCatalogResponse = func(room *rooms.Room, roomID string) ([]byte, bool) {
		return []byte(`{"type":"debug_shape_catalog","shapes":{}}`), true
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
		buildDebugShapeCatalogResponse = originalBuilder
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	room := rooms.NewRoom("room-2", rooms.RoomStateInGame, game.New())
	session := &webSocketSession{
		conn:                        serverConn,
		context:                     SessionContext{Room: room, RoomID: room.ID, GamePlayerID: "player-1"},
		rooms:                       rooms.NewRoomManager(),
		debugShapeCatalogSentRoomID: "room-1",
	}
	session.resetDebugShapeCatalogSent()

	if !maybeWriteDebugShapeCatalog(session, session.sessionContext(), "127.0.0.1:1234") {
		t.Fatal("expected debug shape catalog write to succeed after reset")
	}
	assertDebugShapeCatalogPacket(t, clientConn)
}

func TestWriteGameplayLaneProtocolMessageWritesLanePacket(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}
	if !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected SpawnPlayerShip to succeed")
	}

	transport, channels := newReadyGameplayWebRTCTransportForTests()
	room, _ := newActiveRoomForWriterTest(t, gameInstance)
	session := &webSocketSession{
		conn:            serverConn,
		context:         SessionContext{Room: room, RoomID: room.ID, GamePlayerID: playerID},
		rooms:           rooms.NewRoomManager(),
		webrtcTransport: transport,
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to succeed")
	}

	assertPhysicalGameplayChannelWrites(t, channels, false)
	assertNoMessageWithin(t, clientConn)
}

func TestWriteGameplayLaneProtocolMessageUsesWebRTCForLanePackets(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}
	if !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected SpawnPlayerShip to succeed")
	}

	transport, channels := newReadyGameplayWebRTCTransportForTests()
	room, _ := newActiveRoomForWriterTest(t, gameInstance)
	session := &webSocketSession{
		conn:            serverConn,
		context:         SessionContext{Room: room, RoomID: "room-1", GamePlayerID: playerID},
		rooms:           rooms.NewRoomManager(),
		webrtcTransport: transport,
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to succeed")
	}

	assertPhysicalGameplayChannelWrites(t, channels, false)
	assertNoMessageWithin(t, clientConn)
}

func TestWriteGameplayLaneProtocolMessageSkipsWebSocketWithoutWebRTC(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}
	if !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected SpawnPlayerShip to succeed")
	}

	room, _ := newActiveRoomForWriterTest(t, gameInstance)
	session := &webSocketSession{
		conn:    serverConn,
		context: SessionContext{Room: room, RoomID: "room-1", GamePlayerID: playerID},
		rooms:   rooms.NewRoomManager(),
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to succeed while skipping websocket write without webrtc transport")
	}
	assertNoMessageWithin(t, clientConn)
}

func TestWriteGameplayLaneProtocolMessageSkipsWebSocketWhenWebRTCNotReady(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}
	if !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected SpawnPlayerShip to succeed")
	}

	transport, _ := newReadyGameplayWebRTCTransportForTests()
	transport.ready = false
	room, _ := newActiveRoomForWriterTest(t, gameInstance)
	session := &webSocketSession{
		conn:            serverConn,
		context:         SessionContext{Room: room, RoomID: "room-1", GamePlayerID: playerID},
		rooms:           rooms.NewRoomManager(),
		webrtcTransport: transport,
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to succeed while skipping websocket write when webrtc is not ready")
	}
	assertNoMessageWithin(t, clientConn)
}
func TestResetRealtimeStateForCurrentIdentityResetsSameReceiverAcrossMatches(t *testing.T) {
	room := rooms.NewRoom("room-identity", rooms.RoomStateLobby, nil)
	member := room.AddMember(rooms.NewRoomMember("session-identity"))
	member.SetReady(true)
	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected first match to start, got %v", err)
	}
	firstMatchID := room.CurrentMatchID()
	session := &webSocketSession{
		context:       SessionContext{Room: room, RoomID: room.ID, GamePlayerID: member.PlayerID},
		realtimeState: realtime.NewRealtimeSessionState(member.PlayerID, "old-match"),
	}
	session.realtimeState.UpdateLane(realtime.LaneWorld, realtime.Metadata{Lane: realtime.LaneWorld, Sequence: 8})
	resetRealtimeStateForContext(session, session.sessionContext(), room.CurrentMatchID())
	if session.realtimeState.MatchID != firstMatchID {
		t.Fatalf("expected current match ID %q, got %q", firstMatchID, session.realtimeState.MatchID)
	}
	if _, ok := session.realtimeState.LaneState(realtime.LaneWorld); ok {
		t.Fatal("expected changed match identity to clear prior world lane state")
	}
	if got := realtime.NextLaneSequence(realtime.RealtimeLaneState{}, false); got != 1 {
		t.Fatalf("expected new match baseline sequence 1, got %d", got)
	}
	room.GameInstance().Stop()
}

func TestWriteGameplayLaneProtocolMessageDoesNotDrainEventBatchWhenEventLaneSendFails(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}
	if !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected SpawnPlayerShip to succeed")
	}
	if !control.ApplyPlayerDefeat("debug", playerID) {
		t.Fatal("expected KillPlayer to succeed")
	}

	room, matchID := newActiveRoomForWriterTest(t, gameInstance)
	state := realtime.NewRealtimeSessionState(playerID, matchID)
	state.UpdateLane(realtime.LaneWorld, realtime.Metadata{Lane: realtime.LaneWorld, Sequence: 7, BaselineID: "world-before", SnapshotID: "world-before", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneWorld)
	state.UpdateLane(realtime.LaneOverlay, realtime.Metadata{Lane: realtime.LaneOverlay, Sequence: 7, BaselineID: "overlay-before", SnapshotID: "overlay-before", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneOverlay)
	state.UpdateLane(realtime.LaneSession, realtime.Metadata{Lane: realtime.LaneSession, Sequence: 7, BaselineID: "session-before", SnapshotID: "session-before", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneSession)

	beforeWorld, _ := state.LaneState(realtime.LaneWorld)
	beforeOverlay, _ := state.LaneState(realtime.LaneOverlay)
	beforeSession, _ := state.LaneState(realtime.LaneSession)
	beforePending := len(gameInstance.PendingPresentationEvents(playerID))

	transport, channels := newReadyGameplayWebRTCTransportForTests()
	channels["event"].sendErr = errors.New("webrtc send failed")
	session := &webSocketSession{
		conn:            serverConn,
		context:         SessionContext{Room: room, RoomID: "room-1", GamePlayerID: playerID},
		rooms:           rooms.NewRoomManager(),
		realtimeState:   state,
		webrtcTransport: transport,
	}

	if writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to fail when webrtc send fails")
	}
	assertNoMessageWithin(t, clientConn)

	afterWorld, _ := session.realtimeState.LaneState(realtime.LaneWorld)
	afterOverlay, _ := session.realtimeState.LaneState(realtime.LaneOverlay)
	afterSession, _ := session.realtimeState.LaneState(realtime.LaneSession)
	if afterWorld.Sequence <= beforeWorld.Sequence {
		t.Fatalf("expected world metadata to advance, got %#v want > %#v", afterWorld, beforeWorld)
	}
	if afterOverlay.Sequence <= beforeOverlay.Sequence {
		t.Fatalf("expected overlay metadata to advance, got %#v want > %#v", afterOverlay, beforeOverlay)
	}
	if afterSession.Sequence <= beforeSession.Sequence {
		t.Fatalf("expected session metadata to advance, got %#v want > %#v", afterSession, beforeSession)
	}
	if got := len(gameInstance.PendingPresentationEvents(playerID)); got != beforePending {
		t.Fatalf("expected pending presentation events to remain undrained, got %d want %d", got, beforePending)
	}
}

func TestWriteGameplayLaneProtocolMessageAdvancesMetadataAndDrainsEventBatchOnWebRTCSendSuccess(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}
	if !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected SpawnPlayerShip to succeed")
	}
	if !control.ApplyPlayerDefeat("debug", playerID) {
		t.Fatal("expected KillPlayer to succeed")
	}

	room, matchID := newActiveRoomForWriterTest(t, gameInstance)
	state := realtime.NewRealtimeSessionState(playerID, matchID)
	state.UpdateLane(realtime.LaneWorld, realtime.Metadata{Lane: realtime.LaneWorld, Sequence: 1, BaselineID: "world-before", SnapshotID: "world-before", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneWorld)
	state.UpdateLane(realtime.LaneOverlay, realtime.Metadata{Lane: realtime.LaneOverlay, Sequence: 1, BaselineID: "overlay-before", SnapshotID: "overlay-before", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneOverlay)
	state.UpdateLane(realtime.LaneSession, realtime.Metadata{Lane: realtime.LaneSession, Sequence: 1, BaselineID: "session-before", SnapshotID: "session-before", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneSession)

	beforeWorld, _ := state.LaneState(realtime.LaneWorld)
	beforeOverlay, _ := state.LaneState(realtime.LaneOverlay)
	beforeSession, _ := state.LaneState(realtime.LaneSession)
	beforePending := len(gameInstance.PendingPresentationEvents(playerID))
	transport, channels := newReadyGameplayWebRTCTransportForTests()
	session := &webSocketSession{
		conn:            serverConn,
		context:         SessionContext{Room: room, RoomID: "room-1", GamePlayerID: playerID},
		rooms:           rooms.NewRoomManager(),
		realtimeState:   state,
		webrtcTransport: transport,
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to succeed")
	}

	afterWorld, _ := session.realtimeState.LaneState(realtime.LaneWorld)
	afterOverlay, _ := session.realtimeState.LaneState(realtime.LaneOverlay)
	afterSession, _ := session.realtimeState.LaneState(realtime.LaneSession)
	if afterWorld.Sequence <= beforeWorld.Sequence {
		t.Fatalf("expected world sequence to advance, got %d want > %d", afterWorld.Sequence, beforeWorld.Sequence)
	}
	if afterOverlay.Sequence <= beforeOverlay.Sequence {
		t.Fatalf("expected overlay sequence to advance, got %d want > %d", afterOverlay.Sequence, beforeOverlay.Sequence)
	}
	if afterSession.Sequence <= beforeSession.Sequence {
		t.Fatalf("expected session sequence to advance, got %d want > %d", afterSession.Sequence, beforeSession.Sequence)
	}
	assertPhysicalGameplayChannelWrites(t, channels, true)
	assertNoMessageWithin(t, clientConn)
	if got := len(gameInstance.PendingPresentationEvents(playerID)); got != beforePending-1 {
		t.Fatalf("expected one pending presentation event to be drained, got %d want %d", got, beforePending-1)
	}
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneWorld, "world_full")
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneOverlay, "overlay_full")
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneSession, "session_full")
}

func TestWriteGameplayLaneProtocolMessageStoresBaselineProjectionAfterSuccessfulWrite(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}
	if !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected SpawnPlayerShip to succeed")
	}

	room, matchID := newActiveRoomForWriterTest(t, gameInstance)
	state := realtime.NewRealtimeSessionState(playerID, matchID)
	state.UpdateLane(realtime.LaneWorld, realtime.Metadata{Lane: realtime.LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneWorld)
	state.UpdateLane(realtime.LaneOverlay, realtime.Metadata{Lane: realtime.LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneOverlay)
	state.UpdateLane(realtime.LaneSession, realtime.Metadata{Lane: realtime.LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneSession)

	if projection, ok := state.BaselineProjection(realtime.LaneWorld); ok || projection != nil {
		t.Fatalf("expected no stored world projection before write, got %#v, %t", projection, ok)
	}
	if projection, ok := state.BaselineProjection(realtime.LaneOverlay); ok || projection != nil {
		t.Fatalf("expected no stored overlay projection before write, got %#v, %t", projection, ok)
	}
	if projection, ok := state.BaselineProjection(realtime.LaneSession); ok || projection != nil {
		t.Fatalf("expected no stored session projection before write, got %#v, %t", projection, ok)
	}

	transport, channels := newReadyGameplayWebRTCTransportForTests()
	session := &webSocketSession{
		conn:            serverConn,
		context:         SessionContext{Room: room, RoomID: room.ID, GamePlayerID: playerID},
		rooms:           rooms.NewRoomManager(),
		realtimeState:   state,
		webrtcTransport: transport,
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to succeed")
	}

	assertPhysicalGameplayChannelWrites(t, channels, false)
	assertNoMessageWithin(t, clientConn)
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneWorld, "world_full")
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneOverlay, "overlay_full")
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneSession, "session_full")
}

func assertLanePacketText(t *testing.T, raw string) {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("expected valid json packet: %v", err)
	}
	packetType, _ := payload["type"].(string)
	if packetType == "" {
		packetType, _ = payload["t"].(string)
	}
	switch packetType {
	case "world_full", "world_delta", "overlay_full", "overlay_delta", "session_full", "session_delta", "event_batch", "resync_request", "resync_required", "wf", "wd", "of", "od", "sf", "sd":
	default:
		t.Fatalf("expected lane packet type, got %v", packetType)
	}
}

func assertPhysicalGameplayChannelWrites(t *testing.T, channels map[string]*fakeWebRTCDataChannel, wantEvent bool) {
	t.Helper()

	total := 0
	for _, lane := range []string{webRTCGameplayChannelLaneWorld, "overlay", "session"} {
		texts := channels[lane].sentTexts
		if len(texts) == 0 {
			continue
		}
		total += len(texts)
		assertPhysicalLaneChannelWrites(t, lane, texts)
	}
	if wantEvent {
		assertPhysicalLaneChannelWrites(t, "event", channels["event"].sentTexts)
		total += len(channels["event"].sentTexts)
	} else if len(channels["event"].sentTexts) > 0 {
		total += len(channels["event"].sentTexts)
		assertPhysicalLaneChannelWrites(t, "event", channels["event"].sentTexts)
	}
	if total == 0 {
		t.Fatal("expected gameplay packets on physical channels")
	}
}

func assertPhysicalLaneChannelWrites(t *testing.T, lane string, texts []string) {
	t.Helper()
	if len(texts) == 0 {
		t.Fatalf("expected packets on %s channel", lane)
	}
	for _, raw := range texts {
		assertPhysicalLanePacketText(t, lane, raw)
	}
}

func assertPhysicalLanePacketText(t *testing.T, lane string, raw string) {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("expected valid json packet: %v", err)
	}
	packetType, _ := payload["type"].(string)
	if packetType == "" {
		packetType, _ = payload["t"].(string)
	}
	switch lane {
	case webRTCGameplayChannelLaneWorld:
		if !strings.HasPrefix(packetType, "world_") && packetType != "wf" && packetType != "wd" {
			t.Fatalf("expected world packet on world channel, got %v", packetType)
		}
	case "overlay":
		if !strings.HasPrefix(packetType, "overlay_") && packetType != "of" && packetType != "od" {
			t.Fatalf("expected overlay packet on overlay channel, got %v", packetType)
		}
	case "session":
		if !strings.HasPrefix(packetType, "session_") && packetType != "sf" && packetType != "sd" {
			t.Fatalf("expected session packet on session channel, got %v", packetType)
		}
	case "event":
		if packetType != "event_batch" && packetType != "eb" {
			t.Fatalf("expected event batch on event channel, got %v", packetType)
		}
	default:
		t.Fatalf("unexpected lane %q", lane)
	}
}

func assertStoredBaselineProjectionType(t *testing.T, state realtime.RealtimeSessionState, lane realtime.Lane, wantType string) {
	t.Helper()
	projection, ok := state.BaselineProjection(lane)
	if !ok {
		t.Fatalf("expected stored projection for lane=%q", lane)
	}

	switch lane {
	case realtime.LaneWorld:
		packet, ok := projection.(realtime.WorldWireFullPacket)
		if !ok {
			t.Fatalf("expected world projection to be realtime.WorldWireFullPacket, got %#v", projection)
		}
		if packet.Type != wantType {
			t.Fatalf("expected world projection type=%q, got %q", wantType, packet.Type)
		}
	case realtime.LaneOverlay:
		packet, ok := projection.(realtime.OverlayWireFullPacket)
		if !ok {
			t.Fatalf("expected overlay projection to be realtime.OverlayWireFullPacket, got %#v", projection)
		}
		if packet.Type != wantType {
			t.Fatalf("expected overlay projection type=%q, got %q", wantType, packet.Type)
		}
	case realtime.LaneSession:
		packet, ok := projection.(realtime.SessionWireFullPacket)
		if !ok {
			t.Fatalf("expected session projection to be realtime.SessionWireFullPacket, got %#v", projection)
		}
		if packet.Type != wantType {
			t.Fatalf("expected session projection type=%q, got %q", wantType, packet.Type)
		}
	default:
		t.Fatalf("unexpected lane for stored projection assertion: %q", lane)
	}
}

func newWebSocketTestConn(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	serverConnCh := make(chan *websocket.Conn, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		serverConnCh <- conn
	}))
	t.Cleanup(server.Close)

	clientConn, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	serverConn := <-serverConnCh
	return serverConn, clientConn
}

func TestWriteGameplayLaneProtocolMessageRejectsStaleRoomContextBeforeSending(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	t.Cleanup(func() { canSendDebugShapeCatalog = originalCanSend })

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) || !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{VisibleWorldWidth: 1280, VisibleWorldHeight: 720}) {
		t.Fatal("expected production-valid player setup")
	}
	room, matchID := newActiveRoomForWriterTest(t, gameInstance)
	state := realtime.NewRealtimeSessionState(playerID, matchID)
	state.UpdateLane(realtime.LaneWorld, realtime.Metadata{Lane: realtime.LaneWorld, Sequence: 9, BaselineID: "before", SnapshotID: "before", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneWorld)
	state.StoreBaselineProjection(realtime.LaneWorld, "projection-before")
	beforeLane, _ := state.LaneState(realtime.LaneWorld)

	transport, channels := newReadyGameplayWebRTCTransportForTests()
	session := &webSocketSession{conn: serverConn, context: SessionContext{Room: room, RoomID: room.ID, GamePlayerID: playerID}, rooms: rooms.NewRoomManager(), realtimeState: state, webrtcTransport: transport}
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		if err := room.MarkGameOver(); err != nil {
			t.Fatalf("advance room to game over: %v", err)
		}
		return false
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected stale operation to be rejected without closing writer")
	}
	assertNoMessageWithin(t, clientConn)
	for lane, channel := range channels {
		if len(channel.sentTexts) != 0 {
			t.Fatalf("stale write sent %d packets on %s", len(channel.sentTexts), lane)
		}
	}
	afterLane, _ := session.realtimeState.LaneState(realtime.LaneWorld)
	if afterLane != beforeLane {
		t.Fatalf("stale write advanced realtime lane state: before=%#v after=%#v", beforeLane, afterLane)
	}
	if projection, ok := session.realtimeState.BaselineProjection(realtime.LaneWorld); !ok || projection != "projection-before" {
		t.Fatalf("stale write changed baseline projection: %#v, %t", projection, ok)
	}
	if session.realtimeState.MatchID != matchID {
		t.Fatalf("stale write replaced realtime state identity: %q", session.realtimeState.MatchID)
	}
}

func assertDebugShapeCatalogPacket(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected debug shape catalog packet: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(msg, &payload); err != nil {
		t.Fatalf("expected valid json packet: %v", err)
	}
	if got := payload["type"]; got != "debug_shape_catalog" {
		t.Fatalf("expected debug shape catalog packet, got %v", got)
	}
}

func assertLanePacket(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected lane packet: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(msg, &payload); err != nil {
		t.Fatalf("expected valid json packet: %v", err)
	}

	packetType, _ := payload["type"].(string)
	if packetType == "" {
		packetType, _ = payload["t"].(string)
	}
	switch packetType {
	case "world_full", "world_delta", "overlay_full", "overlay_delta", "session_full", "session_delta", "event_batch", "resync_request", "resync_required", "wf", "wd", "of", "od", "sf", "sd":
	default:
		t.Fatalf("expected lane packet type, got %v", packetType)
	}
}

func assertNoMessageWithin(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	defer conn.SetReadDeadline(time.Time{})
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected no duplicate debug shape catalog packet")
	}
}
