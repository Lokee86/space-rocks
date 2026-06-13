extends RefCounted
class_name TransmissionFlow

var pregame_menu: Control
var screen_display: Control
var active_transmission: Control


func configure(pregame_menu_ref: Control) -> void:
	pregame_menu = pregame_menu_ref
	screen_display = pregame_menu.get_node_or_null("%ScreenDisplay") as Control if pregame_menu != null else null
	if screen_display == null and pregame_menu != null:
		var transmission_screen := pregame_menu.get_node_or_null(
			"CenterContainer/Window8/MainInterface/ScreenStack/TransmissionScreen"
		)
		if transmission_screen != null:
			screen_display = transmission_screen.get_node_or_null("ScreenDisplay") as Control
	if screen_display == null and pregame_menu != null:
		screen_display = pregame_menu.find_child("ScreenDisplay", true, false) as Control


func mount(transmission_scene: PackedScene) -> Control:
	if screen_display == null:
		return null

	clear()

	var transmission := transmission_scene.instantiate() as Control
	if transmission == null:
		return null

	screen_display.add_child(transmission)
	active_transmission = transmission
	return active_transmission


func clear() -> void:
	if screen_display == null:
		active_transmission = null
		return

	for child in screen_display.get_children():
		screen_display.remove_child(child)
		child.queue_free()

	active_transmission = null


func has_active_transmission() -> bool:
	if active_transmission != null and is_instance_valid(active_transmission):
		return true
	if screen_display != null and screen_display.get_child_count() > 0:
		return true
	return false


func get_screen_display() -> Control:
	return screen_display
