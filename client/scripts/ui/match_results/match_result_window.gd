extends Control
class_name MatchResultWindow

const PlayerScoreRowScene := preload("res://scenes/ui/elements/player_score_row.tscn")
const TeamPresentation := preload("res://scripts/teams/team_presentation.gd")
const InterfaceLabel := preload("res://assets/ui/label_settings/interface_label.tres")
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
	(%ReplayLabel as Control).visible = not _is_multiplayer
	_apply_replay_availability()


func set_replay_available(available: bool) -> void:
	_replay_available = available
	_apply_replay_availability()


func clear_rows() -> void:
	for child in (%ScoreContainer as Control).get_children():
		child.queue_free()


func apply_rows(rows: Array) -> void:
	clear_rows()
	var normalized := rows.duplicate(true)
	var team_mode := normalized.any(func(row): return row is Dictionary and not str(row.get("team_id", "")).is_empty())
	(%TeamHeader as Control).visible = team_mode
	if team_mode:
		_apply_team_rows(normalized)
	else:
		normalized.sort_custom(_compare_player_rows)
		for row in normalized:
			_add_player_row(row)


func _apply_team_rows(rows: Array) -> void:
	var teams := {}
	for row in rows:
		if not (row is Dictionary):
			continue
		var team_id := str(row.get("team_id", ""))
		if not teams.has(team_id):
			teams[team_id] = {"score": 0, "players": []}
		teams[team_id]["score"] = int(teams[team_id]["score"]) + int(row.get("score", 0))
		teams[team_id]["players"].append(row)
	var team_ids := teams.keys()
	team_ids.sort_custom(func(left, right):
		var left_score := int(teams[left]["score"])
		var right_score := int(teams[right]["score"])
		return left_score > right_score if left_score != right_score else str(left) < str(right)
	)
	for team_id in team_ids:
		_add_team_header(str(team_id), int(teams[team_id]["score"]))
		var players: Array = teams[team_id]["players"]
		players.sort_custom(_compare_player_rows)
		for row in players:
			_add_player_row(row)


func _compare_player_rows(left, right) -> bool:
	var left_score := int(left.get("score", 0))
	var right_score := int(right.get("score", 0))
	if left_score != right_score:
		return left_score > right_score
	var left_deaths := int(left.get("ship_deaths", 0))
	var right_deaths := int(right.get("ship_deaths", 0))
	if left_deaths != right_deaths:
		return left_deaths < right_deaths
	return str(left.get("game_player_id", "")) < str(right.get("game_player_id", ""))


func _add_team_header(team_id: String, team_score: int) -> void:
	var header := HBoxContainer.new()
	header.custom_minimum_size.y = 36
	var swatch := ColorRect.new()
	swatch.custom_minimum_size = Vector2(18, 18)
	swatch.size_flags_vertical = Control.SIZE_SHRINK_CENTER
	swatch.color = TeamPresentation.color(team_id)
	swatch.mouse_filter = Control.MOUSE_FILTER_IGNORE
	header.add_child(swatch)
	var label := Label.new()
	label.label_settings = InterfaceLabel
	label.text = "%s  —  %d" % [TeamPresentation.display_name(team_id), team_score]
	header.add_child(label)
	(%ScoreContainer as Control).add_child(header)


func _add_player_row(row) -> void:
	if not (row is Dictionary):
		return
	var score_row_instance: Node = PlayerScoreRowScene.instantiate()
	var score_row := score_row_instance as PlayerScoreRow
	if score_row == null:
		ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
			"Player score row scene must instantiate its presentation root",
			{},
			{"subsystem": "match_results", "failure_mode": "wrong_row_type", "resource_kind": "scene", "expected_type": "PlayerScoreRow", "actual_type": score_row_instance.get_class(), "resource_path": PlayerScoreRowScene.resource_path}
		)
		score_row_instance.queue_free()
		return
	score_row.apply_row(row)
	(%ScoreContainer as Control).add_child(score_row)


func _on_lobby_replay_pressed() -> void:
	lobby_replay_requested.emit()
func _on_menu_pressed() -> void:
	menu_requested.emit()
func _on_quit_pressed() -> void:
	quit_requested.emit()


func _connect_button(node_name: String, method_name: String) -> void:
	var button := find_child(node_name, true, false) as BaseButton
	if button != null and not button.pressed.is_connected(Callable(self, method_name)):
		button.pressed.connect(Callable(self, method_name))


func _apply_replay_availability() -> void:
	var replay_button := find_child("LobbyReplayButton", true, false) as BaseButton
	if replay_button != null:
		replay_button.disabled = not _is_multiplayer and not _replay_available
