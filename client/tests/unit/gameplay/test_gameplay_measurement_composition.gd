extends GutTest

const GameplayFlowComposer := preload("res://scripts/gameplay/runtime/gameplay_flow_composer.gd")
const ClientMeasurementCoordinator := preload("res://scripts/devtools/measurement/client_measurement_coordinator.gd")
const GameplayRespawnFlow := preload("res://scripts/gameplay/respawn/gameplay_respawn_flow.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")

var nodes_to_free: Array[Node] = []


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

	func start(_metadata: Dictionary = {}) -> void:
		start_calls += 1

	func stop(_reason: String = "") -> Dictionary:
		stop_calls += 1
		return {"status": "completed"}

	func reset() -> Dictionary:
		reset_calls += 1
		return {}


class FakeReportWriter extends RefCounted:
	var writes: Array = []

	func write(report: Dictionary, run_id: String = "") -> Dictionary:
		writes.append({"report": report.duplicate(true), "run_id": run_id})
		return {"success": true, "path": "user://measurements/fake.json", "error": ""}


class FakeInputContext extends RefCounted:
	func configure_spectate_routes(_open_menu: Callable, _cycle_target: Callable) -> void:
		pass

	func reset() -> void:
		pass


class FakeDevtoolsContext extends RefCounted:
	func get_world_telemetry_context():
		return null

	func apply_gameplay_state(_state: Dictionary) -> void:
		pass


class FakeProcessFlow extends RefCounted:
	func process(_delta: float, _required_lane_baselines_synced: bool) -> void:
		pass


class FakeWorldSync extends WorldSync:
	func visual_position_for_server_position(value):
		return value

	func server_position_for_visual_position(value):
		return value

	func target_source():
		return RefCounted.new()

	func remote_player_nodes() -> Dictionary:
		return {}


class FakeRuntimeContext extends RefCounted:
	var respawn_flow: GameplayRespawnFlow = GameplayRespawnFlow.new()

	func request_respawn(_required_lane_baselines_synced: bool) -> void:
		pass


func after_each() -> void:
	for node in nodes_to_free:
		if is_instance_valid(node):
			node.free()
	nodes_to_free.clear()


func _tracked(node: Node) -> Node:
	nodes_to_free.append(node)
	return node


func test_coordinator_exports_finalized_report_and_local_reset_preserves_result() -> void:
	var connection := FakeConnection.new()
	var context := FakeMeasurementContext.new()
	var writer := FakeReportWriter.new()
	var coordinator := ClientMeasurementCoordinator.new(connection, context, writer)

	var start_request_id := coordinator.start("soak")
	connection.measurement_started_received.emit({"request_id": start_request_id, "run_id": "run-1"})
	var stop_request_id := coordinator.stop()
	connection.measurement_stopped_received.emit({
		"request_id": stop_request_id,
		"run_id": "run-1",
		"report": {"ticks": {"count": 12}},
	})

	assert_eq(writer.writes.size(), 1)
	assert_eq(writer.writes[0]["run_id"], "run-1")
	assert_true(coordinator.get_latest_export_result()["success"])

	coordinator.reset_local_state()
	assert_eq(connection.sent_packets.size(), 2)
	assert_eq(context.reset_calls, 1)
	assert_true(coordinator.get_latest_export_result()["success"])

	var next_start_request_id := coordinator.start("next")
	connection.measurement_started_received.emit({"request_id": next_start_request_id, "run_id": "run-2"})
	assert_eq(coordinator.get_latest_export_result(), {})


func test_composer_wires_measurement_api_and_gameplay_reset_is_local() -> void:
	var connection := FakeConnection.new()
	var runtime_context := FakeRuntimeContext.new()
	var world_sync := FakeWorldSync.new()
	var composer := GameplayFlowComposer.new()
	composer.configure(
		connection,
		_tracked(Node2D.new()),
		_tracked(Player.new()),
		null,
		null,
		runtime_context,
		world_sync,
		null,
		FakeInputContext.new(),
		FakeDevtoolsContext.new(),
		FakeProcessFlow.new()
	)

	assert_not_null(composer.get_client_measurement_coordinator())
	assert_not_null(composer.client_measurement_report_writer)
	assert_eq(composer.get_client_measurement_coordinator().measurement_context, composer.get_client_measurement_context())

	var start_request_id := composer.start_measurement("composition")
	assert_eq(composer.get_measurement_state()["pending_request_ids"]["start"], start_request_id)
	connection.measurement_started_received.emit({"request_id": start_request_id, "run_id": "run-1"})
	composer.process(0.25, true)
	assert_eq(connection.sent_packets.size(), 2)
	assert_eq(connection.sent_packets[1]["type"], Packets.TYPE_MEASUREMENT_SNAPSHOT_REQUEST)
	composer.reset()

	assert_eq(connection.sent_packets.size(), 2)
	assert_eq(composer.get_measurement_state()["active_run_id"], "")
	assert_eq(composer.get_measurement_state()["recording"], false)

	var next_start_request_id := composer.start_measurement("reset-api")
	connection.measurement_started_received.emit({"request_id": next_start_request_id, "run_id": "run-2"})
	var reset_request_id := composer.reset_measurement()
	assert_eq(connection.sent_packets[3], Packets.measurement_reset_packet(reset_request_id, "run-2"))
