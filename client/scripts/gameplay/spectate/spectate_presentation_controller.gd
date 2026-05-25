extends RefCounted

const SpectateCycleViewPolicy = preload("res://scripts/gameplay/spectate/spectate_cycle_view_policy.gd")

var hud_controller
var camera_follow
var world_sync


func configure(hud_controller_object, camera_follow_object = null, world_sync_object = null) -> void:
	hud_controller = hud_controller_object
	camera_follow = camera_follow_object
	world_sync = world_sync_object


func set_cycle_view_hint(session_mode, current_room_state, is_game_over, is_spectating) -> void:
	if hud_controller == null:
		return

	hud_controller.set_session_mode(session_mode)
	var cycle_view_available := SpectateCycleViewPolicy.is_cycle_view_available(
		session_mode,
		current_room_state,
		is_game_over,
		is_spectating
	)
	hud_controller.set_cycle_view_available(cycle_view_available)


func follow_local_player() -> void:
	if camera_follow != null:
		camera_follow.follow_local_player()


func follow_visual_position(visual_position: Vector2) -> void:
	if camera_follow != null:
		camera_follow.follow_visual_position(visual_position)


func remote_player_visual_positions() -> Dictionary:
	if world_sync == null:
		return {}

	return world_sync.get_remote_player_visual_positions()


func has_spectate_targets(spectate_controller, self_id, player_lifecycle) -> bool:
	return spectate_controller.has_targets(
		self_id,
		remote_player_visual_positions(),
		player_lifecycle
	)


func start_spectating(
	spectate_controller,
	self_id,
	player_lifecycle,
	hide_game_menu_callback,
	update_spectate_camera_callback,
	refresh_cycle_view_hint_callback
) -> bool:
	return spectate_controller.start_spectating(
		self_id,
		remote_player_visual_positions(),
		player_lifecycle,
		hide_game_menu_callback,
		update_spectate_camera_callback,
		refresh_cycle_view_hint_callback
	)


func stop_spectating(
	spectate_controller,
	show_game_over_menu: bool,
	show_game_menu_callback: Callable,
	refresh_cycle_view_hint_callback: Callable
) -> void:
	spectate_controller.stop_spectating(
		show_game_over_menu,
		hud_controller != null && hud_controller.is_game_over,
		Callable(self, "follow_local_player"),
		show_game_menu_callback,
		refresh_cycle_view_hint_callback
	)


func update_spectate_camera(spectate_controller, self_id, player_lifecycle, stop_spectating_callback) -> void:
	spectate_controller.update_camera(
		self_id,
		remote_player_visual_positions(),
		player_lifecycle,
		camera_follow,
		stop_spectating_callback
	)


func cycle_spectate_target(
	spectate_controller,
	self_id,
	player_lifecycle,
	stop_spectating_callback,
	follow_visual_position_callback
) -> void:
	spectate_controller.cycle_target(
		self_id,
		remote_player_visual_positions(),
		player_lifecycle,
		stop_spectating_callback,
		follow_visual_position_callback
	)
