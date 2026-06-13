extends RefCounted
class_name MatchEndFlow

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")

signal replay_requested
signal return_to_lobby_requested
signal return_to_pregame_requested
signal quit_to_main_menu_requested

var hud_flow
var menu_flow
var event_flow
var match_results_flow
var session_context
var match_result_provider: Callable
var room_state_provider: Callable
var room_match_over_handled := false


func configure(hud_flow_ref, menu_flow_ref, session_context_ref = null) -> void:
	hud_flow = hud_flow_ref
	menu_flow = menu_flow_ref
	session_context = session_context_ref


func configure_event_flow(event_flow_ref) -> void:
	event_flow = event_flow_ref


func configure_match_results_flow(match_results_flow_ref) -> void:
	match_results_flow = match_results_flow_ref
	if match_results_flow == null:
		return
	_connect_match_results_signal("replay_requested", Callable(self, "_on_replay_requested"))
	_connect_match_results_signal("return_to_lobby_requested", Callable(self, "_on_return_to_lobby_requested"))
	_connect_match_results_signal("return_to_pregame_requested", Callable(self, "_on_return_to_pregame_requested"))
	_connect_match_results_signal("quit_to_main_menu_requested", Callable(self, "_on_quit_to_main_menu_requested"))


func configure_room_state_provider(provider: Callable) -> void:
	room_state_provider = provider


func configure_match_result_provider(provider: Callable) -> void:
	match_result_provider = provider


func refresh_match_end_state() -> void:
	if room_state_provider.is_null():
		return

	var room_state := str(room_state_provider.call())
	if room_state == Constants.ROOM_STATE_GAME_OVER:
		handle_room_match_over()


func handle_local_player_eliminated(event: Dictionary) -> void:
	var lives := int(event.get(Packets.FIELD_LIVES, 0))
	if hud_flow != null && hud_flow.has_method("apply_lives"):
		hud_flow.apply_lives(lives)
	if hud_flow != null && hud_flow.has_method("set_game_over"):
		hud_flow.set_game_over()
	if menu_flow != null && menu_flow.has_method("set_game_over"):
		menu_flow.set_game_over()
	if event_flow != null && event_flow.has_method("play_game_over_sound_after_delay"):
		event_flow.play_game_over_sound_after_delay()


func handle_room_match_over() -> void:
	if room_match_over_handled:
		return
	room_match_over_handled = true
	hide_hud_for_match_over()
	if menu_flow != null && menu_flow.has_method("set_match_over_overlay_enabled"):
		menu_flow.set_match_over_overlay_enabled(true)
	if menu_flow != null && menu_flow.has_method("set_game_over"):
		menu_flow.set_game_over()
	if event_flow != null && event_flow.has_method("play_game_over_sound_after_delay"):
		event_flow.play_game_over_sound_after_delay()
	if match_results_flow != null && match_results_flow.has_method("show_results"):
		match_results_flow.show_results(_current_session_mode(), _current_match_result_rows())


func has_stale_dead_presentation() -> bool:
	if hud_flow != null && _flow_flag_true(hud_flow, "is_dead"):
		return true
	if hud_flow != null && _flow_flag_true(hud_flow, "is_game_over"):
		return true
	if menu_flow != null && _flow_flag_true(menu_flow, "is_game_over"):
		return true
	return false


func handle_alive_restored() -> void:
	if hud_flow != null && hud_flow.has_method("set_alive"):
		hud_flow.set_alive()
	if menu_flow != null && menu_flow.has_method("set_alive"):
		menu_flow.set_alive()


func reset() -> void:
	room_match_over_handled = false
	if hud_flow != null && hud_flow.has_method("clear_match_over_visibility_lock"):
		hud_flow.clear_match_over_visibility_lock()
	if menu_flow != null && menu_flow.has_method("set_match_over_overlay_enabled"):
		menu_flow.set_match_over_overlay_enabled(false)


func hide_hud_for_match_over() -> void:
	if hud_flow == null:
		return
	if hud_flow.has_method("hide_for_match_over"):
		hud_flow.hide_for_match_over()
		return
	if hud_flow.hud != null && hud_flow.hud.has_method("hide"):
		hud_flow.hud.hide()


func _current_session_mode() -> String:
	if session_context != null && !str(session_context.active_mode).is_empty():
		return session_context.active_mode
	return Constants.SESSION_MODE_SINGLE_PLAYER


func _current_match_result_rows() -> Array:
	if match_result_provider.is_null():
		return []

	var match_result = match_result_provider.call()
	if match_result == null:
		return []

	var players = []
	if match_result is Dictionary:
		players = match_result.get("players", [])
	elif match_result is Object and match_result.has_method("get"):
		players = match_result.get("players")

	if players == null:
		return []

	var rows: Array = []
	for player in players:
		if player is Dictionary:
			rows.append({
				"game_player_id": player.get("game_player_id", player.get("player_id", "Player")),
				"score": player.get("score", 0),
				"ship_deaths": player.get("ship_deaths", 0),
				"won": player.get("won", false),
				"kills": player.get("kills", 0),
			})
	return rows


func _flow_flag_true(flow, flag_name: String) -> bool:
	if flow == null:
		return false
	if !flow.has_method("get"):
		return false
	return bool(flow.get(flag_name))


func _on_replay_requested() -> void:
	replay_requested.emit()


func _on_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()


func _on_return_to_pregame_requested() -> void:
	return_to_pregame_requested.emit()


func _on_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()


func _connect_match_results_signal(signal_name: StringName, handler: Callable) -> void:
	if match_results_flow.has_signal(signal_name) and !match_results_flow.is_connected(signal_name, handler):
		match_results_flow.connect(signal_name, handler)
