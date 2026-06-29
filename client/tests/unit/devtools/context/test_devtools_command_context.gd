extends GutTest

const DevtoolsCommandContext := preload("res://scripts/devtools/context/devtools_command_context.gd")
const DevtoolsStateContext := preload("res://scripts/devtools/context/devtools_state_context.gd")
const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")


class FakeConnectionService:
	var sent_packets: Array = []

	func send_packet(packet) -> void:
		sent_packets.append(packet)


class FakeDebugFlow:
	var calls: Array = []

	func process(has_received_state: bool) -> void:
		calls.append(has_received_state)


class FakeStateContext:
	var gameplay_state := false
	var local_player_id := ""

	func has_gameplay_state() -> bool:
		return gameplay_state

	func get_local_player_id() -> String:
		return local_player_id


class FakeDevConnectionService:
	var configured := true
	var respawn_calls: Array = []

	func is_configured() -> bool:
		return configured

	func send_respawn_player(target_scope: String, target_player_id: String) -> void:
		respawn_calls.append({
			"target_scope": target_scope,
			"target_player_id": target_player_id,
		})


class FakeRespawnMarker:
	var call_count := 0

	func mark() -> void:
		call_count += 1


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
	state_context.set_has_received_gameplay_state(true)

	context.request_set_game_target("Player-2")

	assert_eq(connection.sent_packets.size(), 1)
	var packet = connection.sent_packets[0]
	assert_eq(packet.type, "set_target_player_request")
	assert_eq(packet.target_kind, "player")
	assert_eq(packet.target_id, "Player-2")


func test_request_respawn_player_marks_local_respawn_confirmation_for_local_and_all_targets() -> void:
	var state_context := FakeStateContext.new()
	state_context.gameplay_state = true
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