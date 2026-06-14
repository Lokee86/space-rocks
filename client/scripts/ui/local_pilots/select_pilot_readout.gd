extends Control

const GUEST_DISPLAY_NAME := "GUEST"

signal load_requested(item: Dictionary)
signal create_requested
signal edit_requested(item: Dictionary)
signal delete_requested(item: Dictionary)

@onready var pilot_list_view: DiscreteListView = %PilotListView
@onready var load_button: Button = %LoadButton
@onready var create_button: Button = %CreateButton
@onready var reset_button: Button = %ResetButton
@onready var delete_button: Button = %DeleteButton

var selected_item: Dictionary = {"identity_kind": "guest", "display_name": GUEST_DISPLAY_NAME}


func _ready() -> void:
	load_button.pressed.connect(_on_load_pressed)
	create_button.pressed.connect(_on_create_pressed)
	reset_button.pressed.connect(_on_edit_pressed)
	delete_button.pressed.connect(_on_delete_pressed)
	pilot_list_view.selection_changed.connect(_on_list_selection_changed)
	_update_action_buttons()


func populate_pilots(local_pilots: Array) -> void:
	var items: Array = []

	for local_pilot in local_pilots:
		items.append(_build_local_pilot_item(local_pilot))

	var guest_item := {"identity_kind": "guest", "display_name": GUEST_DISPLAY_NAME}
	items.append(guest_item)
	pilot_list_view.set_items(items)
	pilot_list_view.select_index(items.size() - 1)
	selected_item = guest_item.duplicate(true)
	_update_action_buttons()


func _build_local_pilot_item(local_pilot: Dictionary) -> Dictionary:
	var display_name := str(local_pilot.get("display_name", ""))
	return {
		"identity_kind": "local_profile",
		"local_profile_id": local_pilot.get("local_profile_id"),
		"display_name": display_name,
	}


func _on_list_selection_changed(item: Dictionary) -> void:
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
