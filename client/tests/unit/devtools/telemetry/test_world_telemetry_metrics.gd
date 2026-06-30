extends GutTest

const WorldTelemetryMetrics = preload("res://scripts/devtools/telemetry/world_telemetry_metrics.gd")


func test_counts_players_asteroids_pickups_and_bullets_from_lane_native_dictionaries() -> void:
	var metrics := WorldTelemetryMetrics.new()

	metrics.apply_gameplay_state({
		"world": {
			"ships": {"p1": {}, "p2": {}},
			"asteroids": {"a1": {}, "a2": {}, "a3": {}},
			"pickups": {"pk1": {}},
			"bullets": {"b1": {}},
		},
	})

	assert_eq(metrics.players, 2)
	assert_eq(metrics.asteroids, 3)
	assert_eq(metrics.pickups, 1)
	assert_eq(metrics.bullets, 1)


func test_enemies_count_prefers_server_enemies_dictionary() -> void:
	var metrics := WorldTelemetryMetrics.new()

	metrics.apply_gameplay_state({
		"server_enemies": {"e1": {}, "e2": {}},
		"enemies": {"fallback_enemy": {}},
	})

	assert_eq(metrics.enemies, 2)


func test_enemies_falls_back_to_enemies_when_server_enemies_missing() -> void:
	var metrics := WorldTelemetryMetrics.new()

	metrics.apply_gameplay_state({
		"enemies": {"e1": {}, "e2": {}, "e3": {}},
	})

	assert_eq(metrics.enemies, 3)


func test_missing_or_non_dictionary_sources_result_in_zero_counts() -> void:
	var metrics := WorldTelemetryMetrics.new()

	metrics.apply_gameplay_state({
		"world": {
			"ships": ["not", "a", "dictionary"],
			"asteroids": 42,
			"pickups": null,
			"bullets": false,
		},
		"server_enemies": "invalid",
		"enemies": false,
	})

	assert_eq(metrics.players, 0)
	assert_eq(metrics.asteroids, 0)
	assert_eq(metrics.pickups, 0)
	assert_eq(metrics.bullets, 0)
	assert_eq(metrics.enemies, 0)


func test_packet_interval_is_unavailable_after_first_packet() -> void:
	var metrics := WorldTelemetryMetrics.new()

	metrics.apply_gameplay_state({})

	assert_eq(metrics.packet_interval_ms, -1)


func test_packet_interval_becomes_available_after_second_packet() -> void:
	var metrics := WorldTelemetryMetrics.new()

	metrics.apply_gameplay_state({})
	metrics.apply_gameplay_state({})

	assert_true(metrics.packet_interval_ms >= 0)


func test_jitter_is_unavailable_until_two_intervals_exist() -> void:
	var metrics := WorldTelemetryMetrics.new()

	metrics.apply_gameplay_state({})
	metrics.apply_gameplay_state({})
	assert_eq(metrics.jitter_ms, -1)

	metrics.apply_gameplay_state({})
	assert_true(metrics.jitter_ms >= 0)


func test_reset_clears_timing_state() -> void:
	var metrics := WorldTelemetryMetrics.new()

	metrics.apply_gameplay_state({"server_sent_msec": 12345})
	metrics.apply_gameplay_state({"server_sent_msec": 12346})
	metrics.apply_gameplay_state({"server_sent_msec": 12347})
	metrics.reset()

	assert_eq(metrics.server_sent_msec, -1)
	assert_eq(metrics.latest_packet_arrival_msec, -1)
	assert_eq(metrics.previous_packet_arrival_msec, -1)
	assert_eq(metrics.packet_interval_ms, -1)
	assert_eq(metrics.previous_packet_interval_ms, -1)
	assert_eq(metrics.jitter_ms, -1)


func test_server_sent_msec_is_preserved_when_present() -> void:
	var metrics := WorldTelemetryMetrics.new()
	var offset_ms := 1000
	var sent_msec := Time.get_ticks_msec() + offset_ms - 100

	metrics.set_network_metrics({"server_clock_offset_ms": offset_ms})
	metrics.apply_gameplay_state({"server_sent_msec": sent_msec})
	var telemetry := metrics.snapshot()

	assert_eq(metrics.server_sent_msec, sent_msec)
	assert_eq(telemetry["server_sent_msec"], sent_msec)


func test_missing_server_sent_msec_makes_packet_age_unavailable() -> void:
	var metrics := WorldTelemetryMetrics.new()
	var offset_ms := 1000
	var sent_msec := Time.get_ticks_msec() + offset_ms - 100

	metrics.set_network_metrics({"server_clock_offset_ms": offset_ms})
	metrics.apply_gameplay_state({"server_sent_msec": sent_msec})
	metrics.apply_gameplay_state({})
	var telemetry := metrics.snapshot()

	assert_eq(metrics.server_sent_msec, -1)
	assert_eq(telemetry["server_sent_msec"], -1)
	assert_eq(telemetry["packet_age_ms"], -1)
