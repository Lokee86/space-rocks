extends GutTest

const ClientMeasurementCoordinator := preload("res://scripts/devtools/measurement/client_measurement_coordinator.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


class FakeConnection extends RefCounted:
	signal measurement_started_received(packet: Dictionary)
	signal measurement_snapshot_received(packet: Dictionary)
	signal measurement_stopped_received(packet: Dictionary)
	signal tooling_error_received(packet: Dictionary)

	var sent_packets: Array = []

	func send_tooling_packet(packet: Dictionary) -> void:
		sent_packets.append(packet.duplicate(true))


class FakeMeasurementContext extends RefCounted:
	var start_calls := 0
	var stop_calls := 0
	var reset_calls := 0
	var started_metadata: Dictionary = {}
	var stop_reason := ""

	func start(metadata: Dictionary = {}) -> void:
		start_calls += 1
		started_metadata = metadata.duplicate(true)

	func stop(reason: String = "") -> Dictionary:
		stop_calls += 1
		stop_reason = reason
		return {"status": "completed", "reason": reason, "frames": {"count": 3}}

	func reset() -> Dictionary:
		reset_calls += 1
		return {"status": "partial", "partial_reason": "reset"}


func test_start_snapshot_stop_tracks_state_and_combines_bounded_reports() -> void:
	var connection := FakeConnection.new()
	var context := FakeMeasurementContext.new()
	var coordinator := ClientMeasurementCoordinator.new()
	coordinator.configure(connection, context)

	var start_request_id := coordinator.start("soak", {"source": "unit"})
	assert_eq(start_request_id, "client-measurement-1")
	assert_eq(connection.sent_packets, [Packets.measurement_start_packet(start_request_id, "soak")])
	assert_eq(context.start_calls, 0)
	assert_false(coordinator.is_recording())

	connection.measurement_started_received.emit({
		"type": Packets.TYPE_MEASUREMENT_STARTED,
		"request_id": start_request_id,
		"run_id": "run-1",
		"scenario_label": "soak",
	})
	assert_eq(context.start_calls, 1)
	assert_eq(context.started_metadata["source"], "unit")
	assert_eq(context.started_metadata["scenario_label"], "soak")
	assert_eq(coordinator.active_run_id, "run-1")
	assert_true(coordinator.is_recording())

	var snapshot_request_id := coordinator.snapshot()
	assert_eq(connection.sent_packets[1], Packets.measurement_snapshot_request_packet(snapshot_request_id, "run-1"))
	connection.measurement_snapshot_received.emit({
		"type": Packets.TYPE_MEASUREMENT_SNAPSHOT,
		"request_id": snapshot_request_id,
		"run_id": "run-1",
		"report": {"ticks": {"count": 10}},
	})
	assert_eq(coordinator.latest_server_snapshot["ticks"]["count"], 10)

	var stop_request_id := coordinator.stop()
	assert_eq(connection.sent_packets[2], Packets.measurement_stop_packet(stop_request_id, "run-1"))
	connection.measurement_stopped_received.emit({
		"type": Packets.TYPE_MEASUREMENT_STOPPED,
		"request_id": stop_request_id,
		"run_id": "run-1",
		"complete": true,
		"report": {"ticks": {"count": 12}},
	})

	assert_eq(context.stop_calls, 1)
	assert_false(coordinator.is_recording())
	assert_eq(coordinator.active_run_id, "")
	assert_eq(coordinator.latest_server_report["ticks"]["count"], 12)
	assert_eq(coordinator.latest_combined_report.keys(), ["version", "run_id", "client", "server"])
	assert_eq(coordinator.latest_combined_report["version"], 1)
	assert_eq(coordinator.latest_combined_report["run_id"], "run-1")
	assert_eq(coordinator.latest_combined_report["client"]["frames"]["count"], 3)
	assert_eq(coordinator.latest_combined_report["server"]["ticks"]["count"], 12)


func test_mismatched_responses_and_errors_leave_active_state_intact() -> void:
	var connection := FakeConnection.new()
	var context := FakeMeasurementContext.new()
	var coordinator := ClientMeasurementCoordinator.new(connection, context)
	var start_request_id := coordinator.start("scenario")

	connection.measurement_started_received.emit({
		"request_id": "wrong-request",
		"run_id": "run-wrong",
	})
	assert_eq(context.start_calls, 0)
	assert_eq(coordinator.active_run_id, "")
	assert_true(coordinator.pending_request_ids.has("start"))

	connection.tooling_error_received.emit({
		"request_id": "wrong-request",
		"run_id": "",
		"error_code": "wrong_error",
	})
	assert_true(coordinator.pending_request_ids.has("start"))
	assert_eq(coordinator.last_tooling_error["error_code"], "unmatched_tooling_error")

	connection.measurement_started_received.emit({
		"request_id": start_request_id,
		"run_id": "run-1",
	})
	var stop_request_id := coordinator.stop()
	connection.measurement_stopped_received.emit({
		"request_id": stop_request_id,
		"run_id": "run-wrong",
		"report": {"ignored": true},
	})

	assert_true(coordinator.is_recording())
	assert_eq(coordinator.active_run_id, "run-1")
	assert_eq(context.stop_calls, 0)
	assert_eq(coordinator.latest_server_report, {})
	assert_eq(coordinator.last_tooling_error["error_code"], "mismatched_measurement_run_id")

	connection.tooling_error_received.emit({
		"request_id": stop_request_id,
		"run_id": "run-1",
		"error_code": "measurement_stop_failed",
		"message": "failed",
	})
	assert_true(coordinator.is_recording())
	assert_eq(coordinator.active_run_id, "run-1")
	assert_false(coordinator.pending_request_ids.has("stop"))
	assert_eq(coordinator.last_tooling_error["error_code"], "measurement_stop_failed")


func test_reset_restarts_local_window_without_changing_authoritative_run() -> void:
	var connection := FakeConnection.new()
	var context := FakeMeasurementContext.new()
	var coordinator := ClientMeasurementCoordinator.new(connection, context)
	var start_request_id := coordinator.start("scenario", {"seed": 7})
	connection.measurement_started_received.emit({"request_id": start_request_id, "run_id": "run-1"})

	var first_reset_request_id := coordinator.reset()
	var second_reset_request_id := coordinator.reset()

	assert_eq(connection.sent_packets[1], Packets.measurement_reset_packet(first_reset_request_id, "run-1"))
	assert_eq(connection.sent_packets[2], Packets.measurement_reset_packet(second_reset_request_id, "run-1"))
	assert_ne(first_reset_request_id, second_reset_request_id)
	assert_false(coordinator.pending_request_ids.has("reset"))
	assert_eq(context.reset_calls, 2)
	assert_eq(context.start_calls, 3)
	assert_true(coordinator.is_recording())
	assert_eq(coordinator.active_run_id, "run-1")

	connection.tooling_error_received.emit({
		"request_id": second_reset_request_id,
		"run_id": "run-1",
		"error_code": "measurement_reset_failed",
	})
	assert_true(coordinator.is_recording())
	assert_eq(coordinator.active_run_id, "run-1")
	assert_eq(coordinator.last_tooling_error["error_code"], "measurement_reset_failed")
