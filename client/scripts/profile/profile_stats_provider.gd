extends RefCounted
class_name ProfileStatsProvider

const PlayerDataProfileApiClientScript := preload("res://scripts/profile/player_data_profile_api_client.gd")
const GuestTransientStatsProviderScript := preload("res://scripts/profile/guest_transient_stats_provider.gd")
const PregameMenuMode := preload("res://scripts/ui/menu_flow/pregame_menu_mode.gd")

var auth_session_controller
var player_data_profile_api_client
var guest_stats_provider
var cached_account_stats := {}


func configure(auth_session_controller_ref, player_data_profile_api_client_ref = null, guest_stats_provider_ref = null) -> void:
	auth_session_controller = auth_session_controller_ref
	player_data_profile_api_client = player_data_profile_api_client_ref if player_data_profile_api_client_ref != null else PlayerDataProfileApiClientScript.new()
	guest_stats_provider = guest_stats_provider_ref if guest_stats_provider_ref != null else GuestTransientStatsProviderScript.new()


func load_stats(context: Dictionary):
	var profile: Dictionary = await load_profile(context)
	return profile.get("stats", _zero_stats())


func load_profile(context: Dictionary) -> Dictionary:
	var identity_kind := str(context.get("identity_kind", ""))
	var play_mode := _play_mode_for_context(context, identity_kind)
	var local_profile_id := str(context.get("local_profile_id", ""))
	match identity_kind:
		"guest", "local_profile":
			return await _load_profile_profile(play_mode, identity_kind, local_profile_id, "")
		"authenticated_account":
			return await _load_account_profile(play_mode, local_profile_id)
		_:
			return _empty_profile(identity_kind)


func _load_account_profile(play_mode: String, local_profile_id: String) -> Dictionary:
	var token := _session_token()
	if token.is_empty():
		if !cached_account_stats.is_empty():
			return _profile_from_stats("Pilot", "authenticated_account", cached_account_stats)
		return _empty_profile("authenticated_account")

	return await _load_profile_profile(play_mode, "authenticated_account", local_profile_id, token)


func _load_profile_profile(play_mode: String, identity_kind: String, local_profile_id: String, token: String) -> Dictionary:
	var result = await player_data_profile_api_client.load_profile(play_mode, identity_kind, local_profile_id, token)
	if result != null and result.ok and result.body is Dictionary:
		var body: Dictionary = result.body
		if body.has("profile") and body["profile"] is Dictionary:
			var profile: Dictionary = body["profile"]
			if profile.has("stats") and profile["stats"] is Dictionary:
				var normalized := _normalize_stats(profile["stats"])
				var normalized_profile := {
					"callsign": str(profile.get("callsign", "Guest")),
					"activity_status": str(profile.get("activity_status", "OFFLINE")),
					"identity_kind": str(profile.get("identity_kind", identity_kind)),
					"stats": normalized,
				}
				if identity_kind == "authenticated_account":
					cached_account_stats = normalized.duplicate(true)
				return normalized_profile

	if identity_kind == "authenticated_account" and !cached_account_stats.is_empty():
		return _profile_from_stats("Pilot", "authenticated_account", cached_account_stats)
	return _empty_profile(identity_kind)


func _play_mode_for_context(context: Dictionary, identity_kind: String) -> String:
	var play_mode := str(context.get("play_mode", ""))
	if !play_mode.is_empty():
		return play_mode
	if identity_kind == "authenticated_account":
		return PregameMenuMode.MULTIPLAYER
	return PregameMenuMode.SINGLE_PLAYER


func _normalize_stats(stats: Dictionary) -> Dictionary:
	return {
		"total_score": int(stats.get("total_score", 0)),
		"high_score": int(stats.get("high_score", 0)),
		"ship_deaths": int(stats.get("ship_deaths", 0)),
		"games_played": int(stats.get("games_played", 0)),
		"wins": int(stats.get("wins", 0)),
	}


func _zero_stats() -> Dictionary:
	return {
		"total_score": 0,
		"high_score": 0,
		"ship_deaths": 0,
		"games_played": 0,
		"wins": 0,
	}


func _profile_from_stats(callsign: String, identity_kind: String, stats: Dictionary) -> Dictionary:
	return {
		"callsign": callsign,
		"activity_status": "ACTIVE" if identity_kind == "authenticated_account" else "OFFLINE",
		"identity_kind": identity_kind,
		"stats": stats.duplicate(true),
	}


func _empty_profile(identity_kind: String) -> Dictionary:
	return _profile_from_stats("Guest", identity_kind if identity_kind != "" else "guest", _zero_stats())


func _session_token() -> String:
	if auth_session_controller == null or !auth_session_controller.has_method("get_session"):
		return ""

	var session = auth_session_controller.get_session()
	if session == null or !session.has_method("is_signed_in") or !session.is_signed_in():
		return ""

	return str(session.token)
