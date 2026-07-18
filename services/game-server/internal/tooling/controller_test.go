package tooling

import (
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
	networkingtooling "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
	protocol "github.com/Lokee86/space-rocks/services/game-server/internal/protocol/tooling"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestControllerOwnsIndependentRunsAndDetachesOnFinalization(t *testing.T) {
	manager, gameInstance, room := measurementTestRoom(t)
	clock := &measurementTestClock{now: time.Unix(100, 0)}
	controller := NewController(Dependencies{
		Rooms: manager,
		RunOptions: []measurement.RunOption{
			measurement.WithClock(clock.Now),
			measurement.WithSampleInterval(time.Hour),
		},
	})

	startedA, err := controller.Start(networkingtooling.Context{SessionID: "session-a", RoomID: room.ID}, protocol.MeasurementStart{RequestID: "start-a"})
	if err != nil {
		t.Fatalf("start session-a: %v", err)
	}
	startedB, err := controller.Start(networkingtooling.Context{SessionID: "session-b", RoomID: room.ID}, protocol.MeasurementStart{RequestID: "start-b"})
	if err != nil {
		t.Fatalf("start session-b: %v", err)
	}
	if startedA.RunID == startedB.RunID || !gameInstance.HasRuntimeMeasurements() {
		t.Fatalf("expected independent attached runs: %#v %#v", startedA, startedB)
	}

	gameInstance.Step(1.0 / 60.0)
	if _, err := controller.Stop(networkingtooling.Context{SessionID: "session-b", RoomID: room.ID}, protocol.MeasurementStop{RunID: startedA.RunID}); err == nil {
		t.Fatal("expected session-b to reject session-a run")
	}

	stoppedA, err := controller.Stop(networkingtooling.Context{SessionID: "session-a", RoomID: room.ID}, protocol.MeasurementStop{RequestID: "stop-a", RunID: startedA.RunID})
	if err != nil {
		t.Fatalf("stop session-a: %v", err)
	}
	if !stoppedA.Complete || stoppedA.Partial || !gameInstance.HasRuntimeMeasurements() {
		t.Fatalf("stopping one run should preserve the other observer: %#v", stoppedA)
	}
	if _, ok := controller.runs["session-a"]; ok {
		t.Fatal("successful stop should remove the finalized session run")
	}

	gameInstance.Step(1.0 / 60.0)
	if err := controller.FinalizePartial(networkingtooling.Context{SessionID: "session-b", RoomID: room.ID}, "connection_closed"); err != nil {
		t.Fatalf("finalize session-b: %v", err)
	}
	if err := controller.FinalizePartial(networkingtooling.Context{SessionID: "session-b", RoomID: room.ID}, "connection_closed"); err != nil {
		t.Fatalf("repeat finalization: %v", err)
	}
	if gameInstance.HasRuntimeMeasurements() {
		t.Fatal("expected all observers to detach after finalization")
	}
	if _, ok := controller.runs["session-b"]; ok {
		t.Fatal("successful partial finalization should remove the finalized session run")
	}

	startedAgain, err := controller.Start(networkingtooling.Context{SessionID: "session-b", RoomID: room.ID}, protocol.MeasurementStart{RequestID: "start-b-again"})
	if err != nil {
		t.Fatalf("start a later run for the same session: %v", err)
	}
	if startedAgain.RunID == startedB.RunID {
		t.Fatal("later run for the same session should receive a new run ID")
	}
	if _, err := controller.Stop(networkingtooling.Context{SessionID: "session-b", RoomID: room.ID}, protocol.MeasurementStop{RunID: startedAgain.RunID}); err != nil {
		t.Fatalf("stop later same-session run: %v", err)
	}
	if _, ok := controller.runs["session-b"]; ok || gameInstance.HasRuntimeMeasurements() {
		t.Fatal("later same-session run should also clean up ownership and observer")
	}
}

func TestControllerResetKeepsRunAttachedAndClearsMeasurements(t *testing.T) {
	manager, gameInstance, room := measurementTestRoom(t)
	controller := NewController(Dependencies{Rooms: manager})
	context := networkingtooling.Context{SessionID: "session-a", RoomID: room.ID}
	started, err := controller.Start(context, protocol.MeasurementStart{})
	if err != nil {
		t.Fatalf("start measurement: %v", err)
	}

	gameInstance.Step(1.0 / 60.0)
	beforeReset, err := controller.Snapshot(context, protocol.MeasurementSnapshotRequest{RunID: started.RunID})
	if err != nil || beforeReset.Report["ticks"] == nil {
		t.Fatalf("snapshot before reset: packet=%#v err=%v", beforeReset, err)
	}
	if err := controller.Reset(context, protocol.MeasurementReset{RunID: started.RunID}); err != nil {
		t.Fatalf("reset measurement: %v", err)
	}
	if !gameInstance.HasRuntimeMeasurements() {
		t.Fatal("reset should keep the observer attached")
	}

	gameInstance.Step(1.0 / 60.0)
	afterReset, err := controller.Snapshot(context, protocol.MeasurementSnapshotRequest{RunID: started.RunID})
	if err != nil {
		t.Fatalf("snapshot after reset: %v", err)
	}
	if afterReset.Sequence != 1 {
		t.Fatalf("reset did not clear prior ticks: %#v", afterReset)
	}
	if err := controller.FinalizePartial(context, "connection_closed"); err != nil {
		t.Fatalf("finalize measurement: %v", err)
	}
	if gameInstance.HasRuntimeMeasurements() {
		t.Fatal("finalization should detach the observer")
	}
}

func TestControllerObservesPacketWritesForActiveSession(t *testing.T) {
	manager, _, room := measurementTestRoom(t)
	controller := NewController(Dependencies{Rooms: manager})
	context := networkingtooling.Context{SessionID: "session-packets", RoomID: room.ID}
	_, err := controller.Start(context, protocol.MeasurementStart{})
	if err != nil {
		t.Fatalf("start measurement: %v", err)
	}

	controller.ObservePacketWrite(context, "world", "world_full", 42)
	controller.ObservePacketWrite(context, "world", "world_full", 18)
	controller.ObservePacketWrite(networkingtooling.Context{SessionID: "other-session", RoomID: room.ID}, "world", "world_full", 99)

	report := controller.runs[context.SessionID].snapshot()
	if len(report.Packets) != 1 {
		t.Fatalf("expected one packet summary, got %#v", report.Packets)
	}
	packet := report.Packets[0]
	if packet.Lane != "world" || packet.PacketFamily != "world_full" || packet.PacketCount != 2 || packet.EncodedBytesTotal != 60 || packet.MaximumEncodedBytes != 42 {
		t.Fatalf("unexpected packet summary: %#v", packet)
	}

	if err := controller.FinalizePartial(context, "connection_closed"); err != nil {
		t.Fatalf("finalize measurement: %v", err)
	}
}

func TestControllerPersistsStoppedReportAndReturnsExportResult(t *testing.T) {
	manager, gameInstance, room := measurementTestRoom(t)
	writer := &measurementReportWriterStub{path: "measurement-results/game-server/report.json"}
	controller := NewController(Dependencies{Rooms: manager, ReportWriter: writer})
	context := networkingtooling.Context{SessionID: "session-export", RoomID: room.ID}
	started, err := controller.Start(context, protocol.MeasurementStart{RequestID: "start-export"})
	if err != nil {
		t.Fatalf("start measurement: %v", err)
	}
	gameInstance.Step(1.0 / 60.0)

	stopped, err := controller.Stop(context, protocol.MeasurementStop{RequestID: "stop-export", RunID: started.RunID})
	if err != nil {
		t.Fatalf("stop measurement: %v", err)
	}
	if len(writer.reports) != 1 || writer.reports[0].Context.RunID != started.RunID {
		t.Fatalf("expected finalized report persistence, got %#v", writer.reports)
	}
	export, ok := stopped.Report["server_export"].(map[string]any)
	if !ok || export["success"] != true || export["path"] != writer.path || export["error"] != "" {
		t.Fatalf("unexpected server export result: %#v", stopped.Report["server_export"])
	}
}

func TestControllerStopReturnsReportWhenPersistenceFails(t *testing.T) {
	manager, gameInstance, room := measurementTestRoom(t)
	writer := &measurementReportWriterStub{err: errors.New("disk unavailable")}
	controller := NewController(Dependencies{Rooms: manager, ReportWriter: writer})
	context := networkingtooling.Context{SessionID: "session-export-failure", RoomID: room.ID}
	started, err := controller.Start(context, protocol.MeasurementStart{})
	if err != nil {
		t.Fatalf("start measurement: %v", err)
	}
	gameInstance.Step(1.0 / 60.0)

	stopped, err := controller.Stop(context, protocol.MeasurementStop{RunID: started.RunID})
	if err != nil {
		t.Fatalf("stop should preserve the in-memory report after persistence failure: %v", err)
	}
	if !stopped.Complete || stopped.Report["ticks"] == nil {
		t.Fatalf("expected completed in-memory report: %#v", stopped)
	}
	export, ok := stopped.Report["server_export"].(map[string]any)
	if !ok || export["success"] != false || export["path"] != "" || export["error"] != "disk unavailable" {
		t.Fatalf("unexpected failed server export result: %#v", stopped.Report["server_export"])
	}
	if _, ok := controller.runs[context.SessionID]; ok {
		t.Fatal("persistence failure should not retain finalized run ownership")
	}
}

func TestControllerPersistsPartialReport(t *testing.T) {
	manager, gameInstance, room := measurementTestRoom(t)
	writer := &measurementReportWriterStub{path: "measurement-results/game-server/partial.json"}
	controller := NewController(Dependencies{Rooms: manager, ReportWriter: writer})
	context := networkingtooling.Context{SessionID: "session-partial-export", RoomID: room.ID}
	if _, err := controller.Start(context, protocol.MeasurementStart{}); err != nil {
		t.Fatalf("start measurement: %v", err)
	}
	gameInstance.Step(1.0 / 60.0)

	if err := controller.FinalizePartial(context, "connection_closed"); err != nil {
		t.Fatalf("finalize partial measurement: %v", err)
	}
	if len(writer.reports) != 1 || writer.reports[0].Complete || writer.reports[0].StopReason != measurement.StopReasonDisconnected {
		t.Fatalf("expected persisted partial report, got %#v", writer.reports)
	}
}

type measurementReportWriterStub struct {
	path    string
	err     error
	reports []measurement.ServerReport
}

func (writer *measurementReportWriterStub) Write(report measurement.ServerReport) (string, error) {
	writer.reports = append(writer.reports, report)
	return writer.path, writer.err
}

func measurementTestRoom(t *testing.T) (*rooms.RoomManager, *game.Game, *rooms.Room) {
	t.Helper()
	manager := rooms.NewRoomManagerWithCleanupDelay(time.Hour)
	t.Cleanup(manager.StopAll)
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create measurement room: %v", err)
	}
	gameInstance := game.NewWithSeed(7)
	room.SetGameInstance(gameInstance)
	room.State = rooms.RoomStateInGame
	return manager, gameInstance, room
}

type measurementTestClock struct {
	now time.Time
}

func (clock *measurementTestClock) Now() time.Time {
	return clock.now
}
