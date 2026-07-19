extends GutTest

const DevtoolsCommandContext := preload("res://scripts/devtools/context/devtools_command_context.gd")
const DevtoolsStateContext := preload("res://scripts/devtools/context/devtools_state_context.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")
const GameplayDebugFlow := preload("res://scripts/devtools/gameplay_debug_flow.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")
const PresentationEventCapture := preload("res://tests/unit/logging/presentation_event_capture.gd")


class FakeConnectionService:
	var sent_packets: Array = []

	func send_packet(packet: Dictionary, _trace_id: String = "") -> void:
		sent_packets.append(packet)

	func send_tooling_packet(packet: Dictionary) -> void:
		sent_packets.append(packet)


class FakeDebugFlow:
	var calls: Array = []

	func process(required_lane_baselines_synced: bool) -> void:
		calls.append(required_lane_baselines_synced)

	func set_score(target_scope: String, target_player_id: String, score: int) -> void:
		calls.append({"command": "debug_set_score", "target_scope": target_scope, "target_player_id": target_player_id, "value": score})

	func add_score(target_scope: String, target_player_id: String, amount: int) -> void:
		calls.append({"command": "debug_add_score", "target_scope": target_scope, "target_player_id": target_player_id, "value": amount})

	func set_lives(target_scope: String, target_player_id: String, lives: int) -> void:
		calls.append({"command": "debug_set_lives", "target_scope": target_scope, "target_player_id": target_player_id, "value": lives})

	func add_lives(target_scope: String, target_player_id: String, amount: int) -> void:
		calls.append({"command": "debug_add_lives", "target_scope": target_scope, "target_player_id": target_player_id, "value": amount})


class FakeStateContext:
	var lane_baseline_sync := false
	var local_player_id := ""

	func has_lane_baseline_sync() -> bool:
		return lane_baseline_sync

	func get_local_player_id() -> String:
		return local_player_id


class FakeDevConnectionService:
	var configured := true
	var respawn_calls: Array = []

	func is_configured() -> bool:
		return configured

	func send_respawn_player(target_scope: String, target_player_id: String, operation_trace = null) -> void:
		respawn_calls.append({
			"target_scope": target_scope,
			"target_player_id": target_player_id,
			"operation_trace": operation_trace,
		})


class FakeRespawnMarker:
	var call_count := 0

	func mark() -> void:
		call_count += 1


func before_each() -> void:
	ClientLogger.reset_for_tests()


func after_each() -> void:
	ClientLogger.reset_for_tests()


func test_process_delegates_to_debug_flow() -> void:
	var debug_flow := FakeDebugFlow.new()
	var state_context := FakeStateContext.new()
	var context := DevtoolsCommandContext.new()
	context.configure(debug_flow, state_context)

	context.process(true)

	assert_eq(debug_flow.calls.size(), 1)
	assert_true(debug_flow.calls[0])


func test_request_set_game_target_sends_set_target_player_request_packet() -> void:
	var connection := FakeConnectionService.new()
	var state_context := DevtoolsStateContext.new()
	var context := DevtoolsCommandContext.new()
	context.configure(null, state_context)
	context.configure_connection(connection)
	state_context.set_has_lane_baseline_sync(true)

	context.request_set_game_target("Player-2")

	assert_eq(connection.sent_packets.size(), 1)
	var packet = connection.sent_packets[0]
	assert_eq(packet.type, "set_target_player_request")
	assert_eq(packet.target_kind, "player")
	assert_eq(packet.target_id, "Player-2")


