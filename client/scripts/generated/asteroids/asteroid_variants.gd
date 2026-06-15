class_name AsteroidVariants
extends RefCounted

const VARIANTS := [
	{
		"id": "asteroid_1",
		"index": 0,
		"texture": "res://assets/asteroids/asteroid1.png",
		"collision_shape": "asteroid:0",
		"stats_profile": "standard",
		"drop_table": "basicasteroids",
		"timed_spawn_weight": 1.0,
		"fragment_spawn_weight": 1.0,
		"debug_spawn_weight": 1.0,
	},
	{
		"id": "asteroid_2",
		"index": 1,
		"texture": "res://assets/asteroids/asteroid2.png",
		"collision_shape": "asteroid:0",
		"stats_profile": "standard",
		"drop_table": "basicasteroids",
		"timed_spawn_weight": 1.0,
		"fragment_spawn_weight": 1.0,
		"debug_spawn_weight": 1.0,
	},
	{
		"id": "asteroid_3",
		"index": 2,
		"texture": "res://assets/asteroids/asteroid3.png",
		"collision_shape": "asteroid:0",
		"stats_profile": "standard",
		"drop_table": "basicasteroids",
		"timed_spawn_weight": 1.0,
		"fragment_spawn_weight": 1.0,
		"debug_spawn_weight": 1.0,
	},
	{
		"id": "asteroid_4",
		"index": 3,
		"texture": "res://assets/asteroids/asteroid4.png",
		"collision_shape": "asteroid:0",
		"stats_profile": "standard",
		"drop_table": "basicasteroids",
		"timed_spawn_weight": 1.0,
		"fragment_spawn_weight": 1.0,
		"debug_spawn_weight": 1.0,
	},
	{
		"id": "asteroid_5",
		"index": 4,
		"texture": "res://assets/asteroids/asteroid5.png",
		"collision_shape": "asteroid:0",
		"stats_profile": "standard",
		"drop_table": "basicasteroids",
		"timed_spawn_weight": 1.0,
		"fragment_spawn_weight": 1.0,
		"debug_spawn_weight": 1.0,
	},
	{
		"id": "asteroid_6",
		"index": 5,
		"texture": "res://assets/asteroids/asteroid6.png",
		"collision_shape": "asteroid:0",
		"stats_profile": "standard",
		"drop_table": "basicasteroids",
		"timed_spawn_weight": 1.0,
		"fragment_spawn_weight": 1.0,
		"debug_spawn_weight": 1.0,
	},
	{
		"id": "asteroid_7",
		"index": 6,
		"texture": "res://assets/asteroids/asteroid7.png",
		"collision_shape": "asteroid:0",
		"stats_profile": "standard",
		"drop_table": "basicasteroids",
		"timed_spawn_weight": 1.0,
		"fragment_spawn_weight": 1.0,
		"debug_spawn_weight": 1.0,
	},
	{
		"id": "asteroid_8",
		"index": 7,
		"texture": "res://assets/asteroids/asteroid8.png",
		"collision_shape": "asteroid:0",
		"stats_profile": "standard",
		"drop_table": "basicasteroids",
		"timed_spawn_weight": 1.0,
		"fragment_spawn_weight": 1.0,
		"debug_spawn_weight": 1.0,
	},
]

static func count() -> int:
	return VARIANTS.size()

static func texture_path_for_index(index: int) -> String:
	return String(_variant_for_index(index).get("texture", ""))

static func collision_shape_for_index(index: int) -> String:
	return String(_variant_for_index(index).get("collision_shape", ""))

static func timed_spawn_weight_for_index(index: int) -> float:
	return float(_variant_for_index(index).get("timed_spawn_weight", 0.0))

static func fragment_spawn_weight_for_index(index: int) -> float:
	return float(_variant_for_index(index).get("fragment_spawn_weight", 0.0))

static func debug_spawn_weight_for_index(index: int) -> float:
	return float(_variant_for_index(index).get("debug_spawn_weight", 0.0))

static func _variant_for_index(index: int) -> Dictionary:
	if VARIANTS.is_empty():
		return {}

	var wrapped_index := index % VARIANTS.size()
	if wrapped_index < 0:
		wrapped_index += VARIANTS.size()

	return VARIANTS[wrapped_index]
