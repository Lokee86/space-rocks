extends Control

signal confirm_requested(callsign: String)
signal cancel_requested

const EMPTY_CALLSIGN_FEEDBACK_TEXT := "PLEASE ENTER A CALLSIGN"
const INVALID_CALLSIGN_FEEDBACK_TEXT := "A-Z _- ONLY"
const FEEDBACK_DURATION_SECONDS := 1.5
const CALLSIGN_PATTERN := "^[A-Za-z0-9_-]+$"

@onready var create_label: Label = %CreateLabel
@onready var edit_label: Label = %EditLabel
@onready var callsign_input: LineEdit = %CallsignInput
@onready var confirm_button: Button = %ConfirmButton
@onready var cancel_button: Button = %CancelButton
var _callsign_regex: RegEx = RegEx.new()


func _ready() -> void:
	_callsign_regex.compile(CALLSIGN_PATTERN)
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
		callsign_input.placeholder_text = ""
		callsign_input.call_deferred("grab_focus")


func _on_confirm_pressed() -> void:
	if callsign_input == null:
		return

	var callsign := callsign_input.text.strip_edges()
	if callsign.is_empty():
		_show_input_feedback(EMPTY_CALLSIGN_FEEDBACK_TEXT)
		return
	
	if _callsign_regex.search(callsign) == null:
		_show_input_feedback(INVALID_CALLSIGN_FEEDBACK_TEXT)
		return

	confirm_requested.emit(callsign)


func _on_cancel_pressed() -> void:
	cancel_requested.emit()


func _show_input_feedback(message: String) -> void:
	if callsign_input == null:
		return

	callsign_input.clear()
	callsign_input.placeholder_text = message
	callsign_input.call_deferred("grab_focus")

	await get_tree().create_timer(FEEDBACK_DURATION_SECONDS).timeout
	if callsign_input != null and callsign_input.placeholder_text == message:
		callsign_input.placeholder_text = ""
