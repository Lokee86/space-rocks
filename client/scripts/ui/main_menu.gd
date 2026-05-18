extends Control

signal single_player_pressed

@onready var single_player_button: TextureButton = _find_texture_button("SinglePlayerButton")
@onready var quit_button: TextureButton = _find_texture_button("QuitButton")


func _ready() -> void:
	if single_player_button != null:
		single_player_button.pressed.connect(_start_new_game)
	else:
		push_error("Main menu is missing SinglePlayerButton.")

	if quit_button != null:
		quit_button.pressed.connect(_quit)
	else:
		push_error("Main menu is missing QuitButton.")


func _start_new_game() -> void:
	single_player_pressed.emit()


func _quit() -> void:
	get_tree().quit()


func _find_texture_button(button_name: String) -> TextureButton:
	return find_child(button_name, true, false) as TextureButton
