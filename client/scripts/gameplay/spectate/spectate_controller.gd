extends RefCounted

const SpectateTargetsScript = preload("res://scripts/gameplay/spectate/spectate_targets.gd")

var is_spectating := false
var current_spectate_target_id := ""


func is_active() -> bool:
	return is_spectating


func set_active(active: bool) -> void:
	is_spectating = active


func current_target_id() -> String:
	return current_spectate_target_id


func set_current_target_id(target_id: String) -> void:
	current_spectate_target_id = target_id


func start_spectating(
	self_id: String,
	remote_positions: Dictionary,
	player_lifecycle: Dictionary,
	hide_game_menu: Callable,
	update_spectate_camera: Callable,
	refresh_cycle_view_hint: Callable
) -> bool:
	current_spectate_target_id = SpectateTargetsScript.select_target(
		self_id,
		current_spectate_target_id,
		remote_positions,
		player_lifecycle
	)
	if current_spectate_target_id == "":
		is_spectating = false
		return false

	is_spectating = true
	hide_game_menu.call()
	update_spectate_camera.call()
	refresh_cycle_view_hint.call()
	return true


func has_targets(self_id: String, remote_positions: Dictionary, player_lifecycle: Dictionary) -> bool:
	return SpectateTargetsScript.select_target(
		self_id,
		"",
		remote_positions,
		player_lifecycle
	) != ""


func stop_spectating(
	show_game_over_menu: bool,
	is_game_over: bool,
	follow_local_player: Callable,
	show_game_menu: Callable,
	refresh_cycle_view_hint: Callable
) -> void:
	if !is_spectating && current_spectate_target_id == "":
		return

	is_spectating = false
	current_spectate_target_id = ""
	follow_local_player.call()
	if show_game_over_menu && is_game_over:
		show_game_menu.call()
	refresh_cycle_view_hint.call()


func cycle_target(
	self_id: String,
	remote_positions: Dictionary,
	player_lifecycle: Dictionary,
	stop_spectating: Callable,
	follow_visual_position: Callable
) -> void:
	if !is_spectating:
		return

	current_spectate_target_id = SpectateTargetsScript.cycle_target(
		self_id,
		current_spectate_target_id,
		remote_positions,
		player_lifecycle
	)
	if current_spectate_target_id == "":
		stop_spectating.call(true)
		return

	follow_visual_position.call(remote_positions[current_spectate_target_id])


func update_camera(
	self_id: String,
	remote_positions: Dictionary,
	player_lifecycle: Dictionary,
	camera_follow,
	stop_spectating: Callable
) -> void:
	if !is_spectating:
		return

	current_spectate_target_id = SpectateTargetsScript.select_target(
		self_id,
		current_spectate_target_id,
		remote_positions,
		player_lifecycle
	)
	if current_spectate_target_id == "":
		stop_spectating.call(true)
		return

	if camera_follow != null:
		camera_follow.follow_visual_position(remote_positions[current_spectate_target_id])
