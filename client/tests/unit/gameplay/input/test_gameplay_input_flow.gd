extends GutTest

const GameplayInputFlow := preload("res://scripts/gameplay/input/gameplay_input_flow.gd")


class FakeConnectionService:
	extends RefCounted

	var sent_packets: Array[Dictionary] = []

	func send_input_packet(packet: Dictionary) -> void:
		sent_packets.append(packet.duplicate(true))


class FakePlayer:
	extends RefCounted

	var input_packet := {
		"type": "input",
		"input": {
			"forward": false,
			"back": false,
			"right": false,
			"left": false,
			"primary_fire": false,
			"secondary_fire": false,
		},
	}

	func get_input_packet() -> Dictionary:
		return input_packet.duplicate(true)


class FakeMenuFlow:
	extends RefCounted

	var is_gameplay_paused := false


func test_process_sends_initial_input_then_suppresses_unchanged_frames() -> void:
	var now_msec := [0]
	var flow := GameplayInputFlow.new(func() -> int: return now_msec[0])
	var connection := FakeConnectionService.new()
	var player := FakePlayer.new()
	flow.configure(connection, player, FakeMenuFlow.new())

	flow.process(true)
	now_msec[0] = GameplayInputFlow.INPUT_HEARTBEAT_MSEC - 1
	flow.process(true)

	assert_eq(connection.sent_packets.size(), 1)


func test_process_sends_changed_input_immediately() -> void:
	var now_msec := [0]
	var flow := GameplayInputFlow.new(func() -> int: return now_msec[0])
	var connection := FakeConnectionService.new()
	var player := FakePlayer.new()
	flow.configure(connection, player, FakeMenuFlow.new())

	flow.process(true)
	player.input_packet["input"]["forward"] = true
	now_msec[0] = 1
	flow.process(true)

	assert_eq(connection.sent_packets.size(), 2)
	assert_true(connection.sent_packets[1]["input"]["forward"])


func test_process_sends_unchanged_input_on_heartbeat() -> void:
	var now_msec := [0]
	var flow := GameplayInputFlow.new(func() -> int: return now_msec[0])
	var connection := FakeConnectionService.new()
	var player := FakePlayer.new()
	flow.configure(connection, player, FakeMenuFlow.new())

	flow.process(true)
	now_msec[0] = GameplayInputFlow.INPUT_HEARTBEAT_MSEC
	flow.process(true)

	assert_eq(connection.sent_packets.size(), 2)


func test_reset_forces_next_input_send() -> void:
	var now_msec := [0]
	var flow := GameplayInputFlow.new(func() -> int: return now_msec[0])
	var connection := FakeConnectionService.new()
	var player := FakePlayer.new()
	flow.configure(connection, player, FakeMenuFlow.new())

	flow.process(true)
	flow.reset()
	flow.process(true)

	assert_eq(connection.sent_packets.size(), 2)
