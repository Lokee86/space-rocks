extends GutTest

const DevConnectionService := preload("res://scripts/devtools/dev_connection_service.gd")
const ClientOperationTrace := preload("res://scripts/observability/client_operation_trace.gd")


class FakeConnectionService:
	var sent_packets: Array = []

	func send_tooling_packet(packet: Dictionary) -> void:
		sent_packets.append(packet)


func test_spawn_command_uses_tooling_transport_and_correlates_request() -> void:
	var fake_connection := FakeConnectionService.new()
	var service := DevConnectionService.new()
	service.configure(fake_connection)
	var trace_id := "00000000-0000-4000-8000-000000000811"
	var operation_trace := ClientOperationTrace.new("devtools.spawn", func() -> String: return trace_id)

	service.send_spawn_from_placement_result({
		"action_name": &"spawn_asteroid",
		"server_position": Vector2(12.0, 34.0),
	}, operation_trace)

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "debug_spawn_entity")
	assert_eq(packet["request_id"], trace_id)
	assert_eq(packet["trace_id"], trace_id)
	assert_eq(packet["entity_type"], "asteroid")


func test_respawn_command_uses_tooling_transport_and_correlates_request() -> void:
	var fake_connection := FakeConnectionService.new()
	var service := DevConnectionService.new()
	service.configure(fake_connection)
	var trace_id := "00000000-0000-4000-8000-000000000812"
	var operation_trace := ClientOperationTrace.new("devtools.respawn_player", func() -> String: return trace_id)

	service.send_respawn_player("single_player", "player-2", operation_trace)

	assert_eq(fake_connection.sent_packets.size(), 1)
	var packet: Dictionary = fake_connection.sent_packets[0]
	assert_eq(packet["type"], "debug_respawn_player")
	assert_eq(packet["request_id"], trace_id)
	assert_eq(packet["trace_id"], trace_id)
	assert_eq(packet["target_player_id"], "player-2")
