@tool
extends SceneTree

const OUTPUT_PATH := "res://../shared/collisions/collision_shapes.json"
const BULLET_SCENE := "res://scenes/bullet.tscn"
const PLAYER_SCENE := "res://scenes/player.tscn"
const ASTEROID_SCENE := "res://scenes/asteroid.tscn"


func _init() -> void:
	var data := {
		"bullet": _export_bullet(),
		"ship": _export_player(),
		"asteroids": _export_asteroids(),
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
		push_error("Unsupported bullet shape: %s" % shape.get_class())
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