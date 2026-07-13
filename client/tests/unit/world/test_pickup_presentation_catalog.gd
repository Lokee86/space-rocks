extends GutTest

const PickupPresentationCatalog := preload("res://scripts/world/pickups/pickup_presentation_catalog.gd")
const PickupPresentation := preload("res://scripts/entities/pickup.gd")


func test_scene_for_class_returns_powerup_scene() -> void:
	var scene: PackedScene = PickupPresentationCatalog.scene_for_class("powerup")

	assert_not_null(scene)
	var root := scene.instantiate()
	add_child_autofree(root)
	assert_true(root is PickupPresentation)


func test_scene_for_class_returns_weapon_scene() -> void:
	var scene: PackedScene = PickupPresentationCatalog.scene_for_class("weapon")

	assert_not_null(scene)
	var root := scene.instantiate()
	add_child_autofree(root)
	assert_true(root is PickupPresentation)


func test_scene_for_class_returns_null_for_unknown_class() -> void:
	var scene = PickupPresentationCatalog.scene_for_class("unknown")

	assert_null(scene)


func test_available_pickup_types_includes_expected_entries() -> void:
	var pickup_types: Array[String] = PickupPresentationCatalog.available_pickup_types()

	assert_true(pickup_types.has("1_up"))
	assert_true(pickup_types.has("torpedo"))
