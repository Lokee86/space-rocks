class_name LocalPilotFlow
extends RefCounted

const SelectPilotReadoutScene := preload("res://scenes/ui/transmission_displays/select_pilot_readout.tscn")
const EnterPilotIdScene := preload("res://scenes/ui/transmission_displays/sub-transmissions/enter_pilot_id.tscn")

var transmission_flow
var callsign_updated_callable: Callable


func configure(transmission_flow_ref, callsign_updated_callable_ref: Callable = Callable()) -> void:
	transmission_flow = transmission_flow_ref
	callsign_updated_callable = callsign_updated_callable_ref


func show_selector() -> Control:
	if transmission_flow == null:
		return null

	var selector: Control = transmission_flow.mount(SelectPilotReadoutScene)
	if selector == null:
		return null

	if selector.has_method("populate_pilots"):
		selector.populate_pilots([])

	if selector.has_signal("load_requested"):
		selector.connect("load_requested", Callable(self, "_on_load_requested"))
	if selector.has_signal("create_requested"):
		selector.connect("create_requested", Callable(self, "_on_create_requested"))

	return selector


func _on_load_requested(item: Dictionary) -> void:
	if item.get("identity_kind", "") == "guest" and callsign_updated_callable.is_valid():
		callsign_updated_callable.call("Guest")


func _on_create_requested() -> void:
	if transmission_flow == null:
		return

	var mounted_scene: Control = transmission_flow.mount_subpanel(EnterPilotIdScene)
	if mounted_scene == null:
		return

	if mounted_scene.has_method("configure_create"):
		mounted_scene.configure_create()

	if mounted_scene.has_signal("cancel_requested"):
		mounted_scene.connect("cancel_requested", Callable(self, "_on_subpanel_cancel_requested"))
	if mounted_scene.has_signal("confirm_requested"):
		mounted_scene.connect("confirm_requested", Callable(self, "_on_create_confirmed"))


func _on_subpanel_cancel_requested() -> void:
	if transmission_flow == null:
		return

	transmission_flow.clear_subpanel()


func _on_create_confirmed(callsign: String) -> void:
	return
