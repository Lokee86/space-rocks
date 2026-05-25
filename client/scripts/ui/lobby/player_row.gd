extends HBoxContainer

@onready var player_name_label: Label = find_child("PlayerNameLabel", true, false) as Label
@onready var player_ready_label: Label = find_child("PlayerReadyLabel", true, false) as Label
@onready var ready_green: CanvasItem = find_child("ReadyGreen", true, false) as CanvasItem
@onready var ready_red: CanvasItem = find_child("ReadyRed", true, false) as CanvasItem


func _ready() -> void:
	if player_name_label == null:
		player_name_label = find_child("PlayerNameLabel", true, false) as Label
	if player_ready_label == null:
		player_ready_label = find_child("PlayerReadyLabel", true, false) as Label
	if ready_green == null:
		ready_green = find_child("ReadyGreen", true, false) as CanvasItem
	if ready_red == null:
		ready_red = find_child("ReadyRed", true, false) as CanvasItem
	_report_missing_node(player_name_label, "PlayerNameLabel")
	_report_missing_node(player_ready_label, "PlayerReadyLabel")
	_report_missing_node(ready_green, "ReadyGreen")
	_report_missing_node(ready_red, "ReadyRed")


func set_member(member_name, is_ready, member_connected := true) -> void:
	if player_name_label == null:
		player_name_label = find_child("PlayerNameLabel", true, false) as Label
	if player_ready_label == null:
		player_ready_label = find_child("PlayerReadyLabel", true, false) as Label
	if ready_green == null:
		ready_green = find_child("ReadyGreen", true, false) as CanvasItem
	if ready_red == null:
		ready_red = find_child("ReadyRed", true, false) as CanvasItem

	var ready := bool(is_ready)
	var connected := bool(member_connected)

	if player_name_label != null:
		player_name_label.text = str(member_name)
	if player_ready_label != null:
		player_ready_label.text = "Ready" if ready else "Not Ready"
	if ready_green != null:
		ready_green.visible = ready && connected
	if ready_red != null:
		ready_red.visible = !ready || !connected


func _report_missing_node(node: Node, node_name: String) -> void:
	if node == null:
		push_warning("V2 PlayerRow missing node: %s" % node_name)
