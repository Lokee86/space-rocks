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


class FakeProfileReadout:
	extends Control

	var applied_profile: Dictionary = {}

	func apply_profile(profile: Dictionary) -> void:
		applied_profile = profile.duplicate(true)


class FakeTransmissionFlow:
	extends RefCounted

	var last_scene: PackedScene
	var mounted_readout := FakeProfileReadout.new()

	func mount(transmission_scene: PackedScene) -> Control:
		last_scene = transmission_scene
		return mounted_readout


func test_show_profile_combines_context_and_stats() -> void:
	var context_provider := FakeContextProvider.new()
	var stats_provider := FakeStatsProvider.new()
	context_provider.context["play_mode"] = PregameMenuMode.MULTIPLAYER
	context_provider.context["identity_kind"] = "authenticated_account"
	context_provider.context["callsign"] = "Ada"
	context_provider.context["activity_status"] = "ACTIVE"
	stats_provider.profile = {
		"callsign": "Ada",
		"activity_status": "ACTIVE",
		"identity_kind": "authenticated_account",
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

	var mounted = await flow.show_profile(PregameMenuMode.MULTIPLAYER)

	assert_eq(context_provider.last_mode, PregameMenuMode.MULTIPLAYER)
	assert_eq(stats_provider.last_context, context_provider.context)
	assert_eq(mounted.applied_profile, {
		"callsign": "Ada",
		"activity_status": "ACTIVE",
		"total_score": 100,
		"high_score": 75,
		"games_played": 4,
		"wins": 2,
		"ship_deaths": 3,
	})


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
	assert_eq(mounted.applied_profile["callsign"], "Guest")
	assert_eq(mounted.applied_profile["activity_status"], "OFFLINE")
