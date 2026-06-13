extends RefCounted
class_name GameplayMenuFlow

signal quit_to_main_menu_requested
signal return_to_pregame_requested(session_mode: String)
signal return_to_lobby_requested
signal spectate_requested

const Constants = preload("res://scripts/generated/constants/constants.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")
const GameMenuScene := preload("res://scenes/ui/dialogs/game_menu.tscn")

const GAME_OVER_CONTAINER_PATH := "CenterContainer/GameOverContainer"
const GAME_OVER_MARGIN_CONTAINER_PATH := "CenterContainer/GameOverContainer/MarginContainer"
const CYCLE_VIEW_PATH := "CenterContainer/GameOverContainer/MarginContainer2/CycleView"
const GAME_MENU_PATH := "CenterContainer/GameOverContainer/MarginContainer2/GameMenu"

var hud: Control
var game_over_container
var game_over_margin_container
var cycle_view
var game_menu
var overlay_parent
var overlay_game_menu
var uses_match_over_overlay_menu := false
var connection_service
var player: Player
var spectate_menu_state
var session_context
var room_state_provider: Callable
var is_gameplay_paused := false
var is_game_over := false


func configure(
	hud_ref: Control,
	connection_service_ref = null,
	player_ref: Player = null,
	session_context_ref = null
) -> void:
	hud = hud_ref
	connection_service = connection_service_ref
	player = player_ref
	session_context = session_context_ref
	if hud != null:
		game_over_container = hud.get_node_or_null(GAME_OVER_CONTAINER_PATH)
		game_over_margin_container = hud.get_node_or_null(GAME_OVER_MARGIN_CONTAINER_PATH)
		cycle_view = hud.get_node_or_null(CYCLE_VIEW_PATH)
		game_menu = hud.get_node_or_null(GAME_MENU_PATH)
		_log_missing_live_pause_paths()
	hide_live_pause_menu()
	is_gameplay_paused = false
	is_game_over = false
	if game_menu == null:
		return

	var resume_callable := Callable(self, "_on_resume_requested")
	if game_menu.has_signal("resume_requested") && !game_menu.resume_requested.is_connected(resume_callable):
		game_menu.resume_requested.connect(resume_callable)

	var quit_callable := Callable(self, "_on_quit_requested")
	if game_menu.has_signal("quit_requested") && !game_menu.quit_requested.is_connected(quit_callable):
		game_menu.quit_requested.connect(quit_callable)

	var menu_callable := Callable(self, "_on_menu_requested")
	if game_menu.has_signal("menu_requested") && !game_menu.menu_requested.is_connected(menu_callable):
		game_menu.menu_requested.connect(menu_callable)

	var lobby_callable := Callable(self, "_on_lobby_requested")
	if game_menu.has_signal("lobby_requested") && !game_menu.lobby_requested.is_connected(lobby_callable):
		game_menu.lobby_requested.connect(lobby_callable)

	var spectate_callable := Callable(self, "_on_spectate_requested")
	if game_menu.has_signal("spectate_requested") && !game_menu.spectate_requested.is_connected(spectate_callable):
		game_menu.spectate_requested.connect(spectate_callable)


func configure_spectate_menu_state(spectate_menu_state_ref) -> void:
	spectate_menu_state = spectate_menu_state_ref


func configure_overlay_parent(parent: Node) -> void:
	overlay_parent = parent


func set_match_over_overlay_enabled(enabled: bool) -> void:
	uses_match_over_overlay_menu = enabled
	if enabled:
		_ensure_overlay_game_menu()
	if !enabled && overlay_game_menu != null:
		overlay_game_menu.hide()


func configure_lifecycle_routes(quit_route: Callable, return_to_lobby_route: Callable) -> void:
	if !quit_route.is_null() && !quit_to_main_menu_requested.is_connected(quit_route):
		quit_to_main_menu_requested.connect(quit_route)
	if (
		!return_to_lobby_route.is_null()
		&& !return_to_lobby_requested.is_connected(return_to_lobby_route)
	):
		return_to_lobby_requested.connect(return_to_lobby_route)


func configure_room_state_provider(provider: Callable) -> void:
	room_state_provider = provider


