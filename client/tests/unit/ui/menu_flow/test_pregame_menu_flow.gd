extends GutTest

const PregameMenuFlow := preload("res://scripts/ui/menu_flow/pregame_menu_flow.gd")
const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")


class FakePregameMenu:
	extends PregameMenu

	var single_player_calls := 0
	var multiplayer_calls := 0
	var last_callsign := ""

	func _ready() -> void:
		pass

	func show_single_player_mode() -> void:
		single_player_calls += 1

	func show_multiplayer_mode() -> void:
		multiplayer_calls += 1

	func set_callsign(callsign: String) -> void:
		last_callsign = callsign


class FakeProfileContextProvider:
	extends ProfileContextProvider

	var last_mode := ""

	func context_for_mode(mode: String) -> Dictionary:
		last_mode = mode
		if mode == PregameMenuMode.MULTIPLAYER:
			return {
				"callsign": "Ada",
				"identity_kind": "authenticated_account",
				"activity_status": "ACTIVE",
			}
		return {
			"callsign": "Guest",
			"identity_kind": "guest",
			"activity_status": "OFFLINE",
		}


class FakeTransmissionFlow:
	extends TransmissionFlow

	var active := false
	var clear_calls := 0
	var clear_primary_calls := 0
	var mounted_primary: Control

	func has_active_subpanel() -> bool:
		return false

	func has_active_transmission() -> bool:
		return active

	func mount_primary(scene: PackedScene) -> Control:
		mounted_primary = scene.instantiate() as Control
		active = mounted_primary != null
		return mounted_primary

	func clear() -> void:
		clear_calls += 1
		active = false

	func clear_primary() -> void:
		clear_primary_calls += 1
		active = false
		mounted_primary = null


class FakeProfileFlow:
	extends ProfileFlow

	var show_calls := 0
	var last_mode := ""

	func show_profile(mode: String) -> ProfileReadout:
		show_calls += 1
		last_mode = mode
		return null


class LocalProfileContextProvider:
	extends ProfileContextProvider

	func context_for_mode(_mode: String) -> Dictionary:
		return {
			"callsign": "Ranger",
			"identity_kind": "local_profile",
			"local_profile_id": "pilot-local-1",
			"activity_status": "ACTIVE",
		}


class LoadoutRequestProbe:
	extends RefCounted

	var calls: Array = []

	func capture(local_profile_id: String, play_mode: String, mode_id: String) -> void:
		calls.append([local_profile_id, play_mode, mode_id])


class LoadoutSubmitProbe:
	extends RefCounted

	var selections: Array = []

	func capture(selection: Dictionary) -> void:
		selections.append(selection.duplicate(true))


class ReturnToMainMenuProbe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


class StartSinglePlayerProbe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


class Probe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


class ConfigProbe:
	extends RefCounted

	var calls := 0
	var last_config := {}

	func mark_called(config: Dictionary) -> void:
		calls += 1
		last_config = config.duplicate(true)


func test_show_single_player_sets_current_mode_and_calls_menu() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()

	add_child_autofree(menu)
	flow.configure(menu, Callable(), Callable(), Callable(), Callable(), Callable(), Callable(), profile_context_provider)
	await flow.show_single_player()

	assert_eq(flow.current_mode, PregameMenuMode.SINGLE_PLAYER)
	assert_eq(menu.single_player_calls, 1)
	assert_eq(menu.multiplayer_calls, 0)
	assert_eq(menu.last_callsign, "Guest")
	assert_eq(profile_context_provider.last_mode, PregameMenuMode.SINGLE_PLAYER)


func test_show_multiplayer_sets_current_mode_and_calls_menu() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()

	add_child_autofree(menu)
	flow.configure(menu, Callable(), Callable(), Callable(), Callable(), Callable(), Callable(), profile_context_provider)
	flow.show_multiplayer()

	assert_eq(flow.current_mode, PregameMenuMode.MULTIPLAYER)
	assert_eq(menu.single_player_calls, 0)
	assert_eq(menu.multiplayer_calls, 1)
	assert_eq(menu.last_callsign, "Ada")
	assert_eq(profile_context_provider.last_mode, PregameMenuMode.MULTIPLAYER)


