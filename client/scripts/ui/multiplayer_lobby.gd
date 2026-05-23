extends Control

signal ready_requested
signal start_game_requested
signal leave_requested

const ClientLogger = preload("res://scripts/logging/logger.gd")

@export_node_path("Label") var room_code_label_path: NodePath
@export_node_path("Label") var room_status_label_path: NodePath
@export_node_path("Container") var player_list_container_path: NodePath
@export_node_path("BaseButton") var ready_button_path: NodePath
@export_node_path("BaseButton") var start_game_button_path: NodePath
@export_node_path("BaseButton") var leave_button_path: NodePath
@export var player_row_scene: PackedScene
@export var fake_data_enabled := false

@onready var room_code_label: Label = get_node_or_null(room_code_label_path) as Label
@onready var room_status_label: Label = get_node_or_null(room_status_label_path) as Label
@onready var player_list_container: Container = get_node_or_null(player_list_container_path) as Container
@onready var ready_button: BaseButton = get_node_or_null(ready_button_path) as BaseButton
@onready var start_game_button: BaseButton = get_node_or_null(start_game_button_path) as BaseButton
@onready var leave_button: BaseButton = get_node_or_null(leave_button_path) as BaseButton

var local_ready := false
var fake_members := []


func _ready() -> void:
	_validate_required_nodes()
	_connect_lobby_buttons()

	if fake_data_enabled:
		_show_fake_lobby_data()


func set_room_code(room_code) -> void:
	if room_code_label == null:
		return

	room_code_label.text = str(room_code)


func set_status(text) -> void:
	if room_status_label == null:
		return

	room_status_label.text = str(text)


func set_members(members) -> void:
	if player_list_container == null || player_row_scene == null:
		return

	for child in player_list_container.get_children():
		player_list_container.remove_child(child)
		child.queue_free()

	for member in members:
		var row := player_row_scene.instantiate()
		player_list_container.add_child(row)
		_apply_member_to_row(row, member)


func set_local_ready(is_ready) -> void:
	local_ready = bool(is_ready)
	_update_ready_button_text()


func set_start_enabled(enabled) -> void:
	if start_game_button == null:
		return

	start_game_button.disabled = !bool(enabled)


func set_fake_data_enabled(enabled: bool) -> void:
	fake_data_enabled = enabled

	if fake_data_enabled:
		_show_fake_lobby_data()


func _update_ready_button_text() -> void:
	if ready_button == null:
		return

	var text := "UNREADY" if local_ready else "READY"
	if "text" in ready_button:
		ready_button.text = text
		return

	var label := ready_button.find_child("Ready", true, false) as Label
	if label != null:
		label.text = text


func _apply_member_to_row(row: Node, member) -> void:
	var member_name := _member_name(member)
	var member_ready := _member_ready(member)
	var member_connected := _member_connected(member)

	if row.has_method("set_member"):
		row.set_member(member_name, member_ready, member_connected)
		return

	var name_label := row.find_child("PlayerNameLabel", true, false) as Label
	if name_label != null:
		name_label.text = member_name

	var ready_label := row.find_child("PlayerReadyLabel", true, false) as Label
	if ready_label != null:
		ready_label.text = _member_ready_text(member_ready, member_connected)

	var ready_green := row.find_child("ReadyGreen", true, false) as CanvasItem
	if ready_green != null:
		ready_green.visible = member_connected && member_ready

	var ready_red := row.find_child("ReadyRed", true, false) as CanvasItem
	if ready_red != null:
		ready_red.visible = !member_connected || !member_ready


func _member_name(member) -> String:
	if member is Dictionary:
		return str(member.get("name", member.get("player_name", member.get("member_id", member.get("id", "Player")))))

	return str(member)


func _member_ready(member) -> bool:
	if member is Dictionary:
		return bool(member.get("ready", member.get("is_ready", false)))

	return false


func _member_connected(member) -> bool:
	if member is Dictionary:
		return bool(member.get("connected", true))

	return true


func _member_ready_text(is_ready: bool, is_connected: bool) -> String:
	if !is_connected:
		return "Joining"
	if is_ready:
		return "Ready"

	return "Not Ready"


func _validate_required_nodes() -> void:
	if room_code_label == null:
		push_error("Multiplayer lobby is missing RoomCodeValueLabel.")
	if room_status_label == null:
		push_error("Multiplayer lobby is missing RoomStatusValueLabel.")
	if player_list_container == null:
		push_error("Multiplayer lobby is missing PlayerListContainer.")
	if ready_button == null:
		push_error("Multiplayer lobby is missing ReadyButton.")
	if start_game_button == null:
		push_error("Multiplayer lobby is missing StartGameButton.")
	if leave_button == null:
		push_error("Multiplayer lobby is missing LeaveButton.")
	if player_row_scene == null:
		push_error("Multiplayer lobby is missing player_row_scene.")


# Temporary local-only behavior for lobby scene smoke testing before networking.
func _show_fake_lobby_data() -> void:
	fake_members = [
		{"name": "Player 1", "ready": local_ready},
		{"name": "Player 2", "ready": true},
	]
	set_room_code("TEST")
	set_status("Waiting for players...")
	set_members(fake_members)
	set_start_enabled(local_ready)


func _connect_lobby_buttons() -> void:
	if ready_button != null && !ready_button.pressed.is_connected(_on_ready_pressed):
		ready_button.pressed.connect(_on_ready_pressed)
	if start_game_button != null && !start_game_button.pressed.is_connected(_on_start_game_pressed):
		start_game_button.pressed.connect(_on_start_game_pressed)
	if leave_button != null && !leave_button.pressed.is_connected(_on_leave_pressed):
		leave_button.pressed.connect(_on_leave_pressed)


func _on_ready_pressed() -> void:
	print("[lobby] ReadyButton pressed")
	ready_requested.emit()

	if !fake_data_enabled:
		return

	set_local_ready(!local_ready)
	if fake_members.size() > 0 && fake_members[0] is Dictionary:
		fake_members[0]["ready"] = local_ready
	set_members(fake_members)
	set_start_enabled(local_ready)


func _on_start_game_pressed() -> void:
	print("[lobby] StartGameButton pressed")
	start_game_requested.emit()

	if !fake_data_enabled:
		return

	set_status("Start requested...")


func _on_leave_pressed() -> void:
	ClientLogger.lobby_debug("LeaveButton pressed")
	leave_requested.emit()
	set_status("Leave requested...")
