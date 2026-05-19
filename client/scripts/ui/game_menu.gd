extends Control
class_name GameMenu

signal resume_requested
signal quit_requested

@onready var resume_button: TextureButton = find_child("ResumeButton", true, false) as TextureButton
@onready var quit_button: TextureButton = find_child("QuitButton", true, false) as TextureButton


func _ready() -> void:
	if resume_button != null:
		resume_button.pressed.connect(resume_requested.emit)
	else:
		push_error("Game menu is missing ResumeButton.")

	if quit_button != null:
		quit_button.pressed.connect(quit_requested.emit)
	else:
		push_error("Game menu is missing QuitButton.")