func test_back_requested_calls_return_to_main_menu_once() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var return_probe := ReturnToMainMenuProbe.new()

	add_child_autofree(menu)
	flow.configure(menu, Callable(return_probe, "mark_called"), Callable(), Callable(), Callable(), Callable(), Callable(), profile_context_provider)

	menu.back_requested.emit()

	assert_eq(return_probe.calls, 1)


func test_back_clears_active_transmission_before_returning_to_main_menu() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var return_probe := ReturnToMainMenuProbe.new()
	var transmission_flow := FakeTransmissionFlow.new()
	transmission_flow.active = true

	add_child_autofree(menu)
	flow.configure(menu, Callable(return_probe, "mark_called"), Callable(), Callable(), Callable(), Callable(), Callable(), profile_context_provider, null, transmission_flow)

	menu.back_requested.emit()

	assert_eq(transmission_flow.clear_calls, 1)
	assert_eq(return_probe.calls, 0)


func test_back_returns_to_main_menu_when_no_active_transmission() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var return_probe := ReturnToMainMenuProbe.new()
	var transmission_flow := FakeTransmissionFlow.new()

	add_child_autofree(menu)
	flow.configure(menu, Callable(return_probe, "mark_called"), Callable(), Callable(), Callable(), Callable(), Callable(), profile_context_provider, null, transmission_flow)

	menu.back_requested.emit()

	assert_eq(transmission_flow.clear_calls, 0)
	assert_eq(return_probe.calls, 1)


func test_play_endless_requested_calls_start_single_player_when_single_player_mode() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var start_probe := StartSinglePlayerProbe.new()

	add_child_autofree(menu)
	flow.configure(menu, Callable(), Callable(start_probe, "mark_called"), Callable(), Callable(), Callable(), Callable(), profile_context_provider)
	await flow.show_single_player()

	menu.play_endless_requested.emit()

	assert_eq(start_probe.calls, 1)


func test_play_endless_requested_does_not_call_start_single_player_when_multiplayer_mode() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var start_probe := StartSinglePlayerProbe.new()

	add_child_autofree(menu)
	flow.configure(menu, Callable(), Callable(start_probe, "mark_called"), Callable(), Callable(), Callable(), Callable(), profile_context_provider)
	flow.show_multiplayer()

	menu.play_endless_requested.emit()

	assert_eq(start_probe.calls, 0)


func test_multiplayer_create_opens_setup_then_confirms_configured_room() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var clear_probe := Probe.new()
	var create_probe := ConfigProbe.new()
	var transmission_flow := FakeTransmissionFlow.new()

	add_child_autofree(menu)
	flow.configure(
		menu,
		Callable(),
		Callable(),
		Callable(create_probe, "mark_called"),
		Callable(),
		Callable(),
		Callable(clear_probe, "mark_called"),
		profile_context_provider,
		null,
		transmission_flow)
	flow.show_multiplayer()

	menu.create_game_requested.emit()

	assert_not_null(transmission_flow.mounted_primary)
	assert_eq(clear_probe.calls, 0)
	assert_eq(create_probe.calls, 0)
	var config := {
		"team_structure": "auto_balanced",
		"team_assignment_mode": "",
		"team_count": 3,
		"max_players": 8,
	}
	transmission_flow.mounted_primary.emit_signal("create_requested", config)

	assert_eq(clear_probe.calls, 1)
	assert_eq(create_probe.calls, 1)
	assert_eq(create_probe.last_config, config)
	transmission_flow.mounted_primary.free()
	transmission_flow.mounted_primary = null


func test_multiplayer_join_calls_show_join_dialog() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var join_probe := Probe.new()

	add_child_autofree(menu)
	flow.configure(
		menu,
		Callable(),
		Callable(),
		Callable(),
		Callable(join_probe, "mark_called"),
		Callable(),
		Callable(),
		profile_context_provider)
	flow.show_multiplayer()

	menu.join_game_requested.emit()

	assert_eq(join_probe.calls, 1)


func test_loadout_requested_uses_current_local_profile_and_mode() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := LocalProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var request_probe := LoadoutRequestProbe.new()

	add_child_autofree(menu)
	flow.configure(
		menu,
		Callable(),
		Callable(),
		Callable(),
		Callable(),
		Callable(),
		Callable(),
		profile_context_provider,
		null,
		null,
		Callable(request_probe, "capture")
	)
	await flow.show_single_player()

	menu.loadout_requested.emit()

	assert_eq(request_probe.calls, [["pilot-local-1", PregameMenuMode.SINGLE_PLAYER, "arcade_survival"]])


