extends RefCounted
class_name PickupPresentationCatalog

const PICKUP_CLASS_POWERUP := "powerup"
const PICKUP_CLASS_WEAPON := "weapon"

const POWERUP_PICKUP_SCENE := preload("res://scenes/pickups/powerup_pickup.tscn")
const WEAPON_PICKUP_SCENE := preload("res://scenes/pickups/weapon_pickup.tscn")


static func scene_for_class(pickup_class: String) -> PackedScene:
	if pickup_class == PICKUP_CLASS_POWERUP:
		return POWERUP_PICKUP_SCENE
	if pickup_class == PICKUP_CLASS_WEAPON:
		return WEAPON_PICKUP_SCENE

	return null


static func available_pickup_types() -> Array[String]:
	var pickup_types: Array[String] = []

	_collect_pickup_types_from_scene(POWERUP_PICKUP_SCENE, pickup_types)
	_collect_pickup_types_from_scene(WEAPON_PICKUP_SCENE, pickup_types)

	pickup_types.sort()
	return pickup_types


static func _collect_pickup_types_from_scene(scene: PackedScene, pickup_types: Array[String]) -> void:
	if scene == null:
		return

	var scene_root := scene.instantiate()
	if scene_root == null:
		return

	var badge := scene_root.get_node_or_null("Badge")
	if badge == null:
		scene_root.queue_free()
		return

	for child in badge.get_children():
		if child is CanvasItem:
			var pickup_type := str(child.name)
			if pickup_type not in pickup_types:
				pickup_types.append(pickup_type)

	scene_root.queue_free()
