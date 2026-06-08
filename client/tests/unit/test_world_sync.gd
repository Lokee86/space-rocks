extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const WorldStateFixture := preload("res://tests/fixtures/world_state_fixture.gd")
const WorldSyncScript := preload("res://scripts/world/world_sync.gd")
const PlayerScene := preload("res://scenes/player.tscn")

var game_owner: Node2D
var local_player: Player
var view_anchor: Node2D
var bullets_layer: Node2D
var asteroids_layer: Node2D
var pickups_layer: Node2D
var world_sync: WorldSyncScript


func before_each() -> void:
	game_owner = Node2D.new()
	add_child(game_owner)

	local_player = PlayerScene.instantiate()
	view_anchor = Node2D.new()
	bullets_layer = Node2D.new()
	asteroids_layer = Node2D.new()
	pickups_layer = Node2D.new()

	game_owner.add_child(local_player)
	game_owner.add_child(view_anchor)
	game_owner.add_child(bullets_layer)
	game_owner.add_child(asteroids_layer)
	game_owner.add_child(pickups_layer)

	world_sync = WorldSyncScript.new()
	world_sync.configure(
		game_owner,
		local_player,
		view_anchor,
		bullets_layer,
		asteroids_layer,
		pickups_layer
	)


func after_each() -> void:
	world_sync = null
	if game_owner != null:
		game_owner.free()
		game_owner = null


func test_apply_state_creates_player_nodes() -> void:
	_apply_fixture_state()

	assert_true(_player_nodes().has(WorldStateFixture.LOCAL_PLAYER_ID))
	assert_true(_player_nodes().has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_eq(_player_nodes()[WorldStateFixture.LOCAL_PLAYER_ID], local_player)
	assert_eq(
		_player_nodes()[WorldStateFixture.LOCAL_PLAYER_ID].position,
		Vector2(100.0, 120.0)
	)
	assert_eq(
		_player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID].position,
		Vector2(220.0, 240.0)
	)
	assert_eq(
		_player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID].get_parent(),
		game_owner
	)


func test_local_player_draws_above_remote_players() -> void:
	_apply_fixture_state()

	assert_gt(
		_player_nodes()[WorldStateFixture.LOCAL_PLAYER_ID].z_index,
		_player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID].z_index
	)


func test_apply_state_creates_asteroid_nodes() -> void:
	_apply_fixture_state()

	assert_true(_asteroid_nodes().has(WorldStateFixture.ASTEROID_ID))
	assert_eq(
		_asteroid_nodes()[WorldStateFixture.ASTEROID_ID].get_parent(),
		asteroids_layer
	)
	assert_eq(
		_asteroid_nodes()[WorldStateFixture.ASTEROID_ID].global_position,
		Vector2(320.0, 340.0)
	)


func test_apply_state_creates_bullet_nodes() -> void:
	_apply_fixture_state()

	assert_true(_projectile_nodes().has(WorldStateFixture.BULLET_ID))
	assert_eq(
		_projectile_nodes()[WorldStateFixture.BULLET_ID].get_parent(),
		bullets_layer
	)
	assert_eq(
		_projectile_nodes()[WorldStateFixture.BULLET_ID].global_position,
		Vector2(420.0, 440.0)
	)


func test_apply_state_starts_bullet_firing_sound_on_first_projectile_creation() -> void:
	_apply_fixture_state()

	var bullet_node: Node = _projectile_nodes()[WorldStateFixture.BULLET_ID]
	var firing_sound := bullet_node.get_node_or_null("FiringSound")

	assert_not_null(firing_sound)
	assert_true(firing_sound is AudioStreamPlayer2D)


