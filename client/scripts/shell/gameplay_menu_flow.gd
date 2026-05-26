extends RefCounted
class_name GameplayMenuFlow

signal quit_to_main_menu_requested

const Constants = preload("res://scripts/constants/constants.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")

const GAME_OVER_CONTAINER_PATH := "CenterContainer/GameOverContainer"
const GAME_OVER_MARGIN_CONTAINER_PATH := "CenterContainer/GameOverContainer/MarginContainer"
const CYCLE_VIEW_PATH := "CenterContainer/GameOverContainer/MarginContainer2/CycleView"
const GAME_MENU_PATH := "CenterContainer/GameOverContainer/MarginContainer2/GameMenu"

var hud: Control
var game_over_container
var game_over_margin_container
var cycle_view
var game_menu
var connection_service
var player: Player
var is_gameplay_paused := false
var is_game_over := false


func configure(hud_ref: Control, connection_service_ref = null, player_ref: Player = null) -> void:
	hud = hud_ref
	connection_service = connection_service_ref
	player = player_ref
	if hud != null:
		game_over_container = hud.get_node_or_null(GAME_OVER_CONTAINER_PATH)
		game_over_margin_container = hud.get_node_or_null(GAME_OVER_MARGIN_CONTAINER_PATH)
		cycle_view = hud.get_node_or_null(CYCLE_VIEW_PATH)
		game_menu = hud.get_node_or_null(GAME_MENU_PATH)
		_log_missing_live_pause_paths()
	if game_menu == null:
		return

	var resume_callable := Callable(self, "_on_resume_requested")
	if game_menu.has_signal("resume_requested") && !game_menu.resume_requested.is_connected(resume_callable):
		game_menu.resume_requested.connect(resume_callable)

	var quit_callable := Callable(self, "_on_quit_requested")
	if game_menu.has_signal("quit_requested") && !game_menu.quit_requested.is_connected(quit_callable):
		game_menu.quit_requested.connect(quit_callable)


func hide_menu() -> void:
	close_menu()


func close_menu() -> void:
	hide_live_pause_menu()
	is_gameplay_paused = false


func show_menu() -> bool:
	if game_menu == null:
		return false

	game_menu.configure_for_state(Constants.SESSION_MODE_SINGLE_PLAYER, false, "", false)
	if !show_live_pause_menu():
		return false
	is_gameplay_paused = true
	return true


func set_game_over() -> void:
	is_game_over = true
	is_gameplay_paused = false
	if game_menu != null:
		game_menu.configure_for_state(Constants.SESSION_MODE_SINGLE_PLAYER, true, "", false)


func set_alive() -> void:
	is_game_over = false


func can_open_live_pause_menu() -> bool:
	return !is_game_over


func show_live_pause_menu() -> bool:
	if !_has_live_pause_paths():
		return false

	game_over_container.show()
	game_over_margin_container.hide()
	cycle_view.hide()
	game_menu.show()
	return true


func hide_live_pause_menu() -> void:
	if game_menu != null:
		game_menu.hide()
	if game_over_container != null:
		game_over_container.hide()


func handle_open_menu_pressed(has_initial_spawn: bool) -> bool:
	if !Input.is_action_just_pressed("OpenMenu"):
		return false

	if is_menu_visible():
		close_menu()
		if connection_service != null:
			connection_service.send_resume_player_request()
		return true

	return open_live_pause_from_request(has_initial_spawn)


func open_live_pause_from_request(has_initial_spawn: bool) -> bool:
	if !has_initial_spawn:
		return false
	if !can_open_live_pause_menu():
		return false
	if !_has_live_pause_paths():
		return false
	if !show_menu():
		return false
	if player != null:
		player.set_afterburner_active(false)
	if connection_service != null:
		connection_service.send_pause_player_request()
	return true


func is_menu_visible() -> bool:
	return game_menu != null && game_menu.visible


func _on_resume_requested() -> void:
	close_menu()
	if connection_service != null:
		connection_service.send_resume_player_request()


func _on_quit_requested() -> void:
	close_menu()
	quit_to_main_menu_requested.emit()


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
