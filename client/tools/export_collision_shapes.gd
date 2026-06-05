@tool
extends SceneTree

const OUTPUT_PATH := "res://../shared/collisions/collision_shapes.json"
const BULLET_SCENE := "res://scenes/bullet.tscn"
const PLAYER_SCENE := "res://scenes/player.tscn"
const ASTEROID_SCENE := "res://scenes/asteroid.tscn"
const PICKUP_ONE_UP_SCENE := "res://scenes/pickups/1_up.tscn"
const PICKUP_TOML_PATH := "res://../shared/constants/server_entities.toml"
const PICKUP_SECTION := "[constants.server.pickups]"
const PICKUP_RADIUS_KEY := "pickup_one_up_collision_radius"


func _init() -> void:
	var data := {
		"bullet": _export_bullet(),
		"ship": _export_player(),
		"asteroids": _export_asteroids(),
		"pickups": {
			"1_up": _export_pickup_one_up(),
		},
	}

	_export_pickup_one_up_collision_radius()

	var output := FileAccess.open(OUTPUT_PATH, FileAccess.WRITE)
	if output == null:
		push_error("Failed to open %s: %s" % [OUTPUT_PATH, FileAccess.get_open_error()])
		quit(1)
		return

	output.store_string(JSON.stringify(data, "\t") + "\n")
	print("Exported collision shapes to %s" % OUTPUT_PATH)
	quit()


func _export_bullet() -> Dictionary:
	var scene := load(BULLET_SCENE) as PackedScene
	if scene == null:
		push_error("Failed to load %s" % BULLET_SCENE)
		quit(1)
		return {}

	var root := scene.instantiate()
	var collision_shape := root.get_node("CollisionShape2D") as CollisionShape2D
	if collision_shape == null:
		push_error("Missing CollisionShape2D in %s" % BULLET_SCENE)
		root.queue_free()
		quit(1)
		return {}

	var shape := collision_shape.shape
	var exported := {
		"name": collision_shape.name,
	}

	if shape is CircleShape2D:
		exported["type"] = "circle"
		exported["radius"] = shape.radius
	elif shape is CapsuleShape2D:
		exported["type"] = "capsule"
		exported["radius"] = shape.radius
		exported["height"] = shape.height
	elif shape is RectangleShape2D:
		exported["type"] = "rectangle"
		exported["size"] = [shape.size.x, shape.size.y]
	else:
		var shape_class := "<null>" if shape == null else shape.get_class()
		push_error("Unsupported bullet shape: %s" % shape_class)
		root.queue_free()
		quit(1)
		return {}

	root.queue_free()
	return exported


func _export_player() -> Dictionary:
	var scene := load(PLAYER_SCENE) as PackedScene
	if scene == null:
		push_error("Failed to load %s" % PLAYER_SCENE)
		quit(1)
		return {}

	var root := scene.instantiate()
	var collision_polygon := root.get_node("CollisionPolygon2D") as CollisionPolygon2D
	if collision_polygon == null:
		push_error("Missing CollisionPolygon2D in %s" % PLAYER_SCENE)
		root.queue_free()
		quit(1)
		return {}

	var exported := _export_polygon(collision_polygon)
	root.queue_free()
	return exported


func _export_asteroids() -> Array:
	var scene := load(ASTEROID_SCENE) as PackedScene
	if scene == null:
		push_error("Failed to load %s" % ASTEROID_SCENE)
		quit(1)
		return []

	var root := scene.instantiate()
	var variants := root.get_node("CollisionVariants")
	var exported := []

	for child in variants.get_children():
		if child is CollisionPolygon2D:
			exported.append(_export_polygon(child))

	root.queue_free()
	return exported