func test_request_respawn_player_marks_local_respawn_confirmation_for_local_and_all_targets() -> void:
	var state_context := FakeStateContext.new()
	state_context.lane_baseline_sync = true
	state_context.local_player_id = "player-1"
	var dev_connection_service := FakeDevConnectionService.new()
	var marker := FakeRespawnMarker.new()
	var context := DevtoolsCommandContext.new()
	context.configure(null, state_context)
	context.configure_dev_connection(dev_connection_service)
	context.configure_local_respawn_confirmation_marker(Callable(marker, "mark"))

	context.request_respawn_player(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "player-1")

	assert_eq(dev_connection_service.respawn_calls.size(), 1)
	assert_eq(dev_connection_service.respawn_calls[0]["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_eq(dev_connection_service.respawn_calls[0]["target_player_id"], "player-1")
	assert_true(dev_connection_service.respawn_calls[0]["operation_trace"] is ClientOperationTrace)
	assert_eq(marker.call_count, 1)

	context.request_respawn_player(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "player-2")

	assert_eq(dev_connection_service.respawn_calls.size(), 2)
	assert_eq(dev_connection_service.respawn_calls[1]["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER)
	assert_eq(dev_connection_service.respawn_calls[1]["target_player_id"], "player-2")
	assert_eq(marker.call_count, 1)

	context.request_respawn_player(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "")

	assert_eq(dev_connection_service.respawn_calls.size(), 3)
	assert_eq(dev_connection_service.respawn_calls[2]["target_scope"], DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS)
	assert_eq(dev_connection_service.respawn_calls[2]["target_player_id"], "")
	assert_eq(marker.call_count, 2)
func test_create_operation_trace_returns_independent_deterministic_traces() -> void:
	var state := {"index": 0}
	var ids := [
		"00000000-0000-4000-8000-000000000041",
		"00000000-0000-4000-8000-000000000042",
	]
	var context := DevtoolsCommandContext.new()
	context.configure(null, null, func(action_name: String):
		var trace_id: String = ids[state["index"]]
		state["index"] += 1
		return ClientOperationTrace.new(action_name, func() -> String: return trace_id)
	)

	var first := context.create_operation_trace("toggle_invincible")
	var second := context.create_operation_trace("clear_bullets")

	assert_eq(first.trace_id(), ids[0])
	assert_eq(second.trace_id(), ids[1])
	assert_ne(first.trace_id(), second.trace_id())


func test_set_score_missing_single_player_target_emits_rejection_without_dispatch() -> void:
	_assert_counter_rejection("debug_set_score", "00000000-0000-4000-8000-000000000711", func(context): context.request_set_score(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "", 42))


func test_add_score_missing_single_player_target_emits_rejection_without_dispatch() -> void:
	_assert_counter_rejection("debug_add_score", "00000000-0000-4000-8000-000000000712", func(context): context.request_add_score(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "", 5))


func test_set_lives_missing_single_player_target_emits_rejection_without_dispatch() -> void:
	_assert_counter_rejection("debug_set_lives", "00000000-0000-4000-8000-000000000713", func(context): context.request_set_lives(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "", 3))


func test_add_lives_missing_single_player_target_emits_rejection_without_dispatch() -> void:
	_assert_counter_rejection("debug_add_lives", "00000000-0000-4000-8000-000000000714", func(context): context.request_add_lives(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "", 1))


func test_suppressed_counter_does_not_emit_request_or_rejection_event() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var debug_flow := FakeDebugFlow.new()
	var state_context := FakeStateContext.new()
	var context := DevtoolsCommandContext.new()
	context.configure(debug_flow, state_context)

	context.request_set_score(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "", 42)

	assert_eq(debug_flow.calls.size(), 0)
	assert_eq(writer.written_lines.size(), 0)


func test_missing_debug_flow_emits_dependency_without_request_event() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var state_context := FakeStateContext.new()
	state_context.lane_baseline_sync = true
	var context := DevtoolsCommandContext.new()
	context.configure(null, state_context)

	context.request_clear_bullets()

	assert_push_error_count(1)
	assert_eq(writer.written_lines.size(), 1)
	var record: Dictionary = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], "client_dependency_unavailable")
	assert_ne(record["event"], "devtools_command_requested")


func test_valid_counter_commands_reuse_one_context_trace_for_event_and_packet() -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	var state_context := FakeStateContext.new()
	state_context.lane_baseline_sync = true
	var trace_ids := [
		"00000000-0000-4000-8000-000000000721",
		"00000000-0000-4000-8000-000000000722",
		"00000000-0000-4000-8000-000000000723",
		"00000000-0000-4000-8000-000000000724",
	]
	var trace_state := {"index": 0}
	var trace_factory := func(operation_name: String):
		var trace_id: String = trace_ids[trace_state["index"]]
		trace_state["index"] += 1
		return ClientOperationTrace.new(operation_name, func() -> String: return trace_id)
	debug_flow.configure(connection, trace_factory)
	var context := DevtoolsCommandContext.new()
	context.configure(debug_flow, state_context, trace_factory)

	context.request_set_score(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "", 42)
	context.request_add_score(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "player-2", 5)
	context.request_set_lives(DevtoolsTargetResolver.TARGET_SCOPE_ALL_PLAYERS, "", 3)
	context.request_add_lives(DevtoolsTargetResolver.TARGET_SCOPE_SINGLE_PLAYER, "player-2", 1)

	assert_eq(trace_state["index"], trace_ids.size())
	assert_eq(connection.sent_packets.size(), trace_ids.size())
	assert_eq(writer.written_lines.size(), trace_ids.size())
	for index in range(trace_ids.size()):
		assert_eq(connection.sent_packets[index]["request_id"], trace_ids[index])
		assert_eq(connection.sent_packets[index]["trace_id"], trace_ids[index])
		var record: Dictionary = JSON.parse_string(writer.written_lines[index])
		assert_eq(record["event"], "devtools_command_requested")
		assert_eq(record["trace_id"], trace_ids[index])


func _assert_counter_rejection(command_type: String, trace_id: String, request: Callable) -> void:
	var writer := PresentationEventCapture.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var debug_flow := FakeDebugFlow.new()
	var state_context := FakeStateContext.new()
	state_context.lane_baseline_sync = true
	var context := DevtoolsCommandContext.new()
	context.configure(debug_flow, state_context, func(operation_name: String):
		return ClientOperationTrace.new(operation_name, func() -> String: return trace_id)
	)

	request.call(context)

	assert_eq(debug_flow.calls.size(), 0)
	assert_eq(writer.written_lines.size(), 1)
	var record: Dictionary = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], "devtools_command_rejected")
	assert_eq(record["trace_id"], trace_id)
	assert_eq(record["fields"]["command_type"], command_type)
	assert_eq(record["fields"]["reason"], "target_required")
