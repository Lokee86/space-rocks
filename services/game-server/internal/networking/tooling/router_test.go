package tooling

import (
	"testing"

	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
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
	router.Handle(Context{}, sender, map[string]any{"type": protocol.PacketTypeTelemetryPing, "sequence": "not-an-int"})

	if len(sender.packets) != 2 {
		t.Fatalf("expected two tooling errors, got %#v", sender.packets)
	}
	for _, packet := range sender.packets {
		if packet["type"] != protocol.PacketTypeToolingError {
			t.Fatalf("expected tooling_error, got %#v", packet)
		}
	}
}

func TestRouterRunIDsAreConnectionScopedAndCloseFinalizesOnce(t *testing.T) {
	controller := &testMeasurementController{}
	routerA := NewRouter(controller, nil)
	routerB := NewRouter(controller, nil)
	senderA := &testSender{}
	senderB := &testSender{}

	routerA.Handle(Context{SessionID: "session-a"}, senderA, map[string]any{
		"type":       protocol.PacketTypeMeasurementStart,
		"request_id": "start-a",
	})
	routerB.Handle(Context{SessionID: "session-b"}, senderB, map[string]any{
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

	routerA.Handle(Context{SessionID: "session-a"}, senderA, map[string]any{"type": protocol.PacketTypeTelemetrySubscribe})
	routerA.Tick(Context{SessionID: "session-a"}, senderA)
	routerB.Tick(Context{SessionID: "session-b"}, senderB)

	if len(senderA.packets) != 1 || senderA.packets[0]["type"] != protocol.PacketTypeTelemetrySnapshot {
		t.Fatalf("expected one subscribed snapshot, got %#v", senderA.packets)
	}
	if len(senderB.packets) != 0 {
		t.Fatalf("unexpected snapshot for unsubscribed connection: %#v", senderB.packets)
	}
}
