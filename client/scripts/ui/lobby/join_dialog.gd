extends Control

signal join_requested(room_code: String)
signal cancel_requested

var room_code_input: LineEdit
var status_label: Label


func _ready() -> void:
	room_code_input = get_node_or_null("%RoomCodeInput") as LineEdit
	status_label = get_node_or_null("%StatusLabel") as Label

	var join_button := get_node_or_null("%JoinButton") as BaseButton
	if join_button != null:
		join_button.pressed.connect(_on_join_pressed)

	var cancel_button := get_node_or_null("%CancelButton") as BaseButton
	if cancel_button != null:
		cancel_button.pressed.connect(_on_cancel_pressed)


func _on_join_pressed() -> void:
	var room_code := ""
	if room_code_input != null:
		room_code = room_code_input.text.strip_edges()
	join_requested.emit(room_code)


func _on_cancel_pressed() -> void:
	cancel_requested.emit()


func set_status(message: String) -> void:
	if status_label != null:
		status_label.text = message


func clear_status() -> void:
	set_status("")
