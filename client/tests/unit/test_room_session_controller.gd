extends GutTest

const RoomSessionController := preload("res://scripts/session/room_session_controller.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")


class FakeSessionContext:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1

	func activate_requested_mode() -> void:
		pass

	func should_show_multiplayer_lobby(_room_state: String) -> bool:
		return false


class FakeConnectionService:
	extends RefCounted

	func send_set_ready_request(_ready: bool) -> void:
		pass

	func send_start_game_request() -> void:
		pass

	func send_add_bot_request() -> void:
		pass

	func send_remove_room_member_request(_player_id: String) -> void:
		pass

	func send_leave_room_request() -> void:
		pass

	var active_operation_type := ""
	var active_operation_trace_id := ""
	var clear_room_operation_calls := 0

	func active_room_operation_type() -> String:
		return active_operation_type

	func active_room_operation_trace_id() -> String:
		return active_operation_trace_id

	func clear_room_operation_context() -> void:
		clear_room_operation_calls += 1
		active_operation_type = ""
		active_operation_trace_id = ""


class FakeShellBootFlow:
	extends RefCounted

	var clear_calls := 0

	func clear() -> void:
		clear_calls += 1


class Probe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


class OperationProbe:
	extends RefCounted

	var calls := 0
	var operation := ""
	var message := ""

	func capture(operation_value: String, message_value: String) -> void:
		calls += 1
		operation = operation_value
		message = message_value


class FakeWriter:
	extends RefCounted

	var written_lines: Array[String] = []

	func write_line(line: String) -> void:
		written_lines.append(line)

	func close() -> void:
		pass


func before_each() -> void:
	ClientLogger.reset_for_tests()


func after_each() -> void:
	ClientLogger.reset_for_tests()




func test_handle_room_snapshot_caches_current_match_id() -> void:
	var setup := _create_controller()
	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_IN_GAME,
		Packets.FIELD_CURRENT_MATCH_ID: "match-1",
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
	})
	assert_eq(setup.controller.current_match_id(), "match-1")
	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_GAME_OVER,
		Packets.FIELD_CURRENT_MATCH_ID: "match-1",
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
	})
	assert_eq(setup.controller.current_match_id(), "match-1")


func test_handle_room_snapshot_clears_current_match_id_when_snapshot_is_empty() -> void:
	var setup := _create_controller()
	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_IN_GAME,
		Packets.FIELD_CURRENT_MATCH_ID: "match-1",
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
	})
	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_LOBBY,
		Packets.FIELD_ROOM_CODE: "",
		Packets.FIELD_MEMBERS: [],
	})
	assert_eq(setup.controller.current_match_id(), "")


func test_lobby_return_cleanup_clears_session_context_and_shell_boot_flow() -> void:
	var setup := _create_controller()

	setup.controller.lobby_return_flow.return_after_leave()

	assert_eq(setup.session_context.clear_calls, 1)
	assert_eq(setup.shell_boot_flow.clear_calls, 1)


func test_configure_lobby_leave_return_destination_passes_destination_to_lobby_return_flow() -> void:
	var setup := _create_controller()
	var destination_probe := Probe.new()

	setup.controller.configure_lobby_leave_return_destination(Callable(destination_probe, "mark_called"))
	setup.controller.lobby_return_flow.return_after_leave()

	assert_eq(destination_probe.calls, 1)


func test_handle_room_snapshot_caches_valid_match_result() -> void:
	var setup := _create_controller()
	var match_result := {
		Packets.FIELD_MATCH_ID: "room-match-1",
		Packets.FIELD_MODE: "single_player",
		Packets.FIELD_PLAYERS: [
			{
				Packets.FIELD_GAME_PLAYER_ID: "player-1",
				Packets.FIELD_SCORE: 450,
				Packets.FIELD_SHIP_DEATHS: 2,
				Packets.FIELD_WON: false,
			}
		],
	}

	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_GAME_OVER,
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
		Packets.FIELD_MATCH_RESULT: match_result,
	})

	assert_eq(setup.controller.current_match_result(), match_result)


func test_handle_room_snapshot_clears_match_result_for_empty_snapshot_object() -> void:
	var setup := _create_controller()
	var match_result := {
		Packets.FIELD_MATCH_ID: "room-match-1",
		Packets.FIELD_MODE: "single_player",
		Packets.FIELD_PLAYERS: [],
	}

	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_GAME_OVER,
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
		Packets.FIELD_MATCH_RESULT: match_result,
	})
	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_LOBBY,
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
		Packets.FIELD_MATCH_RESULT: {
			Packets.FIELD_MATCH_ID: "",
			Packets.FIELD_MODE: "",
			Packets.FIELD_PLAYERS: [],
		},
	})

	assert_eq(setup.controller.current_match_result(), {})


