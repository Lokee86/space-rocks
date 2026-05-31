extends GutTest

const GameplayDebugFlow := preload("res://scripts/devtools/gameplay_debug_flow.gd")


class FakeConnectionService:
	var sent_packets: Array = []

	func send_packet(packet) -> void:
		sent_packets.append(packet)


func test_toggle_freeze_world_with_target_sends_targeted_packet() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.toggle_freeze_world("collisions")

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "toggle_debug_freeze_world")
	assert_eq(packet["freeze_target"], "collisions")


func test_toggle_freeze_world_without_target_sends_aggregate_packet() -> void:
	var fake_connection := FakeConnectionService.new()
	var debug_flow := GameplayDebugFlow.new()
	debug_flow.configure(fake_connection)

	debug_flow.toggle_freeze_world()

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "toggle_debug_freeze_world")
	assert_false(packet.has("freeze_target"))