func test_apply_state_creates_torpedo_scene_for_torpedo_projectile_type() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_BULLETS] = {
		WorldStateFixture.BULLET_ID: {
			Packets.FIELD_X: 420.0,
			Packets.FIELD_Y: 440.0,
			Packets.FIELD_ROTATION: 1.25,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
	}

	_apply_state(state)

	assert_true(_projectile_nodes().has(WorldStateFixture.BULLET_ID))
	assert_eq(_projectile_nodes()[WorldStateFixture.BULLET_ID].name, "Torpedo")
	assert_true(_projectile_nodes()[WorldStateFixture.BULLET_ID] is Node2D)


func test_apply_state_starts_torpedo_firing_sound_on_first_projectile_creation() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_BULLETS] = {
		WorldStateFixture.BULLET_ID: {
			Packets.FIELD_X: 420.0,
			Packets.FIELD_Y: 440.0,
			Packets.FIELD_ROTATION: 1.25,
			Packets.FIELD_PROJECTILE_TYPE: "torpedo",
		},
	}

	_apply_state(state)

	var torpedo_node: Node = _projectile_nodes()[WorldStateFixture.BULLET_ID]
	var firing_sound := torpedo_node.get_node_or_null("FiringSound")

	assert_not_null(firing_sound)
	assert_true(firing_sound is AudioStreamPlayer2D)


func test_apply_state_defaults_unknown_projectile_type_to_bullet_scene() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_BULLETS] = {
		WorldStateFixture.BULLET_ID: {
			Packets.FIELD_X: 420.0,
			Packets.FIELD_Y: 440.0,
			Packets.FIELD_ROTATION: 1.25,
			Packets.FIELD_PROJECTILE_TYPE: "mystery",
		},
	}

	_apply_state(state)

	assert_true(_projectile_nodes().has(WorldStateFixture.BULLET_ID))
	assert_eq(_projectile_nodes()[WorldStateFixture.BULLET_ID].name, "Bullet")
	assert_true(_projectile_nodes()[WorldStateFixture.BULLET_ID] is CharacterBody2D)


func test_apply_state_exposes_pickup_target_positions() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_PICKUPS] = {
		"pickup-1": {
			Packets.FIELD_ID: "pickup-1",
			Packets.FIELD_TYPE: "1_up",
			Packets.FIELD_PICKUP_CLASS: "powerup",
			Packets.FIELD_X: 520.0,
			Packets.FIELD_Y: 540.0,
		},
	}

	_apply_state(state)

	var pickup_positions: Dictionary = world_sync.target_source().pickup_positions()
	assert_true(pickup_positions.has("pickup-1"))
	assert_eq(pickup_positions["pickup-1"]["visual_position"], Vector2(520.0, 540.0))
	assert_eq(pickup_positions["pickup-1"]["server_position"], Vector2(520.0, 540.0))


func test_apply_state_reuses_existing_entity_nodes() -> void:
	_apply_fixture_state()
	var local_node = _player_nodes()[WorldStateFixture.LOCAL_PLAYER_ID]
	var remote_node = _player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID]
	var asteroid_node = _asteroid_nodes()[WorldStateFixture.ASTEROID_ID]
	var bullet_node = _projectile_nodes()[WorldStateFixture.BULLET_ID]
	var owner_child_count := game_owner.get_child_count()
	var asteroid_child_count := asteroids_layer.get_child_count()
	var bullet_child_count := bullets_layer.get_child_count()

	_apply_state(_updated_state())

	assert_eq(_player_nodes().size(), 2)
	assert_eq(_asteroid_nodes().size(), 1)
	assert_eq(_projectile_nodes().size(), 1)
	assert_eq(_player_nodes()[WorldStateFixture.LOCAL_PLAYER_ID], local_node)
	assert_eq(_player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID], remote_node)
	assert_eq(_asteroid_nodes()[WorldStateFixture.ASTEROID_ID], asteroid_node)
	assert_eq(_projectile_nodes()[WorldStateFixture.BULLET_ID], bullet_node)
	assert_eq(game_owner.get_child_count(), owner_child_count)
	assert_eq(asteroids_layer.get_child_count(), asteroid_child_count)
	assert_eq(bullets_layer.get_child_count(), bullet_child_count)


