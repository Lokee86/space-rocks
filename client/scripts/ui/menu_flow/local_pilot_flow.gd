class_name LocalPilotFlow
extends RefCounted

const SelectPilotReadoutScene := preload("res://scenes/ui/transmission_displays/select_pilot_readout.tscn")
const EnterPilotIdScene := preload("res://scenes/ui/transmission_displays/sub-transmissions/enter_pilot_id.tscn")
const ConfirmDeleteScene := preload("res://scenes/ui/transmission_displays/sub-transmissions/confirm_delete.tscn")
const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")
const ProfileIdentityKindScript := preload("res://scripts/profile/profile_identity_kind.gd")
const LocalPilotApiClientScript := preload("res://scripts/profile/local_pilot_api_client.gd")

var transmission_flow: TransmissionFlow
var callsign_updated_callable: Callable
var profile_context_provider: ProfileContextProvider
var local_pilot_api_client: LocalPilotApiClient
var selector: SelectPilotReadout
var active_pilot_editor: EnterPilotId
var active_delete_confirmation: ConfirmDelete
var active_edit_item: Dictionary = {}


func configure(transmission_flow_ref: TransmissionFlow = null, callsign_updated_callable_ref: Callable = Callable(), profile_context_provider_ref: ProfileContextProvider = null) -> void:
	transmission_flow = transmission_flow_ref
	callsign_updated_callable = callsign_updated_callable_ref
	profile_context_provider = profile_context_provider_ref
	local_pilot_api_client = LocalPilotApiClientScript.new()


func show_selector() -> SelectPilotReadout:
	if transmission_flow == null:
		return null

	selector = transmission_flow.mount(SelectPilotReadoutScene) as SelectPilotReadout
	if selector == null:
		transmission_flow.clear()
		return null

	if not selector.load_requested.is_connected(_on_load_requested):
		selector.load_requested.connect(_on_load_requested)
	if not selector.create_requested.is_connected(_on_create_requested):
		selector.create_requested.connect(_on_create_requested)
	if not selector.edit_requested.is_connected(_on_edit_requested):
		selector.edit_requested.connect(_on_edit_requested)
	if not selector.delete_requested.is_connected(_on_delete_requested):
		selector.delete_requested.connect(_on_delete_requested)

	_refresh_selector()
	return selector


func apply_saved_default() -> void:
	if local_pilot_api_client == null:
		_apply_guest_default()
		return

	var result = await local_pilot_api_client.get_default_profile()
	if result == null or !result.ok or !(result.body is Dictionary):
		_apply_guest_default()
		return

	var body: Dictionary = result.body
	if !body.has("default_profile") or !(body["default_profile"] is Dictionary):
		_apply_guest_default()
		return

	var default_profile: Dictionary = body["default_profile"]
	var identity_kind := str(default_profile.get("identity_kind", ""))
	if identity_kind == ProfileIdentityKindScript.GUEST:
		_apply_guest_default()
		return

	if identity_kind == ProfileIdentityKindScript.LOCAL_PROFILE:
		var local_profile_id := str(default_profile.get("local_profile_id", ""))
		var display_name := str(default_profile.get("display_name", ""))
		if local_profile_id == "" or display_name == "":
			_apply_guest_default()
			return

		if profile_context_provider != null:
			profile_context_provider.select_local_profile(local_profile_id, display_name)
		if callsign_updated_callable.is_valid():
			callsign_updated_callable.call(display_name)
		return

	_apply_guest_default()


func _on_load_requested(item: Dictionary) -> void:
	var identity_kind := str(item.get("identity_kind", ""))
	if identity_kind == ProfileIdentityKindScript.GUEST:
		var guest_result = await local_pilot_api_client.set_default_profile(ProfileIdentityKindScript.GUEST, "")
		if guest_result == null or !guest_result.ok:
			return

		if profile_context_provider != null:
			profile_context_provider.select_guest_profile()
		if callsign_updated_callable.is_valid():
			callsign_updated_callable.call("Guest")
		return

	if identity_kind == ProfileIdentityKindScript.LOCAL_PROFILE:
		var local_profile_id := str(item.get("local_profile_id", ""))
		var display_name := str(item.get("display_name", ""))
		var local_profile_result = await local_pilot_api_client.set_default_profile(ProfileIdentityKindScript.LOCAL_PROFILE, local_profile_id)
		if local_profile_result == null or !local_profile_result.ok:
			return

		if profile_context_provider != null:
			profile_context_provider.select_local_profile(local_profile_id, display_name)
		if callsign_updated_callable.is_valid():
			callsign_updated_callable.call(display_name)


