package tooling

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
	networkingtooling "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	"github.com/google/uuid"
)

var _ networkingtooling.MeasurementController = (*Controller)(nil)

type ServerReportWriter interface {
	Write(measurement.ServerReport) (string, error)
}

type Dependencies struct {
	Rooms        *rooms.RoomManager
	BuildVersion string
	RunOptions   []measurement.RunOption
	ReportWriter ServerReportWriter
}

type Controller struct {
	mu           sync.Mutex
	rooms        *rooms.RoomManager
	buildVersion string
	runs         map[string]*ownedRun
	runOptions   []measurement.RunOption
	reportWriter ServerReportWriter
}

type ownedRun struct {
	mu         sync.Mutex
	run        *measurement.Run
	detach     func()
	detachOnce sync.Once
}

func NewController(deps Dependencies) *Controller {
	return &Controller{
		rooms:        deps.Rooms,
		buildVersion: deps.BuildVersion,
		runs:         make(map[string]*ownedRun),
		runOptions:   append([]measurement.RunOption(nil), deps.RunOptions...),
		reportWriter: deps.ReportWriter,
	}
}

func NewMeasurementController(deps Dependencies) *Controller {
	return NewController(deps)
}

func (controller *Controller) Start(context networkingtooling.Context, request protocol.MeasurementStart) (protocol.MeasurementStarted, error) {
	if context.SessionID == "" {
		return protocol.MeasurementStarted{}, fmt.Errorf("measurement session id is required")
	}
	room, gameInstance, err := controller.resolveRoomGame(context)
	if err != nil {
		return protocol.MeasurementStarted{}, err
	}

	controller.mu.Lock()
	defer controller.mu.Unlock()
	if existing := controller.runs[context.SessionID]; existing != nil && existing.isActive() {
		return protocol.MeasurementStarted{}, fmt.Errorf("a measurement run is already active for this connection")
	}

	runID := uuid.NewString()
	runContext := measurement.RunContext{
		RunID:          runID,
		SessionID:      context.SessionID,
		RoomID:         room.ID,
		MatchID:        room.CurrentMatchID(),
		SimulationSeed: gameInstance.SimulationSeed(),
		BuildVersion:   controller.buildVersion,
	}
	options := []measurement.RunOption{
		measurement.WithRoomCountProvider(controller.rooms.RoomCount),
	}
	options = append(options, controller.runOptions...)
	run := measurement.NewRun(runContext, options...)
	startedAt := run.Snapshot().StartedAt
	state := &ownedRun{run: run}
	state.detach = gameInstance.AttachRuntimeMeasurement(run)
	controller.runs[context.SessionID] = state

	return protocol.MeasurementStarted{
		Type:          protocol.PacketTypeMeasurementStarted,
		RequestID:     request.RequestID,
		RunID:         runID,
		TimestampMsec: int(startedAt.UnixMilli()),
		ScenarioLabel: request.ScenarioLabel,
	}, nil
}

func (controller *Controller) Stop(context networkingtooling.Context, request protocol.MeasurementStop) (protocol.MeasurementStopped, error) {
	state, err := controller.runFor(context, request.RunID)
	if err != nil {
		return protocol.MeasurementStopped{}, err
	}
	report := state.finalize(measurement.StopReasonComplete)
	exportResult, _ := controller.persistReport(report)
	payload, err := reportMap(report)
	if err != nil {
		return protocol.MeasurementStopped{}, err
	}
	payload["server_export"] = exportResult
	controller.removeRun(context.SessionID, state)
	return protocol.MeasurementStopped{
		Type:          protocol.PacketTypeMeasurementStopped,
		RequestID:     request.RequestID,
		RunID:         report.Context.RunID,
		TimestampMsec: int(report.EndedAt.UnixMilli()),
		Partial:       !report.Complete,
		Complete:      report.Complete,
		Report:        payload,
	}, nil
}

func (controller *Controller) Reset(context networkingtooling.Context, request protocol.MeasurementReset) error {
	state, err := controller.runFor(context, request.RunID)
	if err != nil {
		return err
	}
	if !state.reset() {
		return fmt.Errorf("measurement run is already finalized")
	}
	return nil
}

func (controller *Controller) Snapshot(context networkingtooling.Context, request protocol.MeasurementSnapshotRequest) (protocol.MeasurementSnapshot, error) {
	state, err := controller.runFor(context, request.RunID)
	if err != nil {
		return protocol.MeasurementSnapshot{}, err
	}
	report := state.snapshot()
	payload, err := reportMap(report)
	if err != nil {
		return protocol.MeasurementSnapshot{}, err
	}
	return protocol.MeasurementSnapshot{
		Type:          protocol.PacketTypeMeasurementSnapshot,
		RequestID:     request.RequestID,
		RunID:         report.Context.RunID,
		TimestampMsec: int(report.EndedAt.UnixMilli()),
		Sequence:      int(report.Ticks.Count),
		Partial:       !report.Complete,
		Complete:      report.Complete,
		Metrics:       measurementMetrics(report),
		Report:        payload,
	}, nil
}

