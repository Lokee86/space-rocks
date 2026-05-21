extends GutTest

const Packets := preload("res://scripts/networking/packets.gd")
const WorldStateFixture := preload("res://tests/fixtures/world_state_fixture.gd")
const WorldSyncScript := preload("res://scripts/networking/world_sync.gd")
const PlayerScene := preload("res://scenes/player.tscn")

var game_owner: Node2D
var local_player: Player
var bullets_layer: Node2D
var asteroids_layer: Node2D
var world_sync: WorldSync


func before_each() -> void:
	game_owner = Node2D.new()
	add_child(game_owner)

	local_player = PlayerScene.instantiate()
	bullets_layer = Node2D.new()
	asteroids_layer = Node2D.new()

	game_owner.add_child(local_player)
	game_owner.add_child(bullets_layer)
	game_owner.add_child(asteroids_layer)

	world_sync = WorldSyncScript.new()
	world_sync.configure(game_owner, local_player, bullets_layer, asteroids_layer)


func after_each() -> void:
	world_sync = null
	if game_owner != null:
		game_owner.free()
		game_owner = null


func test_apply_state_creates_player_nodes() -> void:
	_apply_fixture_state()

	assert_true(world_sync.player_nodes.has(WorldStateFixture.LOCAL_PLAYER_ID))
	assert_true(world_sync.player_nodes.has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_eq(world_sync.player_nodes[WorldStateFixture.LOCAL_PLAYER_ID], local_player)
	assert_eq(
		world_sync.player_nodes[WorldStateFixture.LOCAL_PLAYER_ID].position,
		Vector2(100.0, 120.0)
	)
	assert_eq(
		world_sync.player_nodes[WorldStateFixture.REMOTE_PLAYER_ID].position,
		Vector2(220.0, 240.0)
	)
	assert_eq(
		world_sync.player_nodes[WorldStateFixture.REMOTE_PLAYER_ID].get_parent(),
		game_owner
	)


func test_apply_state_creates_asteroid_nodes() -> void:
	_apply_fixture_state()

	assert_true(world_sync.asteroid_nodes.has(WorldStateFixture.ASTEROID_ID))
	assert_eq(
		world_sync.asteroid_nodes[WorldStateFixture.ASTEROID_ID].get_parent(),
		asteroids_layer
	)
	assert_eq(
		world_sync.asteroid_nodes[WorldStateFixture.ASTEROID_ID].global_position,
		Vector2(320.0, 340.0)
	)


func test_apply_state_creates_bullet_nodes() -> void:
	_apply_fixture_state()

	assert_true(world_sync.bullet_nodes.has(WorldStateFixture.BULLET_ID))
	assert_eq(
		world_sync.bullet_nodes[WorldStateFixture.BULLET_ID].get_parent(),
		bullets_layer
	)
	assert_eq(
		world_sync.bullet_nodes[WorldStateFixture.BULLET_ID].global_position,
		Vector2(420.0, 440.0)
	)


func test_apply_state_reuses_existing_entity_nodes() -> void:
	_apply_fixture_state()
	var local_node = world_sync.player_nodes[WorldStateFixture.LOCAL_PLAYER_ID]
	var remote_node = world_sync.player_nodes[WorldStateFixture.REMOTE_PLAYER_ID]
	var asteroid_node = world_sync.asteroid_nodes[WorldStateFixture.ASTEROID_ID]
	var bullet_node = world_sync.bullet_nodes[WorldStateFixture.BULLET_ID]
	var owner_child_count := game_owner.get_child_count()
	var asteroid_child_count := asteroids_layer.get_child_count()
	var bullet_child_count := bullets_layer.get_child_count()

	_apply_state(_updated_state())

	assert_eq(world_sync.player_nodes.size(), 2)
	assert_eq(world_sync.asteroid_nodes.size(), 1)
	assert_eq(world_sync.bullet_nodes.size(), 1)
	assert_eq(world_sync.player_nodes[WorldStateFixture.LOCAL_PLAYER_ID], local_node)
	assert_eq(world_sync.player_nodes[WorldStateFixture.REMOTE_PLAYER_ID], remote_node)
	assert_eq(world_sync.asteroid_nodes[WorldStateFixture.ASTEROID_ID], asteroid_node)
	assert_eq(world_sync.bullet_nodes[WorldStateFixture.BULLET_ID], bullet_node)
	assert_eq(game_owner.get_child_count(), owner_child_count)
	assert_eq(asteroids_layer.get_child_count(), asteroid_child_count)
	assert_eq(bullets_layer.get_child_count(), bullet_child_count)


func test_apply_state_updates_existing_entity_targets() -> void:
	_apply_fixture_state()
	_apply_state(_updated_state())

	assert_eq(
		world_sync.target_player_positions[WorldStateFixture.LOCAL_PLAYER_ID],
		Vector2(150.0, 170.0)
	)
	assert_eq(world_sync.target_player_rotations[WorldStateFixture.LOCAL_PLAYER_ID], 0.5)
	assert_eq(
		world_sync.target_player_positions[WorldStateFixture.REMOTE_PLAYER_ID],
		Vector2(260.0, 280.0)
	)
	assert_eq(world_sync.target_player_rotations[WorldStateFixture.REMOTE_PLAYER_ID], 1.75)
	assert_eq(
		world_sync.target_asteroid_positions[WorldStateFixture.ASTEROID_ID],
		Vector2(360.0, 380.0)
	)
	assert_eq(
		world_sync.target_bullet_positions[WorldStateFixture.BULLET_ID],
		Vector2(460.0, 480.0)
	)
	assert_eq(world_sync.target_bullet_rotations[WorldStateFixture.BULLET_ID], 1.25)


func test_interpolate_moves_existing_entities_toward_updated_state() -> void:
	_apply_fixture_state()
	_apply_state(_updated_state())
	world_sync.interpolate(999.0)

	assert_eq(
		world_sync.player_nodes[WorldStateFixture.LOCAL_PLAYER_ID].position,
		Vector2(150.0, 170.0)
	)
	assert_eq(world_sync.player_nodes[WorldStateFixture.LOCAL_PLAYER_ID].rotation, 0.5)
	assert_eq(
		world_sync.player_nodes[WorldStateFixture.REMOTE_PLAYER_ID].position,
		Vector2(260.0, 280.0)
	)
	assert_eq(world_sync.player_nodes[WorldStateFixture.REMOTE_PLAYER_ID].rotation, 1.75)
	assert_eq(
		world_sync.asteroid_nodes[WorldStateFixture.ASTEROID_ID].global_position,
		Vector2(360.0, 380.0)
	)
	assert_eq(
		world_sync.bullet_nodes[WorldStateFixture.BULLET_ID].global_position,
		Vector2(460.0, 480.0)
	)
	assert_eq(world_sync.bullet_nodes[WorldStateFixture.BULLET_ID].rotation, 1.25)


func test_apply_state_removes_stale_remote_player_node() -> void:
	_apply_fixture_state()
	var remote_node = world_sync.player_nodes[WorldStateFixture.REMOTE_PLAYER_ID]

	_apply_state(_state_without_remote_player())

	assert_false(world_sync.player_nodes.has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_false(world_sync.initialized_players.has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_false(world_sync.target_player_positions.has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_false(world_sync.target_player_rotations.has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_true(remote_node.is_queued_for_deletion())


func test_apply_state_removes_stale_asteroid_node() -> void:
	_apply_fixture_state()
	var asteroid_node = world_sync.asteroid_nodes[WorldStateFixture.ASTEROID_ID]

	_apply_state(_state_without_asteroid())

	assert_false(world_sync.asteroid_nodes.has(WorldStateFixture.ASTEROID_ID))
	assert_false(world_sync.initialized_asteroids.has(WorldStateFixture.ASTEROID_ID))
	assert_false(world_sync.target_asteroid_positions.has(WorldStateFixture.ASTEROID_ID))
	assert_true(asteroid_node.is_queued_for_deletion())


func test_apply_state_removes_stale_bullet_node() -> void:
	_apply_fixture_state()
	var bullet_node = world_sync.bullet_nodes[WorldStateFixture.BULLET_ID]

	_apply_state(_state_without_bullet())

	assert_false(world_sync.bullet_nodes.has(WorldStateFixture.BULLET_ID))
	assert_false(world_sync.initialized_bullets.has(WorldStateFixture.BULLET_ID))
	assert_false(world_sync.target_bullet_positions.has(WorldStateFixture.BULLET_ID))
	assert_false(world_sync.target_bullet_rotations.has(WorldStateFixture.BULLET_ID))
	assert_true(bullet_node.is_queued_for_deletion())


func test_apply_state_missing_asteroid_scale_warns_once_and_does_not_crash() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_ASTEROIDS] = {
		WorldStateFixture.ASTEROID_ID: _asteroid_state_without_scale(),
	}

	_apply_state(state)
	_apply_state(state)

	assert_true(world_sync.asteroid_nodes.has(WorldStateFixture.ASTEROID_ID))
	assert_true(world_sync.warned_missing_asteroid_scale.has(WorldStateFixture.ASTEROID_ID))


func test_apply_state_applies_asteroid_packet_scale() -> void:
	var state := WorldStateFixture.state()

	assert_true(state[Packets.FIELD_ASTEROIDS][WorldStateFixture.ASTEROID_ID].has(Packets.FIELD_SCALE))

	_apply_state(state)

	assert_eq(
		world_sync.asteroid_nodes[WorldStateFixture.ASTEROID_ID].scale,
		Vector2.ONE * 1.25
	)

	state[Packets.FIELD_ASTEROIDS] = {
		WorldStateFixture.ASTEROID_ID: WorldStateFixture.asteroid_state(320.0, 340.0, 1, 1.75),
	}
	_apply_state(state)

	assert_eq(
		world_sync.asteroid_nodes[WorldStateFixture.ASTEROID_ID].scale,
		Vector2.ONE * 1.75
	)


func _apply_fixture_state() -> void:
	_apply_state(WorldStateFixture.state())


func _apply_state(state: Dictionary) -> void:
	world_sync.apply_state(
		state[Packets.FIELD_SELF_ID],
		state[Packets.FIELD_PLAYERS],
		state[Packets.FIELD_BULLETS],
		state[Packets.FIELD_ASTEROIDS],
		false
	)


func _updated_state() -> Dictionary:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_PLAYERS] = {
		WorldStateFixture.LOCAL_PLAYER_ID: WorldStateFixture.player_state(150.0, 170.0, 0.5, 15),
		WorldStateFixture.REMOTE_PLAYER_ID: WorldStateFixture.player_state(260.0, 280.0, 1.75, 25),
	}
	state[Packets.FIELD_ASTEROIDS] = {
		WorldStateFixture.ASTEROID_ID: WorldStateFixture.asteroid_state(360.0, 380.0, 2, 1.5),
	}
	state[Packets.FIELD_BULLETS] = {
		WorldStateFixture.BULLET_ID: WorldStateFixture.bullet_state(460.0, 480.0, 1.25),
	}
	return state


func _state_without_remote_player() -> Dictionary:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_PLAYERS] = {
		WorldStateFixture.LOCAL_PLAYER_ID: WorldStateFixture.player_state(100.0, 120.0, 0.25, 10),
	}
	return state


func _state_without_asteroid() -> Dictionary:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_ASTEROIDS] = {}
	return state


func _state_without_bullet() -> Dictionary:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_BULLETS] = {}
	return state


func _asteroid_state_without_scale() -> Dictionary:
	var asteroid := WorldStateFixture.asteroid_state(320.0, 340.0, 1, 1.25)
	asteroid.erase(Packets.FIELD_SCALE)
	return asteroid
