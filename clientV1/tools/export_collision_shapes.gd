@tool
extends SceneTree

const OUTPUT_PATH := "res://../shared/collisions/collision_shapes.json"
const BULLET_SCENE := "res://scenes/bullet.tscn"
const ASTEROID_SCENE := "res://scenes/asteroid.tscn"

func _init() -> void:
	var data := {
		"bullet": _export_bullet(),
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
		quit(1)
		return {}

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
			exported.append({
				"name": child.name,
				"type": "polygon",
				"points": _export_points(child.polygon),
			})

	root.queue_free()
	return exported


func _export_points(points: PackedVector2Array) -> Array:
	var exported := []
	for point in points:
		exported.append([point.x, point.y])

	return exported
