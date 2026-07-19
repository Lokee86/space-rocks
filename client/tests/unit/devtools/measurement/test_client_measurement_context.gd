extends GutTest

const ClientMeasurementContext := preload("res://scripts/devtools/measurement/client_measurement_context.gd")


class FakeWorldSync:
	func entity_counts() -> Dictionary:
		return {"players": 2, "bullets": 3, "asteroids": 4, "pickups": 1}

	func scene_node_count() -> int:
		return 24


class FakeTelemetry:
	func telemetry_snapshot() -> Dictionary:
		return {"rtt_ms": 18, "jitter_ms": 3, "packet_age_ms": 7, "packet_staleness_ms": 9}


class FakeConnection:
	var metrics := {"rtt_ms": 20, "jitter_ms": 4}

	func network_metrics_snapshot() -> Dictionary:
		return metrics.duplicate(true)


func test_periodic_sample_includes_existing_telemetry_and_world_pressure() -> void:
	var context := ClientMeasurementContext.new()
	context.configure(FakeConnection.new(), null, FakeWorldSync.new(), FakeTelemetry.new())
	context.start({"source": "unit"})
	context.process_frame(1.0)

	var result := context.stop()
	var sample: Dictionary = result["resource_samples"]["samples"][0]

	assert_eq(sample["entity_counts"]["players"], 2)
	assert_eq(sample["entity_counts"]["asteroids"], 4)
	assert_eq(sample["node_count"], 24)
	assert_eq(sample["network_metrics"]["rtt_ms"], 20)
	assert_eq(sample["network_metrics"]["packet_age_ms"], 7)
	assert_true(result["long_match_store_metrics"].has("total_entries"))
	assert_eq(result["metadata"]["source"], "unit")


func test_network_counters_are_relative_to_measurement_start() -> void:
	var connection := FakeConnection.new()
	connection.metrics = {
		"packets_in": 100,
		"packets_out": 50,
		"bytes_in": 1000,
		"bytes_out": 500,
		"webrtc_lanes": {
			"world": {"packets_in": 80, "bytes_in": 800},
		},
	}
	var context := ClientMeasurementContext.new()
	context.configure(connection, null, FakeWorldSync.new(), FakeTelemetry.new())
	context.start()

	connection.metrics["packets_in"] = 112
	connection.metrics["packets_out"] = 57
	connection.metrics["bytes_in"] = 1300
	connection.metrics["bytes_out"] = 650
	connection.metrics["webrtc_lanes"]["world"]["packets_in"] = 89
	connection.metrics["webrtc_lanes"]["world"]["bytes_in"] = 920
	context.process_frame(1.0)

	var result := context.stop()
	var samples: Array = result["resource_samples"]["samples"]
	assert_eq(samples[0]["network_metrics"]["packets_in"], 0)
	assert_eq(samples[0]["network_metrics"]["bytes_out"], 0)
	assert_eq(samples[1]["network_metrics"]["packets_in"], 12)
	assert_eq(samples[1]["network_metrics"]["packets_out"], 7)
	assert_eq(samples[1]["network_metrics"]["bytes_in"], 300)
	assert_eq(samples[1]["network_metrics"]["bytes_out"], 150)
	assert_eq(samples[1]["network_metrics"]["webrtc_lanes"]["world"]["packets_in"], 9)
	assert_eq(samples[1]["network_metrics"]["webrtc_lanes"]["world"]["bytes_in"], 120)
	assert_eq(samples[2]["network_metrics"]["packets_in"], 12)
