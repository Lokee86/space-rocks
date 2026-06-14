extends Control

signal confirm_requested(item: Dictionary)
signal cancel_requested

@onready var confirm_button: Button = %ConfirmButton
@onready var cancel_button: Button = %CancelButton

var _pending_item: Dictionary = {}


func _ready() -> void:
	if confirm_button != null:
		confirm_button.pressed.connect(_on_confirm_pressed)
	if cancel_button != null:
		cancel_button.pressed.connect(_on_cancel_pressed)


func configure_delete(item: Dictionary) -> void:
	_pending_item = item.duplicate(true)


func _on_confirm_pressed() -> void:
	confirm_requested.emit(_pending_item.duplicate(true))


func _on_cancel_pressed() -> void:
	cancel_requested.emit()
