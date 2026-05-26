extends Control

const Constants := preload("res://scripts/constants/constants.gd")

signal create_room_requested
signal join_room_requested(room_code: String)
signal canceled

var room_code_input: LineEdit
var status_label: Label


func _ready() -> void:
	room_code_input = find_child("RoomCodeInput", true, false) as LineEdit
	status_label = find_child("StatusLabel", true, false) as Label
	var create_room_button := find_child("CreateButton", true, false) as BaseButton
	var join_room_button := find_child("JoinButton", true, false) as BaseButton
	var cancel_button := find_child("CancelButton", true, false) as BaseButton

	if room_code_input == null:
		push_error("Missing input: RoomCodeInput")
	if status_label == null:
		push_error("Missing label: StatusLabel")

	if create_room_button == null:
		push_error("Missing button: CreateRoomButton")
	else:
		create_room_button.pressed.connect(_on_create_room_pressed)

	if join_room_button == null:
		push_error("Missing button: JoinRoomButton")
	else:
		join_room_button.pressed.connect(_on_join_room_pressed)

	if cancel_button == null:
		push_error("Missing button: CancelButton")
	else:
		cancel_button.pressed.connect(_on_cancel_pressed)


func _on_create_room_pressed() -> void:
	print("V2 multiplayer dialog create pressed")
	create_room_requested.emit()


func _on_join_room_pressed() -> void:
	var room_code := ""
	if room_code_input != null:
		room_code = room_code_input.text.strip_edges()
	print("V2 multiplayer dialog join pressed: %s" % room_code)
	set_status(Constants.DIALOG_STATUS_JOINING_ROOM)
	join_room_requested.emit(room_code)


func _on_cancel_pressed() -> void:
	print("V2 multiplayer dialog cancel pressed")
	canceled.emit()


func set_status(message: String) -> void:
	if status_label != null:
		status_label.text = message


func show_join_error(message: String) -> void:
	set_status(message)
