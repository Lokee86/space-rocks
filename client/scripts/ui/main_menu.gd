extends Control

signal single_player_pressed
signal multiplayer_create_requested
signal multiplayer_join_requested(room_code: String)
signal multiplayer_room_requested(room_id: String)

const MULTIPLAYER_DIALOG_SCENE := preload("res://scenes/ui/dialogs/multiplayer_dialog.tscn")

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
	if multiplayer_dialog.has_signal("create_room_requested"):
		multiplayer_dialog.create_room_requested.connect(_request_multiplayer_create)
	if multiplayer_dialog.has_signal("join_room_requested"):
		multiplayer_dialog.join_room_requested.connect(_request_multiplayer_join)
	if multiplayer_dialog.has_signal("canceled"):
		multiplayer_dialog.canceled.connect(_clear_multiplayer_dialog)
	if multiplayer_dialog.has_signal("submitted"):
		multiplayer_dialog.submitted.connect(_submit_legacy_multiplayer_room)
	add_child(multiplayer_dialog)


func _request_multiplayer_create() -> void:
	multiplayer_dialog = null
	multiplayer_create_requested.emit()


func _request_multiplayer_join(room_code: String) -> void:
	print("[main_menu] join requested room_code=", room_code)
	multiplayer_dialog = null
	multiplayer_join_requested.emit(room_code)


func _clear_multiplayer_dialog() -> void:
	multiplayer_dialog = null


func show_multiplayer_error(message: String) -> bool:
	if multiplayer_dialog == null || !is_instance_valid(multiplayer_dialog):
		return false
	if !multiplayer_dialog.has_method("set_status"):
		return false

	multiplayer_dialog.set_status(message)
	return true


func _submit_legacy_multiplayer_room(room_id: String) -> void:
	_clear_multiplayer_dialog()
	multiplayer_room_requested.emit(room_id)


func _quit() -> void:
	get_tree().quit()


func _find_texture_button(button_name: String) -> TextureButton:
	return find_child(button_name, true, false) as TextureButton
