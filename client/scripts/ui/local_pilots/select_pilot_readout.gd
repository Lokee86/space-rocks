extends Control

const PilotSelectRowScene := preload("res://scenes/ui/elements/pilot_select_row.tscn")

signal load_requested(item: Dictionary)
signal create_requested
signal edit_requested(item: Dictionary)
signal delete_requested(item: Dictionary)

@onready var pilot_list_container: VBoxContainer = %PilotListContainer
@onready var load_button: Button = %LoadButton
@onready var create_button: Button = %CreateButton
@onready var reset_button: Button = %ResetButton
@onready var delete_button: Button = %DeleteButton

var selected_item: Dictionary = {"identity_kind": "guest", "display_name": "Play as Guest"}
var selected_row: Control


func _ready() -> void:
	load_button.pressed.connect(_on_load_pressed)
	create_button.pressed.connect(_on_create_pressed)
	reset_button.pressed.connect(_on_edit_pressed)
	delete_button.pressed.connect(_on_delete_pressed)
	_update_action_buttons()


func populate_pilots(local_pilots: Array) -> void:
	for child in pilot_list_container.get_children():
		child.queue_free()

	selected_row = null

	for local_pilot in local_pilots:
		var local_pilot_data := _build_local_pilot_item(local_pilot)
		_add_row(local_pilot_data["display_name"], local_pilot_data)

	var guest_item := {"identity_kind": "guest", "display_name": "Play as Guest"}
	selected_row = _add_row("Play as Guest", guest_item)
	_select_item(guest_item)
	if selected_row != null:
		selected_row.call_deferred("grab_focus")


func _build_local_pilot_item(local_pilot: Dictionary) -> Dictionary:
	var display_name := str(local_pilot.get("display_name", ""))
	return {
		"identity_kind": "local_profile",
		"local_profile_id": local_pilot.get("local_profile_id"),
		"display_name": display_name,
	}


func _add_row(display_text: String, item_data: Dictionary) -> Control:
	var row := PilotSelectRowScene.instantiate()
	pilot_list_container.add_child(row)
	row.configure(display_text, item_data)
	row.selected.connect(_on_row_selected)
	return row


func _on_row_selected(item: Dictionary) -> void:
	_select_item(item)


func _select_item(item: Dictionary) -> void:
	selected_item = item.duplicate(true)
	_update_action_buttons()


func _update_action_buttons() -> void:
	var is_guest: bool = selected_item.get("identity_kind", "guest") == "guest"
	load_button.disabled = false
	create_button.disabled = false
	reset_button.disabled = is_guest
	delete_button.disabled = is_guest


func _on_load_pressed() -> void:
	load_requested.emit(selected_item.duplicate(true))


func _on_create_pressed() -> void:
	create_requested.emit()


func _on_edit_pressed() -> void:
	edit_requested.emit(selected_item.duplicate(true))


func _on_delete_pressed() -> void:
	delete_requested.emit(selected_item.duplicate(true))