func test_loadout_submit_forwards_selection() -> void:
	var menu := FakePregameMenu.new()
	var flow := PregameMenuFlow.new()
	var submit_probe := LoadoutSubmitProbe.new()
	var selection := {"selected_owned_ship_id": "ship-owned-1"}

	add_child_autofree(menu)
	flow.configure(
		menu,
		Callable(),
		Callable(),
		Callable(),
		Callable(),
		Callable(),
		Callable(),
		null,
		null,
		null,
		Callable(),
		Callable(submit_probe, "capture")
	)

	flow._on_loadout_submit_requested(selection)

	assert_eq(submit_probe.selections, [selection])


func test_profile_requested_calls_profile_flow_with_single_player_mode() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var profile_flow := FakeProfileFlow.new()

	add_child_autofree(menu)
	flow.configure(menu, Callable(), Callable(), Callable(), Callable(), Callable(), Callable(), profile_context_provider, profile_flow)
	await flow.show_single_player()

	menu.profile_requested.emit()
	await get_tree().process_frame

	assert_eq(profile_flow.show_calls, 1)
	assert_eq(profile_flow.last_mode, PregameMenuMode.SINGLE_PLAYER)


func test_profile_requested_calls_profile_flow_with_multiplayer_mode() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var profile_flow := FakeProfileFlow.new()

	add_child_autofree(menu)
	flow.configure(menu, Callable(), Callable(), Callable(), Callable(), Callable(), Callable(), profile_context_provider, profile_flow)
	flow.show_multiplayer()

	menu.profile_requested.emit()
	await get_tree().process_frame

	assert_eq(profile_flow.show_calls, 1)
	assert_eq(profile_flow.last_mode, PregameMenuMode.MULTIPLAYER)


func test_multiplayer_logout_calls_logout_and_return_to_main() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var logout_probe := Probe.new()
	var return_probe := ReturnToMainMenuProbe.new()

	add_child_autofree(menu)
	flow.configure(
		menu,
		Callable(return_probe, "mark_called"),
		Callable(),
		Callable(),
		Callable(),
		Callable(logout_probe, "mark_called"),
		Callable(),
		profile_context_provider)
	flow.show_multiplayer()

	menu.logout_requested.emit()

	assert_eq(logout_probe.calls, 1)
	assert_eq(return_probe.calls, 1)


func test_multiplayer_create_does_nothing_in_single_player_mode() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var clear_probe := Probe.new()
	var create_probe := Probe.new()

	add_child_autofree(menu)
	flow.configure(
		menu,
		Callable(),
		Callable(),
		Callable(create_probe, "mark_called"),
		Callable(),
		Callable(),
		Callable(clear_probe, "mark_called"),
		profile_context_provider)
	await flow.show_single_player()

	menu.create_game_requested.emit()

	assert_eq(clear_probe.calls, 0)
	assert_eq(create_probe.calls, 0)


func test_multiplayer_join_does_nothing_in_single_player_mode() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var join_probe := Probe.new()

	add_child_autofree(menu)
	flow.configure(
		menu,
		Callable(),
		Callable(),
		Callable(),
		Callable(join_probe, "mark_called"),
		Callable(),
		Callable(),
		profile_context_provider)
	await flow.show_single_player()

	menu.join_game_requested.emit()

	assert_eq(join_probe.calls, 0)


func test_multiplayer_logout_does_nothing_in_single_player_mode() -> void:
	var menu := FakePregameMenu.new()
	var profile_context_provider := FakeProfileContextProvider.new()
	var flow := PregameMenuFlow.new()
	var logout_probe := Probe.new()
	var return_probe := ReturnToMainMenuProbe.new()

	add_child_autofree(menu)
	flow.configure(
		menu,
		Callable(return_probe, "mark_called"),
		Callable(),
		Callable(),
		Callable(),
		Callable(logout_probe, "mark_called"),
		Callable(),
		profile_context_provider)
	await flow.show_single_player()

	menu.logout_requested.emit()

	assert_eq(logout_probe.calls, 0)
	assert_eq(return_probe.calls, 0)
