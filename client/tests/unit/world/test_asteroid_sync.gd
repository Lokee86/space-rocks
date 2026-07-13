extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const AsteroidSync := preload("res://scripts/world/asteroid_sync.gd")
const AsteroidPresentation := preload("res://scripts/entities/asteroid.gd")
const AsteroidScene := preload("res://scenes/asteroid.tscn")

func test_asteroid_scene_root_satisfies_asteroid_presentation_contract() -> void:
	var node := AsteroidScene.instantiate()
	add_child_autofree(node)
	assert_true(node is Node2D)
	assert_true(node is AsteroidPresentation)

func test_unknown_hot_asteroid_update_does_not_create_node() -> void:
	var asteroid_sync := _new_asteroid_sync()
	var asteroid_id := "asteroid-unknown"

	asteroid_sync.apply_asteroid(
		asteroid_id,
		{
			Packets.FIELD_X: 320.0,
			Packets.FIELD_Y: 340.0,
			Packets.FIELD_ROTATION: 0.75,
		},
		Vector2.ZERO,
		Vector2.ZERO,
		false
	)

	assert_false(asteroid_sync.asteroid_nodes.has(asteroid_id))
	assert_false(asteroid_sync.initialized_asteroids.has(asteroid_id))
	assert_false(asteroid_sync.get("target_asteroid_positions").has(asteroid_id))
	assert_false(asteroid_sync.get("asteroid_server_positions").has(asteroid_id))
	assert_false(asteroid_sync.get("asteroid_visual_positions").has(asteroid_id))


func test_hot_asteroid_update_after_lifecycle_delete_is_ignored() -> void:
	var asteroid_sync := _new_asteroid_sync()
	var asteroid_id := "asteroid-1"

	asteroid_sync.apply_asteroid(
		asteroid_id,
		{
			Packets.FIELD_X: 10.0,
			Packets.FIELD_Y: 20.0,
			Packets.FIELD_ROTATION: 0.25,
			Packets.FIELD_SCALE: 1.0,
			Packets.FIELD_VARIANT: 2,
		},
		Vector2.ZERO,
		Vector2.ZERO
	)
	asteroid_sync.remove_asteroid(asteroid_id)
	asteroid_sync.apply_asteroid(
		asteroid_id,
		{
			Packets.FIELD_X: 320.0,
			Packets.FIELD_Y: 340.0,
			Packets.FIELD_ROTATION: 0.75,
		},
		Vector2.ZERO,
		Vector2.ZERO,
		false
	)

	assert_false(asteroid_sync.asteroid_nodes.has(asteroid_id))
	assert_true(asteroid_sync.is_deleted(asteroid_id))
	assert_false(asteroid_sync.initialized_asteroids.has(asteroid_id))
	assert_false(asteroid_sync.get("target_asteroid_positions").has(asteroid_id))


func test_deleted_asteroid_tombstones_are_bounded_and_recreated() -> void:
	var asteroid_sync := _new_asteroid_sync()
	for index in range(AsteroidSync.DELETED_ASTEROID_ID_CAP + 1):
		asteroid_sync.remove_asteroid("asteroid-%d" % index)

	assert_eq(asteroid_sync.deleted_asteroid_ids.size(), AsteroidSync.DELETED_ASTEROID_ID_CAP)
	assert_false(asteroid_sync.deleted_asteroid_ids.has("asteroid-0"))
	assert_true(asteroid_sync.deleted_asteroid_ids.has("asteroid-%d" % AsteroidSync.DELETED_ASTEROID_ID_CAP))

	var retained_id := "asteroid-%d" % AsteroidSync.DELETED_ASTEROID_ID_CAP
	var state := {
		Packets.FIELD_X: 10.0,
		Packets.FIELD_Y: 20.0,
		Packets.FIELD_ROTATION: 0.5,
		Packets.FIELD_SCALE: 1.0,
		Packets.FIELD_VARIANT: 2,
	}
	asteroid_sync.apply_asteroid(retained_id, state, Vector2.ZERO, Vector2.ZERO)

	assert_false(asteroid_sync.deleted_asteroid_ids.has(retained_id))
	assert_false(asteroid_sync._deleted_asteroid_id_order.has(retained_id))
	assert_true(asteroid_sync.asteroid_nodes.has(retained_id))

	asteroid_sync.remove_asteroid(retained_id)
	assert_true(asteroid_sync.deleted_asteroid_ids.has(retained_id))
	assert_true(asteroid_sync._deleted_asteroid_id_order.has(retained_id))
	assert_eq(asteroid_sync.deleted_asteroid_ids.size(), AsteroidSync.DELETED_ASTEROID_ID_CAP)

	asteroid_sync.reset()
	assert_true(asteroid_sync.deleted_asteroid_ids.is_empty())
	assert_true(asteroid_sync._deleted_asteroid_id_order.is_empty())


func _new_asteroid_sync() -> AsteroidSync:
	var asteroid_sync := AsteroidSync.new()
	var asteroids_layer := Node2D.new()
	add_child_autofree(asteroids_layer)
	asteroid_sync.configure(asteroids_layer)
	return asteroid_sync
