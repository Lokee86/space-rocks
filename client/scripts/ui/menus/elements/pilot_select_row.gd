extends HBoxContainer

signal selected(item: Dictionary)

@export var normal_style: StyleBox
@export var hover_style: StyleBox
@export var focus_style: StyleBox
@export var pressed_style: StyleBox

@onready var panel: PanelContainer = $PanelContainer

var item: Dictionary = {}
var is_hovered := false
var is_pressed := false


func _ready() -> void:
	focus_mode = Control.FOCUS_ALL
	mouse_filter = Control.MOUSE_FILTER_PASS

	panel.mouse_filter = Control.MOUSE_FILTER_IGNORE

	mouse_entered.connect(_on_mouse_entered)
	mouse_exited.connect(_on_mouse_exited)
	focus_entered.connect(_apply_style)
	focus_exited.connect(_apply_style)

	_apply_style()


func configure(display_text: String, item_data: Dictionary) -> void:
	item = item_data.duplicate(true)

	var label := get_node_or_null("%Label") as Label
	if label == null:
		label = get_node_or_null("PanelContainer/Label") as Label
	if label != null:
		label.text = display_text


func _gui_input(event: InputEvent) -> void:
	if event is InputEventMouseButton and event.button_index == MOUSE_BUTTON_LEFT:
		if event.pressed:
			is_pressed = true
			grab_focus()
			_apply_style()
			accept_event()
		else:
			var was_pressed := is_pressed
			is_pressed = false
			_apply_style()

			if was_pressed and is_hovered:
				selected.emit(item.duplicate(true))

			accept_event()


func _on_mouse_entered() -> void:
	is_hovered = true
	_apply_style()


func _on_mouse_exited() -> void:
	is_hovered = false
	is_pressed = false
	_apply_style()


func _apply_style() -> void:
	var style := normal_style

	if is_pressed and pressed_style != null:
		style = pressed_style
	elif has_focus() and focus_style != null:
		style = focus_style
	elif is_hovered and hover_style != null:
		style = hover_style

	if style != null:
		panel.add_theme_stylebox_override(&"panel", style)
