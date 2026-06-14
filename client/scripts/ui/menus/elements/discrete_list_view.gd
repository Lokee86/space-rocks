class_name DiscreteListView
extends Control

signal selection_changed(item: Dictionary)

@export var row_scene: PackedScene
@export var row_step_px := 50.0

@onready var row_viewport: Control = %RowViewport
@onready var rows: VBoxContainer = %Rows
@onready var scroll_bar: VScrollBar = %VScrollBar

var items: Array = []
var selected_index := -1
var top_visible_index := 0
var visible_row_count := 1
var _is_syncing_scrollbar := false


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_STOP
	scroll_bar.value_changed.connect(_on_scroll_bar_value_changed)
	_refresh_visible_row_count()
	_render_rows()


func _notification(what: int) -> void:
	if what == NOTIFICATION_RESIZED:
		_refresh_visible_row_count()
		_set_top_visible_index(top_visible_index)
		_render_rows()


func set_items(new_items: Array) -> void:
	items.clear()

	for item in new_items:
		if item is Dictionary:
			items.append((item as Dictionary).duplicate(true))
		else:
			items.append(item)

	selected_index = 0 if not items.is_empty() else -1
	top_visible_index = 0

	_refresh_visible_row_count()
	_sync_scrollbar()
	_render_rows()

	if selected_index != -1:
		selection_changed.emit(get_selected_item())


func select_index(index: int) -> void:
	if items.is_empty():
		selected_index = -1
	else:
		selected_index = clamp(index, 0, items.size() - 1)

	_set_top_visible_index(_get_top_visible_index_for_selected_item())
	_render_rows()
	selection_changed.emit(get_selected_item())


func get_selected_item() -> Dictionary:
	if selected_index < 0 or selected_index >= items.size():
		return {}

	var selected_item: Dictionary = items[selected_index]
	if selected_item is Dictionary:
		return (selected_item as Dictionary).duplicate(true)

	return {}


func _gui_input(event: InputEvent) -> void:
	if event is InputEventMouseButton and event.pressed:
		if event.button_index == MOUSE_BUTTON_WHEEL_DOWN:
			_set_top_visible_index(top_visible_index + 1)
			accept_event()
		elif event.button_index == MOUSE_BUTTON_WHEEL_UP:
			_set_top_visible_index(top_visible_index - 1)
			accept_event()


func _refresh_visible_row_count() -> void:
	if row_viewport == null or row_step_px <= 0.0:
		visible_row_count = 1
		return

	visible_row_count = max(1, int(floor(row_viewport.size.y / row_step_px)))


func _set_top_visible_index(index: int) -> void:
	top_visible_index = clamp(index, 0, max(0, items.size() - visible_row_count))
	_sync_scrollbar()
	_render_rows()


func _sync_scrollbar() -> void:
	if scroll_bar == null:
		return

	var item_count := items.size()
	var max_top_index: int = max(0, item_count - visible_row_count)

	_is_syncing_scrollbar = true
	scroll_bar.min_value = 0
	scroll_bar.max_value = item_count
	scroll_bar.step = 1
	scroll_bar.page = visible_row_count
	scroll_bar.value = top_visible_index
	scroll_bar.visible = max_top_index > 0
	_is_syncing_scrollbar = false


func _on_scroll_bar_value_changed(value: float) -> void:
	if _is_syncing_scrollbar:
		return

	_set_top_visible_index(int(value))


func _render_rows() -> void:
	if rows == null:
		return

	for child in rows.get_children():
		child.queue_free()

	if row_scene == null:
		return

	if items.is_empty():
		return

	var last_index: int = min(items.size() - 1, top_visible_index + visible_row_count - 1)

	for absolute_index in range(top_visible_index, last_index + 1):
		var item_variant: Dictionary = items[absolute_index]
		if not (item_variant is Dictionary):
			continue

		var item := item_variant as Dictionary
		var display_name := str(item.get("display_name", item.get("name", "")))
		var row = row_scene.instantiate()
		rows.add_child(row)

		if row.has_method("configure"):
			row.call("configure", display_name, item.duplicate(true))

		if row.has_signal("selected"):
			row.connect("selected", Callable(self, "_on_row_selected").bind(absolute_index))

		if absolute_index == selected_index:
			row.call_deferred("grab_focus")


func _on_row_selected(_item: Dictionary, absolute_index: int) -> void:
	selected_index = clamp(absolute_index, 0, items.size() - 1)
	selection_changed.emit(get_selected_item())
	_render_rows()


func _get_top_visible_index_for_selected_item() -> int:
	if items.is_empty() or selected_index < 0:
		return 0

	var max_top_index: int = max(0, items.size() - visible_row_count)

	if selected_index < top_visible_index:
		return selected_index

	if selected_index >= top_visible_index + visible_row_count:
		return clamp(selected_index - visible_row_count + 1, 0, max_top_index)

	return top_visible_index
