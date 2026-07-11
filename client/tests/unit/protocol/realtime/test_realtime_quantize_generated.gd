extends GutTest

const RealtimeQuantize := preload("res://scripts/protocol/realtime/realtime_quantize.gd")

func test_generated_session_player_paths_decode() -> void:
	var decoded: Dictionary = RealtimeQuantize.decode_session_value({
		"players": {"player-1": {"respawn_cooldown": 2500.0, "spawn_x": 30.0, "spawn_y": -20.0}}
	})
	var player: Dictionary = decoded["players"]["player-1"]
	assert_eq(player["respawn_cooldown"], 2.5)
	assert_eq(player["spawn_x"], 3.0)
	assert_eq(player["spawn_y"], -2.0)

func test_generated_overlay_cooldown_paths_decode() -> void:
	var decoded: Dictionary = RealtimeQuantize.decode_overlay_value({
		"respawn_cooldown": 2000.0,
		"primary_cooldown_remaining": 1500.0,
		"secondary_cooldown_remaining": 500.0,
	})
	assert_eq(decoded["respawn_cooldown"], 2.0)
	assert_eq(decoded["primary_cooldown_remaining"], 1.5)
	assert_eq(decoded["secondary_cooldown_remaining"], 0.5)

func test_generated_ship_paths_decode_position_and_rotation() -> void:
	var decoded: Dictionary = RealtimeQuantize.decode_world_ship_record({"x": 20.0, "y": -10.0, "rotation": 500.0})
	assert_eq(decoded["x"], 2.0)
	assert_eq(decoded["y"], -1.0)
	assert_eq(decoded["rotation"], 0.5)

func test_generated_pickup_paths_decode_age_and_lifespan() -> void:
	var decoded: Dictionary = RealtimeQuantize.decode_world_pickup_record({"age_seconds": 3000.0, "lifespan_seconds": 5000.0})
	assert_eq(decoded["age_seconds"], 3.0)
	assert_eq(decoded["lifespan_seconds"], 5.0)

func test_generated_ship_death_event_paths_decode() -> void:
	var decoded: Dictionary = RealtimeQuantize.decode_event_record({
		"type": "ship_death",
		"x": 20.0,
		"y": -10.0,
		"respawn_delay": 2500.0,
	})
	assert_eq(decoded["x"], 2.0)
	assert_eq(decoded["y"], -1.0)
	assert_eq(decoded["respawn_delay"], 2.5)

func test_generated_position_decodes_integer_and_unknown_integer_is_preserved() -> void:
	var generated: Dictionary = RealtimeQuantize.decode_session_value({
		"players": {"player-1": {"spawn_x": 30}}
	})
	var unknown: Dictionary = RealtimeQuantize.decode_session_value({"unknown": 30})
	assert_eq(generated["players"]["player-1"]["spawn_x"], 3.0)
	assert_eq(unknown["unknown"], 30)

func test_unknown_float_path_is_preserved() -> void:
	var decoded: Dictionary = RealtimeQuantize.decode_session_value({"unknown": 1000.0})
	assert_eq(decoded["unknown"], 1000.0)

func test_unquantized_session_float_fields_are_preserved() -> void:
	var decoded: Dictionary = RealtimeQuantize.decode_session_value({
		"players": {"player-1": {"lives": 3.0, "score": 125.0}},
		"total_asteroids": 9.0,
	})
	assert_eq(decoded["players"]["player-1"]["lives"], 3.0)
	assert_eq(decoded["players"]["player-1"]["score"], 125.0)
	assert_eq(decoded["total_asteroids"], 9.0)
