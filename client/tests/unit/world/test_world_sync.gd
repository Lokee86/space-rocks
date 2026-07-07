extends GutTest

const WorldSync := preload("res://scripts/world/world_sync.gd")
const WorldLaneState := preload("res://scripts/protocol/realtime/world_lane_state.gd")

class FakePlayerRenderApi:
	extends RefCounted

	func remove_missing(_server_players: Dictionary, _self_id: String) -> void:
		pass

	func apply_state(_self_id: String, _server_players: Dictionary) -> void:
		pass

	func visual_position() -> Vector2:
		return Vector2.ZERO

	func server_position() -> Vector2:
		return Vector2.ZERO


class FakeProjectileSync:
	extends RefCounted

	var apply_projectile_calls: Array = []
	var remove_projectile_calls: Array = []
	var apply_all_calls := 0
	var remove_missing_calls := 0

	func apply_projectile(bullet_id: String, state: Dictionary, local_visual_position: Vector2, local_server_position: Vector2) -> void:
		apply_projectile_calls.append({
			"bullet_id": bullet_id,
			"state": state,
			"local_visual_position": local_visual_position,
			"local_server_position": local_server_position,
		})

	func remove_projectile(bullet_id: String) -> void:
		remove_projectile_calls.append(bullet_id)

	func apply(_server_bullets: Dictionary, _local_visual_position: Vector2, _local_server_position: Vector2) -> void:
		apply_all_calls += 1

	func remove_missing(_server_bullets: Dictionary) -> void:
		remove_missing_calls += 1


class FakeAsteroidSync:
	extends RefCounted

	var apply_asteroid_calls: Array = []
	var remove_asteroid_calls: Array = []
	var apply_all_calls := 0
	var remove_missing_calls := 0

	func apply_asteroid(asteroid_id: String, state: Dictionary, local_visual_position: Vector2, local_server_position: Vector2) -> void:
		apply_asteroid_calls.append({
			"asteroid_id": asteroid_id,
			"state": state,
			"local_visual_position": local_visual_position,
			"local_server_position": local_server_position,
		})

	func remove_asteroid(asteroid_id: String) -> void:
		remove_asteroid_calls.append(asteroid_id)

	func apply(_server_asteroids: Dictionary, _local_visual_position: Vector2, _local_server_position: Vector2) -> void:
		apply_all_calls += 1

	func remove_missing(_server_asteroids: Dictionary) -> void:
		remove_missing_calls += 1


class FakePickupSync:
	extends RefCounted

	var apply_calls: Array = []
	var remove_missing_calls := 0

	func apply(server_pickups: Dictionary, local_visual_position: Vector2, local_server_position: Vector2) -> void:
		apply_calls.append({
			"server_pickups": server_pickups,
			"local_visual_position": local_visual_position,
			"local_server_position": local_server_position,
		})

	func remove_missing(_server_pickups: Dictionary) -> void:
		remove_missing_calls += 1


func test_apply_world_lane_state_uses_direct_bullet_change_sets() -> void:
	var world_sync := WorldSync.new()
	var fake_player_render_api := FakePlayerRenderApi.new()
	var fake_projectile_sync := FakeProjectileSync.new()
	var fake_asteroid_sync := FakeAsteroidSync.new()
	var fake_pickup_sync := FakePickupSync.new()
	var world_lane_state := WorldLaneState.new()

	world_sync.player_render_api = fake_player_render_api
	world_sync.projectile_sync = fake_projectile_sync
	world_sync.asteroid_sync = fake_asteroid_sync
	world_sync.pickup_sync = fake_pickup_sync
	world_lane_state.bullets["bullet-1"] = {
		"id": "bullet-1",
		"x": 10.0,
		"y": 20.0,
		"rotation": 1.25,
	}
	world_lane_state.dirty_bullet_ids["bullet-1"] = true
	world_lane_state.bullet_full_sync_required = false

	world_sync.apply_world_lane_state(world_lane_state)

	assert_eq(fake_projectile_sync.apply_projectile_calls.size(), 1)
	assert_eq(fake_projectile_sync.apply_projectile_calls[0]["bullet_id"], "bullet-1")
	assert_eq(fake_projectile_sync.apply_projectile_calls[0]["state"]["id"], "bullet-1")
	assert_eq(fake_projectile_sync.apply_all_calls, 0)
	assert_eq(fake_projectile_sync.remove_missing_calls, 0)
	assert_true(world_lane_state.dirty_bullet_ids.is_empty())


func test_apply_world_lane_state_uses_direct_asteroid_change_sets() -> void:
	var world_sync := WorldSync.new()
	var fake_player_render_api := FakePlayerRenderApi.new()
	var fake_projectile_sync := FakeProjectileSync.new()
	var fake_asteroid_sync := FakeAsteroidSync.new()
	var fake_pickup_sync := FakePickupSync.new()
	var world_lane_state := WorldLaneState.new()

	world_sync.player_render_api = fake_player_render_api
	world_sync.projectile_sync = fake_projectile_sync
	world_sync.asteroid_sync = fake_asteroid_sync
	world_sync.pickup_sync = fake_pickup_sync
	world_lane_state.asteroids["asteroid-1"] = {
		"id": "asteroid-1",
		"x": 10.0,
		"y": 20.0,
		"rotation": 1.25,
		"size": 2,
		"health": 50,
	}
	world_lane_state.dirty_asteroid_ids["asteroid-1"] = true
	world_lane_state.asteroid_full_sync_required = false

	world_sync.apply_world_lane_state(world_lane_state)

	assert_eq(fake_asteroid_sync.apply_asteroid_calls.size(), 1)
	assert_eq(fake_asteroid_sync.apply_asteroid_calls[0]["asteroid_id"], "asteroid-1")
	assert_eq(fake_asteroid_sync.apply_all_calls, 0)
	assert_eq(fake_asteroid_sync.remove_missing_calls, 0)
	assert_true(world_lane_state.dirty_asteroid_ids.is_empty())
