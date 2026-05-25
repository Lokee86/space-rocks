extends Control

const MultiplayerDialogScene := preload("res://scenes/ui/dialogs/multiplayer_dialog.tscn")

signal single_player_pressed
signal multiplayer_create_requested
signal multiplayer_join_requested(room_code: String)
signal multiplayer_room_requested(room_id: String)

var multiplayer_dialog: Control


func _ready() -> void:
	var single_player_button := find_child("SinglePlayerButton", true, false) as TextureButton
	var multiplayer_button := find_child("MultiplayerButton", true, false) as TextureButton
	var quit_button := find_child("QuitButton", true, false) as TextureButton

	if single_player_button == null:
		push_error("Missing button: SinglePlayerButton")
	else:
		single_player_button.pressed.connect(_on_single_player_pressed)

	if multiplayer_button == null:
		push_error("Missing button: MultiplayerButton")
	else:
		multiplayer_button.pressed.connect(_on_multiplayer_pressed)

	if quit_button == null:
		push_error("Missing button: QuitButton")
	else:
		quit_button.pressed.connect(_on_quit_pressed)


func _on_single_player_pressed() -> void:
	print("V2 main menu single player pressed")
	single_player_pressed.emit()


func _on_multiplayer_pressed() -> void:
	print("V2 main menu multiplayer pressed")
	_open_multiplayer_dialog()


func _on_quit_pressed() -> void:
	get_tree().quit()


func _open_multiplayer_dialog() -> void:
	if multiplayer_dialog != null && is_instance_valid(multiplayer_dialog):
		multiplayer_dialog.show()
		return

	multiplayer_dialog = MultiplayerDialogScene.instantiate()
	add_child(multiplayer_dialog)
	multiplayer_dialog.create_room_requested.connect(_on_dialog_create_room_requested)
	multiplayer_dialog.join_room_requested.connect(_on_dialog_join_room_requested)
	multiplayer_dialog.canceled.connect(_on_dialog_canceled)
	multiplayer_dialog.show()


func _on_dialog_create_room_requested() -> void:
	print("V2 main menu relaying dialog create room requested")
	multiplayer_create_requested.emit()
	_clear_multiplayer_dialog()


func _on_dialog_join_room_requested(room_code: String) -> void:
	print("V2 main menu relaying dialog join room requested: %s" % room_code)
	multiplayer_join_requested.emit(room_code)
	_clear_multiplayer_dialog()


func _on_dialog_canceled() -> void:
	_clear_multiplayer_dialog()


func _clear_multiplayer_dialog() -> void:
	if multiplayer_dialog != null && is_instance_valid(multiplayer_dialog):
		multiplayer_dialog.queue_free()
	multiplayer_dialog = null
