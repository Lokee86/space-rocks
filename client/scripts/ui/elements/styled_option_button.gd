extends OptionButton
class_name StyledOptionButton

const InterfaceFont := preload("res://assets/fonts/Modeseven-L3n5.ttf")


func _ready() -> void:
	custom_minimum_size = Vector2(maxf(custom_minimum_size.x, 96.0), maxf(custom_minimum_size.y, 30.0))
	focus_mode = Control.FOCUS_ALL
	fit_to_longest_item = false
	_apply_theme()


func select_value(value: Variant) -> bool:
	for index in range(item_count):
		var metadata = get_item_metadata(index)
		if typeof(metadata) == typeof(value) and metadata == value:
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
	control_theme.set_font("font", "OptionButton", InterfaceFont)
	control_theme.set_font_size("font_size", "OptionButton", 18)
	control_theme.set_color("font_color", "OptionButton", Color(0.08, 0.08, 0.08))
	control_theme.set_color("font_hover_color", "OptionButton", Color(0.02, 0.02, 0.02))
	control_theme.set_color("font_pressed_color", "OptionButton", Color(0.02, 0.02, 0.02))
	control_theme.set_color("font_disabled_color", "OptionButton", Color(0.35, 0.35, 0.35))
	control_theme.set_stylebox("normal", "OptionButton", _flat_style(Color(0.66, 0.69, 0.67), Color(0.13, 0.14, 0.14)))
	control_theme.set_stylebox("hover", "OptionButton", _flat_style(Color(0.78, 0.81, 0.79), Color(0.05, 0.06, 0.06)))
	control_theme.set_stylebox("pressed", "OptionButton", _flat_style(Color(0.54, 0.58, 0.56), Color(0.02, 0.03, 0.03)))
	control_theme.set_stylebox("disabled", "OptionButton", _flat_style(Color(0.48, 0.5, 0.49), Color(0.18, 0.18, 0.18)))
	control_theme.set_stylebox("focus", "OptionButton", _outline_style())
	control_theme.set_constant("arrow_margin", "OptionButton", 8)

	var popup_panel := _flat_style(Color(0.18, 0.19, 0.19), Color(0.04, 0.04, 0.04))
	popup_panel.set_content_margin_all(4.0)
	var popup_hover := _flat_style(Color(0.72, 0.75, 0.73), Color(0.08, 0.08, 0.08))
	control_theme.set_font("font", "PopupMenu", InterfaceFont)
	control_theme.set_font_size("font_size", "PopupMenu", 18)
	control_theme.set_color("font_color", "PopupMenu", Color(0.94, 0.94, 0.9))
	control_theme.set_color("font_hover_color", "PopupMenu", Color(0.05, 0.05, 0.05))
	control_theme.set_stylebox("panel", "PopupMenu", popup_panel)
	control_theme.set_stylebox("hover", "PopupMenu", popup_hover)
	theme = control_theme
	get_popup().theme = control_theme


func _flat_style(background: Color, border: Color) -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = background
	style.border_color = border
	style.set_border_width_all(2)
	style.set_content_margin(SIDE_LEFT, 8.0)
	style.set_content_margin(SIDE_TOP, 4.0)
	style.set_content_margin(SIDE_RIGHT, 8.0)
	style.set_content_margin(SIDE_BOTTOM, 4.0)
	return style


func _outline_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0, 0, 0, 0)
	style.border_color = Color(0.85, 0.85, 0.72)
	style.set_border_width_all(2)
	return style
