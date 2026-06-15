extends GutTest

const SessionNetworkController := preload("res://scripts/session/session_network_controller.gd")
const ShellBootFlow := preload("res://scripts/boot/shell_boot_flow.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


class FakeConnectionService:
	extends Node

	signal connected
	signal closed
	signal packet_parse_failed(text: String)
	signal unknown_packet_received(packet: Dictionary)
	signal websocket_auth_result_received(packet: Dictionary)

	var websocket_auth_authenticated := false
	var sent_single_player := 0
	var last_local_profile_id := ""
	var sent_create_room := 0
	var sent_join_room_codes: Array[String] = []

	func is_websocket_auth_authenticated() -> bool:
		return websocket_auth_authenticated

	func send_start_single_player_request(local_profile_id := "") -> void:
		last_local_profile_id = local_profile_id
		sent_single_player += 1

	func send_create_room_request() -> void:
		sent_create_room += 1

	func send_join_room_request(room_code: String) -> void:
		sent_join_room_codes.append(room_code)

	func emit_connected() -> void:
		connected.emit()

	func emit_websocket_auth_result(authenticated: bool) -> void:
		websocket_auth_authenticated = authenticated
		websocket_auth_result_received.emit({
			"authenticated": authenticated,
			"user_id": 7 if authenticated else null,
			"display_name": "Ada" if authenticated else "",
		})


func test_connection_sends_single_player_request_without_websocket_auth() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := _create_shell_boot_flow(connection)
	flow.request_single_player()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()

	assert_eq(connection.sent_single_player, 1)
	assert_eq(connection.last_local_profile_id, "")
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_connection_does_not_send_create_room_before_websocket_auth() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := _create_shell_boot_flow(connection)
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()

	assert_eq(connection.sent_create_room, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)


func test_connection_sends_create_room_after_websocket_auth_success() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := _create_shell_boot_flow(connection)
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.emit_websocket_auth_result(true)

	assert_eq(connection.sent_create_room, 1)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_connection_does_not_send_create_room_after_websocket_auth_failure() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := _create_shell_boot_flow(connection)
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.emit_websocket_auth_result(false)

	assert_eq(connection.sent_create_room, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)


func test_connection_sends_create_room_after_websocket_auth_unavailable() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := _create_shell_boot_flow(connection)
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.websocket_auth_result_received.emit({
		"authenticated": false,
		"error_code": "token_verification_unavailable",
	})

	assert_eq(connection.sent_create_room, 1)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_connection_does_not_send_create_room_after_invalid_token_auth_failure() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := _create_shell_boot_flow(connection)
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.websocket_auth_result_received.emit({
		"authenticated": false,
		"error_code": "invalid_token",
	})

	assert_eq(connection.sent_create_room, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)


func _create_shell_boot_flow(connection: FakeConnectionService) -> ShellBootFlow:
	return ShellBootFlow.new(connection, "ws://example", Callable())


func _create_controller(connection: FakeConnectionService, flow: ShellBootFlow) -> SessionNetworkController:
	var controller := SessionNetworkController.new()
	controller.configure(connection, flow, Callable(), {})
	controller.connect_connection_signals()
	return controller
