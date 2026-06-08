@tool
extends SceneTree

const OUTPUT_PATH := "res://../shared/collisions/collision_shapes.json"
const BULLET_SCENE := "res://scenes/bullet.tscn"
const PLAYER_SCENE := "res://scenes/player.tscn"
const ASTEROID_SCENE := "res://scenes/asteroid.tscn"
const PICKUP_POWERUP_SCENE := "res://scenes/pickups/powerup_pickup.tscn"
const PICKUP_WEAPON_SCENE := "res://scenes/pickups/weapon_pickup.tscn"


func _init() -> void:
	var data := {
		"bullet": _export_bullet(),
		"ship": _export_player(),
		"asteroids": _export_asteroids(),
		"pickups": {
			"powerup": _export_pickup_shape(PICKUP_POWERUP_SCENE),
			"weapon": _export_pickup_shape(PICKUP_WEAPON_SCENE),
		},
	}

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

	var scene_root := scene.instantiate()
	var collision_shape := scene_root.get_node("CollisionShape2D") as CollisionShape2D
	if collision_shape == null:
		push_error("Missing CollisionShape2D in %s" % BULLET_SCENE)
		scene_root.queue_free()
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
		scene_root.queue_free()
		quit(1)
		return {}

	scene_root.queue_free()
	return exported


func _export_player() -> Dictionary:
	var scene := load(PLAYER_SCENE) as PackedScene
	if scene == null:
		push_error("Failed to load %s" % PLAYER_SCENE)
		quit(1)
		return {}

	var scene_root := scene.instantiate()
	var collision_polygon := scene_root.get_node("CollisionPolygon2D") as CollisionPolygon2D
	if collision_polygon == null:
		push_error("Missing CollisionPolygon2D in %s" % PLAYER_SCENE)
		scene_root.queue_free()
		quit(1)
		return {}

	var exported := _export_polygon(collision_polygon)
	scene_root.queue_free()
	return exported


func _export_asteroids() -> Array:
	var scene := load(ASTEROID_SCENE) as PackedScene
	if scene == null:
		push_error("Failed to load %s" % ASTEROID_SCENE)
		quit(1)
		return []

	var scene_root := scene.instantiate()
	var variants := scene_root.get_node("CollisionVariants")
	var exported := []

	for child in variants.get_children():
		if child is CollisionPolygon2D:
			exported.append(_export_polygon(child))

	scene_root.queue_free()
	return exported


func _export_pickup_shape(scene_path: String) -> Dictionary:
	var pickup_scene := _pickup_scene(scene_path)
	if pickup_scene.is_empty():
		return {}

	var scene_root: Node = pickup_scene[0]
	var collision_shape: CollisionShape2D = pickup_scene[1]
	var shape := collision_shape.shape
	if !(shape is CircleShape2D):
		var shape_class := "<null>" if shape == null else shape.get_class()
		push_error("Unsupported pickup shape in %s: %s" % [scene_path, shape_class])
		scene_root.queue_free()
		quit(1)
		return {}

	var exported := {
		"name": collision_shape.name,
		"type": "circle",
		"radius": shape.radius,
		"offset": [collision_shape.position.x, collision_shape.position.y],
	}

	scene_root.queue_free()
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


func _pickup_scene(scene_path: String) -> Array:
	var scene := load(scene_path) as PackedScene
	if scene == null:
		push_error("Failed to load %s" % scene_path)
		quit(1)
		return []

	var scene_root := scene.instantiate()
	var collision_shape := scene_root.get_node("CollisionShape2D") as CollisionShape2D
	if collision_shape == null:
		push_error("Missing CollisionShape2D in %s" % scene_path)
		scene_root.queue_free()
		quit(1)
		return []

	return [scene_root, collision_shape]
