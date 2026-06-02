extends RefCounted
class_name GameplayRuntimeContext

const WorldSyncScript = preload("res://scripts/world/world_sync.gd")
const TARGET_PLAYER_PICK_RADIUS := 32.0
const TARGET_ASTEROID_BASE_PICK_RADIUS := 32.0
const TARGET_BULLET_PICK_RADIUS := 12.0

var world_sync
var player
var event_flow
var death_flow
var respawn_flow
var hud_flow


func configure_world(
	game_owner: Node2D,
	player_ref: Player,
	bullets: Node2D,
	asteroids: Node2D,
	pause_state_tracker = null
) -> void:
	player = player_ref
	world_sync = WorldSyncScript.new()
	world_sync.configure(game_owner, player_ref, bullets, asteroids, pause_state_tracker)


func configure_events(
	game_owner: Node2D,
	hud: Control,
	hud_flow_ref,
	menu_flow_ref
) -> void:
	event_flow = GameplayEventFlow.new()
	event_flow.configure(
		game_owner,
		hud,
		Callable(world_sync, "visual_position_for_server_position")
	)
	death_flow = GameplayDeathFlow.new()
	death_flow.configure(hud_flow_ref, menu_flow_ref, event_flow, player)
	event_flow.self_death_event.connect(Callable(death_flow, "apply_self_death_event"))


func configure_respawn(connection_service_ref, hud_flow_ref) -> void:
	hud_flow = hud_flow_ref
	respawn_flow = GameplayRespawnFlow.new()
	respawn_flow.configure(connection_service_ref, hud_flow_ref)


func reset() -> void:
	if player != null:
		player.hide()
	if world_sync != null:
		world_sync.reset()
	if event_flow != null:
		event_flow.reset()
	if respawn_flow != null:
		respawn_flow.reset()


func process(delta: float) -> void:
	if world_sync != null:
		world_sync.interpolate(delta)


func request_respawn(has_received_state: bool) -> void:
	if respawn_flow != null:
		respawn_flow.request_respawn(has_received_state)


func apply_world_state(state: Dictionary, has_received_state: bool) -> void:
	if world_sync == null:
		return

	world_sync.apply_state(
		state["self_id"],
		state["server_players"],
		state["server_bullets"],
		state["server_asteroids"],
		has_received_state
	)


func apply_server_events(state: Dictionary) -> void:
	if event_flow != null:
		event_flow.apply_server_events(state["server_events"], state["self_id"])


func apply_respawn_alive_restore(state: Dictionary, menu_flow_ref) -> void:
	if hud_flow == null || respawn_flow == null:
		return

	var has_stale_dead_presentation: bool = false
	has_stale_dead_presentation = bool(hud_flow.is_dead) || bool(hud_flow.is_game_over)
	if menu_flow_ref != null:
		has_stale_dead_presentation = has_stale_dead_presentation || bool(menu_flow_ref.is_game_over)

	if !respawn_flow.should_restore_alive_hud(state, player, has_stale_dead_presentation):
		return

	if world_sync != null:
		if world_sync.has_method("clear_view_reference_player"):
			world_sync.clear_view_reference_player()
		if world_sync.has_method("clear_view_target_player"):
			world_sync.clear_view_target_player()
	hud_flow.set_alive()
	if menu_flow_ref != null:
		menu_flow_ref.set_alive()
	respawn_flow.clear_awaiting_confirmation()


func current_camera() -> Camera2D:
	if player == null:
		return null
	return player.get_node_or_null("Camera2D") as Camera2D


func remote_player_visual_positions() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.get_remote_player_visual_positions()


func remote_player_hues() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.get_remote_player_hues()


func remote_player_nodes() -> Dictionary:
	if world_sync == null:
		return {}
	return world_sync.remote_player_nodes()


func server_position_for_visual_position(visual_position: Vector2) -> Vector2:
	if world_sync == null:
		return visual_position
	return world_sync.server_position_for_visual_position(visual_position)


func target_visual_candidates() -> Array:
	var candidates: Array = []
	if world_sync == null:
		return candidates

	var player_positions: Dictionary = world_sync.player_target_positions()
	for player_id in player_positions.keys():
		var position_entry = player_positions[player_id]
		if not (position_entry is Dictionary):
			continue
		if not position_entry.has("visual_position"):
			continue
		if not position_entry.has("server_position"):
			continue

		var candidate := TargetVisualCandidate.new()
		candidate.target_kind = "player"
		candidate.target_id = String(player_id)
		candidate.visual_position = position_entry["visual_position"]
		candidate.server_position = position_entry["server_position"]
		candidate.pick_radius = TARGET_PLAYER_PICK_RADIUS
		candidates.append(candidate)

	var asteroid_positions: Dictionary = world_sync.asteroid_target_positions()
	for asteroid_id in asteroid_positions.keys():
		var position_entry = asteroid_positions[asteroid_id]
		if not (position_entry is Dictionary):
			continue
		if not position_entry.has("visual_position"):
			continue
		if not position_entry.has("server_position"):
			continue

		var visual_scale := 1.0
		if position_entry.has("visual_scale"):
			visual_scale = float(position_entry["visual_scale"])

		var asteroid_candidate := TargetVisualCandidate.new()
		asteroid_candidate.target_kind = "asteroid"
		asteroid_candidate.target_id = String(asteroid_id)
		asteroid_candidate.visual_position = position_entry["visual_position"]
		asteroid_candidate.server_position = position_entry["server_position"]
		asteroid_candidate.pick_radius = TARGET_ASTEROID_BASE_PICK_RADIUS * visual_scale
		candidates.append(asteroid_candidate)

	var bullet_positions: Dictionary = world_sync.bullet_target_positions()
	for bullet_id in bullet_positions.keys():
		var position_entry = bullet_positions[bullet_id]
		if not (position_entry is Dictionary):
			continue
		if not position_entry.has("visual_position"):
			continue
		if not position_entry.has("server_position"):
			continue

		var bullet_candidate := TargetVisualCandidate.new()
		bullet_candidate.target_kind = "bullet"
		bullet_candidate.target_id = String(bullet_id)
		bullet_candidate.visual_position = position_entry["visual_position"]
		bullet_candidate.server_position = position_entry["server_position"]
		bullet_candidate.pick_radius = TARGET_BULLET_PICK_RADIUS
		candidates.append(bullet_candidate)

	return candidates