func test_handle_room_snapshot_clears_match_result_when_field_missing() -> void:
	var setup := _create_controller()

	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_GAME_OVER,
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
		Packets.FIELD_MATCH_RESULT: {
			Packets.FIELD_MATCH_ID: "room-match-1",
			Packets.FIELD_MODE: "single_player",
			Packets.FIELD_PLAYERS: [],
		},
	})
	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_LOBBY,
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
	})

	assert_eq(setup.controller.current_match_result(), {})


func test_initial_room_snapshot_clears_transition_ui_and_active_room_operation() -> void:
	var setup := _create_controller()
	var transition_probe := Probe.new()
	setup.controller.configure_room_transition_completed(Callable(transition_probe, "mark_called"))
	setup.connection_service.active_operation_type = Constants.BOOT_REQUEST_JOIN_ROOM
	setup.connection_service.active_operation_trace_id = "trace-join"

	setup.controller.handle_room_snapshot({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: Constants.ROOM_STATE_LOBBY,
		Packets.FIELD_ROOM_CODE: "ROOM1",
		Packets.FIELD_MEMBERS: [],
	})

	assert_eq(transition_probe.calls, 1)
	assert_eq(setup.connection_service.clear_room_operation_calls, 1)
	assert_eq(setup.connection_service.active_operation_trace_id, "")


func test_room_error_emits_bounded_failure_and_clears_matching_operation() -> void:
	var writer := FakeWriter.new()
	ClientLogger._set_file_writer_for_tests(writer)
	var setup := _create_controller()
	var operation_probe := OperationProbe.new()
	setup.controller.configure_room_operation_failed(Callable(operation_probe, "capture"))
	setup.connection_service.active_operation_type = Constants.BOOT_REQUEST_CREATE_ROOM
	var trace_id := "00000000-0000-4000-8000-000000000022"
	setup.connection_service.active_operation_trace_id = trace_id
	var server_message := "unrestricted server message"

	setup.controller.handle_room_error({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_ERROR,
		Packets.FIELD_TRACE_ID: trace_id,
		Packets.FIELD_ERROR_CODE: "room_full",
		Packets.FIELD_MESSAGE: server_message,
	})

	assert_eq(setup.connection_service.clear_room_operation_calls, 1)
	assert_eq(operation_probe.calls, 1)
	assert_eq(operation_probe.operation, "create_room")
	assert_eq(operation_probe.message, "Room is full.")
	assert_eq(writer.written_lines.size(), 1)
	var record = JSON.parse_string(writer.written_lines[0])
	assert_eq(record["event"], ObservabilityContract.EVENT_ROOM_OPERATION_FAILED)
	assert_eq(record["trace_id"], trace_id)
	assert_eq(record["error_code"], "room_full")
	assert_eq(record["fields"]["operation"], "create_room")
	assert_false(record["fields"].has("message"))
	assert_false(writer.written_lines[0].contains(server_message))


func test_stale_room_error_does_not_clear_newer_operation() -> void:
	var setup := _create_controller()
	setup.connection_service.active_operation_type = Constants.BOOT_REQUEST_JOIN_ROOM
	setup.connection_service.active_operation_trace_id = "00000000-0000-4000-8000-000000000023"

	setup.controller.handle_room_error({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_ERROR,
		Packets.FIELD_TRACE_ID: "00000000-0000-4000-8000-000000000024",
		Packets.FIELD_ERROR_CODE: "room_not_found",
		Packets.FIELD_MESSAGE: "stale server message",
	})

	assert_eq(setup.connection_service.clear_room_operation_calls, 0)
	assert_eq(setup.connection_service.active_operation_trace_id, "00000000-0000-4000-8000-000000000023")

func _create_controller() -> Dictionary:

	var main_menu := Control.new()
	var canvas_layer := CanvasLayer.new()
	var session_context := FakeSessionContext.new()
	var connection_service := FakeConnectionService.new()
	var shell_boot_flow := FakeShellBootFlow.new()
	var controller := RoomSessionController.new()

	add_child_autofree(main_menu)
	add_child_autofree(canvas_layer)
	main_menu.hide()

	controller.configure(
		main_menu,
		canvas_layer,
		session_context,
		connection_service,
		shell_boot_flow
	)

	return {
		"controller": controller,
		"main_menu": main_menu,
		"session_context": session_context,
		"shell_boot_flow": shell_boot_flow,
		"connection_service": connection_service,
	}
