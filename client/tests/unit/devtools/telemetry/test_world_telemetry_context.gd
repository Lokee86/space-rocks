extends GutTest

const WorldTelemetryContext := preload("res://scripts/devtools/telemetry/world_telemetry_context.gd")
const Packets := preload("res://scripts/networking/packets/packets.gd")


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

	var constants_map: Dictionary = Packets.get_script_constant_map()
	var field_type := Packets.FIELD_TYPE if constants_map.has("FIELD_TYPE") else "type"
	var field_sequence := Packets.FIELD_SEQUENCE if constants_map.has("FIELD_SEQUENCE") else "sequence"
	var field_client_sent_msec := Packets.FIELD_CLIENT_SENT_MSEC if constants_map.has("FIELD_CLIENT_SENT_MSEC") else "client_sent_msec"
	var type_telemetry_ping := Packets.TYPE_TELEMETRY_PING if constants_map.has("TYPE_TELEMETRY_PING") else "telemetry_ping"

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
