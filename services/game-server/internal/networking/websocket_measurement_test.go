package networking

import (
	"errors"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	toolingrouter "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

type sessionMeasurementController struct {
	writes []observedPacketWrite
}

func (controller *sessionMeasurementController) Start(toolingrouter.Context, protocol.MeasurementStart) (protocol.MeasurementStarted, error) {
	return protocol.MeasurementStarted{}, nil
}

func (controller *sessionMeasurementController) Stop(toolingrouter.Context, protocol.MeasurementStop) (protocol.MeasurementStopped, error) {
	return protocol.MeasurementStopped{}, nil
}

func (controller *sessionMeasurementController) Reset(toolingrouter.Context, protocol.MeasurementReset) error {
	return nil
}

func (controller *sessionMeasurementController) Snapshot(toolingrouter.Context, protocol.MeasurementSnapshotRequest) (protocol.MeasurementSnapshot, error) {
	return protocol.MeasurementSnapshot{}, nil
}

func (controller *sessionMeasurementController) FinalizePartial(toolingrouter.Context, string) error {
	return nil
}

func (controller *sessionMeasurementController) ObservePacketWrite(context toolingrouter.Context, lane string, packetFamily string, encodedBytes int) {
	controller.writes = append(controller.writes, observedPacketWrite{
		context:      context,
		lane:         lane,
		packetFamily: packetFamily,
		encodedBytes: encodedBytes,
	})
}

type observedPacketWrite struct {
	context      toolingrouter.Context
	lane         string
	packetFamily string
	encodedBytes int
}

type sessionTelemetryProvider struct{}

func (sessionTelemetryProvider) TelemetrySnapshot(toolingrouter.Context) (protocol.TelemetrySnapshot, error) {
	return protocol.TelemetrySnapshot{Metrics: map[string]any{"fps": 60}}, nil
}

func TestNewWebSocketSessionWithToolingInjectsOptionalDependencies(t *testing.T) {
	roomManager := rooms.NewRoomManager()
	t.Cleanup(roomManager.StopAll)
	controller := &sessionMeasurementController{}
	session := newWebSocketSessionWithTooling(nil, roomManager, nil, nil, controller, sessionTelemetryProvider{})

	if session.packetObserver == nil {
		t.Fatal("expected measurement controller packet observer to be injected")
	}
	if legacy := newWebSocketSession(nil, roomManager, nil, nil); legacy.packetObserver != nil {
		t.Fatal("expected legacy constructor to preserve nil packet observer")
	}
}

func TestWriteGameplayLaneProtocolMessageObservesOnlySuccessfulWrites(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(*rooms.Room) bool { return false }
	t.Cleanup(func() { canSendDebugShapeCatalog = originalCanSend })

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := "player-1"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) || !control.SpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{VisibleWorldWidth: 1280, VisibleWorldHeight: 720}) {
		t.Fatal("expected production-valid player setup")
	}
	room, _ := newActiveRoomForWriterTest(t, gameInstance)
	transport, _ := newReadyGameplayWebRTCTransportForTests()
	controller := &sessionMeasurementController{}
	session := &webSocketSession{
		context:         SessionContext{Room: room, RoomID: room.ID, GamePlayerID: playerID},
		rooms:           rooms.NewRoomManager(),
		webrtcTransport: transport,
		packetObserver:  controller,
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected successful gameplay lane write")
	}
	observedCount := len(controller.writes)
	if observedCount == 0 {
		t.Fatal("expected successful writes to be observed")
	}
	for _, observation := range controller.writes {
		if observation.context.SessionID != session.sessionID || observation.context.RoomID != room.ID || observation.context.GamePlayerID != playerID {
			t.Fatalf("unexpected packet observation context: %#v", observation.context)
		}
		if observation.lane == "" || observation.packetFamily == "" || observation.encodedBytes <= 0 {
			t.Fatalf("unexpected packet observation: %#v", observation)
		}
	}

	failedTransport, failedChannels := newReadyGameplayWebRTCTransportForTests()
	failedChannels[webRTCChannelLaneWorld].sendErr = errors.New("webrtc send failed")
	failedController := &sessionMeasurementController{}
	failedSession := &webSocketSession{
		context:         SessionContext{Room: room, RoomID: room.ID, GamePlayerID: playerID},
		rooms:           rooms.NewRoomManager(),
		webrtcTransport: failedTransport,
		packetObserver:  failedController,
	}
	if writeGameplayLaneProtocolMessage(failedSession, "127.0.0.1:1234") {
		t.Fatal("expected failed gameplay lane write")
	}
	if len(failedController.writes) != 0 {
		t.Fatalf("failed write produced packet observations: %#v", failedController.writes)
	}
}

func TestNewWebSocketSessionUsesTemporaryToolingCapabilities(t *testing.T) {
	roomManager := rooms.NewRoomManager()
	t.Cleanup(roomManager.StopAll)
	session := newWebSocketSession(nil, roomManager, nil, nil)

	if !session.toolingCapabilities.Has(toolingrouter.CapabilityToolingRead) {
		t.Fatal("expected temporary policy to grant tooling.read")
	}
	if !session.toolingCapabilities.Has(toolingrouter.CapabilityToolingControl) {
		t.Fatal("expected temporary policy to grant tooling.control")
	}
	if session.toolingCapabilities.Has(toolingrouter.CapabilityAdminControl) {
		t.Fatal("temporary policy must not grant admin.control")
	}
}

func TestToolingContextAttachesCurrentRoomControllerWithoutParticipation(t *testing.T) {
	room := rooms.NewRoom("room-a", rooms.RoomStateInGame, game.New())
	session := &webSocketSession{
		sessionID:           "session-a",
		context:             SessionContext{Room: room, RoomID: room.ID},
		toolingCapabilities: toolingrouter.NewTemporaryCapabilitySet(),
	}

	context := session.toolingContext()
	if context.SessionID != "session-a" || context.RoomID != room.ID {
		t.Fatalf("unexpected tooling context: %#v", context)
	}
	if context.GamePlayerID != "" {
		t.Fatalf("expected observer context without participant ID, got %q", context.GamePlayerID)
	}
	if context.CommandController == nil {
		t.Fatal("expected current room game to provide a command controller")
	}
	if !context.Capabilities.Has(toolingrouter.CapabilityToolingControl) {
		t.Fatal("expected session capabilities to be preserved")
	}
}

func TestToolingContextWithoutActiveGameHasNoCommandController(t *testing.T) {
	room := rooms.NewRoom("room-a", rooms.RoomStateLobby, nil)
	session := &webSocketSession{
		sessionID:           "session-a",
		context:             SessionContext{Room: room, RoomID: room.ID, GamePlayerID: "player-a"},
		toolingCapabilities: toolingrouter.NewTemporaryCapabilitySet(),
	}

	context := session.toolingContext()
	if context.CommandController != nil {
		t.Fatal("expected no command controller without an active game")
	}
	if context.GamePlayerID != "player-a" {
		t.Fatalf("expected participant identity to be preserved, got %q", context.GamePlayerID)
	}
}