func test_apply_state_updates_existing_entity_targets() -> void:
	_apply_fixture_state()
	_apply_state(_updated_state())

	assert_eq(
		_player_sync().get("target_player_positions")[WorldStateFixture.LOCAL_PLAYER_ID],
		Vector2(150.0, 170.0)
	)
	assert_eq(_player_sync().get("target_player_rotations")[WorldStateFixture.LOCAL_PLAYER_ID], 0.5)
	assert_eq(
		_player_sync().get("target_player_positions")[WorldStateFixture.REMOTE_PLAYER_ID],
		Vector2(260.0, 280.0)
	)
	assert_eq(_player_sync().get("target_player_rotations")[WorldStateFixture.REMOTE_PLAYER_ID], 1.75)
	assert_eq(
		_asteroid_sync().get("target_asteroid_positions")[WorldStateFixture.ASTEROID_ID],
		Vector2(360.0, 380.0)
	)
	assert_eq(
		_projectile_sync().get("target_projectile_positions")[WorldStateFixture.BULLET_ID],
		Vector2(460.0, 480.0)
	)
	assert_eq(_projectile_sync().get("target_projectile_rotations")[WorldStateFixture.BULLET_ID], 1.25)


func test_apply_state_corrects_remote_visual_copy_mismatch_before_interpolation() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_PLAYERS] = {
		WorldStateFixture.LOCAL_PLAYER_ID: WorldStateFixture.player_state(656.0, 320.0, 0.0),
		WorldStateFixture.REMOTE_PLAYER_ID: WorldStateFixture.player_state(656.0, 320.0, 0.0),
	}
	state[Packets.FIELD_ASTEROIDS] = {}
	state[Packets.FIELD_BULLETS] = {}

	_apply_state(state)
	world_sync.interpolate(999.0)

	var rendered_snapshot_a: Vector2 = _player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID].position
	_local_visual_sync().set(
		"local_visual_position",
		Vector2(656.0, 320.0 - Constants.WORLD_HEIGHT)
	)

	_apply_state(state)
	var expected_target := Vector2(656.0, 320.0 - Constants.WORLD_HEIGHT)
	var rendered_snapshot_b: Vector2 = _player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID].position
	var remote_visual_positions: Dictionary = world_sync.get_remote_player_visual_positions()

	assert_eq(_player_sync().get("target_player_positions")[WorldStateFixture.REMOTE_PLAYER_ID], expected_target)
	assert_eq(rendered_snapshot_b, expected_target)
	assert_eq(remote_visual_positions[WorldStateFixture.REMOTE_PLAYER_ID], expected_target)
	assert_gt(abs(expected_target.y - rendered_snapshot_a.y), Constants.WORLD_HEIGHT * 0.5)


func test_interpolate_moves_existing_entities_toward_updated_state() -> void:
	_apply_fixture_state()
	_apply_state(_updated_state())
	world_sync.interpolate(999.0)

	assert_eq(
		_player_nodes()[WorldStateFixture.LOCAL_PLAYER_ID].position,
		Vector2(150.0, 170.0)
	)
	assert_eq(_player_nodes()[WorldStateFixture.LOCAL_PLAYER_ID].rotation, 0.5)
	assert_eq(
		_player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID].position,
		Vector2(260.0, 280.0)
	)
	assert_eq(_player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID].rotation, 1.75)
	assert_eq(
		_asteroid_nodes()[WorldStateFixture.ASTEROID_ID].global_position,
		Vector2(360.0, 380.0)
	)
	assert_eq(
		_projectile_nodes()[WorldStateFixture.BULLET_ID].global_position,
		Vector2(460.0, 480.0)
	)
	assert_eq(_projectile_nodes()[WorldStateFixture.BULLET_ID].rotation, 1.25)


