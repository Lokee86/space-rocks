class_name PregameMenuFlow
extends RefCounted

const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")
const LocalPilotFlowScript := preload("res://scripts/ui/menu_flow/local_pilot_flow.gd")
const MultiplayerRoomSetupScene := preload("res://scenes/ui/transmission_displays/multiplayer_room_setup_readout.tscn")

var pregame_menu: PregameMenu
var return_to_main_menu: Callable
var start_single_player_callable: Callable
var create_room_callable: Callable
var show_join_dialog_callable: Callable
var logout_callable: Callable
var clear_for_room_transition_callable: Callable
var profile_context_provider: ProfileContextProvider
var profile_flow: ProfileFlow
var transmission_flow: TransmissionFlow
var local_pilot_flow: LocalPilotFlow
var room_setup: MultiplayerRoomSetupReadout
var current_mode := ""


func configure(
		pregame_menu_ref: PregameMenu,
		return_to_main_menu_callable: Callable,
		start_single_player_callable_ref: Callable = Callable(),
		create_room_callable_ref: Callable = Callable(),
		show_join_dialog_callable_ref: Callable = Callable(),
		logout_callable_ref: Callable = Callable(),
		clear_for_room_transition_callable_ref: Callable = Callable(),
		profile_context_provider_ref: ProfileContextProvider = null,
		profile_flow_ref: ProfileFlow = null,
		transmission_flow_ref: TransmissionFlow = null) -> void:
	pregame_menu = pregame_menu_ref
	return_to_main_menu = return_to_main_menu_callable
	start_single_player_callable = start_single_player_callable_ref
	create_room_callable = create_room_callable_ref
	show_join_dialog_callable = show_join_dialog_callable_ref
	logout_callable = logout_callable_ref
	clear_for_room_transition_callable = clear_for_room_transition_callable_ref
	profile_context_provider = profile_context_provider_ref
	profile_flow = profile_flow_ref
	transmission_flow = transmission_flow_ref
	local_pilot_flow = LocalPilotFlowScript.new()
	if local_pilot_flow != null:
		local_pilot_flow.configure(transmission_flow, Callable(pregame_menu, "set_callsign"), profile_context_provider)

	if pregame_menu != null:
		if not pregame_menu.back_requested.is_connected(_on_back_requested):
			pregame_menu.back_requested.connect(_on_back_requested)
		if not pregame_menu.play_endless_requested.is_connected(_on_play_endless_requested):
			pregame_menu.play_endless_requested.connect(_on_play_endless_requested)
		if not pregame_menu.create_game_requested.is_connected(_on_create_game_requested):
			pregame_menu.create_game_requested.connect(_on_create_game_requested)
		if not pregame_menu.join_game_requested.is_connected(_on_join_game_requested):
			pregame_menu.join_game_requested.connect(_on_join_game_requested)
		if not pregame_menu.logout_requested.is_connected(_on_logout_requested):
			pregame_menu.logout_requested.connect(_on_logout_requested)
		if not pregame_menu.profile_requested.is_connected(_on_profile_requested):
			pregame_menu.profile_requested.connect(_on_profile_requested)
		if not pregame_menu.select_pilot_requested.is_connected(_on_select_pilot_requested):
			pregame_menu.select_pilot_requested.connect(_on_select_pilot_requested)


func show_single_player() -> void:
	current_mode = PregameMenuMode.SINGLE_PLAYER
	if pregame_menu != null:
		pregame_menu.show_single_player_mode()
	if local_pilot_flow != null:
		await local_pilot_flow.apply_saved_default()
	elif profile_context_provider != null:
		profile_context_provider.select_guest_profile()
	_update_callsign_indicator()


func show_multiplayer() -> void:
	current_mode = PregameMenuMode.MULTIPLAYER
	if pregame_menu != null:
		pregame_menu.show_multiplayer_mode()
	_update_callsign_indicator()


func _on_back_requested() -> void:
	if transmission_flow != null:
		if transmission_flow.has_active_subpanel():
			transmission_flow.clear_subpanel()
			return

		if transmission_flow.has_active_transmission():
			transmission_flow.clear()
			return

	if return_to_main_menu.is_valid():
		return_to_main_menu.call()


func _on_play_endless_requested() -> void:
	if current_mode != PregameMenuMode.SINGLE_PLAYER:
		return
	if start_single_player_callable.is_valid():
		start_single_player_callable.call()


func _on_create_game_requested() -> void:
	if current_mode != PregameMenuMode.MULTIPLAYER or transmission_flow == null:
		return
	room_setup = transmission_flow.mount_primary(MultiplayerRoomSetupScene) as MultiplayerRoomSetupReadout
	if room_setup == null:
		return
	room_setup.connect("create_requested", Callable(self, "_on_room_setup_create_requested"))
	room_setup.connect("cancel_requested", Callable(self, "_on_room_setup_cancel_requested"))


func _on_room_setup_create_requested(config: Dictionary) -> void:
	if room_setup != null:
		room_setup.set_status("Creating room...")
		room_setup.set_pending(true)
	if create_room_callable.is_valid():
		create_room_callable.call(config)


func _on_room_setup_cancel_requested() -> void:
	if transmission_flow != null:
		transmission_flow.clear_primary()
	room_setup = null


func show_room_setup_status(message: String) -> void:
	if room_setup == null or not is_instance_valid(room_setup):
		return
	room_setup.set_pending(false)
	room_setup.set_status(message)


func _on_join_game_requested() -> void:
	if current_mode != PregameMenuMode.MULTIPLAYER:
		return
	if show_join_dialog_callable.is_valid():
		show_join_dialog_callable.call()


func _on_logout_requested() -> void:
	if current_mode != PregameMenuMode.MULTIPLAYER:
		return
	if logout_callable.is_valid():
		logout_callable.call()
	if return_to_main_menu.is_valid():
		return_to_main_menu.call()


func _on_profile_requested() -> void:
	if profile_flow != null:
		await profile_flow.show_profile(current_mode)


func _on_select_pilot_requested() -> void:
	if current_mode != PregameMenuMode.SINGLE_PLAYER:
		return
	if local_pilot_flow != null:
		local_pilot_flow.show_selector()


func _update_callsign_indicator() -> void:
	if pregame_menu == null:
		return
	if profile_context_provider == null:
		return

	var context: Dictionary = profile_context_provider.context_for_mode(current_mode)
	pregame_menu.set_callsign(str(context.get("callsign", "Guest")))


func get_single_player_profile_context() -> Dictionary:
	if profile_context_provider != null:
		return profile_context_provider.context_for_mode(PregameMenuMode.SINGLE_PLAYER)

	return {
		"play_mode": PregameMenuMode.SINGLE_PLAYER,
		"identity_kind": "guest",
		"callsign": "Guest",
	}
