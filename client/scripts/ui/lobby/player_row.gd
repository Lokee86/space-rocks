extends HBoxContainer
class_name PlayerRow

const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

signal remove_requested(player_id: String)
signal team_assignment_requested(player_id: String, team_id: String)

@onready var player_name_label: Label = find_child("PlayerNameLabel", true, false) as Label
@onready var player_ready_label: Label = find_child("PlayerReadyLabel", true, false) as Label
@onready var owner_indicator: CanvasItem = find_child("OwnerIndicator", true, false) as CanvasItem
@onready var ready_green: CanvasItem = find_child("ReadyGreen", true, false) as CanvasItem
@onready var ready_red: CanvasItem = find_child("ReadyRed", true, false) as CanvasItem
@onready var team_selector: Control = %TeamSelector
@onready var remove_button: BaseButton = find_child("RemoveButton", true, false) as BaseButton
var _player_id := ""


func _ready() -> void:
	for requirement in [
		[player_name_label, "PlayerNameLabel", "Label"],
		[player_ready_label, "PlayerReadyLabel", "Label"],
		[owner_indicator, "OwnerIndicator", "CanvasItem"],
		[ready_green, "ReadyGreen", "CanvasItem"],
		[ready_red, "ReadyRed", "CanvasItem"],
		[team_selector, "TeamSelector", "TeamSelector"],
		[remove_button, "RemoveButton", "BaseButton"],
	]:
		_report_missing_node(requirement[0], requirement[1], requirement[2])
	if remove_button != null:
		remove_button.pressed.connect(_on_remove_pressed)
	if team_selector != null:
		team_selector.connect("team_selected", Callable(self, "_on_team_selected"))


func set_member(
	player_id: String,
	member_name,
	is_ready,
	member_connected := true,
	is_owner := false,
	can_remove := false,
	team_id := "",
	team_ids: Array = [],
	can_edit_team := false,
	show_team := false
) -> void:
	_player_id = player_id
	var member_ready := bool(is_ready)
	var connected := bool(member_connected)
	if player_name_label != null:
		player_name_label.text = str(member_name)
	if player_ready_label != null:
		player_ready_label.text = "Ready" if member_ready else "Not Ready"
	if owner_indicator != null:
		owner_indicator.visible = bool(is_owner)
	if ready_green != null:
		ready_green.visible = member_ready and connected
	if ready_red != null:
		ready_red.visible = not member_ready or not connected
	if team_selector != null:
		team_selector.visible = show_team
		if show_team:
			team_selector.call("configure", team_id, team_ids, can_edit_team)
	if remove_button != null:
		remove_button.visible = can_remove
		remove_button.disabled = not can_remove


func _on_team_selected(team_id: String) -> void:
	if not _player_id.is_empty():
		team_assignment_requested.emit(_player_id, team_id)


func _on_remove_pressed() -> void:
	if not _player_id.is_empty():
		remove_requested.emit(_player_id)


func _report_missing_node(node: Node, node_name: String, expected_type: String) -> void:
	if node != null:
		return
	ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Player row is missing a required presentation node",
		{},
		{
			"subsystem": "lobby",
			"failure_mode": "missing_node",
			"node_name": node_name,
			"resource_kind": "ui_node",
			"expected_type": expected_type,
			"actual_type": "null",
		}
	)
