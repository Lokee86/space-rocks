extends RefCounted
class_name WorldSync

const Constants = preload("res://scripts/generated/constants/constants.gd")
const AsteroidSyncScript = preload("res://scripts/world/asteroid_sync.gd")
const ProjectileSyncScript = preload("res://scripts/world/projectile_sync.gd")
const PickupSyncScript = preload("res://scripts/world/pickup_sync.gd")
const PlayerRenderApiScript = preload("res://scripts/world/player_render/player_render_api.gd")
const TargetPositionSource = preload("res://scripts/gameplay/targeting/target_position_source.gd")
const AsteroidTrace = preload("res://scripts/networking/realtime/asteroid_trace.gd")

var asteroid_sync
var projectile_sync
var pickup_sync
var player_render_api
var target_position_source
var world_lane_state
var view_anchor: Node2D
var local_player: Player
var current_self_id := ""
var _measurement_observer: Callable


func configure(
	game_owner: Node2D,
	player: Player,
	view_anchor_ref: Node2D,
	bullets: Node2D,
	asteroids: Node2D,
	pickups: Node2D,
	pause_state_tracker = null
) -> void:
	asteroid_sync = AsteroidSyncScript.new()
	asteroid_sync.configure(asteroids)
	projectile_sync = ProjectileSyncScript.new()
	projectile_sync.configure(bullets)
	pickup_sync = PickupSyncScript.new()
	pickup_sync.configure(pickups)
	local_player = player
	player_render_api = PlayerRenderApiScript.new()
	view_anchor = view_anchor_ref
	player_render_api.configure(game_owner, player, view_anchor_ref, pause_state_tracker)
	target_position_source = TargetPositionSource.new()
	target_position_source.configure(player_render_api, asteroid_sync, projectile_sync, pickup_sync)

	asteroids.z_index = Constants.ASTEROID_Z_INDEX
	pickups.z_index = Constants.PICKUP_Z_INDEX
	bullets.z_index = Constants.BULLET_Z_INDEX


func configure_measurement_observer(observer: Callable) -> void:
	_measurement_observer = observer
	if player_render_api != null and player_render_api.has_method("set_measurement_observer"):
		player_render_api.set_measurement_observer(observer)
	if asteroid_sync != null and asteroid_sync.has_method("set_measurement_observer"):
		asteroid_sync.set_measurement_observer(observer)
	if projectile_sync != null and projectile_sync.has_method("set_measurement_observer"):
		projectile_sync.set_measurement_observer(observer)
	if pickup_sync != null and pickup_sync.has_method("set_measurement_observer"):
		pickup_sync.set_measurement_observer(observer)


func entity_counts() -> Dictionary:
	return {
		"players": player_nodes().size(),
		"bullets": _dictionary_size(projectile_sync, "projectile_nodes"),
		"asteroids": _dictionary_size(asteroid_sync, "asteroid_nodes"),
		"pickups": _dictionary_size(pickup_sync, "pickup_nodes"),
	}


func scene_node_count() -> int:
	if view_anchor == null or not view_anchor.has_method("get_tree"):
		return -1
	var tree := view_anchor.get_tree()
	return tree.get_node_count() if tree != null else -1


func _dictionary_size(owner, property_name: String) -> int:
	if owner == null:
		return 0
	var value = owner.get(property_name)
	return value.size() if value is Dictionary else 0


func _dictionary_property_has(owner, property_name: StringName, key) -> bool:
	if owner == null:
		return false
	for property in owner.get_property_list():
		if StringName(property.get("name", "")) != property_name:
			continue
		var value = owner.get(property_name)
		return value is Dictionary and value.has(key)
	return false


func set_current_self_id(self_id: String) -> void:
	current_self_id = self_id
	if target_position_source != null:
		target_position_source.set_current_self_id(self_id)


