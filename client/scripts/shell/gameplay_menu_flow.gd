extends RefCounted
class_name GameplayMenuFlow

signal quit_to_main_menu_requested

const Constants = preload("res://scripts/constants/constants.gd")

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
		game_over_container = hud.get_node_or_null("CenterContainer/GameOverContainer")
		game_over_margin_container = hud.get_node_or_null("CenterContainer/GameOverContainer/MarginContainer")
		cycle_view = hud.get_node_or_null("CenterContainer/GameOverContainer/MarginContainer2/CycleView")
		game_menu = hud.get_node_or_null("CenterContainer/GameOverContainer/MarginContainer2/GameMenu")
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


func show_menu() -> void:
	if game_menu != null:
		game_menu.configure_for_state(Constants.SESSION_MODE_SINGLE_PLAYER, false, "", false)
		show_live_pause_menu()
	is_gameplay_paused = true


func set_game_over() -> void:
	is_game_over = true
	is_gameplay_paused = false
	if game_menu != null:
		game_menu.configure_for_state(Constants.SESSION_MODE_SINGLE_PLAYER, true, "", false)


func set_alive() -> void:
	is_game_over = false


func can_open_live_pause_menu() -> bool:
	return !is_game_over


func show_live_pause_menu() -> void:
	if game_over_container != null:
		game_over_container.show()
	if game_over_margin_container != null:
		game_over_margin_container.hide()
	if cycle_view != null:
		cycle_view.hide()
	if game_menu != null:
		game_menu.show()


func hide_live_pause_menu() -> void:
	if game_menu != null:
		game_menu.hide()
	if game_over_container != null:
		game_over_container.hide()


func handle_open_menu_pressed(has_initial_spawn: bool) -> bool:
	if !Input.is_action_just_pressed("OpenMenu") || !has_initial_spawn:
		return false
	if !can_open_live_pause_menu():
		return false

	if is_menu_visible():
		close_menu()
		if connection_service != null:
			connection_service.send_resume_player_request()
		return true

	show_menu()
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
