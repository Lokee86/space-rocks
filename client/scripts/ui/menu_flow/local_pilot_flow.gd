class_name LocalPilotFlow
extends RefCounted

const SelectPilotReadoutScene := preload("res://scenes/ui/transmission_displays/select_pilot_readout.tscn")
const EnterPilotIdScene := preload("res://scenes/ui/transmission_displays/sub-transmissions/enter_pilot_id.tscn")
const LocalPilotApiClient := preload("res://scripts/profile/local_pilot_api_client.gd")

var transmission_flow
var callsign_updated_callable: Callable
var profile_context_provider
var local_pilot_api_client
var selector: Control
var active_entry_scene: Control


func configure(transmission_flow_ref, callsign_updated_callable_ref: Callable = Callable(), profile_context_provider_ref = null) -> void:
	transmission_flow = transmission_flow_ref
	callsign_updated_callable = callsign_updated_callable_ref
	profile_context_provider = profile_context_provider_ref
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
	var identity_kind := str(item.get("identity_kind", ""))
	if identity_kind == "guest":
		if profile_context_provider != null and profile_context_provider.has_method("select_guest_profile"):
			profile_context_provider.select_guest_profile()
		if callsign_updated_callable.is_valid():
			callsign_updated_callable.call("Guest")
		return

	if identity_kind == "local_profile":
		var local_profile_id := str(item.get("local_profile_id", ""))
		var display_name := str(item.get("display_name", ""))
		if profile_context_provider != null and profile_context_provider.has_method("select_local_profile"):
			profile_context_provider.select_local_profile(local_profile_id, display_name)
		if callsign_updated_callable.is_valid():
			callsign_updated_callable.call(display_name)


func _on_create_requested() -> void:
	if transmission_flow == null:
		return

	var mounted_scene: Control = transmission_flow.mount_subpanel(EnterPilotIdScene)
	if mounted_scene == null:
		return
	active_entry_scene = mounted_scene

	if mounted_scene.has_method("configure_create"):
		mounted_scene.configure_create()

	if mounted_scene.has_signal("cancel_requested"):
		mounted_scene.connect("cancel_requested", Callable(self, "_on_subpanel_cancel_requested"))
	if mounted_scene.has_signal("confirm_requested"):
		mounted_scene.connect("confirm_requested", Callable(self, "_on_create_confirmed"))


func _on_subpanel_cancel_requested() -> void:
	if transmission_flow == null:
		return

	active_entry_scene = null
	transmission_flow.clear_subpanel()


func _on_create_confirmed(callsign: String) -> void:
	if local_pilot_api_client == null:
		return

	var seed_from_guest_stats := true
	if profile_context_provider != null and profile_context_provider.has_method("context_for_mode"):
		var context: Dictionary = profile_context_provider.context_for_mode("SINGLE_PLAYER")
		var identity_kind := str(context.get("identity_kind", ""))
		if identity_kind == "guest":
			seed_from_guest_stats = true
		elif identity_kind == "local_profile":
			seed_from_guest_stats = false

	if active_entry_scene != null and is_instance_valid(active_entry_scene) and active_entry_scene.has_method("show_create_submitting"):
		active_entry_scene.show_create_submitting()

	var result = await local_pilot_api_client.create_profile(callsign, seed_from_guest_stats)
	if result == null or !result.ok:
		if active_entry_scene != null and is_instance_valid(active_entry_scene) and active_entry_scene.has_method("show_create_failed"):
			active_entry_scene.show_create_failed("CREATE FAILED")
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
