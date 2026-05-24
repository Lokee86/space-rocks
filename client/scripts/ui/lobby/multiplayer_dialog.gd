extends Control

signal create_room_requested
signal join_room_requested(room_code: String)
signal canceled

const EMPTY_ROOM_CODE_MESSAGE := "Enter a room\ncode to join."
const ClientLogger = preload("res://scripts/logging/logger.gd")

@onready var dialog_window: Control = find_child("MultiplayerDialogWindow", true, false) as Control
@onready var room_code_input: LineEdit = _find_line_edit(["RoomCodeInput", "IDentry"])
@onready var create_button: BaseButton = find_child("CreateButton", true, false) as BaseButton
@onready var join_button: BaseButton = find_child("JoinButton", true, false) as BaseButton
@onready var cancel_button: BaseButton = find_child("CancelButton", true, false) as BaseButton
@onready var status_label: Label = find_child("StatusLabel", true, false) as Label


func _ready() -> void:
	if create_button != null:
		create_button.pressed.connect(_request_create_room)
	else:
		push_error("Multiplayer dialog is missing CreateButton.")

	if join_button != null:
		join_button.pressed.connect(_request_join_room)
	else:
		push_error("Multiplayer dialog is missing JoinButton.")

	if cancel_button != null:
		cancel_button.pressed.connect(_cancel)
	else:
		push_error("Multiplayer dialog is missing CancelButton.")

	if room_code_input != null:
		room_code_input.grab_focus()
	else:
		push_error("Multiplayer dialog is missing RoomCodeInput.")

	if status_label == null:
		push_error("Multiplayer dialog is missing StatusLabel.")


func _request_create_room() -> void:
	ClientLogger.lobby_debug("CreateButton pressed")
	create_room_requested.emit()
	queue_free()


func _request_join_room() -> void:
	if room_code_input == null:
		return

	var room_code := room_code_input.text.strip_edges()
	if room_code == "":
		ClientLogger.lobby_debug("JoinButton pressed empty room code")
		_set_status(EMPTY_ROOM_CODE_MESSAGE)
		return

	ClientLogger.lobby_debug("JoinButton pressed room_code=%s" % room_code)
	join_room_requested.emit(room_code)
	queue_free()


func _cancel() -> void:
	canceled.emit()
	queue_free()


func set_status(text: String) -> void:
	_set_status(text)


func _set_status(text: String) -> void:
	if status_label == null:
		return

	status_label.text = text


func _find_line_edit(names: Array) -> LineEdit:
	for node_name in names:
		var line_edit := find_child(node_name, true, false) as LineEdit
		if line_edit != null:
			return line_edit

	return null