func reset() -> void:
	is_gameplay_paused = false
	is_game_over = false
	if game_menu != null:
		game_menu.hide()
	if game_over_container != null:
		game_over_container.hide()
	if game_over_margin_container != null:
		game_over_margin_container.hide()
	if cycle_view != null:
		cycle_view.hide()


func hide_menu() -> void:
	close_menu()


func close_menu() -> void:
	hide_live_pause_menu()
	is_gameplay_paused = false


func show_menu() -> bool:
	var active_game_menu: Control = _active_game_menu()
	if active_game_menu == null:
		return false

	var session_mode := _current_session_mode()
	var is_game_over_menu := uses_match_over_overlay_menu
	var room_state := _current_room_state() if is_game_over_menu else ""
	var has_spectate_targets := _has_spectate_targets() if is_game_over_menu else false
	_log_configure_for_state("show_menu", session_mode, is_game_over_menu, room_state, has_spectate_targets)
	active_game_menu.configure_for_state(session_mode, is_game_over_menu, room_state, has_spectate_targets)
	if uses_match_over_overlay_menu:
		active_game_menu.show()
		if overlay_game_menu != null:
			overlay_game_menu.move_to_front()
		is_gameplay_paused = false
		return true
	if !show_live_pause_menu():
		return false
	is_gameplay_paused = true
	return true


func set_game_over() -> void:
	is_game_over = true
	is_gameplay_paused = false
	_configure_game_menu_for_game_over()


func refresh_game_over_menu_state() -> void:
	if !is_game_over:
		return
	_configure_game_menu_for_game_over()


func set_alive() -> void:
	is_game_over = false
	is_gameplay_paused = false
	if game_menu != null:
		game_menu.hide()
	if game_over_container != null:
		game_over_container.hide()
	if game_over_margin_container != null:
		game_over_margin_container.hide()
	if cycle_view != null:
		cycle_view.hide()


func can_open_live_pause_menu() -> bool:
	return !is_game_over


func show_live_pause_menu() -> bool:
	if !_has_live_pause_paths():
		return false

	game_over_container.show()
	game_over_margin_container.hide()
	cycle_view.hide()
	var active_game_menu: Control = _active_game_menu()
	if active_game_menu != null:
		active_game_menu.show()
	return true


func show_single_player_game_over_menu() -> void:
	_configure_game_menu_for_game_over()


func show_spectating_menu() -> void:
	if !_has_live_pause_paths():
		return

	game_over_container.show()
	game_over_margin_container.show()
	cycle_view.hide()
	game_menu.show()
	var session_mode := _current_session_mode()
	var is_game_over_menu := true
	var room_state := _current_room_state()
	var has_spectate_targets := _has_spectate_targets()
	_log_configure_for_state(
		"show_spectating_menu",
		session_mode,
		is_game_over_menu,
		room_state,
		has_spectate_targets
	)
	game_menu.configure_for_state(
		session_mode,
		is_game_over_menu,
		room_state,
		has_spectate_targets
	)


func hide_live_pause_menu() -> void:
	if game_menu != null:
		game_menu.hide()
	if overlay_game_menu != null:
		overlay_game_menu.hide()
	if game_over_container != null:
		game_over_container.hide()


func handle_open_menu_pressed(has_initial_spawn: bool) -> bool:
	if !Input.is_action_just_pressed("OpenMenu"):
		return false

	if is_menu_visible():
		close_menu()
		if connection_service != null && !is_game_over:
			connection_service.send_pause_request()
		return true

	return open_live_pause_from_request(has_initial_spawn)


func open_live_pause_from_request(has_initial_spawn: bool) -> bool:
	if !has_initial_spawn:
		return false
	if !uses_match_over_overlay_menu && !_has_live_pause_paths():
		return false
	if !show_menu():
		return false
	if player != null:
		player.set_afterburner_active(false)
	if connection_service != null && !is_game_over:
		connection_service.send_pause_request()
	return true


func is_menu_visible() -> bool:
	var active_game_menu := _active_game_menu()
	return active_game_menu != null && active_game_menu.visible


func _on_resume_requested() -> void:
	close_menu()
	if connection_service != null:
		connection_service.send_pause_request()


func _on_quit_requested() -> void:
	close_menu()
	quit_to_main_menu_requested.emit()


