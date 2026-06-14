class_name LocalPilotFlow
extends RefCounted

const SelectPilotReadoutScene := preload("res://scenes/ui/transmission_displays/select_pilot_readout.tscn")

var transmission_flow
var callsign_updated_callable: Callable


func configure(transmission_flow_ref, callsign_updated_callable_ref := Callable()) -> void:
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

	if selector.has_signal("load_requested") and not selector.load_requested.is_connected(_on_load_requested):
		selector.load_requested.connect(_on_load_requested)

	return selector


func _on_load_requested(item: Dictionary) -> void:
	if item.get("identity_kind", "") == "guest" and callsign_updated_callable.is_valid():
		callsign_updated_callable.call("Guest")
