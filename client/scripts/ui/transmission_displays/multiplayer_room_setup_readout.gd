extends VBoxContainer
class_name MultiplayerRoomSetupReadout

signal create_requested(config: Dictionary)
signal cancel_requested

const SESSION_SINGLE_PLAYER := "single_player"
const SESSION_MULTIPLAYER := "multiplayer"
const PRESET_ARCADE_SURVIVAL := "arcade_survival"
const PRESET_SCORE_ATTACK := "score_attack"
const PRESET_DEATHMATCH := "deathmatch"
const LIVES_INFINITE := "infinite"
const TARGET_CUSTOM := "custom"
const CUSTOM_TARGET_ERROR := "ENTER A POSITIVE CUSTOM TARGET"

@onready var title_label: Label = %Title
@onready var game_mode_select = %GameModeSelect
@onready var lives_row: Control = $Content/Fields/LivesRow
@onready var lives_select = %LivesSelect
@onready var target_score_row: Control = %TargetScoreRow
@onready var target_label: Label = $Content/Fields/TargetScoreRow/Row/Label
@onready var target_score_select = %TargetScoreSelect
@onready var custom_target_score_row: Control = %CustomTargetScoreRow
@onready var custom_target_score_input: LineEdit = %CustomTargetScoreInput
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
var is_pending := false
var target_options_preset := ""


func _ready() -> void:
	_refresh_game_mode_options()
	lives_select.replace_items([
		{"label": "1 LIFE", "value": 1},
		{"label": "2 LIVES", "value": 2},
		{"label": "3 LIVES", "value": 3},
		{"label": "5 LIVES", "value": 5},
		{"label": "INFINITE", "value": LIVES_INFINITE},
	], 3)
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
	target_score_select.item_selected.connect(_on_selection_changed)
	team_structure_select.item_selected.connect(_on_selection_changed)
	custom_target_score_input.text_changed.connect(_on_custom_target_changed)
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
	var deathmatch := preset_id == PRESET_DEATHMATCH
	var lives_value = lives_select.selected_value(3)
	var infinite_lives := deathmatch or str(lives_value) == LIVES_INFINITE
	var config := {
		"preset_id": preset_id,
		"starting_lives": 0 if infinite_lives else int(lives_value),
		"infinite_lives": infinite_lives,
		"target_score": _selected_target_score(),
		"target_kills": _selected_target_kills(),
	}
	if session_mode == SESSION_SINGLE_PLAYER:
		if deathmatch:
			config["max_players"] = int(max_players_select.selected_value(8))
		return config

	var structure := "ffa" if deathmatch else str(team_structure_select.selected_value("ffa"))
	config["team_structure"] = structure
	config["team_assignment_mode"] = str(assignment_select.selected_value("owner_assigned")) if structure == "custom" else ""
	config["team_count"] = int(team_count_select.selected_value(2)) if structure == "auto_balanced" else 0
	config["max_players"] = int(max_players_select.selected_value(8))
	return config


func _on_selection_changed(_index: int) -> void:
	_refresh_visibility()


func _on_custom_target_changed(_value: String) -> void:
	_refresh_create_state()


func _apply_session_mode() -> void:
	var single_player := session_mode == SESSION_SINGLE_PLAYER
	title_label.text = "SINGLE PLAYER CONFIGURATION" if single_player else "MATCH CONFIGURATION"
	create_action_label.text = "START GAME" if single_player else "CREATE ROOM"
	_refresh_game_mode_options()


func _refresh_game_mode_options() -> void:
	if game_mode_select == null:
		return
	var selected := str(game_mode_select.selected_value(PRESET_ARCADE_SURVIVAL))
	var items := [
		{"label": "ARCADE SURVIVAL", "value": PRESET_ARCADE_SURVIVAL},
		{"label": "SCORE ATTACK", "value": PRESET_SCORE_ATTACK},
		{"label": "DEATHMATCH", "value": PRESET_DEATHMATCH},
	]
	game_mode_select.replace_items(items, selected)


