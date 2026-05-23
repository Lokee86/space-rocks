extends HBoxContainer

@onready var player_name_label: Label = find_child("PlayerNameLabel", true, false) as Label
@onready var player_ready_label: Label = find_child("PlayerReadyLabel", true, false) as Label
@onready var ready_green: CanvasItem = find_child("ReadyGreen", true, false) as CanvasItem
@onready var ready_red: CanvasItem = find_child("ReadyRed", true, false) as CanvasItem


func _ready() -> void:
	if player_name_label == null:
		push_error("Player row is missing PlayerNameLabel.")
	if ready_green == null:
		push_error("Player row is missing ReadyGreen.")
	if ready_red == null:
		push_error("Player row is missing ReadyRed.")


func set_member(member_name, is_ready, is_connected := true) -> void:
	var ready := bool(is_ready)
	var connected := bool(is_connected)
	var is_ready_visible := connected && ready

	if player_name_label != null:
		player_name_label.text = str(member_name)

	if ready_green != null:
		ready_green.visible = is_ready_visible
	if ready_red != null:
		ready_red.visible = !is_ready_visible

	if player_ready_label != null:
		if !connected:
			player_ready_label.text = "Joining"
		elif ready:
			player_ready_label.text = "Ready"
		else:
			player_ready_label.text = "Not Ready"
		player_ready_label.visible = true
