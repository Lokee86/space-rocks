extends OptionButton
class_name StyledOptionButton

const ButtonNormal := preload("res://assets/ui/ui_buttons/button.png")
const ButtonPressed := preload("res://assets/ui/ui_buttons/button_pressed.png")
const ButtonHover := preload("res://assets/ui/ui_buttons/button_hover.png")
const ButtonDisabled := preload("res://assets/ui/ui_buttons/button_disabled.png")
const DownArrow := preload("res://assets/ui/icons/down_arrow.png")
const InterfaceFont := preload("res://assets/fonts/PixelCaps-LOnE.ttf")


func _ready() -> void:
	custom_minimum_size = Vector2(maxf(custom_minimum_size.x, 110.0), maxf(custom_minimum_size.y, 38.0))
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
		var label := str(item.get("label", ""))
		add_item(label)
		set_item_metadata(item_count - 1, item.get("value"))
	if item_count > 0 and not select_value(selected_metadata):
		select(0)


func _apply_theme() -> void:
	var control_theme := Theme.new()
	control_theme.set_font("font", "OptionButton", InterfaceFont)
	control_theme.set_font_size("font_size", "OptionButton", 16)
	control_theme.set_color("font_color", "OptionButton", Color.BLACK)
	control_theme.set_color("font_hover_color", "OptionButton", Color.BLACK)
	control_theme.set_color("font_pressed_color", "OptionButton", Color.BLACK)
	control_theme.set_color("font_disabled_color", "OptionButton", Color(0.25, 0.25, 0.25))
	control_theme.set_icon("arrow", "OptionButton", DownArrow)
	control_theme.set_stylebox("normal", "OptionButton", _texture_style(ButtonNormal))
	control_theme.set_stylebox("hover", "OptionButton", _texture_style(ButtonHover))
	control_theme.set_stylebox("pressed", "OptionButton", _texture_style(ButtonPressed))
	control_theme.set_stylebox("disabled", "OptionButton", _texture_style(ButtonDisabled))
	control_theme.set_stylebox("focus", "OptionButton", StyleBoxEmpty.new())

	var popup_panel := StyleBoxFlat.new()
	popup_panel.bg_color = Color(0.035, 0.05, 0.055, 0.98)
	popup_panel.border_color = Color(0.84, 0.82, 0.48)
	popup_panel.set_border_width_all(2)
	popup_panel.set_content_margin_all(8)
	var popup_hover := StyleBoxFlat.new()
	popup_hover.bg_color = Color(0.84, 0.82, 0.48)
	popup_hover.set_content_margin_all(4)
	control_theme.set_font("font", "PopupMenu", InterfaceFont)
	control_theme.set_font_size("font_size", "PopupMenu", 16)
	control_theme.set_color("font_color", "PopupMenu", Color(0.95, 0.94, 0.78))
	control_theme.set_color("font_hover_color", "PopupMenu", Color.BLACK)
	control_theme.set_stylebox("panel", "PopupMenu", popup_panel)
	control_theme.set_stylebox("hover", "PopupMenu", popup_hover)
	theme = control_theme
	get_popup().theme = control_theme


func _texture_style(texture: Texture2D) -> StyleBoxTexture:
	var style := StyleBoxTexture.new()
	style.texture = texture
	style.texture_margin_left = 12.0
	style.texture_margin_top = 8.0
	style.texture_margin_right = 12.0
	style.texture_margin_bottom = 8.0
	style.set_content_margin_all(8.0)
	return style