func (controller *Controller) FinalizePartial(context networkingtooling.Context, reason string) error {
	if context.SessionID == "" {
		return nil
	}
	controller.mu.Lock()
	state := controller.runs[context.SessionID]
	controller.mu.Unlock()
	if state == nil {
		return nil
	}
	report := state.finalize(partialStopReason(reason))
	_, writeErr := controller.persistReport(report)
	controller.removeRun(context.SessionID, state)
	return writeErr
}

func (controller *Controller) ObservePacketWrite(context networkingtooling.Context, lane string, packetFamily string, encodedBytes int) {
	if context.SessionID == "" {
		return
	}
	controller.mu.Lock()
	state := controller.runs[context.SessionID]
	controller.mu.Unlock()
	if state == nil {
		return
	}
	state.observePacketWrite(lane, packetFamily, encodedBytes)
}

func (controller *Controller) resolveRoomGame(context networkingtooling.Context) (*rooms.Room, *game.Game, error) {
	if controller.rooms == nil {
		return nil, nil, fmt.Errorf("room manager is not configured")
	}
	room, ok := controller.rooms.Find(context.RoomID)
	if !ok {
		return nil, nil, fmt.Errorf("room %q was not found", context.RoomID)
	}
	gameInstance := room.GameInstance()
	if gameInstance == nil {
		return nil, nil, fmt.Errorf("room %q has no active game", room.ID)
	}
	return room, gameInstance, nil
}

func (controller *Controller) runFor(context networkingtooling.Context, runID string) (*ownedRun, error) {
	if context.SessionID == "" {
		return nil, fmt.Errorf("measurement session id is required")
	}
	controller.mu.Lock()
	state := controller.runs[context.SessionID]
	controller.mu.Unlock()
	if state == nil {
		return nil, fmt.Errorf("measurement run is not active for this connection")
	}
	report := state.snapshot()
	if runID == "" || runID != report.Context.RunID {
		return nil, fmt.Errorf("run_id is not active for this connection")
	}
	return state, nil
}

func (controller *Controller) removeRun(sessionID string, expected *ownedRun) {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.runs[sessionID] == expected {
		delete(controller.runs, sessionID)
	}
}

func (state *ownedRun) isActive() bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	return state.run.IsActive()
}

func (state *ownedRun) snapshot() measurement.ServerReport {
	state.mu.Lock()
	defer state.mu.Unlock()
	return state.run.Snapshot()
}

func (state *ownedRun) reset() bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	if !state.run.IsActive() {
		return false
	}
	state.run.Reset()
	return true
}

func (state *ownedRun) observePacketWrite(lane string, packetFamily string, encodedBytes int) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.run.ObservePacketWrite(lane, packetFamily, encodedBytes)
}

func (state *ownedRun) finalize(reason measurement.StopReason) measurement.ServerReport {
	state.mu.Lock()
	defer state.mu.Unlock()
	report := state.run.Finalize(reason)
	state.detachOnce.Do(state.detach)
	return report
}

func (controller *Controller) persistReport(report measurement.ServerReport) (map[string]any, error) {
	result := map[string]any{
		"success": false,
		"path":    "",
		"error":   "",
	}
	if controller.reportWriter == nil {
		result["error"] = "measurement report writer is not configured"
		return result, nil
	}
	path, err := controller.reportWriter.Write(report)
	if err != nil {
		result["error"] = err.Error()
		return result, err
	}
	result["success"] = true
	result["path"] = path
	return result, nil
}

func partialStopReason(reason string) measurement.StopReason {
	switch reason {
	case "connection_closed", "disconnected":
		return measurement.StopReasonDisconnected
	case "shutdown":
		return measurement.StopReasonShutdown
	default:
		return measurement.StopReasonPartial
	}
}

func reportMap(report measurement.ServerReport) (map[string]any, error) {
	encoded, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("encode measurement report: %w", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, fmt.Errorf("decode measurement report: %w", err)
	}
	return payload, nil
}

func measurementMetrics(report measurement.ServerReport) map[string]any {
	return map[string]any{
		"duration": report.Duration,
		"ticks":    report.Ticks,
		"samples":  report.Samples,
		"packets":  report.Packets,
	}
}
