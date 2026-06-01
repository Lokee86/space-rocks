extends GutTest

const NetworkTelemetryMetrics := preload("res://scripts/devtools/telemetry/network_telemetry_metrics.gd")
const Packets := preload("res://scripts/networking/packets/packets.gd")


func test_initial_snapshot_reports_unavailable_rtt() -> void:
	var metrics := NetworkTelemetryMetrics.new()

	assert_eq(metrics.snapshot()["rtt_ms"], -1)


func test_next_ping_packet_contains_type_sequence_and_client_sent_msec() -> void:
	var metrics := NetworkTelemetryMetrics.new()
	var constants_map: Dictionary = Packets.get_script_constant_map()
	var type_field := Packets.FIELD_TYPE if constants_map.has("FIELD_TYPE") else "type"
	var sequence_field := Packets.FIELD_SEQUENCE if constants_map.has("FIELD_SEQUENCE") else "sequence"
	var client_sent_field := Packets.FIELD_CLIENT_SENT_MSEC if constants_map.has("FIELD_CLIENT_SENT_MSEC") else "client_sent_msec"
	var telemetry_ping_type := Packets.TYPE_TELEMETRY_PING if constants_map.has("TYPE_TELEMETRY_PING") else "telemetry_ping"

	var packet := metrics.next_ping_packet()

	assert_eq(packet[type_field], telemetry_ping_type)
	assert_eq(int(packet[sequence_field]), 0)
	assert_true(int(packet[client_sent_field]) >= 0)


func test_apply_pong_with_matching_sequence_updates_rtt() -> void:
	var metrics := NetworkTelemetryMetrics.new()
	var constants_map: Dictionary = Packets.get_script_constant_map()
	var sequence_field := Packets.FIELD_SEQUENCE if constants_map.has("FIELD_SEQUENCE") else "sequence"
	var packet := metrics.next_ping_packet()
	var sequence := int(packet[sequence_field])

	metrics.apply_pong({sequence_field: sequence})

	assert_true(metrics.snapshot()["rtt_ms"] >= 0)


func test_apply_pong_with_unknown_sequence_keeps_rtt_unchanged() -> void:
	var metrics := NetworkTelemetryMetrics.new()
	var constants_map: Dictionary = Packets.get_script_constant_map()
	var sequence_field := Packets.FIELD_SEQUENCE if constants_map.has("FIELD_SEQUENCE") else "sequence"
	metrics.next_ping_packet()

	metrics.apply_pong({sequence_field: 99999})

	assert_eq(metrics.snapshot()["rtt_ms"], -1)


func test_reset_clears_pending_and_restores_unavailable_rtt() -> void:
	var metrics := NetworkTelemetryMetrics.new()
	var constants_map: Dictionary = Packets.get_script_constant_map()
	var sequence_field := Packets.FIELD_SEQUENCE if constants_map.has("FIELD_SEQUENCE") else "sequence"
	var first_packet := metrics.next_ping_packet()
	metrics.apply_pong({sequence_field: int(first_packet[sequence_field])})
	assert_true(metrics.snapshot()["rtt_ms"] >= 0)

	metrics.next_ping_packet()
	metrics.reset()
	metrics.apply_pong({sequence_field: 1})

	assert_eq(metrics.snapshot()["rtt_ms"], -1)
