extends GutTest

const ShellBootFlow := preload("res://scripts/boot/shell_boot_flow.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")


class FakeConnectionService:
	extends RefCounted

	var sent_single_player := 0
	var last_local_profile_id := ""
	var sent_create_room := 0
	var sent_join_room_codes: Array[String] = []
	var room_operations: Array = []
	var loadout_requests: Array = []

	func send_start_single_player_request(local_profile_id := "") -> void:
		last_local_profile_id = local_profile_id
		sent_single_player += 1

	func send_create_room_request() -> void:
		sent_create_room += 1

	func send_join_room_request(room_code: String) -> void:
		sent_join_room_codes.append(room_code)

	func send_loadout_options_request(local_profile_id: String, play_mode: String, mode_id: String) -> void:
		loadout_requests.append([local_profile_id, play_mode, mode_id])

	func begin_room_operation(operation_type: String, trace_id: String) -> void:
		room_operations.append({
			"operation_type": operation_type,
			"trace_id": trace_id,
		})


func test_send_pending_loadout_request_consumes_and_sends() -> void:
	var connection := FakeConnectionService.new()
	var flow := ShellBootFlow.new(connection, "ws://example", Callable())

	flow.request_loadout_options("pilot-1", Constants.SESSION_MODE_SINGLE_PLAYER, "arcade_survival")

	assert_true(flow.has_pending_loadout_request())
	assert_true(flow.pending_loadout_request_is_single_player())
	assert_false(flow.pending_loadout_request_is_multiplayer())
	flow.send_pending_loadout_request()
	assert_eq(connection.loadout_requests, [["pilot-1", Constants.SESSION_MODE_SINGLE_PLAYER, "arcade_survival"]])
	assert_false(flow.has_pending_loadout_request())


func test_send_pending_single_player_request_consumes_and_sends() -> void:
	var connection := FakeConnectionService.new()
	var flow := ShellBootFlow.new(connection, "ws://example", Callable())

	flow.request_single_player()
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_SINGLE_PLAYER)
	assert_true(flow.pending_request_is_single_player())
	assert_false(flow.pending_request_is_multiplayer())

	flow.send_pending_boot_request()

	assert_eq(connection.sent_single_player, 1)
	assert_eq(connection.last_local_profile_id, "")
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_pending_create_room_is_multiplayer_without_consuming() -> void:
	var connection := FakeConnectionService.new()
	var flow := ShellBootFlow.new(connection, "ws://example", Callable())

	flow.request_create_room()

	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)
	assert_true(flow.pending_request_is_multiplayer())
	assert_false(flow.pending_request_is_single_player())


func test_pending_join_room_is_multiplayer_without_consuming() -> void:
	var connection := FakeConnectionService.new()
	var flow := ShellBootFlow.new(connection, "ws://example", Callable())

	flow.request_join_room("ABCD")

	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_JOIN_ROOM)
	assert_true(flow.pending_request_is_multiplayer())
	assert_false(flow.pending_request_is_single_player())

func test_send_pending_transfers_trace_to_connection_service_before_request() -> void:
	var connection := FakeConnectionService.new()
	var flow := ShellBootFlow.new(connection, "ws://example", Callable(), func(operation_name: String):
		return ClientOperationTrace.new(
			operation_name,
			func() -> String: return "00000000-0000-4000-8000-000000000007"
		)
	)

	flow.request_create_room()
	flow.send_pending_boot_request()

	assert_eq(connection.room_operations.size(), 1)
	assert_eq(connection.room_operations[0]["operation_type"], Constants.BOOT_REQUEST_CREATE_ROOM)
	assert_eq(connection.room_operations[0]["trace_id"], "00000000-0000-4000-8000-000000000007")
	assert_eq(connection.sent_create_room, 1)