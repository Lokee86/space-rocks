extends GutTest

const NetworkTelemetryMetrics := preload("res://scripts/devtools/telemetry/network_telemetry_metrics.gd")
const Packets := preload("res://scripts/networking/packets/packets.gd")


func test_initial_snapshot_reports_unavailable_rtt() -> void:
	var metrics := NetworkTelemetryMetrics.new()

	assert_eq(metrics.snapshot()["rtt_ms"], -1)


func test_next_ping_packet_contains_type_sequence_and_client_sent_msec() -> void:
	var metrics := NetworkTelemetryMetrics.new()

	var packet := metrics.next_ping_packet()

	assert_eq(packet[Packets.FIELD_TYPE], Packets.TYPE_TELEMETRY_PING)
	assert_eq(int(packet[Packets.FIELD_SEQUENCE]), 0)
	assert_true(int(packet[Packets.FIELD_CLIENT_SENT_MSEC]) >= 0)


func test_apply_pong_with_matching_sequence_updates_rtt() -> void:
	var metrics := NetworkTelemetryMetrics.new()
	var packet := metrics.next_ping_packet()
	var sequence := int(packet[Packets.FIELD_SEQUENCE])

	metrics.apply_pong({Packets.FIELD_SEQUENCE: sequence})

	assert_true(metrics.snapshot()["rtt_ms"] >= 0)


func test_apply_pong_with_unknown_sequence_keeps_rtt_unchanged() -> void:
	var metrics := NetworkTelemetryMetrics.new()
	metrics.next_ping_packet()

	metrics.apply_pong({Packets.FIELD_SEQUENCE: 99999})

	assert_eq(metrics.snapshot()["rtt_ms"], -1)


func test_reset_clears_pending_and_restores_unavailable_rtt() -> void:
	var metrics := NetworkTelemetryMetrics.new()
	var first_packet := metrics.next_ping_packet()
	metrics.apply_pong({Packets.FIELD_SEQUENCE: int(first_packet[Packets.FIELD_SEQUENCE])})
	assert_true(metrics.snapshot()["rtt_ms"] >= 0)

	metrics.next_ping_packet()
	metrics.reset()
	metrics.apply_pong({Packets.FIELD_SEQUENCE: 1})

	assert_eq(metrics.snapshot()["rtt_ms"], -1)
