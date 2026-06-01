extends GutTest

const TargetRequestFlow = preload("res://scripts/gameplay/input/target_request_flow.gd")
const TargetVisualCandidate = preload("res://scripts/gameplay/input/target_visual_candidate.gd")

class FakeConnectionService:
	var sent_packets: Array = []

	func send(packet) -> void:
		sent_packets.append(packet)

func test_deselect_target_sends_clear_target_request_packet() -> void:
	var connection := FakeConnectionService.new()
	var flow := TargetRequestFlow.new()
	flow.configure(
		connection,
		Callable(),
		Callable(),
		Callable()
	)

	flow.deselect_target()

	assert(connection.sent_packets.size() == 1)
	assert(connection.sent_packets[0].type == "clear_target_request")

func test_select_target_with_no_visual_target_sends_no_packet() -> void:
	var connection := FakeConnectionService.new()
	var flow := TargetRequestFlow.new()
	flow.configure(
		connection,
		func() -> Array:
			return [],
		func() -> Vector2:
			return Vector2(10, 20),
		func(_visual_position: Vector2) -> Vector2:
			return Vector2(100, 200)
	)

	flow.select_target()

	assert(connection.sent_packets.is_empty())

func test_select_target_with_valid_visual_target_sends_select_target_claim_packet() -> void:
	var connection := FakeConnectionService.new()
	var flow := TargetRequestFlow.new()
	var candidate := TargetVisualCandidate.new()
	candidate.target_kind = "asteroid"
	candidate.target_id = "asteroid-7"
	candidate.visual_position = Vector2(20, 40)
	candidate.pick_radius = 25.0
	candidate.visible = true
	candidate.pick_rank = 1

	var mouse_visual_position := Vector2(25, 44)
	var converted_server_position := Vector2(250, 440)
	flow.configure(
		connection,
		func() -> Array:
			return [candidate],
		func() -> Vector2:
			return mouse_visual_position,
		func(_visual_position: Vector2) -> Vector2:
			return converted_server_position
	)

	flow.select_target()

	assert(connection.sent_packets.size() == 1)
	var sent_packet = connection.sent_packets[0]
	assert(sent_packet.type == "select_target_at_position_request")
	assert(sent_packet.x == converted_server_position.x)
	assert(sent_packet.y == converted_server_position.y)
	assert(sent_packet.target_kind == candidate.target_kind)
	assert(sent_packet.target_id == candidate.target_id)
