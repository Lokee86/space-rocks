class_name MultiplayerLobby
extends Control

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")
const LobbyPlayerListView := preload("res://scripts/ui/lobby/lobby_player_list_view.gd")
const LobbyStatusViewModel := preload("res://scripts/ui/lobby/lobby_status_view_model.gd")
const TeamPresentation := preload("res://scripts/teams/team_presentation.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")

signal ready_requested(ready: bool)
signal start_game_requested
signal add_bot_requested
signal remove_member_requested(player_id: String)
signal team_assignment_requested(player_id: String, team_id: String)
signal leave_requested

@export var player_row_scene: PackedScene
@onready var window_frame: Control = %WindowFrame
@onready var room_code_label: Label = %RoomCodeValueLabel
@onready var room_status_label: Label = %RoomStatusValueLabel
@onready var mode_value_label: Label = %ModeValueLabel
@onready var assignment_row: Control = %AssignmentInfoRow
@onready var assignment_value_label: Label = %AssignmentValueLabel
@onready var team_header: Control = %TeamHeader
@onready var player_list_container: Container = %PlayerListContainer
@onready var ready_button: BaseButton = %ReadyButton
@onready var start_game_button: BaseButton = %StartGameButton
@onready var add_bot_button: BaseButton = %AddBotButton
@onready var leave_button: BaseButton = %LeaveButton
var local_ready := false


func _ready() -> void:
	ready_button.pressed.connect(_on_ready_pressed)
	start_game_button.disabled = true
	start_game_button.pressed.connect(_on_start_game_pressed)
	add_bot_button.visible = false
	add_bot_button.pressed.connect(_on_add_bot_pressed)
	leave_button.pressed.connect(_on_leave_pressed)
	get_viewport().size_changed.connect(_update_window_size)
	_update_ready_button_text()
	_update_window_size()


func apply_lobby_state(
	room_code: String,
	room_state: String,
	local_player_id: String,
	owner_id: String,
	max_players: int,
	members: Array,
	team_structure := "ffa",
	team_assignment_mode := "",
	team_count := 0,
	team_assignments_locked := false,
	can_start := false
) -> void:
	room_code_label.text = room_code
	room_status_label.text = LobbyStatusViewModel.status_text(room_state, local_player_id, owner_id, members, can_start)
	mode_value_label.text = TeamPresentation.structure_name(team_structure)
	assignment_row.visible = team_structure == "custom"
	assignment_value_label.visible = team_structure == "custom"
	assignment_value_label.text = TeamPresentation.assignment_name(team_assignment_mode)
	team_header.visible = team_structure != "ffa"
	local_ready = LobbyMemberViewModel.is_local_ready(local_player_id, members)
	_update_ready_button_text()
	var local_is_owner := not local_player_id.is_empty() and local_player_id == owner_id
	add_bot_button.visible = local_is_owner
	add_bot_button.disabled = not local_is_owner or members.size() >= max_players
	LobbyPlayerListView.render(
		player_list_container,
		player_row_scene,
		local_player_id,
		owner_id,
		members,
		team_structure,
		team_assignment_mode,
		team_count,
		team_assignments_locked,
		Callable(self, "_on_remove_member_requested"),
		Callable(self, "_on_team_assignment_requested")
	)


func set_start_enabled(enabled: bool) -> void:
	start_game_button.disabled = not enabled


func _update_window_size() -> void:
	# The lobby fills the transmission display; its internal roster and actions adapt.
	if window_frame != null:
		window_frame.custom_minimum_size = Vector2.ZERO


func _update_ready_button_text() -> void:
	var ready_label := ready_button.find_child("Ready", true, false) as Label
	if ready_label != null:
		ready_label.text = Constants.READY_BUTTON_TEXT_UNREADY if local_ready else Constants.READY_BUTTON_TEXT_READY


func _on_ready_pressed() -> void:
	ready_requested.emit(not local_ready)

func _on_start_game_pressed() -> void:
	start_game_requested.emit()

func _on_add_bot_pressed() -> void:
	add_bot_requested.emit()

func _on_remove_member_requested(player_id: String) -> void:
	remove_member_requested.emit(player_id)

func _on_team_assignment_requested(player_id: String, team_id: String) -> void:
	team_assignment_requested.emit(player_id, team_id)

func _on_leave_pressed() -> void:
	leave_requested.emit()
