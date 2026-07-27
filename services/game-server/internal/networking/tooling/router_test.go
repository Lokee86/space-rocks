package tooling

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

type testSender struct {
	packets []map[string]any
}

func (sender *testSender) SendToolingJSON(packet map[string]any) error {
	sender.packets = append(sender.packets, packet)
	return nil
}

type testMeasurementController struct {
	partialCalls int
}

func (controller *testMeasurementController) Start(Context, protocol.MeasurementStart) (protocol.MeasurementStarted, error) {
	return protocol.MeasurementStarted{RunID: "run-a"}, nil
}

func (controller *testMeasurementController) Stop(Context, protocol.MeasurementStop) (protocol.MeasurementStopped, error) {
	return protocol.MeasurementStopped{RunID: "run-a", Complete: true}, nil
}

func (controller *testMeasurementController) Reset(Context, protocol.MeasurementReset) error {
	return nil
}

func (controller *testMeasurementController) Snapshot(Context, protocol.MeasurementSnapshotRequest) (protocol.MeasurementSnapshot, error) {
	return protocol.MeasurementSnapshot{RunID: "run-a"}, nil
}

func (controller *testMeasurementController) FinalizePartial(Context, string) error {
	controller.partialCalls++
	return nil
}

type testTelemetryProvider struct{}

func (testTelemetryProvider) TelemetrySnapshot(Context) (protocol.TelemetrySnapshot, error) {
	return protocol.TelemetrySnapshot{Metrics: map[string]any{"fps": 60}}, nil
}

type testCommandController struct {
	calls    int
	playerID string
	command  devtools.DebugCommand
	applied  bool
}

func (controller *testCommandController) HandleCommand(playerID string, command devtools.DebugCommand) bool {
	controller.calls++
	controller.playerID = playerID
	controller.command = command
	return controller.applied
}

func TestRouterSendsTelemetryPongThroughToolingSender(t *testing.T) {
	router := NewRouter(nil, nil)
	sender := &testSender{}

	if !router.Handle(Context{SessionID: "session-a"}, sender, map[string]any{
		"type":             protocol.PacketTypeTelemetryPing,
		"request_id":       "request-1",
		"sequence":         4,
		"client_sent_msec": 100,
	}) {
		t.Fatal("expected tooling packet to be handled")
	}
	if len(sender.packets) != 1 || sender.packets[0]["type"] != protocol.PacketTypeTelemetryPong {
		t.Fatalf("unexpected pong packets: %#v", sender.packets)
	}
	if sender.packets[0]["sequence"] != float64(4) {
		t.Fatalf("unexpected pong sequence: %#v", sender.packets[0]["sequence"])
	}
}

func TestRouterRejectsUnknownAndMalformedPackets(t *testing.T) {
	router := NewRouter(nil, nil)
	sender := &testSender{}

	router.Handle(Context{}, sender, map[string]any{"type": "unknown_tooling_packet", "request_id": "unknown"})
	router.Handle(Context{}, sender, map[string]any{
		"type":       protocol.PacketTypeTelemetryPing,
		"request_id": "malformed",
		"sequence":   "not-an-int",
	})

	if len(sender.packets) != 2 {
		t.Fatalf("expected two tooling errors, got %#v", sender.packets)
	}
	for _, packet := range sender.packets {
		if packet["type"] != protocol.PacketTypeToolingError {
			t.Fatalf("expected tooling_error, got %#v", packet)
		}
	}
}

func TestRouterPreflightRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name      string
		context   Context
		packet    map[string]any
		errorCode string
		requestID string
	}{
		{
			name: "missing request id",
			packet: map[string]any{
				"type": protocol.PacketTypeTelemetryPing,
			},
			errorCode: "request_id_required",
		},
		{
			name: "missing room attachment",
			packet: map[string]any{
				"type":       protocol.PacketTypeMeasurementStart,
				"request_id": "request-room",
			},
			errorCode: "room_required",
			requestID: "request-room",
		},
		{
			name: "server packet submitted by client",
			packet: map[string]any{
				"type":       protocol.PacketTypeTelemetryPong,
				"request_id": "request-server",
			},
			errorCode: "server_packet_not_allowed",
			requestID: "request-server",
		},
		{
			name: "missing capability",
			context: Context{
				RoomID:       "room-a",
				Capabilities: CapabilitySet{},
			},
			packet: map[string]any{
				"type":       devtools.PacketTypeDebugKillPlayer,
				"request_id": "request-capability",
			},
			errorCode: "capability_required",
			requestID: "request-capability",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := NewRouter(nil, nil)
			sender := &testSender{}

			router.Handle(test.context, sender, test.packet)

			if len(sender.packets) != 1 {
				t.Fatalf("expected one tooling error, got %#v", sender.packets)
			}
			response := sender.packets[0]
			if response["type"] != protocol.PacketTypeToolingError {
				t.Fatalf("expected tooling_error, got %#v", response)
			}
			if response["error_code"] != test.errorCode {
				t.Fatalf("error_code = %v, want %s", response["error_code"], test.errorCode)
			}
			if response["request_id"] != test.requestID {
				t.Fatalf("request_id = %v, want %s", response["request_id"], test.requestID)
			}
		})
	}
}

func TestRouterUnauthorizedCommandDoesNotCallController(t *testing.T) {
	controller := &testCommandController{applied: true}
	router := NewRouter(nil, nil)
	sender := &testSender{}

	router.Handle(Context{
		RoomID:            "room-a",
		Capabilities:      CapabilitySet{},
		CommandController: controller,
	}, sender, map[string]any{
		"type":             devtools.PacketTypeDebugKillPlayer,
		"request_id":       "unauthorized-command",
		"target_player_id": "Player-2",
	})

	if controller.calls != 0 {
		t.Fatalf("controller calls = %d, want 0", controller.calls)
	}
	assertToolingError(t, sender, "capability_required", "unauthorized-command")
}

func TestRouterAuthorizedCommandSendsCorrelatedResult(t *testing.T) {
	controller := &testCommandController{applied: true}
	router := NewRouter(nil, nil)
	sender := &testSender{}

	router.Handle(Context{
		RoomID:            "room-a",
		Capabilities:      NewTemporaryCapabilitySet(),
		CommandController: controller,
	}, sender, map[string]any{
		"type":             devtools.PacketTypeDebugKillPlayer,
		"request_id":       "authorized-command",
		"target_player_id": "Player-2",
	})

	if controller.calls != 1 {
		t.Fatalf("controller calls = %d, want 1", controller.calls)
	}
	if controller.playerID != "" {
		t.Fatalf("controller player ID = %q, want empty value passed through", controller.playerID)
	}
	if controller.command.Type != devtools.PacketTypeDebugKillPlayer || controller.command.RequestID != "authorized-command" || controller.command.TargetPlayerID != "Player-2" {
		t.Fatalf("unexpected decoded command: %#v", controller.command)
	}
	if len(sender.packets) != 1 {
		t.Fatalf("expected one result, got %#v", sender.packets)
	}
	result := sender.packets[0]
	if result["type"] != protocol.PacketTypeToolingCommandResult || result["request_id"] != "authorized-command" || result["command_type"] != devtools.PacketTypeDebugKillPlayer || result["applied"] != true {
		t.Fatalf("unexpected command result: %#v", result)
	}
}

