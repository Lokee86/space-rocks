extends VBoxContainer
class_name MultiplayerRoomSetupReadout

signal create_requested(config: Dictionary)
signal cancel_requested

@onready var mode_select = %ModeSelect
@onready var assignment_row: Control = %AssignmentRow
@onready var assignment_select = %AssignmentSelect
@onready var team_count_row: Control = %TeamCountRow
@onready var team_count_select = %TeamCountSelect
@onready var max_players_select = %MaxPlayersSelect
@onready var create_button: BaseButton = %CreateButton
@onready var cancel_button: BaseButton = %CancelButton
@onready var status_label: Label = %StatusLabel


func _ready() -> void:
	mode_select.replace_items([
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
	mode_select.item_selected.connect(_on_mode_selected)
	create_button.pressed.connect(_on_create_pressed)
	cancel_button.pressed.connect(_on_cancel_pressed)
	_refresh_visibility()


func current_config() -> Dictionary:
	var structure := str(mode_select.selected_value("ffa"))
	var assignment_mode := ""
	var team_count := 0
	if structure == "custom":
		assignment_mode = str(assignment_select.selected_value("owner_assigned"))
	elif structure == "auto_balanced":
		team_count = int(team_count_select.selected_value(2))
	return {
		"team_structure": structure,
		"team_assignment_mode": assignment_mode,
		"team_count": team_count,
		"max_players": int(max_players_select.selected_value(8)),
	}


func _on_mode_selected(_index: int) -> void:
	_refresh_visibility()


func _refresh_visibility() -> void:
	var structure := str(mode_select.selected_value("ffa"))
	assignment_row.visible = structure == "custom"
	team_count_row.visible = structure == "auto_balanced"


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
	if mode_select != null:
		mode_select.disabled = pending
	if assignment_select != null:
		assignment_select.disabled = pending
	if team_count_select != null:
		team_count_select.disabled = pending
	if max_players_select != null:
		max_players_select.disabled = pending


func _on_cancel_pressed() -> void:
	cancel_requested.emit()
