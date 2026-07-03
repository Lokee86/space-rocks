extends GutTest

const NetworkRuntimeMetrics := preload("res://scripts/networking/network_runtime_metrics.gd")


func test_default_snapshot_reports_zeroed_network_runtime_metrics() -> void:
	var metrics := NetworkRuntimeMetrics.new()
	var snapshot := metrics.snapshot()

	assert_eq(snapshot["transport"], "websocket")
	assert_eq(snapshot["packets_in"], 0)
	assert_eq(snapshot["packets_out"], 0)
	assert_eq(snapshot["bytes_in"], 0)
	assert_eq(snapshot["bytes_out"], 0)
	assert_eq(snapshot["last_in_packet_bytes"], 0)
	assert_eq(snapshot["last_out_packet_bytes"], 0)
	assert_eq(snapshot["max_in_packet_bytes"], 0)
	assert_eq(snapshot["max_out_packet_bytes"], 0)
	assert_eq(snapshot["decode_failures"], 0)
	assert_eq(snapshot["encode_failures"], 0)
	assert_eq(snapshot["send_failures"], 0)
	assert_eq(snapshot["last_packet_type_in"], "")
	assert_eq(snapshot["last_packet_type_out"], "")


func test_reset_restores_default_snapshot_values() -> void:
	var metrics := NetworkRuntimeMetrics.new()
	metrics.transport = "webrtc"
	metrics.packets_in = 3
	metrics.packets_out = 4
	metrics.bytes_in = 120
	metrics.bytes_out = 240
	metrics.last_in_packet_bytes = 64
	metrics.last_out_packet_bytes = 96
	metrics.max_in_packet_bytes = 80
	metrics.max_out_packet_bytes = 128
	metrics.decode_failures = 1
	metrics.encode_failures = 2
	metrics.send_failures = 3
	metrics.last_packet_type_in = "world_delta"
	metrics.last_packet_type_out = "player_input"

	metrics.reset()

	var snapshot := metrics.snapshot()
	assert_eq(snapshot["transport"], "websocket")
	assert_eq(snapshot["packets_in"], 0)
	assert_eq(snapshot["packets_out"], 0)
	assert_eq(snapshot["bytes_in"], 0)
	assert_eq(snapshot["bytes_out"], 0)
	assert_eq(snapshot["last_in_packet_bytes"], 0)
	assert_eq(snapshot["last_out_packet_bytes"], 0)
	assert_eq(snapshot["max_in_packet_bytes"], 0)
	assert_eq(snapshot["max_out_packet_bytes"], 0)
	assert_eq(snapshot["decode_failures"], 0)
	assert_eq(snapshot["encode_failures"], 0)
	assert_eq(snapshot["send_failures"], 0)
	assert_eq(snapshot["last_packet_type_in"], "")
	assert_eq(snapshot["last_packet_type_out"], "")


func test_observe_inbound_updates_counters_and_max_values() -> void:
	var metrics := NetworkRuntimeMetrics.new()

	metrics.observe_inbound(64, "world_delta")
	metrics.observe_inbound(32, "event_batch")

	var snapshot := metrics.snapshot()
	assert_eq(snapshot["packets_in"], 2)
	assert_eq(snapshot["bytes_in"], 96)
	assert_eq(snapshot["last_in_packet_bytes"], 32)
	assert_eq(snapshot["max_in_packet_bytes"], 64)
	assert_eq(snapshot["last_packet_type_in"], "event_batch")


func test_observe_outbound_updates_counters_and_clamps_negative_byte_values() -> void:
	var metrics := NetworkRuntimeMetrics.new()

	metrics.observe_outbound(48, "player_input")
	metrics.observe_outbound(-5, "telemetry_ping")

	var snapshot := metrics.snapshot()
	assert_eq(snapshot["packets_out"], 2)
	assert_eq(snapshot["bytes_out"], 48)
	assert_eq(snapshot["last_out_packet_bytes"], 0)
	assert_eq(snapshot["max_out_packet_bytes"], 48)
	assert_eq(snapshot["last_packet_type_out"], "telemetry_ping")


func test_observe_decode_failure_increments_failure_counter() -> void:
	var metrics := NetworkRuntimeMetrics.new()

	metrics.observe_decode_failure(64)
	metrics.observe_decode_failure(-5)

	assert_eq(metrics.snapshot()["decode_failures"], 2)


func test_observe_encode_failure_increments_counter_and_preserves_last_packet_type_when_empty() -> void:
	var metrics := NetworkRuntimeMetrics.new()

	metrics.observe_encode_failure("player_input")
	metrics.observe_encode_failure("")

	var snapshot := metrics.snapshot()
	assert_eq(snapshot["encode_failures"], 2)
	assert_eq(snapshot["last_packet_type_out"], "player_input")


func test_observe_send_failure_updates_failure_and_outbound_packet_tracking() -> void:
	var metrics := NetworkRuntimeMetrics.new()

	metrics.observe_send_failure(72, "telemetry_ping")
	metrics.observe_send_failure(-1)

	var snapshot := metrics.snapshot()
	assert_eq(snapshot["send_failures"], 2)
	assert_eq(snapshot["last_out_packet_bytes"], 0)
	assert_eq(snapshot["max_out_packet_bytes"], 72)
	assert_eq(snapshot["last_packet_type_out"], "")
