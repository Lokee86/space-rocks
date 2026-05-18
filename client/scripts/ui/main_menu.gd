extends Control

signal single_player_pressed
signal multiplayer_room_requested(room_id: String)

const MULTIPLAYER_DIALOG_SCENE := preload("res://scenes/ui/multiplayer_dialog.tscn")

@onready var single_player_button: TextureButton = _find_texture_button("SinglePlayerButton")
@onready var multiplayer_button: TextureButton = _find_texture_button("MultiplayerButton")
@onready var quit_button: TextureButton = _find_texture_button("QuitButton")

var multiplayer_dialog: Control


func _ready() -> void:
	if single_player_button != null:
		single_player_button.pressed.connect(_start_new_game)
	else:
		push_error("Main menu is missing SinglePlayerButton.")

	if multiplayer_button != null:
		multiplayer_button.pressed.connect(_open_multiplayer_dialog)
	else:
		push_error("Main menu is missing MultiplayerButton.")

	if quit_button != null:
		quit_button.pressed.connect(_quit)
	else:
		push_error("Main menu is missing QuitButton.")


func _start_new_game() -> void:
	single_player_pressed.emit()


func _open_multiplayer_dialog() -> void:
	if multiplayer_dialog != null && is_instance_valid(multiplayer_dialog):
		return

	multiplayer_dialog = MULTIPLAYER_DIALOG_SCENE.instantiate()
	multiplayer_dialog.connect("submitted", _submit_multiplayer_room)
	add_child(multiplayer_dialog)


func _submit_multiplayer_room(room_id: String) -> void:
	multiplayer_dialog = null
	multiplayer_room_requested.emit(room_id)


func _quit() -> void:
	get_tree().quit()


func _find_texture_button(button_name: String) -> TextureButton:
	return find_child(button_name, true, false) as TextureButton
