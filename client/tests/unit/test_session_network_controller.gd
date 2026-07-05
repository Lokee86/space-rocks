extends GutTest

const SessionNetworkController := preload("res://scripts/session/session_network_controller.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


class FakeConnectionService:
	extends Node

	signal connected
	signal closed
	signal packet_parse_failed(text: String)
	signal unknown_packet_received(packet: Dictionary)
	signal websocket_auth_result_received(packet: Dictionary)
	signal webrtc_ready_received(packet: Dictionary)

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

	func emit_webrtc_ready() -> void:
		webrtc_ready_received.emit({"type": "webrtc_ready"})


class FakeShellBootFlow:
	extends RefCounted

	var send_calls := 0
	var pending_request := Constants.BOOT_REQUEST_NONE
	var single_player_local_profile_id := ""
	var create_room_calls := 0
	var join_room_codes: Array[String] = []

	func request_single_player(local_profile_id := "") -> void:
		pending_request = Constants.BOOT_REQUEST_SINGLE_PLAYER
		single_player_local_profile_id = local_profile_id

	func request_create_room() -> void:
		pending_request = Constants.BOOT_REQUEST_CREATE_ROOM
		create_room_calls += 1

	func request_join_room(room_code: String) -> void:
		pending_request = Constants.BOOT_REQUEST_JOIN_ROOM
		join_room_codes.append(room_code)

	func pending_request_is_single_player() -> bool:
		return pending_request == Constants.BOOT_REQUEST_SINGLE_PLAYER

	func pending_request_is_multiplayer() -> bool:
		return pending_request == Constants.BOOT_REQUEST_CREATE_ROOM or pending_request == Constants.BOOT_REQUEST_JOIN_ROOM

	func pending_request_type() -> String:
		return pending_request

	func send_pending_boot_request() -> void:
		send_calls += 1
		if pending_request == Constants.BOOT_REQUEST_SINGLE_PLAYER:
			pending_request = Constants.BOOT_REQUEST_NONE
			return
		if pending_request == Constants.BOOT_REQUEST_CREATE_ROOM:
			pending_request = Constants.BOOT_REQUEST_NONE
			return
		if pending_request == Constants.BOOT_REQUEST_JOIN_ROOM:
			pending_request = Constants.BOOT_REQUEST_NONE




func test_connection_sends_single_player_request_without_websocket_auth() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := FakeShellBootFlow.new()
	flow.request_single_player()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	assert_eq(flow.send_calls, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_SINGLE_PLAYER)
	connection.emit_webrtc_ready()

	assert_eq(flow.send_calls, 1)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_connection_does_not_send_create_room_before_websocket_auth() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := FakeShellBootFlow.new()
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()

	assert_eq(flow.send_calls, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)


func test_connection_sends_create_room_after_websocket_auth_success() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := FakeShellBootFlow.new()
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.emit_websocket_auth_result(true)
	assert_eq(flow.send_calls, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)
	connection.emit_webrtc_ready()

	assert_eq(flow.send_calls, 1)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_connection_does_not_send_create_room_after_websocket_auth_failure() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := FakeShellBootFlow.new()
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.emit_websocket_auth_result(false)

	assert_eq(flow.send_calls, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)


func test_connection_sends_create_room_after_websocket_auth_unavailable() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := FakeShellBootFlow.new()
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.websocket_auth_result_received.emit({
		"authenticated": false,
		"error_code": "token_verification_unavailable",
	})

	assert_eq(flow.send_calls, 1)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_connection_does_not_send_create_room_after_invalid_token_auth_failure() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := FakeShellBootFlow.new()
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.websocket_auth_result_received.emit({
		"authenticated": false,
		"error_code": "invalid_token",
	})

	assert_eq(flow.send_calls, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)


func _create_controller(connection: FakeConnectionService, flow: FakeShellBootFlow) -> SessionNetworkController:
	var controller := SessionNetworkController.new()
	controller.configure(connection, flow, Callable(), {})
	controller.connect_connection_signals()
	return controller


func test_connection_waits_for_webrtc_ready_before_multiplayer_auth_then_sends_once() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := FakeShellBootFlow.new()
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.emit_websocket_auth_result(true)

	assert_eq(flow.send_calls, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)

	connection.emit_webrtc_ready()

	assert_eq(flow.send_calls, 1)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)


func test_connection_waits_for_webrtc_ready_before_multiplayer_ready_then_sends_once() -> void:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var flow := FakeShellBootFlow.new()
	flow.request_create_room()
	var controller := _create_controller(connection, flow)

	connection.emit_connected()
	connection.emit_webrtc_ready()

	assert_eq(flow.send_calls, 0)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_CREATE_ROOM)

	connection.emit_websocket_auth_result(true)

	assert_eq(flow.send_calls, 1)
	assert_eq(flow.pending_request_type(), Constants.BOOT_REQUEST_NONE)
