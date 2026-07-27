extends OptionButton
class_name StyledOptionButton

const TransmissionNormal := preload("res://assets/ui/FlatBoxStyles/transmission_box.tres")
const TransmissionPressed := preload("res://assets/ui/FlatBoxStyles/transmission_pressed.tres")
const TransmissionFocus := preload("res://assets/ui/FlatBoxStyles/transmission_focus.tres")
const TransmissionFont := preload("res://assets/fonts/DigitalNormal-xO6j.otf")


func _ready() -> void:
	custom_minimum_size = Vector2(maxf(custom_minimum_size.x, 110.0), maxf(custom_minimum_size.y, 34.0))
	focus_mode = Control.FOCUS_ALL
	fit_to_longest_item = false
	_apply_theme()


func select_value(value: Variant) -> bool:
	for index in range(item_count):
		if get_item_metadata(index) == value:
			select(index)
			return true
	return false


func selected_value(default_value: Variant = null) -> Variant:
	if selected < 0 or selected >= item_count:
		return default_value
	return get_item_metadata(selected)


func replace_items(items: Array, selected_metadata: Variant = null) -> void:
	clear()
	for item in items:
		if not (item is Dictionary):
			continue
		add_item(str(item.get("label", "")))
		set_item_metadata(item_count - 1, item.get("value"))
	if item_count > 0 and not select_value(selected_metadata):
		select(0)


func _apply_theme() -> void:
	var control_theme := Theme.new()
	control_theme.set_font("font", "OptionButton", TransmissionFont)
	control_theme.set_font_size("font_size", "OptionButton", 22)
	control_theme.set_color("font_color", "OptionButton", Color(0, 1, 1))
	control_theme.set_color("font_hover_color", "OptionButton", Color.WHITE)
	control_theme.set_color("font_pressed_color", "OptionButton", Color.WHITE)
	control_theme.set_color("font_disabled_color", "OptionButton", Color(0.45, 0.55, 0.55))
	control_theme.set_stylebox("normal", "OptionButton", TransmissionNormal)
	control_theme.set_stylebox("hover", "OptionButton", TransmissionFocus)
	control_theme.set_stylebox("pressed", "OptionButton", TransmissionPressed)
	control_theme.set_stylebox("disabled", "OptionButton", TransmissionNormal)
	control_theme.set_stylebox("focus", "OptionButton", TransmissionFocus)

	var popup_panel := StyleBoxFlat.new()
	popup_panel.bg_color = Color(0.015, 0.09, 0.1, 0.98)
	popup_panel.border_color = Color(0, 1, 1)
	popup_panel.set_border_width_all(1)
	popup_panel.set_content_margin_all(5)
	var popup_hover := StyleBoxFlat.new()
	popup_hover.bg_color = Color(0, 0.35, 0.38)
	popup_hover.set_content_margin_all(3)
	control_theme.set_font("font", "PopupMenu", TransmissionFont)
	control_theme.set_font_size("font_size", "PopupMenu", 21)
	control_theme.set_color("font_color", "PopupMenu", Color(0, 1, 1))
	control_theme.set_color("font_hover_color", "PopupMenu", Color.WHITE)
	control_theme.set_stylebox("panel", "PopupMenu", popup_panel)
	control_theme.set_stylebox("hover", "PopupMenu", popup_hover)
	theme = control_theme
	get_popup().theme = control_theme
