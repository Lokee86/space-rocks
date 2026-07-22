extends GutTest

const SessionNetworkController := preload("res://scripts/session/session_network_controller.gd")
const ClientConnectionService := preload("res://scripts/networking/client_connection_service.gd")
const GameplaySessionController := preload("res://scripts/session/gameplay_session_controller.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


class FakeConnectionService extends ClientConnectionService:

	var sent_single_player := 0
	var last_local_profile_id := ""
	var sent_create_room := 0
	var sent_join_room_codes: Array[String] = []
	var events: Array[String] = []

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
		realtime_transport_ready.emit()

	func begin_realtime_match(match_id: String) -> void:
		events.append("begin:%s" % match_id)

	func end_realtime_match() -> void:
		events.append("end")

	func emit_closed() -> void:
		closed.emit()


class FakeRoomSessionController:
	extends RefCounted

	var room_state := Constants.ROOM_STATE_LOBBY
	var match_id := ""

	func current_room_state() -> String:
		return room_state

	func current_match_id() -> String:
		return match_id

	func handle_room_snapshot(packet: Dictionary) -> void:
		room_state = str(packet.get("room_state", room_state))
		match_id = str(packet.get("current_match_id", match_id))

	func handle_room_state_changed(packet: Dictionary) -> void:
		room_state = str(packet.get("room_state", room_state))


class FakeGameplaySessionController extends GameplaySessionController:

	var events: Array[String] = []

	func set_events(shared_events: Array[String]) -> void:
		events = shared_events

	func reset() -> void:
		events.append("gameplay_reset")

	func begin_accepting_gameplay_packets() -> void:
		events.append("gameplay_accept")

	func refresh_match_end_state() -> void:
		pass


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
	controller.configure(connection, flow, {})
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


func _create_match_controller() -> Array:
	var connection := FakeConnectionService.new()
	add_child_autofree(connection)
	var room := FakeRoomSessionController.new()
	var gameplay: FakeGameplaySessionController = autofree(FakeGameplaySessionController.new())
	gameplay.set_events(connection.events)
	var controller := SessionNetworkController.new()
	controller.configure(connection, null, {})
	controller.configure_room_session_controller(room)
	controller.configure_gameplay_session_controller(gameplay)
	controller.connect_connection_signals()
	controller.connect_room_signals()
	return [controller, connection, room, gameplay]


func _match_snapshot(state: String, match_id: String) -> Dictionary:
	return {"room_state": state, "current_match_id": match_id}


func test_match_boundaries_order_idempotency_and_lobby_cleanup() -> void:
	var setup := _create_match_controller()
	var controller: SessionNetworkController = setup[0]
	var connection: FakeConnectionService = setup[1]

	controller._on_room_snapshot_received(_match_snapshot(Constants.ROOM_STATE_IN_GAME, "match-1"))
	assert_eq(connection.events, ["gameplay_reset", "begin:match-1", "gameplay_accept"])
	controller._on_room_snapshot_received(_match_snapshot(Constants.ROOM_STATE_IN_GAME, "match-1"))
	assert_eq(connection.events, ["gameplay_reset", "begin:match-1", "gameplay_accept", "gameplay_accept"])
	controller._on_room_snapshot_received(_match_snapshot(Constants.ROOM_STATE_GAME_OVER, "match-1"))
	controller._on_room_state_changed(_match_snapshot(Constants.ROOM_STATE_GAME_OVER, "match-1"))
	assert_eq(connection.events.size(), 4)
	controller._on_room_snapshot_received(_match_snapshot(Constants.ROOM_STATE_LOBBY, "match-1"))
	assert_eq(connection.events, ["gameplay_reset", "begin:match-1", "gameplay_accept", "gameplay_accept", "gameplay_reset", "end"])
	controller._on_room_snapshot_received(_match_snapshot(Constants.ROOM_STATE_LOBBY, "match-1"))
	controller._on_room_state_changed(_match_snapshot(Constants.ROOM_STATE_LOBBY, "match-1"))
	assert_eq(connection.events.size(), 6)
	controller._on_room_snapshot_received(_match_snapshot(Constants.ROOM_STATE_IN_GAME, "match-2"))
	assert_eq(connection.events.slice(6), ["gameplay_reset", "begin:match-2", "gameplay_accept"])


func test_connection_close_ends_active_match_once() -> void:
	var setup := _create_match_controller()
	var controller: SessionNetworkController = setup[0]
	var connection: FakeConnectionService = setup[1]
	controller._on_room_snapshot_received(_match_snapshot(Constants.ROOM_STATE_IN_GAME, "match-1"))
	connection.events.clear()
	connection.emit_closed()
	assert_eq(connection.events, ["gameplay_reset", "end"])
	controller._on_room_state_changed(_match_snapshot(Constants.ROOM_STATE_LOBBY, "match-1"))
	assert_eq(connection.events, ["gameplay_reset", "end"])