func _export_pickup_one_up() -> Dictionary:
	var pickup_scene := _pickup_one_up_scene()
	if pickup_scene.is_empty():
		return {}

	var root: Node = pickup_scene[0]
	var collision_shape: CollisionShape2D = pickup_scene[1]
	var shape := collision_shape.shape
	if !(shape is CircleShape2D):
		var shape_class := "<null>" if shape == null else shape.get_class()
		push_error("Unsupported pickup shape in %s: %s" % [PICKUP_ONE_UP_SCENE, shape_class])
		root.queue_free()
		quit(1)
		return {}

	var exported := {
		"name": collision_shape.name,
		"type": "circle",
		"radius": shape.radius,
		"offset": [collision_shape.position.x, collision_shape.position.y],
	}

	root.queue_free()
	return exported


func _export_polygon(collision_polygon: CollisionPolygon2D) -> Dictionary:
	return {
		"name": collision_polygon.name,
		"type": "polygon",
		"points": _export_points(collision_polygon.polygon),
	}


func _export_points(points: PackedVector2Array) -> Array:
	var exported := []
	for point in points:
		exported.append([point.x, point.y])

	return exported


func _export_pickup_one_up_collision_radius() -> float:
	var pickup_scene := _pickup_one_up_scene()
	if pickup_scene.is_empty():
		return 0.0

	var root: Node = pickup_scene[0]
	var collision_shape: CollisionShape2D = pickup_scene[1]
	var circle_shape := collision_shape.shape as CircleShape2D
	if circle_shape == null:
		var shape := collision_shape.shape
		var shape_class := "<null>" if shape == null else shape.get_class()
		push_error("Unsupported pickup shape in %s: %s" % [PICKUP_ONE_UP_SCENE, shape_class])
		root.queue_free()
		quit(1)
		return 0.0

	var radius: float = circle_shape.radius
	root.queue_free()

	_update_pickup_radius(radius)
	return radius


func _pickup_one_up_scene() -> Array:
	var scene := load(PICKUP_ONE_UP_SCENE) as PackedScene
	if scene == null:
		push_error("Failed to load %s" % PICKUP_ONE_UP_SCENE)
		quit(1)
		return []

	var root := scene.instantiate()
	var collision_shape := root.get_node("CollisionShape2D") as CollisionShape2D
	if collision_shape == null:
		push_error("Missing CollisionShape2D in %s" % PICKUP_ONE_UP_SCENE)
		root.queue_free()
		quit(1)
		return []

	return [root, collision_shape]


func _update_pickup_radius(radius: float) -> void:
	var file := FileAccess.open(PICKUP_TOML_PATH, FileAccess.READ)
	if file == null:
		push_error("Failed to read %s: %s" % [PICKUP_TOML_PATH, FileAccess.get_open_error()])
		quit(1)
		return

	var text := file.get_as_text()
	file.close()

	if not text.contains(PICKUP_SECTION):
		push_error("Missing TOML section %s in %s" % [PICKUP_SECTION, PICKUP_TOML_PATH])
		quit(1)
		return

	var section_start := text.find(PICKUP_SECTION)
	var next_section_start := text.find("\n[", section_start + PICKUP_SECTION.length())
	if next_section_start == -1:
		next_section_start = text.length()

	var section_text := text.substr(section_start, next_section_start - section_start)
	var prefix := PICKUP_RADIUS_KEY + " = "
	var key_offset := section_text.find(prefix)
	if key_offset == -1:
		push_error("Missing TOML key %s in %s" % [PICKUP_RADIUS_KEY, PICKUP_TOML_PATH])
		quit(1)
		return

	var start := section_start + key_offset
	var line_end := text.find("\n", start)
	if line_end == -1:
		line_end = text.length()

	var updated_text := text.substr(0, start) + prefix + str(radius) + text.substr(line_end)

	var output := FileAccess.open(PICKUP_TOML_PATH, FileAccess.WRITE)
	if output == null:
		push_error("Failed to write %s: %s" % [PICKUP_TOML_PATH, FileAccess.get_open_error()])
		quit(1)
		return

	output.store_string(updated_text)
	output.close()
	print("Updated %s to %s in %s" % [PICKUP_RADIUS_KEY, str(radius), PICKUP_TOML_PATH])
