extends RefCounted
class_name MatchResultsFlow

const MatchResultWindowScene := preload("res://scenes/ui/dialogs/match_result_window.tscn")

signal replay_requested
signal return_to_lobby_requested
signal return_to_pregame_requested
signal quit_to_main_menu_requested

var mount_parent: Node = null
var window: MatchResultWindow = null
var current_session_mode := ""


func configure(mount_parent_ref: Node) -> void:
	mount_parent = mount_parent_ref


func show_results(session_mode: String, rows: Array = []) -> Control:
	clear()
	if mount_parent == null:
		return null

	window = MatchResultWindowScene.instantiate() as MatchResultWindow
	if window == null:
		push_error("Match result window scene must instantiate MatchResultWindow")
		return null
	mount_parent.add_child(window)
	window.move_to_front()
	var replay_callable := Callable(self, "_on_lobby_replay_requested")
	if !window.lobby_replay_requested.is_connected(replay_callable):
		window.lobby_replay_requested.connect(replay_callable)
	var menu_callable := Callable(self, "_on_menu_requested")
	if !window.menu_requested.is_connected(menu_callable):
		window.menu_requested.connect(menu_callable)
	var quit_callable := Callable(self, "_on_quit_requested")
	if !window.quit_requested.is_connected(quit_callable):
		window.quit_requested.connect(quit_callable)
	window.configure_for_mode(session_mode)
	window.apply_rows(rows)
	current_session_mode = session_mode
	return window


func clear() -> void:
	if is_instance_valid(window):
		var parent: Node = window.get_parent()
		if parent != null:
			parent.remove_child(window)
		window.queue_free()
	window = null


func _on_lobby_replay_requested() -> void:
	if current_session_mode == "multiplayer":
		return_to_lobby_requested.emit()
		return
	replay_requested.emit()


func _on_menu_requested() -> void:
	return_to_pregame_requested.emit()


func _on_quit_requested() -> void:
	quit_to_main_menu_requested.emit()
