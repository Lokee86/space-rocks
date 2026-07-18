extends GutTest

const MeasurementOverlaySnapshot := preload("res://scripts/devtools/measurement/client_measurement_overlay_snapshot.gd")


class FakeCoordinator extends RefCounted:
	var state: Dictionary = {}

	func get_state() -> Dictionary:
		return state.duplicate(true)


class FakeClientMeasurementContext extends RefCounted:
	var value: Dictionary = {}

	func snapshot() -> Dictionary:
		return value.duplicate(true)


func test_builds_compact_overlay_metrics_and_tracks_server_snapshot_age() -> void:
	var coordinator := FakeCoordinator.new()
	coordinator.state = {
		"active_run_id": "run-9",
		"recording": true,
		"pending_request_ids": {},
		"last_tooling_error": {},
		"latest_server_snapshot": {
			"ticks": {"average": 2500000, "maximum": 6000000},
			"samples": [{
				"entities": {"players": 3, "asteroids": 12, "projectiles": 7, "pickups": 2},
			}],
			"packets": [
				{"packet_count": 4, "encoded_bytes_total": 400},
				{"packet_count": 6, "encoded_bytes_total": 900},
			],
		},
	}
	var client_context := FakeClientMeasurementContext.new()
	client_context.value = {
		"duration": 1250.0,
		"frame_timing": {"average": 16.5, "maximum": 31.0},
		"network_metrics": {"packets_in": 8, "packets_out": 3, "bytes_in": 800, "bytes_out": 250},
	}
	var adapter := MeasurementOverlaySnapshot.new()
	adapter.configure(coordinator, client_context)

	var metrics := adapter.snapshot(2000)
	assert_eq(metrics["measurement_status"], "recording")
	assert_eq(metrics["measurement_run_id"], "run-9")
	assert_eq(metrics["measurement_elapsed_ms"], 1250.0)
	assert_eq(metrics["measurement_client_frame_average_ms"], 16.5)
	assert_eq(metrics["measurement_client_frame_maximum_ms"], 31.0)
	assert_eq(metrics["measurement_server_tick_average_ms"], 2.5)
	assert_eq(metrics["measurement_server_tick_maximum_ms"], 6.0)
	assert_eq(metrics["measurement_server_players"], 3)
	assert_eq(metrics["measurement_server_asteroids"], 12)
	assert_eq(metrics["measurement_server_projectiles"], 7)
	assert_eq(metrics["measurement_server_pickups"], 2)
	assert_eq(metrics["measurement_client_packets"], 11)
	assert_eq(metrics["measurement_client_bytes"], 1050)
	assert_eq(metrics["measurement_server_packets"], 10)
	assert_eq(metrics["measurement_server_bytes"], 1300)
	assert_eq(metrics["measurement_snapshot_age_ms"], 0)

	metrics = adapter.snapshot(2250)
	assert_eq(metrics["measurement_snapshot_age_ms"], 250)


func test_reports_starting_without_fabricating_unavailable_measurements() -> void:
	var coordinator := FakeCoordinator.new()
	coordinator.state = {
		"active_run_id": "",
		"recording": false,
		"pending_request_ids": {"start": "request-1"},
		"latest_server_snapshot": {},
	}
	var adapter := MeasurementOverlaySnapshot.new()
	adapter.configure(coordinator, FakeClientMeasurementContext.new())

	var metrics := adapter.snapshot(100)
	assert_eq(metrics["measurement_status"], "starting")
	assert_eq(metrics["measurement_run_id"], "")
	assert_eq(metrics["measurement_server_tick_average_ms"], -1.0)
	assert_eq(metrics["measurement_server_packets"], -1)
	assert_eq(metrics["measurement_server_bytes"], -1)
	assert_eq(metrics["measurement_snapshot_age_ms"], -1)