func TestRouterAppliedCommandImmediatelyPushesAuthoritativeDebugStatus(t *testing.T) {
	if !devtools.Enabled() {
		t.Skip("debug readouts are disabled by this build")
	}
	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	const playerID = "player-a"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected player session")
	}
	room := rooms.NewRoom("room-a", rooms.RoomStateInGame, gameInstance)
	router := NewRouter(nil, nil)
	sender := &testSender{}

	router.Handle(Context{
		SessionID:         "session-a",
		RoomID:            room.ID,
		GamePlayerID:      playerID,
		Room:              room,
		Capabilities:      NewTemporaryCapabilitySet(),
		CommandController: devtools.NewController(devtools.Dependencies{Target: control}),
	}, sender, map[string]any{
		"type":          devtools.PacketTypeToggleDebugFreezeWorld,
		"request_id":    "freeze-asteroids",
		"freeze_target": "asteroids",
	})

	if len(sender.packets) != 2 {
		t.Fatalf("expected command result and status push, got %#v", sender.packets)
	}
	if sender.packets[0]["type"] != protocol.PacketTypeToolingCommandResult {
		t.Fatalf("expected command result first, got %#v", sender.packets[0])
	}
	status := sender.packets[1]
	if status["type"] != devtools.PacketTypeDebugStatus || status["request_id"] != "freeze-asteroids" {
		t.Fatalf("unexpected status packet: %#v", status)
	}
	debugStatus, ok := status["debug_status"].(map[string]any)
	if !ok || debugStatus["asteroids_frozen"] != true {
		t.Fatalf("expected authoritative asteroid freeze status, got %#v", status["debug_status"])
	}
}

func TestRouterRejectsUnavailableOrUnappliedCommand(t *testing.T) {
	tests := []struct {
		name       string
		controller CommandController
		errorCode  string
	}{
		{name: "controller unavailable", errorCode: "command_controller_unavailable"},
		{name: "command not applied", controller: &testCommandController{}, errorCode: "command_not_applied"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := NewRouter(nil, nil)
			sender := &testSender{}
			router.Handle(Context{
				RoomID:            "room-a",
				Capabilities:      NewTemporaryCapabilitySet(),
				CommandController: test.controller,
			}, sender, map[string]any{
				"type":       devtools.PacketTypeDebugKillPlayer,
				"request_id": "rejected-command",
			})
			assertToolingError(t, sender, test.errorCode, "rejected-command")
		})
	}
}

func TestRouterRunIDsAreConnectionScopedAndCloseFinalizesOnce(t *testing.T) {
	controller := &testMeasurementController{}
	routerA := NewRouter(controller, nil)
	routerB := NewRouter(controller, nil)
	senderA := &testSender{}
	senderB := &testSender{}

	routerA.Handle(Context{SessionID: "session-a", RoomID: "room-a"}, senderA, map[string]any{
		"type":       protocol.PacketTypeMeasurementStart,
		"request_id": "start-a",
	})
	routerB.Handle(Context{SessionID: "session-b", RoomID: "room-a"}, senderB, map[string]any{
		"type":       protocol.PacketTypeMeasurementStop,
		"request_id": "stop-b",
		"run_id":     "run-a",
	})
	if senderB.packets[0]["error_code"] != "measurement_run_not_owned" {
		t.Fatalf("unexpected cross-connection response: %#v", senderB.packets[0])
	}

	routerA.Close(Context{SessionID: "session-a"})
	routerA.Close(Context{SessionID: "session-a"})
	if controller.partialCalls != 1 {
		t.Fatalf("partial finalization calls = %d, want 1", controller.partialCalls)
	}
}

func TestRouterTelemetrySubscriptionIsPerConnection(t *testing.T) {
	provider := testTelemetryProvider{}
	routerA := NewRouter(nil, provider)
	routerB := NewRouter(nil, provider)
	senderA := &testSender{}
	senderB := &testSender{}

	routerA.Handle(Context{SessionID: "session-a", RoomID: "room-a"}, senderA, map[string]any{
		"type":       protocol.PacketTypeTelemetrySubscribe,
		"request_id": "subscribe-a",
	})
	routerA.Tick(Context{SessionID: "session-a", RoomID: "room-a"}, senderA)
	routerB.Tick(Context{SessionID: "session-b", RoomID: "room-a"}, senderB)

	if len(senderA.packets) != 1 || senderA.packets[0]["type"] != protocol.PacketTypeTelemetrySnapshot {
		t.Fatalf("expected one subscribed snapshot, got %#v", senderA.packets)
	}
	if len(senderB.packets) != 0 {
		t.Fatalf("unexpected snapshot for unsubscribed connection: %#v", senderB.packets)
	}
}