func _apply_guest_default() -> void:
	if profile_context_provider != null:
		profile_context_provider.select_guest_profile()
	if callsign_updated_callable.is_valid():
		callsign_updated_callable.call("Guest")


func _on_create_requested() -> void:
	if transmission_flow == null:
		return

	var mounted_scene := transmission_flow.mount_subpanel(EnterPilotIdScene) as EnterPilotId
	if mounted_scene == null:
		transmission_flow.clear_subpanel()
		return
	active_delete_confirmation = null
	active_pilot_editor = mounted_scene

	mounted_scene.configure_create()

	if not mounted_scene.cancel_requested.is_connected(_on_subpanel_cancel_requested):
		mounted_scene.cancel_requested.connect(_on_subpanel_cancel_requested)
	if not mounted_scene.confirm_requested.is_connected(_on_create_confirmed):
		mounted_scene.confirm_requested.connect(_on_create_confirmed)


func _on_edit_requested(item: Dictionary) -> void:
	if transmission_flow == null:
		return

	var identity_kind := str(item.get("identity_kind", ""))
	if identity_kind != ProfileIdentityKindScript.LOCAL_PROFILE:
		return

	var local_profile_id := str(item.get("local_profile_id", ""))
	if local_profile_id == "":
		return

	var mounted_scene := transmission_flow.mount_subpanel(EnterPilotIdScene) as EnterPilotId
	if mounted_scene == null:
		transmission_flow.clear_subpanel()
		return
	active_delete_confirmation = null
	active_pilot_editor = mounted_scene
	active_edit_item = item.duplicate(true)

	var display_name := str(item.get("display_name", ""))
	mounted_scene.configure_label("ENTER NEW CALLSIGN", display_name)

	if not mounted_scene.cancel_requested.is_connected(_on_subpanel_cancel_requested):
		mounted_scene.cancel_requested.connect(_on_subpanel_cancel_requested)
	if not mounted_scene.confirm_requested.is_connected(_on_edit_confirmed):
		mounted_scene.confirm_requested.connect(_on_edit_confirmed)


func _on_subpanel_cancel_requested() -> void:
	active_pilot_editor = null
	active_delete_confirmation = null
	active_edit_item = {}
	if transmission_flow == null:
		return
	transmission_flow.clear_subpanel()


func _on_create_confirmed(callsign: String) -> void:
	if local_pilot_api_client == null:
		return

	var seed_from_guest_stats := true
	if profile_context_provider != null:
		var context: Dictionary = profile_context_provider.context_for_mode(PregameMenuMode.SINGLE_PLAYER)
		var identity_kind := str(context.get("identity_kind", ""))
		if identity_kind == ProfileIdentityKindScript.GUEST:
			seed_from_guest_stats = true
		elif identity_kind == ProfileIdentityKindScript.LOCAL_PROFILE:
			seed_from_guest_stats = false

	if active_pilot_editor != null and is_instance_valid(active_pilot_editor):
		active_pilot_editor.show_create_submitting()

	var result = await local_pilot_api_client.create_profile(callsign, seed_from_guest_stats)
	if result == null or !result.ok:
		if active_pilot_editor != null and is_instance_valid(active_pilot_editor):
			active_pilot_editor.show_create_failed("CREATE FAILED")
		return

	await _refresh_selector()
	_on_subpanel_cancel_requested()