func test_apply_state_removes_stale_remote_player_node() -> void:
	_apply_fixture_state()
	var remote_node = _player_nodes()[WorldStateFixture.REMOTE_PLAYER_ID]

	_apply_state(_state_without_remote_player())

	assert_false(_player_nodes().has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_false(_player_sync().get("initialized_players").has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_false(_player_sync().get("target_player_positions").has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_false(_player_sync().get("target_player_rotations").has(WorldStateFixture.REMOTE_PLAYER_ID))
	assert_true(remote_node.is_queued_for_deletion())


func test_apply_state_removes_stale_asteroid_node() -> void:
	_apply_fixture_state()
	var asteroid_node = _asteroid_nodes()[WorldStateFixture.ASTEROID_ID]

	_apply_state(_state_without_asteroid())

	assert_false(_asteroid_nodes().has(WorldStateFixture.ASTEROID_ID))
	assert_false(_asteroid_sync().get("initialized_asteroids").has(WorldStateFixture.ASTEROID_ID))
	assert_false(_asteroid_sync().get("target_asteroid_positions").has(WorldStateFixture.ASTEROID_ID))
	assert_true(asteroid_node.is_queued_for_deletion())


func test_apply_state_removes_stale_bullet_node() -> void:
	_apply_fixture_state()
	var bullet_node = _projectile_nodes()[WorldStateFixture.BULLET_ID]

	_apply_state(_state_without_bullet())

	assert_false(_projectile_nodes().has(WorldStateFixture.BULLET_ID))
	assert_false(_projectile_sync().get("initialized_projectiles").has(WorldStateFixture.BULLET_ID))
	assert_false(_projectile_sync().get("target_projectile_positions").has(WorldStateFixture.BULLET_ID))
	assert_false(_projectile_sync().get("target_projectile_rotations").has(WorldStateFixture.BULLET_ID))
	assert_true(bullet_node.is_queued_for_deletion())


func test_apply_state_missing_asteroid_scale_warns_once_and_does_not_crash() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_ASTEROIDS] = {
		WorldStateFixture.ASTEROID_ID: _asteroid_state_without_scale(),
	}

	_apply_state(state)
	_apply_state(state)

	assert_true(_asteroid_nodes().has(WorldStateFixture.ASTEROID_ID))
	assert_true(
		_asteroid_sync().get("warned_missing_asteroid_scale").has(WorldStateFixture.ASTEROID_ID)
	)


func test_apply_state_applies_asteroid_packet_scale() -> void:
	var state := WorldStateFixture.state()

	assert_true(state[Packets.FIELD_ASTEROIDS][WorldStateFixture.ASTEROID_ID].has(Packets.FIELD_SCALE))

	_apply_state(state)

	assert_eq(
		_asteroid_nodes()[WorldStateFixture.ASTEROID_ID].scale,
		Vector2.ONE * 1.25
	)

	state[Packets.FIELD_ASTEROIDS] = {
		WorldStateFixture.ASTEROID_ID: WorldStateFixture.asteroid_state(320.0, 340.0, 1, 1.75),
	}
	_apply_state(state)

	assert_eq(
		_asteroid_nodes()[WorldStateFixture.ASTEROID_ID].scale,
		Vector2.ONE * 1.75
	)


func _apply_fixture_state() -> void:
	_apply_state(WorldStateFixture.state())


func _projectile_nodes() -> Dictionary:
	return _projectile_sync().get("projectile_nodes")


func _player_nodes() -> Dictionary:
	return world_sync.player_nodes()


func _player_sync():
	return _player_render_api().get("player_meaning").get("legacy_player_sync")


func _local_visual_sync():
	return _player_render_api().get("view_anchor_sync").get("legacy_sync")


func _player_render_api():
	return world_sync.get("player_render_api")


func _asteroid_nodes() -> Dictionary:
	return _asteroid_sync().get("asteroid_nodes")


func _asteroid_sync():
	return world_sync.get("asteroid_sync")


func _projectile_sync():
	return world_sync.get("projectile_sync")


func _apply_state(state: Dictionary) -> void:
	world_sync.apply_state(
		state[Packets.FIELD_SELF_ID],
		state[Packets.FIELD_PLAYERS],
		state[Packets.FIELD_BULLETS],
		state[Packets.FIELD_ASTEROIDS],
		state.get(Packets.FIELD_PICKUPS, {})
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
