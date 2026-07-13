extends RefCounted
class_name PickupPresentationCatalog

const PICKUP_CLASS_POWERUP := "powerup"
const PICKUP_CLASS_WEAPON := "weapon"

const POWERUP_PICKUP_SCENE := preload("res://scenes/pickups/powerup_pickup.tscn")
const WEAPON_PICKUP_SCENE := preload("res://scenes/pickups/weapon_pickup.tscn")
const PickupPresentation = preload("res://scripts/entities/pickup.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")


static func scene_for_class(pickup_class: String) -> PackedScene:
	if pickup_class == PICKUP_CLASS_POWERUP:
		return POWERUP_PICKUP_SCENE
	if pickup_class == PICKUP_CLASS_WEAPON:
		return WEAPON_PICKUP_SCENE

	return null


static func available_pickup_types() -> Array[String]:
	var pickup_types: Array[String] = []

	_collect_pickup_types_from_scene(POWERUP_PICKUP_SCENE, PICKUP_CLASS_POWERUP, pickup_types)
	_collect_pickup_types_from_scene(WEAPON_PICKUP_SCENE, PICKUP_CLASS_WEAPON, pickup_types)

	pickup_types.sort()
	return pickup_types


static func _collect_pickup_types_from_scene(
	scene: PackedScene,
	pickup_class: String,
	pickup_types: Array[String]
) -> void:
	if scene == null:
		return

	var scene_root: Node = scene.instantiate()
	if scene_root == null:
		return

	var pickup_root := scene_root as PickupPresentation
	if pickup_root == null:
		var actual_script: Script = scene_root.get_script()
		var actual_script_path := ""
		if actual_script is Script:
			actual_script_path = actual_script.resource_path
		ClientLogger.event(
			ClientLogger.CATEGORY_WORLD_SYNC,
			ClientLogger.LEVEL_ERROR,
			"pickup_presentation_contract_violation",
			"Pickup catalog scene root does not satisfy its presentation contract",
			{
				"pickup_class": pickup_class,
				"actual_class": scene_root.get_class(),
				"actual_script": actual_script_path,
			}
		)
		scene_root.queue_free()
		return

	var badge := pickup_root.get_node_or_null("Badge")
	if badge == null:
		scene_root.queue_free()
		return

	for child in badge.get_children():
		if child is CanvasItem:
			var pickup_type := str(child.name)
			if pickup_type not in pickup_types:
				pickup_types.append(pickup_type)

	scene_root.queue_free()
