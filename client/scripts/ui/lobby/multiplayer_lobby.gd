extends Control

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")
const LobbyPlayerListView := preload("res://scripts/ui/lobby/lobby_player_list_view.gd")
const LobbyStatusViewModel := preload("res://scripts/ui/lobby/lobby_status_view_model.gd")

signal ready_requested(ready: bool)
signal start_game_requested
signal leave_requested

@export_node_path("Label") var room_code_label_path: NodePath
@export_node_path("Label") var room_status_label_path: NodePath
@export_node_path("Container") var player_list_container_path: NodePath
@export_node_path("BaseButton") var ready_button_path: NodePath
@export_node_path("BaseButton") var start_game_button_path: NodePath
@export_node_path("BaseButton") var leave_button_path: NodePath
@export var player_row_scene: PackedScene

@onready var room_code_label: Label = get_node_or_null(room_code_label_path) as Label
@onready var room_status_label: Label = get_node_or_null(room_status_label_path) as Label
@onready var player_list_container: Container = get_node_or_null(player_list_container_path) as Container
@onready var ready_button: BaseButton = get_node_or_null(ready_button_path) as BaseButton
@onready var start_game_button: BaseButton = get_node_or_null(start_game_button_path) as BaseButton
@onready var leave_button: BaseButton = get_node_or_null(leave_button_path) as BaseButton

var local_ready := false


func _ready() -> void:
	if ready_button != null:
		ready_button.pressed.connect(_on_ready_pressed)
	if start_game_button != null:
		start_game_button.disabled = true
		start_game_button.pressed.connect(_on_start_game_pressed)
	if leave_button != null:
		leave_button.pressed.connect(_on_leave_pressed)
	_update_ready_button_text()


func apply_lobby_state(
	room_code: String,
	room_state: String,
	local_member_id: String,
	owner_id: String,
	_max_players: int,
	members: Array,
	can_start := false
) -> void:
	if room_code_label != null:
		room_code_label.text = room_code
	if room_status_label != null:
		room_status_label.text = LobbyStatusViewModel.status_text(
			room_state,
			local_member_id,
			owner_id,
			members,
			can_start
		)
	local_ready = LobbyMemberViewModel.is_local_ready(local_member_id, members)
	_update_ready_button_text()
	LobbyPlayerListView.render(player_list_container, player_row_scene, local_member_id, owner_id, members)


func set_start_enabled(enabled: bool) -> void:
	if start_game_button != null:
		start_game_button.disabled = !enabled


func _update_ready_button_text() -> void:
	if ready_button == null:
		return

	var button_text := "UNREADY" if local_ready else "READY"
	var ready_label := ready_button.find_child("Ready", true, false) as Label
	if ready_label != null:
		ready_label.text = button_text
	elif "text" in ready_button:
		ready_button.text = button_text


func _on_ready_pressed() -> void:
	ready_requested.emit(!local_ready)


func _on_start_game_pressed() -> void:
	start_game_requested.emit()


func _on_leave_pressed() -> void:
	leave_requested.emit()
