extends GutTest

const WorldTelemetryOverlayScene := preload("res://scenes/devtools/world_telemetry_overlay.tscn")


func test_renders_authoritative_server_and_transport_metrics() -> void:
	var overlay := WorldTelemetryOverlayScene.instantiate()
	add_child_autofree(overlay)
	overlay.refresh_metrics({
		"players": 1,
		"asteroids": 2,
		"bullets": 3,
		"server_players": 4,
		"server_enemies": 5,
		"server_asteroids": 6,
		"server_pickups": 7,
		"server_projectiles": 8,
		"server_total_asteroids_spawned": 9,
		"server_room_count": 2,
		"server_match_id": "match-4",
		"server_player_sessions": 4,
		"server_radial_effects": 1,
		"server_heap_allocated_bytes": 1000,
		"server_heap_in_use_bytes": 900,
		"server_system_bytes": 2000,
		"server_goroutines": 12,
		"server_gc_cycles": 3,
		"server_packets_out": 30,
		"server_bytes_out": 4000,
		"server_max_packet_bytes": 700,
		"server_lane_metrics": {
			"tooling": {
				"packet_count": 2,
				"encoded_bytes_total": 300,
				"maximum_encoded_bytes": 200,
			},
		},
	})

	var label := overlay.find_child("MetricsLabel", true, false) as Label
	assert_not_null(label)
	assert_true(label.text.contains("players: 4"))
	assert_true(label.text.contains("enemies: 5"))
	assert_true(label.text.contains("projectiles: 8"))
	assert_true(label.text.contains("asteroids_spawned: 9"))
	assert_true(label.text.contains("match: match-4"))
	assert_true(label.text.contains("packets_out: 30"))
	assert_true(label.text.contains("tooling p/b/max: 2/300/200"))


func test_renders_compact_measurement_section() -> void:
	var overlay := WorldTelemetryOverlayScene.instantiate()
	add_child_autofree(overlay)
	overlay.refresh_metrics({
		"measurement_status": "recording",
		"measurement_run_id": "run-4",
		"measurement_elapsed_ms": 2500.0,
		"measurement_client_frame_average_ms": 16.25,
		"measurement_client_frame_maximum_ms": 29.5,
		"measurement_server_tick_average_ms": 3.0,
		"measurement_server_tick_maximum_ms": 8.0,
		"measurement_server_players": 2,
		"measurement_server_asteroids": 15,
		"measurement_server_projectiles": 6,
		"measurement_server_pickups": 1,
		"measurement_client_packets": 30,
		"measurement_client_bytes": 4000,
		"measurement_server_packets": 24,
		"measurement_server_bytes": 3500,
		"measurement_snapshot_age_ms": 125,
	})

	var label := overlay.find_child("MetricsLabel", true, false) as Label
	assert_not_null(label)
	assert_true(label.text.contains("Measurement"))
	assert_true(label.text.contains("status: recording"))
	assert_true(label.text.contains("run: run-4"))
	assert_true(label.text.contains("elapsed_s: 2.5"))
	assert_true(label.text.contains("client_frame_ms avg/max: 16.25 / 29.50"))
	assert_true(label.text.contains("server_tick_ms avg/max: 3.00 / 8.00"))
	assert_true(label.text.contains("entities p/a/pr/pk: 2/15/6/1"))
	assert_true(label.text.contains("client packets/bytes: 30 / 4000"))
	assert_true(label.text.contains("server packets/bytes: 24 / 3500"))
	assert_true(label.text.contains("snapshot_age_ms: 125"))


func test_renders_unavailable_measurement_values_as_em_dash() -> void:
	var overlay := WorldTelemetryOverlayScene.instantiate()
	add_child_autofree(overlay)
	overlay.refresh_metrics({
		"measurement_status": "starting",
		"measurement_run_id": "",
		"measurement_elapsed_ms": -1,
		"measurement_client_frame_average_ms": -1,
		"measurement_client_frame_maximum_ms": -1,
		"measurement_server_tick_average_ms": -1,
		"measurement_server_tick_maximum_ms": -1,
		"measurement_server_players": -1,
		"measurement_server_asteroids": -1,
		"measurement_server_projectiles": -1,
		"measurement_server_pickups": -1,
		"measurement_client_packets": -1,
		"measurement_client_bytes": -1,
		"measurement_server_packets": -1,
		"measurement_server_bytes": -1,
		"measurement_snapshot_age_ms": -1,
	})

	var label := overlay.find_child("MetricsLabel", true, false) as Label
	assert_true(label.text.contains("status: starting"))
	assert_true(label.text.contains("run: —"))
	assert_true(label.text.contains("elapsed_s: —"))
	assert_true(label.text.contains("snapshot_age_ms: —"))
