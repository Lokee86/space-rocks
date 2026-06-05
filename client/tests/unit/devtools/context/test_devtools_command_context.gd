extends GutTest

const DevtoolsCommandContext := preload("res://scripts/devtools/context/devtools_command_context.gd")
const DevtoolsStateContext := preload("res://scripts/devtools/context/devtools_state_context.gd")


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

	func has_gameplay_state() -> bool:
		return gameplay_state


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
