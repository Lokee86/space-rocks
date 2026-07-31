extends VBoxContainer
class_name MultiplayerRoomSetupReadout

signal create_requested(config: Dictionary)
signal cancel_requested

const SESSION_SINGLE_PLAYER := "single_player"
const SESSION_MULTIPLAYER := "multiplayer"
const PRESET_ARCADE_SURVIVAL := "arcade_survival"
const PRESET_SCORE_ATTACK := "score_attack"
const LIVES_INFINITE := "infinite"

@onready var title_label: Label = %Title
@onready var game_mode_select = %GameModeSelect
@onready var lives_select = %LivesSelect
@onready var target_score_row: Control = %TargetScoreRow
@onready var target_score_select = %TargetScoreSelect
@onready var team_structure_row: Control = %TeamStructureRow
@onready var team_structure_select = %TeamStructureSelect
@onready var assignment_row: Control = %AssignmentRow
@onready var assignment_select = %AssignmentSelect
@onready var team_count_row: Control = %TeamCountRow
@onready var team_count_select = %TeamCountSelect
@onready var max_players_row: Control = %MaxPlayersRow
@onready var max_players_select = %MaxPlayersSelect
@onready var create_button: BaseButton = %CreateButton
@onready var create_action_label: Label = %CreateButton.get_node("Label") as Label
@onready var cancel_button: BaseButton = %CancelButton
@onready var status_label: Label = %StatusLabel

var session_mode := SESSION_MULTIPLAYER


func _ready() -> void:
	game_mode_select.replace_items([
		{"label": "ARCADE SURVIVAL", "value": PRESET_ARCADE_SURVIVAL},
		{"label": "SCORE ATTACK", "value": PRESET_SCORE_ATTACK},
	], PRESET_ARCADE_SURVIVAL)
	lives_select.replace_items([
		{"label": "1 LIFE", "value": 1},
		{"label": "2 LIVES", "value": 2},
		{"label": "3 LIVES", "value": 3},
		{"label": "5 LIVES", "value": 5},
		{"label": "INFINITE", "value": LIVES_INFINITE},
	], 3)
	target_score_select.replace_items([
		{"label": "500", "value": 500},
		{"label": "1,000", "value": 1000},
		{"label": "2,500", "value": 2500},
		{"label": "5,000", "value": 5000},
	], 1000)
	team_structure_select.replace_items([
		{"label": "FREE-FOR-ALL", "value": "ffa"},
		{"label": "CO-OP", "value": "co_op"},
		{"label": "CUSTOM TEAMS", "value": "custom"},
		{"label": "AUTO-BALANCED", "value": "auto_balanced"},
	], "ffa")
	assignment_select.replace_items([
		{"label": "OWNER ASSIGNED", "value": "owner_assigned"},
		{"label": "PLAYER SELECTED", "value": "player_selected"},
	], "owner_assigned")
	var team_count_items := []
	for count in range(2, 9):
		team_count_items.append({"label": "%d TEAMS" % count, "value": count})
	team_count_select.replace_items(team_count_items, 2)
	var max_player_items := []
	for count in range(2, 9):
		max_player_items.append({"label": "%d PLAYERS" % count, "value": count})
	max_players_select.replace_items(max_player_items, 8)
	game_mode_select.item_selected.connect(_on_selection_changed)
	team_structure_select.item_selected.connect(_on_selection_changed)
	create_button.pressed.connect(_on_create_pressed)
	cancel_button.pressed.connect(_on_cancel_pressed)
	_apply_session_mode()
	_refresh_visibility()


func configure_single_player() -> void:
	session_mode = SESSION_SINGLE_PLAYER
	if is_node_ready():
		_apply_session_mode()
		_refresh_visibility()


func configure_multiplayer() -> void:
	session_mode = SESSION_MULTIPLAYER
	if is_node_ready():
		_apply_session_mode()
		_refresh_visibility()


func current_config() -> Dictionary:
	var preset_id := str(game_mode_select.selected_value(PRESET_ARCADE_SURVIVAL))
	var lives_value = lives_select.selected_value(3)
	var infinite_lives := str(lives_value) == LIVES_INFINITE
	var config := {
		"preset_id": preset_id,
		"starting_lives": 0 if infinite_lives else int(lives_value),
		"infinite_lives": infinite_lives,
		"target_score": int(target_score_select.selected_value(1000)) if preset_id == PRESET_SCORE_ATTACK else 0,
	}
	if session_mode == SESSION_SINGLE_PLAYER:
		return config

	var structure := str(team_structure_select.selected_value("ffa"))
	config["team_structure"] = structure
	config["team_assignment_mode"] = str(assignment_select.selected_value("owner_assigned")) if structure == "custom" else ""
	config["team_count"] = int(team_count_select.selected_value(2)) if structure == "auto_balanced" else 0
	config["max_players"] = int(max_players_select.selected_value(8))
	return config


func _on_selection_changed(_index: int) -> void:
	_refresh_visibility()


func _apply_session_mode() -> void:
	var single_player := session_mode == SESSION_SINGLE_PLAYER
	title_label.text = "SINGLE PLAYER CONFIGURATION" if single_player else "MATCH CONFIGURATION"
	create_action_label.text = "START GAME" if single_player else "CREATE ROOM"
	team_structure_row.visible = not single_player
	max_players_row.visible = not single_player


func _refresh_visibility() -> void:
	var preset_id := str(game_mode_select.selected_value(PRESET_ARCADE_SURVIVAL))
	target_score_row.visible = preset_id == PRESET_SCORE_ATTACK
	var multiplayer := session_mode == SESSION_MULTIPLAYER
	var structure := str(team_structure_select.selected_value("ffa"))
	team_structure_row.visible = multiplayer
	assignment_row.visible = multiplayer and structure == "custom"
	team_count_row.visible = multiplayer and structure == "auto_balanced"
	max_players_row.visible = multiplayer


func _on_create_pressed() -> void:
	create_requested.emit(current_config())


func set_status(message: String) -> void:
	if status_label != null:
		status_label.text = message


func set_pending(pending: bool) -> void:
	if create_button != null:
		create_button.disabled = pending
	if cancel_button != null:
		cancel_button.disabled = pending
	for selector in [game_mode_select, lives_select, target_score_select, team_structure_select, assignment_select, team_count_select, max_players_select]:
		if selector != null:
			selector.disabled = pending


func _on_cancel_pressed() -> void:
	cancel_requested.emit()
