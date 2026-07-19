package tooling

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/outbound"
	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

const telemetryInterval = 250 * time.Millisecond
const debugStatusIntervalTicks = 8

type Context struct {
	SessionID         string
	RoomID            string
	GamePlayerID      string
	Room              *rooms.Room
	Capabilities      CapabilitySet
	CommandController CommandController
}

type Sender interface {
	SendToolingJSON(map[string]any) error
}

type MeasurementController interface {
	Start(Context, protocol.MeasurementStart) (protocol.MeasurementStarted, error)
	Stop(Context, protocol.MeasurementStop) (protocol.MeasurementStopped, error)
	Reset(Context, protocol.MeasurementReset) error
	Snapshot(Context, protocol.MeasurementSnapshotRequest) (protocol.MeasurementSnapshot, error)
	FinalizePartial(Context, string) error
}

type TelemetryProvider interface {
	TelemetrySnapshot(Context) (protocol.TelemetrySnapshot, error)
}

type Router struct {
	mu                    sync.Mutex
	measurement           MeasurementController
	telemetry             TelemetryProvider
	subscribed            bool
	activeRunID           string
	lastTelemetry         time.Time
	debugStatusSubscribed bool
	debugStatusRequestID  string
	debugStatusTicks      int
	closed                bool
}

func NewRouter(measurement MeasurementController, telemetry TelemetryProvider) *Router {
	return &Router{measurement: measurement, telemetry: telemetry}
}

func (router *Router) Handle(context Context, sender Sender, packet map[string]any) bool {
	packetType, policy, ok := router.preflight(context, sender, packet)
	if !ok {
		return true
	}
	if policy.Capability == CapabilityToolingControl && policy.Interaction == InteractionCommand {
		return router.handleCommand(context, sender, packet, packetType)
	}

	switch packetType {
	case protocol.PacketTypeDebugStatusSubscribe:
		var request protocol.DebugStatusSubscribe
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		router.mu.Lock()
		router.debugStatusSubscribed = true
		router.debugStatusRequestID = request.RequestID
		router.debugStatusTicks = 0
		router.mu.Unlock()
		return true
	case protocol.PacketTypeDebugStatusUnsubscribe:
		var request protocol.DebugStatusUnsubscribe
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		router.mu.Lock()
		router.debugStatusSubscribed = false
		router.debugStatusRequestID = ""
		router.debugStatusTicks = 0
		router.mu.Unlock()
		return true
	case protocol.PacketTypeDebugShapeCatalogRequest:
		var request protocol.DebugShapeCatalogRequest
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		response, available := outbound.BuildDebugShapeCatalogPacket(context.Room, context.RoomID, request.RequestID)
		if !available {
			router.sendError(sender, request.RequestID, "", "debug_shape_catalog_unavailable", "debug shape catalog is not available for this room")
			return true
		}
		router.send(sender, response)
		return true
	case protocol.PacketTypeTelemetrySubscribe:
		var request protocol.TelemetrySubscribe
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		router.mu.Lock()
		router.subscribed = true
		router.mu.Unlock()
		return true
	case protocol.PacketTypeTelemetryUnsubscribe:
		var request protocol.TelemetryUnsubscribe
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		router.mu.Lock()
		router.subscribed = false
		router.mu.Unlock()
		return true
	case protocol.PacketTypeTelemetryPing:
		var request protocol.TelemetryPing
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		serverReceived := time.Now().UnixMilli()
		response := protocol.TelemetryPong{
			Type:               protocol.PacketTypeTelemetryPong,
			RequestID:          request.RequestID,
			Sequence:           request.Sequence,
			ClientSentMsec:     request.ClientSentMsec,
			ServerReceivedMsec: int(serverReceived),
			ServerSentMsec:     int(time.Now().UnixMilli()),
		}
		router.send(sender, response)
		return true
	case protocol.PacketTypeMeasurementStart:
		var request protocol.MeasurementStart
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		return router.start(context, sender, request)
	case protocol.PacketTypeMeasurementStop:
		var request protocol.MeasurementStop
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		return router.stop(context, sender, request)
	case protocol.PacketTypeMeasurementReset:
		var request protocol.MeasurementReset
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		return router.reset(context, sender, request)
	case protocol.PacketTypeMeasurementSnapshotRequest:
		var request protocol.MeasurementSnapshotRequest
		if !router.decode(packet, &request, sender, packetType) {
			return true
		}
		return router.snapshot(context, sender, request)
	default:
		router.sendError(sender, stringValue(packet, "request_id"), stringValue(packet, "run_id"), "unknown_packet", fmt.Sprintf("unknown tooling packet type %q", packetType))
		return true
	}
}

func (router *Router) Tick(context Context, sender Sender) {
	router.mu.Lock()
	debugStatusDue := false
	debugStatusRequestID := ""
	if router.debugStatusSubscribed {
		debugStatusDue = router.debugStatusTicks == 0
		debugStatusRequestID = router.debugStatusRequestID
		router.debugStatusTicks = (router.debugStatusTicks + 1) % debugStatusIntervalTicks
	}
	telemetryDue := router.subscribed && router.telemetry != nil && time.Since(router.lastTelemetry) >= telemetryInterval
	provider := router.telemetry
	if telemetryDue {
		router.lastTelemetry = time.Now()
	}
	router.mu.Unlock()

	if debugStatusDue {
		if status, available := outbound.BuildDebugStatusPacket(context.Room, context.GamePlayerID, debugStatusRequestID); available {
			router.send(sender, status)
		}
	}
	if !telemetryDue {
		return
	}
	snapshot, err := provider.TelemetrySnapshot(context)
	if err != nil {
		return
	}
	if snapshot.Type == "" {
		snapshot.Type = protocol.PacketTypeTelemetrySnapshot
	}
	router.send(sender, snapshot)
}

