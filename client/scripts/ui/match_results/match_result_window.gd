extends Control
class_name MatchResultWindow

const PlayerScoreRowScene := preload("res://scenes/ui/elements/player_score_row.tscn")

signal lobby_replay_requested
signal menu_requested
signal quit_requested


func _ready() -> void:
	_connect_button("LobbyReplayButton", "_on_lobby_replay_pressed")
	_connect_button("MenuButton", "_on_menu_pressed")
	_connect_button("QuitButton", "_on_quit_pressed")


func configure_for_mode(session_mode: String) -> void:
	var is_multiplayer := str(session_mode) == "multiplayer"
	(%LobbyLabel as Control).visible = is_multiplayer
	(%ReplayLabel as Control).visible = !is_multiplayer


func clear_rows() -> void:
	var score_container := %ScoreContainer as Control
	for child in score_container.get_children():
		if child is PlayerScoreRow:
			child.queue_free()


func apply_rows(rows: Array) -> void:
	clear_rows()

	var score_container := %ScoreContainer as Control
	for row in rows:
		var score_row := PlayerScoreRowScene.instantiate()
		if score_row.has_method("apply_row"):
			score_row.apply_row(row)
		score_container.add_child(score_row)


func _on_lobby_replay_pressed() -> void:
	lobby_replay_requested.emit()


func _on_menu_pressed() -> void:
	menu_requested.emit()


func _on_quit_pressed() -> void:
	quit_requested.emit()


func _connect_button(node_name: String, method_name: String) -> void:
	var button := find_child(node_name, true, false) as BaseButton
	if button != null && !button.pressed.is_connected(Callable(self, method_name)):
		button.pressed.connect(Callable(self, method_name))
