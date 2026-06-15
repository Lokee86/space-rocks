extends RefCounted
class_name ProfileFlow

const ProfileReadoutScene := preload("res://scenes/ui/transmission_displays/profile_readout.tscn")

var profile_context_provider
var profile_stats_provider
var transmission_flow


func configure(profile_context_provider_ref, profile_stats_provider_ref, transmission_flow_ref) -> void:
	profile_context_provider = profile_context_provider_ref
	profile_stats_provider = profile_stats_provider_ref
	transmission_flow = transmission_flow_ref


func show_profile(mode: String) -> Control:
	if profile_context_provider == null or profile_stats_provider == null or transmission_flow == null:
		return null

	var context: Dictionary = profile_context_provider.context_for_mode(mode)
	var profile: Dictionary = await profile_stats_provider.load_profile(context)
	var stats: Dictionary = profile.get("stats", {})
	var profile_for_readout := {
		"callsign": context.get("callsign", profile.get("callsign", "Guest")),
		"activity_status": context.get("activity_status", profile.get("activity_status", "ACTIVE")),
		"total_score": stats.get("total_score", 0),
		"high_score": stats.get("high_score", 0),
		"games_played": stats.get("games_played", 0),
		"wins": stats.get("wins", 0),
		"ship_deaths": stats.get("ship_deaths", 0),
	}

	var mounted := transmission_flow.mount(ProfileReadoutScene) as Control
	if mounted != null and mounted.has_method("apply_profile"):
		mounted.apply_profile(profile_for_readout)
	return mounted