func (router *Router) Close(context Context) {
	router.mu.Lock()
	if router.closed {
		router.mu.Unlock()
		return
	}
	router.closed = true
	router.subscribed = false
	router.debugStatusSubscribed = false
	router.debugStatusRequestID = ""
	router.debugStatusTicks = 0
	runID := router.activeRunID
	router.activeRunID = ""
	controller := router.measurement
	router.mu.Unlock()
	if runID != "" && controller != nil {
		_ = controller.FinalizePartial(context, "connection_closed")
	}
}

func (router *Router) start(context Context, sender Sender, request protocol.MeasurementStart) bool {
	router.mu.Lock()
	if router.activeRunID != "" {
		router.mu.Unlock()
		router.sendError(sender, request.RequestID, router.activeRunID, "measurement_already_active", "a measurement run is already active for this connection")
		return true
	}
	controller := router.measurement
	router.mu.Unlock()
	if controller == nil {
		router.sendError(sender, request.RequestID, "", "measurement_unavailable", "measurement controller is not configured")
		return true
	}
	response, err := controller.Start(context, request)
	if err != nil {
		router.sendError(sender, request.RequestID, "", "measurement_start_failed", err.Error())
		return true
	}
	if response.Type == "" {
		response.Type = protocol.PacketTypeMeasurementStarted
	}
	router.mu.Lock()
	router.activeRunID = response.RunID
	router.mu.Unlock()
	router.send(sender, response)
	return true
}

func (router *Router) stop(context Context, sender Sender, request protocol.MeasurementStop) bool {
	controller, ok := router.controllerForRun(sender, request.RequestID, request.RunID)
	if !ok {
		return true
	}
	response, err := controller.Stop(context, request)
	if err != nil {
		router.sendError(sender, request.RequestID, request.RunID, "measurement_stop_failed", err.Error())
		return true
	}
	router.mu.Lock()
	router.activeRunID = ""
	router.mu.Unlock()
	if response.Type == "" {
		response.Type = protocol.PacketTypeMeasurementStopped
	}
	router.send(sender, response)
	return true
}

func (router *Router) reset(context Context, sender Sender, request protocol.MeasurementReset) bool {
	controller, ok := router.controllerForRun(sender, request.RequestID, request.RunID)
	if !ok {
		return true
	}
	if err := controller.Reset(context, request); err != nil {
		router.sendError(sender, request.RequestID, request.RunID, "measurement_reset_failed", err.Error())
	}
	return true
}

func (router *Router) snapshot(context Context, sender Sender, request protocol.MeasurementSnapshotRequest) bool {
	controller, ok := router.controllerForRun(sender, request.RequestID, request.RunID)
	if !ok {
		return true
	}
	response, err := controller.Snapshot(context, request)
	if err != nil {
		router.sendError(sender, request.RequestID, request.RunID, "measurement_snapshot_failed", err.Error())
		return true
	}
	if response.Type == "" {
		response.Type = protocol.PacketTypeMeasurementSnapshot
	}
	router.send(sender, response)
	return true
}

func (router *Router) controllerForRun(sender Sender, requestID string, runID string) (MeasurementController, bool) {
	router.mu.Lock()
	activeRunID := router.activeRunID
	controller := router.measurement
	router.mu.Unlock()
	if controller == nil {
		router.sendError(sender, requestID, runID, "measurement_unavailable", "measurement controller is not configured")
		return nil, false
	}
	if runID == "" || runID != activeRunID {
		router.sendError(sender, requestID, runID, "measurement_run_not_owned", "run_id is not active for this connection")
		return nil, false
	}
	return controller, true
}

func (router *Router) decode(packet map[string]any, target any, sender Sender, packetType string) bool {
	encoded, err := json.Marshal(packet)
	if err == nil {
		err = json.Unmarshal(encoded, target)
	}
	if err == nil {
		return true
	}
	router.sendError(sender, stringValue(packet, "request_id"), stringValue(packet, "run_id"), "malformed_packet", fmt.Sprintf("invalid %s packet: %v", packetType, err))
	return false
}

func (router *Router) send(sender Sender, packet any) {
	encoded, err := json.Marshal(packet)
	if err != nil {
		return
	}
	var payload map[string]any
	if json.Unmarshal(encoded, &payload) != nil {
		return
	}
	_ = sender.SendToolingJSON(payload)
}

func (router *Router) sendError(sender Sender, requestID string, runID string, code string, message string) {
	router.send(sender, protocol.ToolingError{Type: protocol.PacketTypeToolingError, RequestID: requestID, RunID: runID, ErrorCode: code, Message: message})
}

func stringValue(packet map[string]any, key string) string {
	value, _ := packet[key].(string)
	return value
}