func TestRouterTelemetrySubscriptionClearsWhenRoomAttachmentChanges(t *testing.T) {
	provider := testTelemetryProvider{}
	router := NewRouter(nil, provider)
	sender := &testSender{}

	router.Handle(Context{SessionID: "session-a", RoomID: "room-a"}, sender, map[string]any{
		"type":       protocol.PacketTypeTelemetrySubscribe,
		"request_id": "subscribe-a",
	})
	router.Tick(Context{SessionID: "session-a", RoomID: "room-b"}, sender)

	if len(sender.packets) != 0 {
		t.Fatalf("unexpected snapshot after room attachment changed: %#v", sender.packets)
	}
	if router.subscribed || router.telemetryRoomID != "" {
		t.Fatalf("telemetry subscription was not cleared: subscribed=%v room=%q", router.subscribed, router.telemetryRoomID)
	}
}

func TestRouterDebugShapeCatalogRequestSendsCorrelatedResponse(t *testing.T) {
	if !devtools.Enabled() {
		t.Skip("debug readouts are disabled by this build")
	}
	router := NewRouter(nil, nil)
	sender := &testSender{}
	room := rooms.NewRoom("room-a", rooms.RoomStateInGame, game.New())

	router.Handle(Context{
		RoomID:       room.ID,
		Room:         room,
		Capabilities: NewTemporaryCapabilitySet(),
	}, sender, map[string]any{
		"type":       protocol.PacketTypeDebugShapeCatalogRequest,
		"request_id": "catalog-request",
	})

	if len(sender.packets) != 1 {
		t.Fatalf("expected one catalog response, got %#v", sender.packets)
	}
	response := sender.packets[0]
	if response["type"] != devtools.PacketTypeDebugShapeCatalog || response["request_id"] != "catalog-request" {
		t.Fatalf("unexpected catalog response: %#v", response)
	}
	shapes, ok := response["shapes"].(map[string]any)
	if !ok || len(shapes) == 0 {
		t.Fatalf("expected non-empty shape catalog, got %#v", response["shapes"])
	}
}

func TestRouterDebugStatusSubscriptionPushesAndUnsubscribes(t *testing.T) {
	if !devtools.Enabled() {
		t.Skip("debug readouts are disabled by this build")
	}
	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	const playerID = "player-a"
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected player session")
	}
	room := rooms.NewRoom("room-a", rooms.RoomStateInGame, gameInstance)
	context := Context{
		SessionID:    "session-a",
		RoomID:       room.ID,
		GamePlayerID: playerID,
		Room:         room,
		Capabilities: NewTemporaryCapabilitySet(),
	}
	router := NewRouter(nil, nil)
	sender := &testSender{}

	router.Handle(context, sender, map[string]any{
		"type":       protocol.PacketTypeDebugStatusSubscribe,
		"request_id": "status-subscription",
	})
	router.Tick(context, sender)
	if len(sender.packets) != 1 {
		t.Fatalf("expected immediate status push, got %#v", sender.packets)
	}
	response := sender.packets[0]
	if response["type"] != devtools.PacketTypeDebugStatus || response["request_id"] != "status-subscription" {
		t.Fatalf("unexpected status response: %#v", response)
	}

	router.Handle(context, sender, map[string]any{
		"type":       protocol.PacketTypeDebugStatusUnsubscribe,
		"request_id": "status-unsubscribe",
	})
	for range debugStatusIntervalTicks + 1 {
		router.Tick(context, sender)
	}
	if len(sender.packets) != 1 {
		t.Fatalf("unexpected status push after unsubscribe: %#v", sender.packets)
	}
}

func assertToolingError(t *testing.T, sender *testSender, errorCode string, requestID string) {
	t.Helper()
	if len(sender.packets) != 1 {
		t.Fatalf("expected one tooling error, got %#v", sender.packets)
	}
	response := sender.packets[0]
	if response["type"] != protocol.PacketTypeToolingError || response["error_code"] != errorCode || response["request_id"] != requestID {
		t.Fatalf("unexpected tooling error: %#v", response)
	}
}
