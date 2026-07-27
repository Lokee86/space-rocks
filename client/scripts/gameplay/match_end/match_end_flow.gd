extends RefCounted
class_name MatchEndFlow

const Constants := preload("res://scripts/generated/constants/constants.gd")

signal replay_requested
signal return_to_lobby_requested
signal return_to_pregame_requested
signal quit_to_main_menu_requested

var hud_flow: GameplayHudFlow = null
var menu_flow: GameplayMenuFlow = null
var event_flow: GameplayEventFlow = null
var match_results_flow: MatchResultsFlow = null
var session_context: ClientSessionContext = null
var match_result_provider: Callable
var room_state_provider: Callable
var room_match_over_handled := false
var local_player_eliminated_handled := false


func configure(
	hud_flow_ref: GameplayHudFlow,
	menu_flow_ref: GameplayMenuFlow,
	session_context_ref: ClientSessionContext = null
) -> void:
	hud_flow = hud_flow_ref
	menu_flow = menu_flow_ref
	session_context = session_context_ref


func configure_event_flow(event_flow_ref: GameplayEventFlow) -> void:
	event_flow = event_flow_ref


func configure_match_results_flow(match_results_flow_ref: MatchResultsFlow) -> void:
	match_results_flow = match_results_flow_ref
	if match_results_flow == null:
		return
	var replay_callable := Callable(self, "_on_replay_requested")
	if !match_results_flow.replay_requested.is_connected(replay_callable):
		match_results_flow.replay_requested.connect(replay_callable)
	var lobby_callable := Callable(self, "_on_return_to_lobby_requested")
	if !match_results_flow.return_to_lobby_requested.is_connected(lobby_callable):
		match_results_flow.return_to_lobby_requested.connect(lobby_callable)
	var pregame_callable := Callable(self, "_on_return_to_pregame_requested")
	if !match_results_flow.return_to_pregame_requested.is_connected(pregame_callable):
		match_results_flow.return_to_pregame_requested.connect(pregame_callable)
	var quit_callable := Callable(self, "_on_quit_to_main_menu_requested")
	if !match_results_flow.quit_to_main_menu_requested.is_connected(quit_callable):
		match_results_flow.quit_to_main_menu_requested.connect(quit_callable)


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


func handle_local_player_eliminated(lives: int) -> void:
	if local_player_eliminated_handled || room_match_over_handled:
		return
	if _current_session_mode() != Constants.SESSION_MODE_MULTIPLAYER:
		return
	local_player_eliminated_handled = true
	if hud_flow != null:
		hud_flow.apply_lives(lives)
	if hud_flow != null:
		hud_flow.set_game_over()
	if menu_flow != null:
		menu_flow.set_game_over()
	if event_flow != null:
		event_flow.play_game_over_sound_after_delay()


func handle_room_match_over() -> void:
	if room_match_over_handled:
		return
	room_match_over_handled = true
	hide_hud_for_match_over()
	if menu_flow != null:
		menu_flow.set_match_over_overlay_enabled(true)
	if menu_flow != null:
		menu_flow.set_game_over()
	if event_flow != null:
		event_flow.play_game_over_sound_after_delay()
	if match_results_flow != null:
		match_results_flow.show_results(_current_session_mode(), _current_match_result_rows())


func has_stale_dead_presentation() -> bool:
	if hud_flow != null && hud_flow.is_dead:
		return true
	if hud_flow != null && hud_flow.is_game_over:
		return true
	if menu_flow != null && menu_flow.is_game_over:
		return true
	return false


func handle_alive_restored() -> void:
	local_player_eliminated_handled = false
	if hud_flow != null:
		hud_flow.set_alive()
	if menu_flow != null:
		menu_flow.set_alive()


func is_game_over() -> bool:
	return room_match_over_handled || local_player_eliminated_handled


func reset() -> void:
	room_match_over_handled = false
	local_player_eliminated_handled = false
	if hud_flow != null:
		hud_flow.clear_match_over_visibility_lock()
	if menu_flow != null:
		menu_flow.set_match_over_overlay_enabled(false)


func hide_hud_for_match_over() -> void:
	if hud_flow == null:
		return
	hud_flow.hide_for_match_over()


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
	if not match_result is Dictionary:
		return []
	players = match_result.get("players", [])

	if not players is Array or players.is_empty():
		return []

	var rows: Array = []
	for player in players:
		if player is Dictionary:
			var row := {
				"game_player_id": player.get("game_player_id", player.get("player_id", "Player")),
				"score": player.get("score", 0),
				"ship_deaths": player.get("ship_deaths", 0),
				"won": player.get("won", false),
			}
			var team_id := str(player.get("team_id", ""))
			if not team_id.is_empty():
				row["team_id"] = team_id
			rows.append(row)
	return rows



func _on_replay_requested() -> void:
	replay_requested.emit()


func _on_return_to_lobby_requested() -> void:
	return_to_lobby_requested.emit()


func _on_return_to_pregame_requested() -> void:
	return_to_pregame_requested.emit()


func _on_quit_to_main_menu_requested() -> void:
	quit_to_main_menu_requested.emit()