func _refresh_visibility() -> void:
	var preset_id := str(game_mode_select.selected_value(PRESET_ARCADE_SURVIVAL))
	var score_attack := preset_id == PRESET_SCORE_ATTACK
	var deathmatch := preset_id == PRESET_DEATHMATCH
	var has_target := score_attack or deathmatch
	_configure_target_options(preset_id)
	lives_row.visible = not deathmatch
	target_score_row.visible = has_target
	custom_target_score_row.visible = has_target and str(target_score_select.selected_value(_default_target(preset_id))) == TARGET_CUSTOM
	var multiplayer := session_mode == SESSION_MULTIPLAYER
	var structure := str(team_structure_select.selected_value("ffa"))
	team_structure_row.visible = multiplayer and not deathmatch
	assignment_row.visible = multiplayer and not deathmatch and structure == "custom"
	team_count_row.visible = multiplayer and not deathmatch and structure == "auto_balanced"
	max_players_row.visible = multiplayer or (session_mode == SESSION_SINGLE_PLAYER and deathmatch)
	_refresh_create_state()


func _configure_target_options(preset_id: String) -> void:
	if preset_id != PRESET_SCORE_ATTACK and preset_id != PRESET_DEATHMATCH:
		target_options_preset = ""
		return
	if target_options_preset == preset_id:
		return
	target_options_preset = preset_id
	custom_target_score_input.clear()
	if preset_id == PRESET_DEATHMATCH:
		target_label.text = "KILL TARGET"
		custom_target_score_input.placeholder_text = "ENTER KILLS"
		target_score_select.replace_items([
			{"label": "5 KILLS", "value": 5},
			{"label": "10 KILLS", "value": 10},
			{"label": "15 KILLS", "value": 15},
			{"label": "25 KILLS", "value": 25},
			{"label": "50 KILLS", "value": 50},
			{"label": "CUSTOM", "value": TARGET_CUSTOM},
		], 10)
		return
	target_label.text = "TARGET SCORE"
	custom_target_score_input.placeholder_text = "ENTER SCORE"
	target_score_select.replace_items([
		{"label": "25,000", "value": 25000},
		{"label": "50,000", "value": 50000},
		{"label": "75,000", "value": 75000},
		{"label": "100,000", "value": 100000},
		{"label": "125,000", "value": 125000},
		{"label": "150,000", "value": 150000},
		{"label": "CUSTOM", "value": TARGET_CUSTOM},
	], 25000)


func _default_target(preset_id: String) -> int:
	return 10 if preset_id == PRESET_DEATHMATCH else 25000


func _refresh_create_state() -> void:
	var target_valid := _has_valid_target()
	create_button.disabled = is_pending or not target_valid
	if not target_valid:
		status_label.text = CUSTOM_TARGET_ERROR
	elif status_label.text == CUSTOM_TARGET_ERROR:
		status_label.text = ""


func _has_valid_target() -> bool:
	var preset_id := str(game_mode_select.selected_value(PRESET_ARCADE_SURVIVAL))
	if preset_id != PRESET_SCORE_ATTACK and preset_id != PRESET_DEATHMATCH:
		return true
	var selected_target = target_score_select.selected_value(_default_target(preset_id))
	if str(selected_target) != TARGET_CUSTOM:
		return int(selected_target) > 0
	return _custom_target_value() > 0


func _selected_target_score() -> int:
	if str(game_mode_select.selected_value(PRESET_ARCADE_SURVIVAL)) != PRESET_SCORE_ATTACK:
		return 0
	return _selected_target_value(PRESET_SCORE_ATTACK)


func _selected_target_kills() -> int:
	if str(game_mode_select.selected_value(PRESET_ARCADE_SURVIVAL)) != PRESET_DEATHMATCH:
		return 0
	return _selected_target_value(PRESET_DEATHMATCH)


func _selected_target_value(preset_id: String) -> int:
	var selected_target = target_score_select.selected_value(_default_target(preset_id))
	if str(selected_target) != TARGET_CUSTOM:
		return int(selected_target)
	return _custom_target_value()


func _custom_target_value() -> int:
	var normalized := custom_target_score_input.text.strip_edges().replace(",", "").replace("_", "").replace(" ", "")
	if not normalized.is_valid_int():
		return 0
	var target := int(normalized)
	return target if target > 0 else 0


func _on_create_pressed() -> void:
	if not _has_valid_target():
		_refresh_create_state()
		return
	create_requested.emit(current_config())


func set_status(message: String) -> void:
	if status_label != null:
		status_label.text = message


func set_pending(pending: bool) -> void:
	is_pending = pending
	if cancel_button != null:
		cancel_button.disabled = pending
	for selector in [game_mode_select, lives_select, target_score_select, team_structure_select, assignment_select, team_count_select, max_players_select]:
		if selector != null:
			selector.disabled = pending
	if custom_target_score_input != null:
		custom_target_score_input.editable = not pending
	_refresh_create_state()


func _on_cancel_pressed() -> void:
	cancel_requested.emit()
