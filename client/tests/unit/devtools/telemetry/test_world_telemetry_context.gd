extends GutTest

const WorldTelemetryContext := preload("res://scripts/devtools/telemetry/world_telemetry_context.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


class FakeConnectionService:
	extends RefCounted

	signal telemetry_pong_received(packet: Dictionary)

	var sent_packets: Array[Dictionary] = []

	func send_packet(packet: Dictionary) -> void:
		sent_packets.append(packet)

	func is_server_connected() -> bool:
		return true


func test_process_sends_ping_and_pong_updates_rtt_metrics() -> void:
	var fake_connection := FakeConnectionService.new()
	var telemetry_context := WorldTelemetryContext.new()
	telemetry_context.configure(fake_connection)
	telemetry_context.toggle_overlay()

	telemetry_context.process(true)

	assert_eq(fake_connection.sent_packets.size(), 1)

	var field_type := Packets.FIELD_TYPE
	var field_sequence := Packets.FIELD_SEQUENCE
	var field_client_sent_msec := Packets.FIELD_CLIENT_SENT_MSEC
	var type_telemetry_ping := Packets.TYPE_TELEMETRY_PING

	var ping_packet := fake_connection.sent_packets[0]
	assert_eq(ping_packet[field_type], type_telemetry_ping)
	assert_true(int(ping_packet[field_sequence]) >= 0)
	assert_true(int(ping_packet[field_client_sent_msec]) >= 0)

	fake_connection.telemetry_pong_received.emit({
		field_sequence: int(ping_packet[field_sequence]),
		field_client_sent_msec: int(ping_packet[field_client_sent_msec]),
	})

	assert_true(telemetry_context.network_metrics.snapshot()["rtt_ms"] >= 0)

	var overlay_node = telemetry_context.overlay_flow.overlay
	telemetry_context.reset()
	if is_instance_valid(overlay_node):
		overlay_node.queue_free()



func test_apply_gameplay_state_updates_lane_counts() -> void:
	var fake_connection := FakeConnectionService.new()
	var telemetry_context := WorldTelemetryContext.new()
	telemetry_context.configure(fake_connection)
	telemetry_context.apply_gameplay_state({
		"world": {
			"ships": {
				"player-1": {},
				"player-2": {},
			},
			"asteroids": {
				"asteroid-1": {},
			},
			"bullets": {
				"bullet-1": {},
				"bullet-2": {},
			},
			"pickups": {
				"pickup-1": {},
			},
		},
		"total_asteroids": 7,
	})

	var snapshot: Dictionary = telemetry_context.telemetry_snapshot()
	assert_eq(snapshot["players"], 2)
	assert_eq(snapshot["asteroids"], 1)
	assert_eq(snapshot["bullets"], 2)
	assert_eq(snapshot["pickups"], 1)
	assert_eq(snapshot["total_asteroids"], 7)
