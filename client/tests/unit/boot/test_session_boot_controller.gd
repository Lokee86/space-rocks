extends GutTest

const SessionBootController := preload("res://scripts/boot/session_boot_controller.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")


class FakeConnectionService:
	extends RefCounted

	var connected := false
	var authenticated := false

	func is_server_connected() -> bool:
		return connected

	func is_websocket_auth_authenticated() -> bool:
		return authenticated


class FakeShellBootFlow:
	extends RefCounted

	var requested: Array = []
	var websocket_urls: Array[String] = []
	var connect_reasons: Array[String] = []
	var send_calls := 0

	func request_loadout_options(local_profile_id: String, play_mode: String, mode_id: String) -> void:
		requested.append([local_profile_id, play_mode, mode_id])

	func set_websocket_url(url: String) -> void:
		websocket_urls.append(url)

	func connect_to_game_server(reason: String) -> String:
		connect_reasons.append(reason)
		return Constants.CONNECT_RESULT_STARTED_CONNECTING

	func send_pending_loadout_request() -> void:
		send_calls += 1


func test_disconnected_loadout_request_starts_preflight_connection() -> void:
	var controller := SessionBootController.new()
	var connection := FakeConnectionService.new()
	var flow := FakeShellBootFlow.new()
	controller.connection_service = connection
	controller.shell_boot_flow = flow

	controller.request_loadout_options("pilot-1", Constants.SESSION_MODE_SINGLE_PLAYER, "arcade_survival")

	assert_eq(flow.requested, [["pilot-1", Constants.SESSION_MODE_SINGLE_PLAYER, "arcade_survival"]])
	assert_eq(flow.websocket_urls, [Constants.SINGLE_PLAYER_WS_URL])
	assert_eq(flow.connect_reasons, ["single_player loadout"])
	assert_eq(flow.send_calls, 0)


func test_connected_single_player_reuses_socket_immediately() -> void:
	var controller := SessionBootController.new()
	var connection := FakeConnectionService.new()
	var flow := FakeShellBootFlow.new()
	connection.connected = true
	controller.connection_service = connection
	controller.shell_boot_flow = flow

	controller.request_loadout_options("pilot-1", Constants.SESSION_MODE_SINGLE_PLAYER, "arcade_survival")

	assert_eq(flow.send_calls, 1)
	assert_true(flow.connect_reasons.is_empty())


func test_connected_multiplayer_waits_until_authentication() -> void:
	var controller := SessionBootController.new()
	var connection := FakeConnectionService.new()
	var flow := FakeShellBootFlow.new()
	connection.connected = true
	controller.connection_service = connection
	controller.shell_boot_flow = flow

	controller.request_loadout_options("", Constants.SESSION_MODE_MULTIPLAYER, "arcade_survival")
	assert_eq(flow.send_calls, 0)

	connection.authenticated = true
	controller.request_loadout_options("", Constants.SESSION_MODE_MULTIPLAYER, "arcade_survival")
	assert_eq(flow.send_calls, 1)
