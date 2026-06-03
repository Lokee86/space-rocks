extends RefCounted

const Packets = preload("res://scripts/networking/packets/packets.gd")
const VisualSyncPositions = preload("res://scripts/world/visual_sync_positions.gd")

var local_player: Player
var player_lifecycle := PlayerSyncLifecycle.new()
var player_interpolation := PlayerSyncInterpolation.new()
var player_targets := PlayerSyncTargets.new()
var player_presentation := PlayerSyncPresentation.new()
var view_target_player_id := ""
var player_hue_presenter := PlayerHuePresenter.new()
var pause_state_tracker


func configure(game_owner: Node2D, player: Player, pause_tracker = null) -> void:
	local_player = player
	pause_state_tracker = pause_tracker
	player_lifecycle.configure(
		game_owner,
		player,
		Callable(self, "apply_local_player_hue"),
		Callable(self, "apply_remote_player_hue")
	)


func reset() -> void:
	player_hue_presenter.reset()
	view_target_player_id = ""

	player_lifecycle.reset()
	player_targets.reset()


func set_view_target_player(player_id: String) -> void:
	view_target_player_id = player_id
	_make_local_camera_current()


func clear_view_target_player() -> void:
	view_target_player_id = ""
	_make_local_camera_current()


func focus_camera_on_player(player_id: String) -> bool:
	if !player_lifecycle.has_player_node(player_id):
		return false

	view_target_player_id = player_id
	_make_local_camera_current()

	return true


func _make_local_camera_current() -> bool:
	if local_player == null:
		return false

	var camera := local_player.get_node_or_null("Camera2D") as Camera2D
	if camera == null:
		return false

	camera.make_current()
	return true


func apply(
	self_id: String,
	server_players: Dictionary,
	local_visual_position: Vector2,
	local_server_position: Vector2
) -> void:
	if local_player != null:
		player_hue_presenter.apply_local_player_hue(local_player)

	var remote_player_ids := []
	for player_id in server_players.keys():
		if player_id == self_id:
			continue
		remote_player_ids.append(player_id)
	remote_player_ids.sort()
	player_hue_presenter.set_remote_player_order(remote_player_ids)

	for player_id in server_players.keys():
		var state: Dictionary = server_players[player_id]
		var player_node: Player = player_lifecycle.get_or_create_player_node(self_id, player_id)
		var visual_position := local_visual_position
		if player_id != self_id:
			visual_position = VisualSyncPositions.relative_to_local_visual(
				local_visual_position,
				local_server_position,
				Vector2(state[Packets.FIELD_X], state[Packets.FIELD_Y])
			)
		var server_rotation: float = state[Packets.FIELD_ROTATION]

		if player_id != self_id:
			var remote_afterburner_active: bool = bool(state.get("thrusting", false)) && (
				pause_state_tracker == null or !pause_state_tracker.is_paused(player_id)
			)
			var is_paused: bool = pause_state_tracker != null and pause_state_tracker.is_paused(player_id)
			player_interpolation.correct_remote_visual_copy_mismatch(
				player_id,
				player_node,
				visual_position,
				player_lifecycle,
				player_targets
			)
			player_hue_presenter.apply_remote_player_hue(player_id, player_node)
			player_presentation.apply_remote_player_presentation(
				player_id,
				self_id,
				player_node,
				is_paused,
				remote_afterburner_active
			)

		player_targets.set_target_player_state(player_id, visual_position, server_rotation)

		if !player_lifecycle.is_initialized(player_id):
			player_lifecycle.mark_initialized(player_id)
			player_node.position = visual_position
			player_node.rotation = server_rotation


func apply_local_player_hue(player: Player) -> void:
	player_hue_presenter.apply_local_player_hue(player)


func apply_remote_player_hue(player_id: String, remote_player: Player) -> void:
	player_hue_presenter.apply_remote_player_hue(player_id, remote_player)


func get_remote_player_hues(current_self_id: String) -> Dictionary:
	return player_hue_presenter.remote_player_hues_without(current_self_id)


func remote_player_nodes(current_self_id: String) -> Dictionary:
	return player_lifecycle.get_remote_player_nodes(current_self_id)


func remove_missing(server_players: Dictionary, self_id: String) -> void:
	var removed_player_ids := player_lifecycle.remove_missing(server_players, self_id)
	for player_id in removed_player_ids:
		if player_id == view_target_player_id:
			clear_view_target_player()
		player_targets.erase_player(player_id)
		player_hue_presenter.remove_player(player_id)


func interpolate(weight: float, current_self_id: String) -> void:
	player_interpolation.interpolate_player_nodes(
		weight,
		current_self_id,
		player_lifecycle,
		player_targets,
		view_target_player_id,
		local_player
	)


func get_remote_player_visual_positions(current_self_id: String) -> Dictionary:
	return player_targets.get_remote_player_visual_positions_without(current_self_id)


func server_hitbox_draw_entries(_current_self_id: String) -> Array:
	return player_targets.build_server_hitbox_draw_entries(_current_self_id, player_lifecycle)
