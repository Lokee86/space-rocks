extends Control
class_name MatchResultWindow

const PlayerScoreRowScene := preload("res://scenes/ui/elements/player_score_row.tscn")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

signal lobby_replay_requested
signal menu_requested
signal quit_requested

var _is_multiplayer := false
var _replay_available := true

func _ready() -> void:
	_connect_button("LobbyReplayButton", "_on_lobby_replay_pressed")
	_connect_button("MenuButton", "_on_menu_pressed")
	_connect_button("QuitButton", "_on_quit_pressed")


func configure_for_mode(session_mode: String) -> void:
	_is_multiplayer = str(session_mode) == "multiplayer"
	(%LobbyLabel as Control).visible = _is_multiplayer
	(%ReplayLabel as Control).visible = !_is_multiplayer
	_apply_replay_availability()


func set_replay_available(available: bool) -> void:
	_replay_available = available
	_apply_replay_availability()


func clear_rows() -> void:
	var score_container := %ScoreContainer as Control
	for child in score_container.get_children():
		if child is PlayerScoreRow:
			child.queue_free()


func apply_rows(rows: Array) -> void:
	clear_rows()

	var score_container := %ScoreContainer as Control
	for row in rows:
		var score_row_instance: Node = PlayerScoreRowScene.instantiate()
		var score_row := score_row_instance as PlayerScoreRow
		if score_row == null:
			ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Player score row scene must instantiate its presentation root",
		{},
		{
			"subsystem": "match_results",
			"failure_mode": "wrong_row_type",
			"resource_kind": "scene",
			"expected_type": "PlayerScoreRow",
			"actual_type": score_row_instance.get_class(),
			"resource_path": PlayerScoreRowScene.resource_path,
		}
	)
			score_row_instance.queue_free()
			continue
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


func _apply_replay_availability() -> void:
	var replay_button := find_child("LobbyReplayButton", true, false) as BaseButton
	if replay_button != null:
		replay_button.disabled = !_is_multiplayer and !_replay_available
