extends RefCounted
class_name TransmissionFlow

var pregame_menu: Control
var primary_screen_display: Control
var subpanel_screen_display: Control
var active_primary_transmission: Control
var active_subpanel_transmission: Control
var screen_display: Control
var active_transmission: Control
var primary_input_locks: Array = []


func configure(pregame_menu_ref: Control) -> void:
	pregame_menu = pregame_menu_ref
	primary_screen_display = pregame_menu.get_node_or_null(
		"CenterContainer/Window8/MainInterface/ScreenStack/TransmissionScreen/TransmissionScreen/ScreenDisplay"
	) as Control if pregame_menu != null else null
	subpanel_screen_display = pregame_menu.get_node_or_null(
		"CenterContainer/Window8/MainInterface/ScreenStack/SubPanel/SubPanelTransmissionScreen2/TransmissionScreen/ScreenDisplay"
	) as Control if pregame_menu != null else null
	if primary_screen_display == null and pregame_menu != null:
		var transmission_screen := pregame_menu.get_node_or_null(
			"CenterContainer/Window8/MainInterface/ScreenStack/TransmissionScreen"
		)
		if transmission_screen != null:
			primary_screen_display = transmission_screen.get_node_or_null("ScreenDisplay") as Control
	if primary_screen_display == null and pregame_menu != null:
		primary_screen_display = pregame_menu.get_node_or_null("%ScreenDisplay") as Control
	if primary_screen_display == null and pregame_menu != null:
		primary_screen_display = pregame_menu.find_child("ScreenDisplay", true, false) as Control
	screen_display = primary_screen_display


func mount(transmission_scene: PackedScene) -> Control:
	return mount_primary(transmission_scene)


func mount_primary(transmission_scene: PackedScene) -> Control:
	if primary_screen_display == null:
		return null

	clear_primary()

	var transmission := transmission_scene.instantiate() as Control
	if transmission == null:
		return null

	primary_screen_display.add_child(transmission)
	active_primary_transmission = transmission
	active_transmission = active_primary_transmission
	if has_active_subpanel():
		lock_primary_input()
	return active_primary_transmission


func mount_subpanel(transmission_scene: PackedScene) -> Control:
	if subpanel_screen_display == null:
		return null

	clear_subpanel()

	var transmission := transmission_scene.instantiate() as Control
	if transmission == null:
		return null

	subpanel_screen_display.add_child(transmission)
	active_subpanel_transmission = transmission
	lock_primary_input()
	return active_subpanel_transmission


func lock_primary_input() -> void:
	unlock_primary_input()

	if active_primary_transmission == null or not is_instance_valid(active_primary_transmission):
		return

	var viewport := active_primary_transmission.get_viewport()
	var focused_control: Control = null
	if viewport != null:
		focused_control = viewport.gui_get_focus_owner()
	if focused_control != null and active_primary_transmission.is_ancestor_of(focused_control):
		focused_control.release_focus()

	_lock_primary_input_node(active_primary_transmission)


func unlock_primary_input() -> void:
	for lock_entry in primary_input_locks:
		if not (lock_entry is Dictionary):
			continue

		var control: Control = lock_entry.get("node") as Control
		if control == null or not is_instance_valid(control):
			continue

		control.mouse_filter = lock_entry.get("mouse_filter", control.mouse_filter)
		control.focus_mode = lock_entry.get("focus_mode", control.focus_mode)

	primary_input_locks.clear()


func _lock_primary_input_node(node: Node) -> void:
	if node == null or not is_instance_valid(node):
		return

	if node is Control:
		var control: Control = node as Control
		primary_input_locks.append({
			"node": control,
			"mouse_filter": control.mouse_filter,
			"focus_mode": control.focus_mode,
		})
		control.mouse_filter = Control.MOUSE_FILTER_IGNORE
		control.focus_mode = Control.FOCUS_NONE

	for child in node.get_children():
		_lock_primary_input_node(child)


func clear() -> void:
	clear_primary()


func clear_primary() -> void:
	unlock_primary_input()

	if primary_screen_display == null:
		active_primary_transmission = null
		active_transmission = null
		return

	for child in primary_screen_display.get_children():
		primary_screen_display.remove_child(child)
		child.queue_free()

	active_primary_transmission = null
	active_transmission = null


func clear_subpanel() -> void:
	if subpanel_screen_display == null:
		active_subpanel_transmission = null
		unlock_primary_input()
		return

	for child in subpanel_screen_display.get_children():
		subpanel_screen_display.remove_child(child)
		child.queue_free()

	active_subpanel_transmission = null
	unlock_primary_input()


func has_active_transmission() -> bool:
	return has_active_primary()


func has_active_primary() -> bool:
	if active_primary_transmission != null and is_instance_valid(active_primary_transmission):
		return true
	if primary_screen_display != null and primary_screen_display.get_child_count() > 0:
		return true
	return false


func has_active_subpanel() -> bool:
	if active_subpanel_transmission != null and is_instance_valid(active_subpanel_transmission):
		return true
	if subpanel_screen_display != null and subpanel_screen_display.get_child_count() > 0:
		return true
	return false


func get_screen_display() -> Control:
	return get_primary_screen_display()


func get_primary_screen_display() -> Control:
	return primary_screen_display


func get_subpanel_screen_display() -> Control:
	return subpanel_screen_display