func _on_edit_confirmed(callsign: String) -> void:
	if local_pilot_api_client == null:
		return

	var identity_kind := str(active_edit_item.get("identity_kind", ""))
	if identity_kind != ProfileIdentityKindScript.LOCAL_PROFILE:
		return

	var local_profile_id := str(active_edit_item.get("local_profile_id", ""))
	if local_profile_id == "":
		return

	if active_pilot_editor != null and is_instance_valid(active_pilot_editor):
		active_pilot_editor.show_submitting("UPDATING...")

	var result = await local_pilot_api_client.update_profile_display_name(local_profile_id, callsign)
	if result == null or !result.ok:
		if active_pilot_editor != null and is_instance_valid(active_pilot_editor):
			active_pilot_editor.show_failed("UPDATE FAILED")
		return

	var should_update_active_profile := false
	if profile_context_provider != null:
		var context: Dictionary = profile_context_provider.context_for_mode(PregameMenuMode.SINGLE_PLAYER)
		if str(context.get("identity_kind", "")) == ProfileIdentityKindScript.LOCAL_PROFILE and str(context.get("local_profile_id", "")) == local_profile_id:
			should_update_active_profile = true

	if should_update_active_profile:
		if profile_context_provider != null:
			profile_context_provider.select_local_profile(local_profile_id, callsign)
		if callsign_updated_callable.is_valid():
			callsign_updated_callable.call(callsign)

	await _refresh_selector()
	if selector != null and is_instance_valid(selector):
		selector.select_item_by_identity(ProfileIdentityKindScript.LOCAL_PROFILE, local_profile_id)

	_on_subpanel_cancel_requested()


func _on_delete_requested(item: Dictionary) -> void:
	if transmission_flow == null:
		return

	var identity_kind := str(item.get("identity_kind", ""))
	if identity_kind != ProfileIdentityKindScript.LOCAL_PROFILE:
		return

	var local_profile_id := str(item.get("local_profile_id", ""))
	if local_profile_id == "":
		return

	var mounted_scene := transmission_flow.mount_subpanel(ConfirmDeleteScene) as ConfirmDelete
	if mounted_scene == null:
		transmission_flow.clear_subpanel()
		return
	active_pilot_editor = null
	active_edit_item = {}
	active_delete_confirmation = mounted_scene

	mounted_scene.configure_delete(item)
	if not mounted_scene.cancel_requested.is_connected(_on_subpanel_cancel_requested):
		mounted_scene.cancel_requested.connect(_on_subpanel_cancel_requested)
	if not mounted_scene.confirm_requested.is_connected(_on_delete_confirmed):
		mounted_scene.confirm_requested.connect(_on_delete_confirmed)


func _on_delete_confirmed(item: Dictionary) -> void:
	if local_pilot_api_client == null:
		return

	var identity_kind := str(item.get("identity_kind", ""))
	if identity_kind != ProfileIdentityKindScript.LOCAL_PROFILE:
		return

	var local_profile_id := str(item.get("local_profile_id", ""))
	if local_profile_id == "":
		return

	var result = await local_pilot_api_client.delete_profile(local_profile_id)
	if result == null or !result.ok:
		return

	var should_apply_guest_default := false
	if profile_context_provider != null:
		var context: Dictionary = profile_context_provider.context_for_mode(PregameMenuMode.SINGLE_PLAYER)
		if str(context.get("identity_kind", "")) == ProfileIdentityKindScript.LOCAL_PROFILE and str(context.get("local_profile_id", "")) == local_profile_id:
			should_apply_guest_default = true

	if should_apply_guest_default:
		_apply_guest_default()

	await _refresh_selector()
	_on_subpanel_cancel_requested()


func _refresh_selector() -> void:
	if selector == null or !is_instance_valid(selector):
		return
	if local_pilot_api_client == null:
		selector.populate_pilots([])
		return

	var result = await local_pilot_api_client.list_profiles()
	if selector == null or !is_instance_valid(selector):
		return

	if result != null and result.ok and result.body is Dictionary:
		var body: Dictionary = result.body
		if body.has("profiles") and body["profiles"] is Array:
			selector.populate_pilots(body["profiles"])
			_select_selector_default_row()
			return

	selector.populate_pilots([])
	_select_selector_default_row()


func _select_selector_default_row() -> void:
	if selector == null or !is_instance_valid(selector):
		return
	var identity_kind := ProfileIdentityKindScript.GUEST
	var local_profile_id := ""
	if profile_context_provider != null:
		var context: Dictionary = profile_context_provider.context_for_mode(PregameMenuMode.SINGLE_PLAYER)
		identity_kind = str(context.get("identity_kind", ProfileIdentityKindScript.GUEST))
		local_profile_id = str(context.get("local_profile_id", ""))

	selector.select_item_by_identity(identity_kind, local_profile_id)