func _on_menu_requested() -> void:
	close_menu()
	return_to_pregame_requested.emit(_current_session_mode())


func _on_lobby_requested() -> void:
	return_to_lobby_requested.emit()


func _on_spectate_requested() -> void:
	if game_menu != null:
		game_menu.hide()
	if cycle_view != null:
		cycle_view.show()
	is_gameplay_paused = false
	spectate_requested.emit()


func _has_spectate_targets() -> bool:
	if spectate_menu_state != null:
		return spectate_menu_state.has_spectate_targets()
	return false


func _current_session_mode() -> String:
	if session_context != null && !str(session_context.active_mode).is_empty():
		return session_context.active_mode
	return Constants.SESSION_MODE_SINGLE_PLAYER


func _current_room_state() -> String:
	if !room_state_provider.is_null():
		return str(room_state_provider.call())
	return ""


func _log_configure_for_state(
	path: String,
	session_mode: String,
	is_game_over_menu: bool,
	room_state: String,
	has_spectate_targets: bool
) -> void:
	ClientLogger.shell_debug(
		"Gameplay menu configure trace: path=%s session_mode=%s game_over=%s room_state=%s has_spectate_targets=%s"
		% [path, session_mode, is_game_over_menu, room_state, has_spectate_targets]
	)


func _has_live_pause_paths() -> bool:
	return (
		game_over_container != null
		&& game_over_margin_container != null
		&& cycle_view != null
		&& game_menu != null
	)


func _log_missing_live_pause_paths() -> void:
	_log_missing_path(GAME_OVER_CONTAINER_PATH, game_over_container)
	_log_missing_path(GAME_OVER_MARGIN_CONTAINER_PATH, game_over_margin_container)
	_log_missing_path(CYCLE_VIEW_PATH, cycle_view)
	_log_missing_path(GAME_MENU_PATH, game_menu)


func _log_missing_path(path: String, node) -> void:
	if node == null:
		ClientLogger.shell_error("GameplayMenuFlow missing expected HUD path: %s" % path)


func _active_game_menu() -> Control:
	if uses_match_over_overlay_menu:
		_ensure_overlay_game_menu()
		if overlay_game_menu != null:
			return overlay_game_menu
	return game_menu


func _ensure_overlay_game_menu() -> void:
	if overlay_game_menu != null || overlay_parent == null:
		return

	overlay_game_menu = GameMenuScene.instantiate()
	overlay_parent.add_child(overlay_game_menu)
	overlay_game_menu.hide()

	var resume_callable := Callable(self, "_on_resume_requested")
	if overlay_game_menu.has_signal("resume_requested") && !overlay_game_menu.resume_requested.is_connected(resume_callable):
		overlay_game_menu.resume_requested.connect(resume_callable)

	var quit_callable := Callable(self, "_on_quit_requested")
	if overlay_game_menu.has_signal("quit_requested") && !overlay_game_menu.quit_requested.is_connected(quit_callable):
		overlay_game_menu.quit_requested.connect(quit_callable)

	var menu_callable := Callable(self, "_on_menu_requested")
	if overlay_game_menu.has_signal("menu_requested") && !overlay_game_menu.menu_requested.is_connected(menu_callable):
		overlay_game_menu.menu_requested.connect(menu_callable)

	var lobby_callable := Callable(self, "_on_lobby_requested")
	if overlay_game_menu.has_signal("lobby_requested") && !overlay_game_menu.lobby_requested.is_connected(lobby_callable):
		overlay_game_menu.lobby_requested.connect(lobby_callable)

	var spectate_callable := Callable(self, "_on_spectate_requested")
	if overlay_game_menu.has_signal("spectate_requested") && !overlay_game_menu.spectate_requested.is_connected(spectate_callable):
		overlay_game_menu.spectate_requested.connect(spectate_callable)


func _configure_game_menu_for_game_over() -> void:
	if !_has_live_pause_paths():
		return

	var session_mode := _current_session_mode()
	var room_state := _current_room_state()
	var has_spectate_targets := _has_spectate_targets()
	_log_configure_for_state(
		"configure_game_menu_for_game_over",
		session_mode,
		true,
		room_state,
		has_spectate_targets
	)
	game_menu.configure_for_state(session_mode, true, room_state, has_spectate_targets)
