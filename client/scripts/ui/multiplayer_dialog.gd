extends Control

signal submitted(room_id: String)

const ROOM_ID_CHARACTERS := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const GENERATED_ROOM_ID_LENGTH := 6

@onready var room_id_entry: LineEdit = find_child("IDentry", true, false) as LineEdit
@onready var submit_button: TextureButton = find_child("SubmitButton", true, false) as TextureButton
@onready var cancel_button: TextureButton = find_child("CancelButton", true, false) as TextureButton


func _ready() -> void:
	if submit_button != null:
		submit_button.pressed.connect(_submit)
	else:
		push_error("Multiplayer dialog is missing SubmitButton.")

	if cancel_button != null:
		cancel_button.pressed.connect(queue_free)
	else:
		push_error("Multiplayer dialog is missing CancelButton.")

	if room_id_entry != null:
		room_id_entry.grab_focus()
	else:
		push_error("Multiplayer dialog is missing IDentry.")


func _submit() -> void:
	if room_id_entry == null:
		return

	var room_id := room_id_entry.text.strip_edges()
	if room_id == "":
		room_id = _generate_room_id()

	submitted.emit(room_id)
	queue_free()


func _generate_room_id() -> String:
	var room_id := ""
	for index in range(GENERATED_ROOM_ID_LENGTH):
		var character_index := randi() % ROOM_ID_CHARACTERS.length()
		room_id += ROOM_ID_CHARACTERS.substr(character_index, 1)

	return room_id
