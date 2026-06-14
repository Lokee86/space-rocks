extends Control

signal confirm_requested(callsign: String)
signal cancel_requested

@onready var create_label: Label = %CreateLabel
@onready var edit_label: Label = %EditLabel
@onready var callsign_input: LineEdit = %CallsignInput
@onready var confirm_button: Button = %ConfirmButton
@onready var cancel_button: Button = %CancelButton


func _ready() -> void:
	if confirm_button != null:
		confirm_button.pressed.connect(_on_confirm_pressed)
	if cancel_button != null:
		cancel_button.pressed.connect(_on_cancel_pressed)


func configure_create() -> void:
	if create_label != null:
		create_label.visible = true
	if edit_label != null:
		edit_label.visible = false
	if callsign_input != null:
		callsign_input.clear()
		callsign_input.call_deferred("grab_focus")


func _on_confirm_pressed() -> void:
	if callsign_input == null:
		confirm_requested.emit("")
		return

	confirm_requested.emit(callsign_input.text.strip_edges())


func _on_cancel_pressed() -> void:
	cancel_requested.emit()