func apply_world_lane_state(world_lane_state_ref) -> void:
	if world_lane_state_ref == null:
		return

	world_lane_state = world_lane_state_ref
	if player_render_api != null:
		player_render_api.remove_missing(world_lane_state.ships, current_self_id)
		player_render_api.apply_state(current_self_id, world_lane_state.ships)
	_rebase_world_entities_to_view_anchor()
	var local_visual_position: Vector2 = player_render_api.visual_position()
	var local_server_position: Vector2 = player_render_api.server_position()
	if projectile_sync != null:
		if world_lane_state.bullet_full_sync_required:
			projectile_sync.remove_missing(world_lane_state.bullets)
			projectile_sync.apply(world_lane_state.bullets, local_visual_position, local_server_position)
			world_lane_state.clear_bullet_change_sets()
		else:
			for bullet_id in world_lane_state.removed_bullet_ids.keys():
				projectile_sync.remove_projectile(str(bullet_id))
			for bullet_id in world_lane_state.dirty_bullet_ids.keys():
				if world_lane_state.bullets.has(bullet_id):
					projectile_sync.apply_projectile(str(bullet_id), world_lane_state.bullets[bullet_id], local_visual_position, local_server_position)
			world_lane_state.clear_bullet_change_sets()
	if asteroid_sync != null:
		if world_lane_state.asteroid_full_sync_required:
			asteroid_sync.remove_missing(world_lane_state.asteroids)
			asteroid_sync.apply(world_lane_state.asteroids, player_render_api.visual_position(), player_render_api.server_position())
			world_lane_state.clear_asteroid_change_sets()
		else:
			for asteroid_id in world_lane_state.removed_asteroid_ids.keys():
				AsteroidTrace.record_event("presentation_remove_requested", {
					"asteroid_id": str(asteroid_id),
					"node_existed": _dictionary_property_has(asteroid_sync, &"asteroid_nodes", asteroid_id),
				})
				asteroid_sync.remove_asteroid(str(asteroid_id))
			for asteroid_id in world_lane_state.dirty_asteroid_ids.keys():
				if world_lane_state.asteroids.has(asteroid_id):
					var source: String = str(world_lane_state.asteroid_dirty_sources.get(asteroid_id, "unknown"))
					var node_existed: bool = _dictionary_property_has(asteroid_sync, &"asteroid_nodes", asteroid_id)
					if source == "hot_update" and not node_existed:
						AsteroidTrace.anomaly("hot_update_recreating_presentation_node", {
							"asteroid_id": str(asteroid_id),
							"state_count": world_lane_state.asteroids.size(),
						})
					if source != "hot_update" or not node_existed:
						AsteroidTrace.record_event("presentation_apply", {
							"asteroid_id": str(asteroid_id),
							"source": source,
							"node_existed": node_existed,
						})
					asteroid_sync.apply_asteroid(
						str(asteroid_id),
						world_lane_state.asteroids[asteroid_id],
						player_render_api.visual_position(),
						player_render_api.server_position(),
						source != "hot_update"
					)
			world_lane_state.clear_asteroid_change_sets()
	if pickup_sync != null:
		pickup_sync.remove_missing(world_lane_state.pickups)
		pickup_sync.apply(world_lane_state.pickups, player_render_api.visual_position(), player_render_api.server_position())


func reset() -> void:
	set_current_self_id("")
	if player_render_api != null:
		player_render_api.reset()
	if projectile_sync != null:
		projectile_sync.reset()
	if asteroid_sync != null:
		asteroid_sync.reset()
	if pickup_sync != null:
		pickup_sync.reset()
	world_lane_state = null
	clear_view_target_player()


func apply_state(
	self_id: String,
	server_players: Dictionary,
	server_bullets: Dictionary,
	server_asteroids: Dictionary,
	server_pickups: Dictionary = {}
) -> void:
	current_self_id = self_id

	if target_position_source != null:
		target_position_source.set_current_self_id(self_id)
	player_render_api.remove_missing(server_players, self_id)
	projectile_sync.remove_missing(server_bullets)
	asteroid_sync.remove_missing(server_asteroids)
	pickup_sync.remove_missing(server_pickups)
	player_render_api.apply_state(self_id, server_players)
	_rebase_world_entities_to_view_anchor()
	projectile_sync.apply(
		server_bullets,
		player_render_api.visual_position(),
		player_render_api.server_position()
	)
	asteroid_sync.apply(
		server_asteroids,
		player_render_api.visual_position(),
		player_render_api.server_position()
	)
	pickup_sync.apply(
		server_pickups,
		player_render_api.visual_position(),
		player_render_api.server_position()
	)


func _rebase_world_entities_to_view_anchor() -> void:
	if player_render_api == null:
		return
	var anchor_visual_position: Vector2 = player_render_api.visual_position()
	var anchor_server_position: Vector2 = player_render_api.server_position()
	for entity_sync in [projectile_sync, asteroid_sync, pickup_sync]:
		if entity_sync != null && entity_sync.has_method("rebase_to_view_anchor"):
			entity_sync.rebase_to_view_anchor(anchor_visual_position, anchor_server_position)


func interpolate(delta: float) -> void:
	var weight := 1.0 - exp(-Constants.PLAYER_INTERPOLATION_SPEED * delta)
	player_render_api.interpolate(weight, current_self_id)
	projectile_sync.interpolate(weight)
	asteroid_sync.interpolate(weight)
	pickup_sync.interpolate(weight)


func get_remote_player_visual_positions() -> Dictionary:
	if player_render_api == null:
		return {}
	return player_render_api.get_remote_player_visual_positions(current_self_id)


func get_remote_player_hues() -> Dictionary:
	if player_render_api == null:
		return {}
	return player_render_api.get_remote_player_hues(current_self_id)


func remote_player_nodes() -> Dictionary:
	if player_render_api == null:
		return {}
	return player_render_api.remote_player_nodes(current_self_id)


func player_nodes() -> Dictionary:
	if player_render_api == null:
		return {}
	return player_render_api.player_nodes()


func focus_camera_on_player(player_id: String) -> bool:
	if player_render_api == null:
		return false
	return player_render_api.focus_camera_on_player(player_id)


func set_view_target_player(player_id: String) -> void:
	if player_render_api != null:
		player_render_api.set_view_target_player(player_id)


func clear_view_target_player() -> void:
	if player_render_api != null:
		player_render_api.clear_view_target_player()


func visual_position_for_server_position(server_position: Vector2) -> Vector2:
	return player_render_api.visual_position_for_server_position(server_position)


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	return player_render_api.server_position_for_visual_position(visual_position)


func target_source():
	return target_position_source
