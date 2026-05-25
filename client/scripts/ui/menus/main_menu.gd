extends Control

signal single_player_pressed
signal multiplayer_create_requested
signal multiplayer_join_requested(room_code: String)
signal multiplayer_room_requested(room_id: String)


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
	multiplayer_create_requested.emit()


func _on_quit_pressed() -> void:
	get_tree().quit()
