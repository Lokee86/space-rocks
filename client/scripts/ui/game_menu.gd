extends Control
class_name GameMenu

signal primary_action_requested
signal menu_requested
signal resume_requested
signal quit_requested

@onready var primary_action_button: BaseButton = _find_button(["PrimaryActionButton", "ResumeButton", "LeftButton"])
@onready var menu_button: BaseButton = _find_button(["MenuButton", "QuitButton"])


func _ready() -> void:
	if primary_action_button != null:
		primary_action_button.pressed.connect(_on_primary_action_pressed)
	else:
		push_error("Game menu is missing PrimaryActionButton.")

	if menu_button != null:
		menu_button.pressed.connect(_on_menu_pressed)
	else:
		push_error("Game menu is missing MenuButton.")


func set_primary_text(text) -> void:
	_set_button_text(primary_action_button, str(text))


func set_primary_enabled(enabled) -> void:
	if primary_action_button == null:
		return

	primary_action_button.disabled = !bool(enabled)


func set_menu_text(text) -> void:
	_set_button_text(menu_button, str(text))


func _on_primary_action_pressed() -> void:
	primary_action_requested.emit()
	resume_requested.emit()


func _on_menu_pressed() -> void:
	menu_requested.emit()
	quit_requested.emit()


func _find_button(names: Array) -> BaseButton:
	for button_name in names:
		var button := find_child(button_name, true, false) as BaseButton
		if button != null:
			return button

	return null


func _set_button_text(button: BaseButton, text: String) -> void:
	if button == null:
		return

	var first_label := true
	for child in button.find_children("*", "Label", true, false):
		var label := child as Label
		label.text = "\n\n%s" % text
		label.visible = first_label
		first_label = false
