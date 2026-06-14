class_name LocalPilotFlow
extends RefCounted

const SelectPilotReadoutScene := preload("res://scenes/ui/transmission_displays/select_pilot_readout.tscn")
const EnterPilotIdScene := preload("res://scenes/ui/transmission_displays/sub-transmissions/enter_pilot_id.tscn")
const LocalPilotApiClient := preload("res://scripts/profile/local_pilot_api_client.gd")

var transmission_flow
var callsign_updated_callable: Callable
var local_pilot_api_client
var selector: Control


func configure(transmission_flow_ref, callsign_updated_callable_ref: Callable = Callable()) -> void:
	transmission_flow = transmission_flow_ref
	callsign_updated_callable = callsign_updated_callable_ref
	local_pilot_api_client = LocalPilotApiClient.new()


func show_selector() -> Control:
	if transmission_flow == null:
		return null

	selector = transmission_flow.mount(SelectPilotReadoutScene)
	if selector == null:
		return null

	if selector.has_signal("load_requested"):
		selector.connect("load_requested", Callable(self, "_on_load_requested"))
	if selector.has_signal("create_requested"):
		selector.connect("create_requested", Callable(self, "_on_create_requested"))

	_refresh_selector()
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
	if local_pilot_api_client == null:
		return

	var seed_from_guest_stats := true
	if selector != null and is_instance_valid(selector):
		var selected_item = selector.get("selected_item")
		if selected_item is Dictionary and str(selected_item.get("identity_kind", "")) != "":
			seed_from_guest_stats = str(selected_item.get("identity_kind", "")) == "guest"

	var result = await local_pilot_api_client.create_profile(callsign, seed_from_guest_stats)
	if result == null or !result.ok:
		var status_code_text := "unknown"
		var error_message_text := "unknown"
		if result != null:
			status_code_text = str(result.get("status_code", "unknown"))
			error_message_text = str(result.get("error_message", result.get("message", "unknown")))
		push_warning("local pilot creation failed: status_code=%s error=%s" % [status_code_text, error_message_text])
		return

	await _refresh_selector()
	_on_subpanel_cancel_requested()


func _refresh_selector() -> void:
	if selector == null or !is_instance_valid(selector):
		return
	if local_pilot_api_client == null:
		if selector.has_method("populate_pilots"):
			selector.populate_pilots([])
		return

	var result = await local_pilot_api_client.list_profiles()
	if selector == null or !is_instance_valid(selector):
		return

	if result != null and result.ok and result.body is Dictionary:
		var body: Dictionary = result.body
		if body.has("profiles") and body["profiles"] is Array and selector.has_method("populate_pilots"):
			selector.populate_pilots(body["profiles"])
			return

	if selector.has_method("populate_pilots"):
		selector.populate_pilots([])
