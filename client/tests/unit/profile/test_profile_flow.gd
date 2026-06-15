extends GutTest

const ProfileFlow := preload("res://scripts/profile/profile_flow.gd")
const ProfileReadoutScene := preload("res://scenes/ui/transmission_displays/profile_readout.tscn")
const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")


class FakeContextProvider:
	extends RefCounted

	var context := {
		"play_mode": "single_player",
		"identity_kind": "guest",
		"callsign": "Guest",
		"activity_status": "OFFLINE",
	}
	var last_mode := ""

	func context_for_mode(mode: String) -> Dictionary:
		last_mode = mode
		return context.duplicate(true)


class FakeStatsProvider:
	extends RefCounted

	var profile := {
		"callsign": "Guest",
		"activity_status": "OFFLINE",
		"identity_kind": "guest",
		"stats": {
			"total_score": 100,
			"high_score": 75,
			"ship_deaths": 3,
			"games_played": 4,
			"wins": 2,
		},
	}
	var last_context: Dictionary = {}

	func load_profile(context: Dictionary) -> Dictionary:
		last_context = context.duplicate(true)
		return profile.duplicate(true)


class FakeTransmissionFlow:
	extends RefCounted

	var last_scene: PackedScene

	func mount(transmission_scene: PackedScene) -> Control:
		last_scene = transmission_scene
		return transmission_scene.instantiate() as Control


func test_show_profile_combines_context_and_stats() -> void:
	var context_provider := FakeContextProvider.new()
	var stats_provider := FakeStatsProvider.new()
	context_provider.context["play_mode"] = PregameMenuMode.SINGLE_PLAYER
	context_provider.context["identity_kind"] = "local_profile"
	context_provider.context["callsign"] = "ACE"
	context_provider.context["activity_status"] = "ACTIVE"
	stats_provider.profile = {
		"callsign": "Local Pilot",
		"activity_status": "LOCAL",
		"identity_kind": "local_profile",
		"stats": {
			"total_score": 100,
			"high_score": 75,
			"ship_deaths": 3,
			"games_played": 4,
			"wins": 2,
		},
	}
	var transmission_flow := FakeTransmissionFlow.new()
	var flow := ProfileFlow.new()
	flow.configure(context_provider, stats_provider, transmission_flow)

	var mounted = await flow.show_profile(PregameMenuMode.SINGLE_PLAYER)

	assert_eq(context_provider.last_mode, PregameMenuMode.SINGLE_PLAYER)
	assert_eq(stats_provider.last_context, context_provider.context)
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/CallsignLabel") as Label).text, "CALLSIGN: ACE")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/ActivityLabel") as Label).text, "STATUS: ACTIVE")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/TotalScoreContainer/VBoxContainer/TotalScoreValueLabel") as Label).text, "100")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/HighScoreContainer/VBoxContainer/HighScoreValueLabel") as Label).text, "75")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/StatContainer/MissionsContainer/VBoxContainer/MissionsValueLabel") as Label).text, "4")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/StatContainer/WinsContainer/VBoxContainer/WinsValueLabel") as Label).text, "2")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/StatContainer/ShipLossesContainer/VBoxContainer/ShipLossesValueLabel") as Label).text, "3")


func test_show_profile_uses_context_readout_fields_with_loaded_stats() -> void:
	var context_provider := FakeContextProvider.new()
	context_provider.context["play_mode"] = PregameMenuMode.SINGLE_PLAYER
	context_provider.context["identity_kind"] = "local_profile"
	context_provider.context["callsign"] = "ACE"
	context_provider.context["activity_status"] = "ACTIVE"

	var stats_provider := FakeStatsProvider.new()
	stats_provider.profile = {
		"callsign": "Local Pilot",
		"activity_status": "LOCAL",
		"identity_kind": "local_profile",
		"stats": {
			"total_score": 240,
			"high_score": 120,
			"ship_deaths": 5,
			"games_played": 9,
			"wins": 6,
		},
	}
	var transmission_flow := FakeTransmissionFlow.new()
	var flow := ProfileFlow.new()
	flow.configure(context_provider, stats_provider, transmission_flow)

	var mounted = await flow.show_profile(PregameMenuMode.SINGLE_PLAYER)

	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/CallsignLabel") as Label).text, "CALLSIGN: ACE")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/ActivityLabel") as Label).text, "STATUS: ACTIVE")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/TotalScoreContainer/VBoxContainer/TotalScoreValueLabel") as Label).text, "240")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/ScoreContainer/HighScoreContainer/VBoxContainer/HighScoreValueLabel") as Label).text, "120")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/StatContainer/MissionsContainer/VBoxContainer/MissionsValueLabel") as Label).text, "9")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/StatContainer/WinsContainer/VBoxContainer/WinsValueLabel") as Label).text, "6")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/StatContainer/ShipLossesContainer/VBoxContainer/ShipLossesValueLabel") as Label).text, "5")


func test_show_profile_mounts_profile_readout_scene() -> void:
	var context_provider := FakeContextProvider.new()
	var stats_provider := FakeStatsProvider.new()
	var transmission_flow := FakeTransmissionFlow.new()
	var flow := ProfileFlow.new()
	flow.configure(context_provider, stats_provider, transmission_flow)

	await flow.show_profile(PregameMenuMode.SINGLE_PLAYER)

	assert_eq(transmission_flow.last_scene, ProfileReadoutScene)


func test_show_profile_calls_apply_profile_on_mounted_readout() -> void:
	var context_provider := FakeContextProvider.new()
	context_provider.context["play_mode"] = PregameMenuMode.SINGLE_PLAYER
	var stats_provider := FakeStatsProvider.new()
	var transmission_flow := FakeTransmissionFlow.new()
	var flow := ProfileFlow.new()
	flow.configure(context_provider, stats_provider, transmission_flow)

	var mounted = await flow.show_profile(PregameMenuMode.SINGLE_PLAYER)

	assert_true(mounted.has_method("apply_profile"))
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/CallsignLabel") as Label).text, "CALLSIGN: Guest")
	assert_eq((mounted.get_node("ReadoutContainer/VBoxContainer/CallsignActivityContainer/ActivityLabel") as Label).text, "STATUS: OFFLINE")
