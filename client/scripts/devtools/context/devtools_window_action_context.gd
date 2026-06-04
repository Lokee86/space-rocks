extends RefCounted
class_name DevtoolsWindowActionContext

var devtools_window_controller
var command_context
var placement_context
var overlay_context


func configure(devtools_window_controller_ref, command_context_ref, placement_context_ref, overlay_context_ref) -> void:
	devtools_window_controller = devtools_window_controller_ref
	command_context = command_context_ref
	placement_context = placement_context_ref
	overlay_context = overlay_context_ref


func connect_signals() -> void:
	if !devtools_window_controller.toggle_invincible_requested.is_connected(Callable(command_context, "request_toggle_invincible")):
		devtools_window_controller.toggle_invincible_requested.connect(Callable(command_context, "request_toggle_invincible"))
	if !devtools_window_controller.toggle_infinite_lives_requested.is_connected(Callable(command_context, "request_toggle_infinite_lives")):
		devtools_window_controller.toggle_infinite_lives_requested.connect(Callable(command_context, "request_toggle_infinite_lives"))
	if !devtools_window_controller.toggle_freeze_world_requested.is_connected(Callable(command_context, "request_toggle_freeze_world")):
		devtools_window_controller.toggle_freeze_world_requested.connect(Callable(command_context, "request_toggle_freeze_world"))
	if !devtools_window_controller.toggle_freeze_player_requested.is_connected(Callable(command_context, "request_toggle_freeze_player")):
		devtools_window_controller.toggle_freeze_player_requested.connect(Callable(command_context, "request_toggle_freeze_player"))
	if !devtools_window_controller.placement_action_requested.is_connected(Callable(placement_context, "request_placement_action")):
		devtools_window_controller.placement_action_requested.connect(Callable(placement_context, "request_placement_action"))
	if !devtools_window_controller.respawn_player_requested.is_connected(Callable(command_context, "request_respawn_player")):
		devtools_window_controller.respawn_player_requested.connect(Callable(command_context, "request_respawn_player"))
	if devtools_window_controller.has_signal("set_score_requested"):
		var set_score_callable := Callable(command_context, "request_set_score")
		if !devtools_window_controller.is_connected("set_score_requested", set_score_callable):
			devtools_window_controller.connect("set_score_requested", set_score_callable)
	if devtools_window_controller.has_signal("add_score_requested"):
		var add_score_callable := Callable(command_context, "request_add_score")
		if !devtools_window_controller.is_connected("add_score_requested", add_score_callable):
			devtools_window_controller.connect("add_score_requested", add_score_callable)
	if devtools_window_controller.has_signal("set_lives_requested"):
		var set_lives_callable := Callable(command_context, "request_set_lives")
		if !devtools_window_controller.is_connected("set_lives_requested", set_lives_callable):
			devtools_window_controller.connect("set_lives_requested", set_lives_callable)
	if devtools_window_controller.has_signal("add_lives_requested"):
		var add_lives_callable := Callable(command_context, "request_add_lives")
		if !devtools_window_controller.is_connected("add_lives_requested", add_lives_callable):
			devtools_window_controller.connect("add_lives_requested", add_lives_callable)
	if devtools_window_controller.has_signal("clear_bullets_requested"):
		var clear_bullets_callable := Callable(command_context, "request_clear_bullets")
		if !devtools_window_controller.is_connected("clear_bullets_requested", clear_bullets_callable):
			devtools_window_controller.connect("clear_bullets_requested", clear_bullets_callable)
	if devtools_window_controller.has_signal("clear_asteroids_requested"):
		var clear_asteroids_callable := Callable(command_context, "request_clear_asteroids")
		if !devtools_window_controller.is_connected("clear_asteroids_requested", clear_asteroids_callable):
			devtools_window_controller.connect("clear_asteroids_requested", clear_asteroids_callable)
	if devtools_window_controller.has_signal("game_target_set_requested"):
		var game_target_set_callable := Callable(command_context, "request_set_game_target")
		if !devtools_window_controller.is_connected("game_target_set_requested", game_target_set_callable):
			devtools_window_controller.connect("game_target_set_requested", game_target_set_callable)
	if devtools_window_controller.has_signal("game_target_clear_requested"):
		var game_target_clear_callable := Callable(command_context, "request_clear_game_target")
		if !devtools_window_controller.is_connected("game_target_clear_requested", game_target_clear_callable):
			devtools_window_controller.connect("game_target_clear_requested", game_target_clear_callable)
	if devtools_window_controller.has_signal("show_server_hitboxes_changed"):
		var show_server_hitboxes_callable := Callable(overlay_context, "set_server_hitboxes_enabled")
		if !devtools_window_controller.is_connected("show_server_hitboxes_changed", show_server_hitboxes_callable):
			devtools_window_controller.connect("show_server_hitboxes_changed", show_server_hitboxes_callable)
