extends Node


func should_gameplay_ui_receive_mouse_event(event: InputEvent, gameplay_ui_root: Control, viewport: Viewport) -> bool:
	if not event is InputEventMouseButton or not event.pressed:
		return false

	if gameplay_ui_root == null or not gameplay_ui_root.visible:
		return false

	if viewport == null:
		return false

	var hovered_control: Control = viewport.gui_get_hovered_control()
	if hovered_control == null:
		return false

	if hovered_control == gameplay_ui_root:
		return true

	return gameplay_ui_root.is_ancestor_of(hovered_control)


func should_hud_receive_mouse_event(event: InputEvent, hud: Control, viewport: Viewport) -> bool:
	if not event is InputEventMouseButton or not event.pressed:
		return false

	if hud == null or not hud.visible:
		return false

	if viewport == null:
		return false

	var hovered_control: Control = viewport.gui_get_hovered_control()
	if hovered_control == hud:
		return true

	return hud.is_ancestor_of(hovered_control)

