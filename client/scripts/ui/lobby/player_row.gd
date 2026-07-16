extends HBoxContainer
class_name PlayerRow

const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

@onready var player_name_label: Label = find_child("PlayerNameLabel", true, false) as Label
@onready var player_ready_label: Label = find_child("PlayerReadyLabel", true, false) as Label
@onready var owner_indicator: CanvasItem = find_child("OwnerIndicator", true, false) as CanvasItem
@onready var ready_green: CanvasItem = find_child("ReadyGreen", true, false) as CanvasItem
@onready var ready_red: CanvasItem = find_child("ReadyRed", true, false) as CanvasItem


func _ready() -> void:
	_report_missing_node(player_name_label, "PlayerNameLabel", "Label")
	_report_missing_node(player_ready_label, "PlayerReadyLabel", "Label")
	_report_missing_node(owner_indicator, "OwnerIndicator", "CanvasItem")
	_report_missing_node(ready_green, "ReadyGreen", "CanvasItem")
	_report_missing_node(ready_red, "ReadyRed", "CanvasItem")


func set_member(member_name, is_ready, member_connected := true, is_owner := false) -> void:
	var member_ready := bool(is_ready)
	var connected := bool(member_connected)

	if player_name_label != null:
		player_name_label.text = str(member_name)
	if player_ready_label != null:
		player_ready_label.text = "Ready" if member_ready else "Not Ready"
	if owner_indicator != null:
		owner_indicator.visible = bool(is_owner)
	if ready_green != null:
		ready_green.visible = member_ready && connected
	if ready_red != null:
		ready_red.visible = !member_ready || !connected


func _report_missing_node(node: Node, node_name: String, expected_type: String) -> void:
	if node == null:
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
