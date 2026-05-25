extends RefCounted

const Packets = preload("res://scripts/networking/packets.gd")

var hud_controller
var network_client
var player
var session_mode_provider: Callable
var is_game_over_provider: Callable
var current_room_state_provider: Callable
var is_room_game_over_provider: Callable
var has_spectate_targets_provider: Callable
var return_to_menu_requested: Callable
var start_spectating: Callable
var open_menu_input_armed := false
var is_gameplay_paused := false


func configure(
	hud_controller_object,
	network_client_object,
	player_node,
	session_mode_provider_callback: Callable,
	is_game_over_provider_callback: Callable,
	current_room_state_provider_callback: Callable,
	is_room_game_over_provider_callback: Callable,
	has_spectate_targets_provider_callback: Callable,
	return_to_menu_requested_callback: Callable,
	start_spectating_callback: Callable
) -> void:
	hud_controller = hud_controller_object
	network_client = network_client_object
	player = player_node
	session_mode_provider = session_mode_provider_callback
	is_game_over_provider = is_game_over_provider_callback
	current_room_state_provider = current_room_state_provider_callback
	is_room_game_over_provider = is_room_game_over_provider_callback
	has_spectate_targets_provider = has_spectate_targets_provider_callback
	return_to_menu_requested = return_to_menu_requested_callback
	start_spectating = start_spectating_callback


func connect_game_menu_signals(game_menu) -> void:
	if game_menu == null:
		return

	if game_menu.has_signal("lobby_requested"):
		var lobby_callable := Callable(self, "on_lobby_requested")
		if !game_menu.lobby_requested.is_connected(lobby_callable):
			game_menu.lobby_requested.connect(lobby_callable)
	if game_menu.has_signal("spectate_requested"):
		var spectate_callable := Callable(self, "on_spectate_requested")
		if !game_menu.spectate_requested.is_connected(spectate_callable):
			game_menu.spectate_requested.connect(spectate_callable)
	var resume_callable := Callable(self, "on_resume_requested")
	if !game_menu.resume_requested.is_connected(resume_callable):
		game_menu.resume_requested.connect(resume_callable)
	var quit_callable := Callable(self, "on_quit_requested")
	if !game_menu.quit_requested.is_connected(quit_callable):
		game_menu.quit_requested.connect(quit_callable)


func show_game_menu() -> void:
	var game_menu = hud_controller.get_game_menu()
	if game_menu == null:
		return

	refresh_game_menu_state(game_menu)
	connect_game_menu_signals(game_menu)
	hud_controller.show_game_menu()


func refresh_game_menu_state(game_menu) -> void:
	if game_menu == null:
		return
	if game_menu.has_method("configure_for_state"):
		game_menu.configure_for_state(
			session_mode_provider.call(),
			is_game_over_provider.call(),
			current_room_state_provider.call(),
			has_spectate_targets_provider.call()
		)


func is_game_menu_open() -> bool:
	return hud_controller != null && hud_controller.is_game_menu_visible()


func hide_game_menu() -> void:
	if hud_controller == null:
		return

	hud_controller.hide_game_menu()


func update_open_menu_input_armed(has_initial_spawn: bool) -> void:
	if open_menu_input_armed || !has_initial_spawn:
		return
	if !Input.is_action_pressed("OpenMenu"):
		open_menu_input_armed = true


func disarm_open_menu_input() -> void:
	open_menu_input_armed = false


func can_pause_server_gameplay(has_initial_spawn: bool) -> bool:
	return network_client.is_connected_to_server() && has_initial_spawn && !hud_controller.is_dead


func set_gameplay_paused(paused: bool) -> void:
	if is_gameplay_paused == paused:
		if !paused:
			hide_game_menu()
			hud_controller.set_suspended(false)
		return

	is_gameplay_paused = paused
	hud_controller.set_suspended(paused)
	if paused:
		player.set_afterburner_active(false)
		network_client.send_packet(Packets.pause_player_packet())
	else:
		hide_game_menu()
		network_client.send_packet(Packets.resume_player_packet())


func handle_open_menu_pressed(has_initial_spawn: bool) -> bool:
	if !Input.is_action_just_pressed("OpenMenu"):
		return false
	if should_block_open_menu_for_game_over():
		return false
	if !open_menu_input_armed && !hud_controller.is_game_over:
		return false

	if is_game_menu_open():
		close_game_menu()
	else:
		open_game_menu(has_initial_spawn)
	return true


func open_game_menu(has_initial_spawn: bool) -> void:
	show_game_menu()
	if can_pause_server_gameplay(has_initial_spawn):
		set_gameplay_paused(true)


func close_game_menu() -> void:
	if is_gameplay_paused:
		set_gameplay_paused(false)
	else:
		hide_game_menu()
		hud_controller.set_suspended(false)


func should_block_open_menu_for_game_over() -> bool:
	return !_is_multiplayer_session() && is_game_over_provider.call()


func _is_multiplayer_session() -> bool:
	return str(session_mode_provider.call()).strip_edges().to_lower() == "multiplayer"


func on_resume_requested() -> void:
	close_game_menu()


func on_lobby_requested() -> void:
	if !_is_multiplayer_session():
		return
	if network_client == null || !network_client.is_connected_to_server():
		return

	network_client.send_return_to_lobby_request()


func on_spectate_requested() -> void:
	if !_is_multiplayer_session() || is_room_game_over_provider.call():
		return
	if !start_spectating.call():
		show_game_menu()


func on_quit_requested() -> void:
	is_gameplay_paused = false
	hud_controller.set_suspended(false)
	if _is_multiplayer_session() && network_client != null && network_client.is_connected_to_server():
		network_client.send_leave_room_request()
	return_to_menu_requested.call()
