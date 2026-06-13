extends RefCounted
class_name MatchResultsFlow

const MatchResultWindowScene := preload("res://scenes/ui/dialogs/match_result_window.tscn")

signal replay_requested
signal return_to_lobby_requested
signal return_to_pregame_requested
signal quit_to_main_menu_requested

var mount_parent
var window
var current_session_mode := ""


func configure(mount_parent_ref: Node) -> void:
	mount_parent = mount_parent_ref


func show_results(session_mode: String, rows: Array = []) -> Control:
	clear()
	if mount_parent == null:
		return null

	window = MatchResultWindowScene.instantiate()
	mount_parent.add_child(window)
	if window != null:
		window.move_to_front()
		_connect_window_signal("lobby_replay_requested", Callable(self, "_on_lobby_replay_requested"))
		_connect_window_signal("menu_requested", Callable(self, "_on_menu_requested"))
		_connect_window_signal("quit_requested", Callable(self, "_on_quit_requested"))
		if window.has_method("configure_for_mode"):
			window.configure_for_mode(session_mode)
		if window.has_method("apply_rows"):
			window.apply_rows(rows)
	current_session_mode = session_mode
	return window


func clear() -> void:
	if is_instance_valid(window):
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


func _connect_window_signal(signal_name: StringName, handler: Callable) -> void:
	if window == null:
		return
	if window.has_signal(signal_name) and !window.is_connected(signal_name, handler):
		window.connect(signal_name, handler)
