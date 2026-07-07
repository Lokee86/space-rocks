extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const AsteroidSync := preload("res://scripts/world/asteroid_sync.gd")

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


func _new_asteroid_sync() -> AsteroidSync:
	var asteroid_sync := AsteroidSync.new()
	var asteroids_layer := Node2D.new()
	add_child_autofree(asteroids_layer)
	asteroid_sync.configure(asteroids_layer)
	return asteroid_sync
