extends GutTest

const GameplayDevtoolsContext = preload("res://scripts/devtools/gameplay_devtools_context.gd")

class FakeConnectionService:
	var sent_packets: Array = []

	func send_packet(packet) -> void:
		sent_packets.append(packet)

func test_request_set_game_target_sends_set_target_player_request_packet() -> void:
	var connection := FakeConnectionService.new()
	var context := GameplayDevtoolsContext.new()
	context.configure(connection)
	context.state_context.set_has_received_gameplay_state(true)

	context.request_set_game_target("Player-2")

	assert_eq(connection.sent_packets.size(), 1)
	var packet = connection.sent_packets[0]
	assert_eq(packet.type, "set_target_player_request")
	assert_eq(packet.target_player_id, "Player-2")
