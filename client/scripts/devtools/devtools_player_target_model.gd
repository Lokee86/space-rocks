extends RefCounted

const TELEMETRY_SOURCE_STATE_PACKET = "state_packet"
const TELEMETRY_SOURCE_SESSION_PACKET = "session_packet"


var self_id := ""
var server_players: Dictionary = {}
var player_sessions: Dictionary = {}
var server_asteroids: Dictionary = {}
var server_bullets: Dictionary = {}
var server_enemies: Dictionary = {}
var has_server_enemies := false
var player_lifecycle: Dictionary = {}
var debug_statuses: Dictionary = {}
var game_target_kind := ""
var game_target_id := ""
var game_target_player_id := ""


func reset() -> void:
	self_id = ""
	server_players = {}
	player_sessions = {}
	server_asteroids = {}
	server_bullets = {}
	server_enemies = {}
	has_server_enemies = false
	player_lifecycle = {}
	debug_statuses = {}
	game_target_kind = ""
	game_target_id = ""
	game_target_player_id = ""


func apply_gameplay_state(state: Dictionary) -> void:
	self_id = str(state.get("self_id", ""))

	var players_value = state.get("server_players", {})
	server_players = players_value if players_value is Dictionary else {}
	var player_sessions_value = state.get("player_sessions", {})
	player_sessions = player_sessions_value if player_sessions_value is Dictionary else {}
	var asteroids_value = state.get("server_asteroids", {})
	server_asteroids = asteroids_value if asteroids_value is Dictionary else {}
	var bullets_value = state.get("server_bullets", {})
	server_bullets = bullets_value if bullets_value is Dictionary else {}
	has_server_enemies = false
	server_enemies = {}
	if state.has("server_enemies"):
		var server_enemies_value = state.get("server_enemies", {})
		if server_enemies_value is Dictionary:
			server_enemies = server_enemies_value
		has_server_enemies = true
	elif state.has("enemies"):
		var enemies_value = state.get("enemies", {})
		if enemies_value is Dictionary:
			server_enemies = enemies_value
		has_server_enemies = true
	game_target_kind = ""
	game_target_id = ""
	game_target_player_id = ""
	if self_id != "":
		var local_player_value = server_players.get(self_id, {})
		if local_player_value is Dictionary:
			game_target_kind = str(local_player_value.get("target_kind", ""))
			game_target_id = str(local_player_value.get("target_id", ""))
			if game_target_kind == "player":
				game_target_player_id = game_target_id
			elif game_target_kind == "" and game_target_id == "":
				var fallback_target_player_id := str(local_player_value.get("target_player_id", ""))
				if fallback_target_player_id != "":
					game_target_kind = "player"
					game_target_id = fallback_target_player_id
					game_target_player_id = fallback_target_player_id

	var lifecycle_value = state.get("player_lifecycle", {})
	player_lifecycle = lifecycle_value if lifecycle_value is Dictionary else {}

	var debug_statuses_value = state.get("debug_statuses", {})
	debug_statuses = debug_statuses_value if debug_statuses_value is Dictionary else {}


func target_rows() -> Array:
	var union_ids: Dictionary = {}
	for player_id in player_lifecycle.keys():
		union_ids[str(player_id)] = true
	for player_id in server_players.keys():
		union_ids[str(player_id)] = true

	var rows: Array = []
	for player_id in union_ids.keys():
		var player_id_text: String = str(player_id)
		var lifecycle_status: String = str(player_lifecycle.get(player_id_text, ""))
		var alive: bool = lifecycle_status == "active"
		if lifecycle_status == "":
			alive = server_players.has(player_id_text)
		var status: String = "ALIVE" if alive else "DEAD"
		var is_self: bool = player_id_text == self_id
		var label: String = "%s: %s" % [player_id_text, status]
		rows.append({
			"player_id": player_id_text,
			"status": status,
			"alive": alive,
			"is_self": is_self,
			"label": label,
		})

	return rows


func active_player_target_rows() -> Array:
	var rows: Array = [_all_players_row()]
	rows.append_array(_game_target_rows())
	for player_id in server_players.keys():
		var player_id_text: String = str(player_id)
		rows.append({
			"player_id": player_id_text,
			"label": player_id_text,
			"is_self": player_id_text == self_id,
		})

	return rows


func local_player_state() -> Dictionary:
	return local_player_state_for_source(TELEMETRY_SOURCE_STATE_PACKET)


func local_player_state_for_source(source: String) -> Dictionary:
	if self_id == "":
		return {}

	match source:
		TELEMETRY_SOURCE_STATE_PACKET:
			var local_state = server_players.get(self_id, null)
			if local_state is Dictionary:
				return local_state
		TELEMETRY_SOURCE_SESSION_PACKET:
			var local_session_state = player_sessions.get(self_id, null)
			if local_session_state is Dictionary:
				return local_session_state

	return {}


func target_state() -> Dictionary:
	return target_state_for_source(TELEMETRY_SOURCE_STATE_PACKET)


func target_state_for_source(source: String) -> Dictionary:
	if game_target_kind == "" or game_target_id == "":
		return {}

	var value = null
	match source:
		TELEMETRY_SOURCE_STATE_PACKET:
			match game_target_kind:
				"player":
					value = server_players.get(game_target_id, null)
				"asteroid":
					value = server_asteroids.get(game_target_id, null)
				"bullet":
					value = server_bullets.get(game_target_id, null)
				"enemy":
					if has_server_enemies:
						value = server_enemies.get(game_target_id, null)
				_:
					return {}
		TELEMETRY_SOURCE_SESSION_PACKET:
			if game_target_kind == "player":
				value = player_sessions.get(game_target_id, null)
		_:
			return {}

	if value is Dictionary:
		return value
	return {}


func invincible_target_rows() -> Array:
	var rows: Array = [_all_players_row()]
	rows.append_array(_game_target_rows())
	for row in target_rows():
		var player_id_text: String = str(row.get("player_id", ""))
		var invincible_on: bool = _player_feature_enabled(player_id_text, "invincible")
		var label: String = "%s: %s" % [player_id_text, "Active" if invincible_on else "Inactive"]

		rows.append({
			"player_id": player_id_text,
			"status": str(row.get("status", "DEAD")),
			"alive": bool(row.get("alive", false)),
			"is_self": bool(row.get("is_self", false)),
			"label": label,
		})

	return rows


func infinite_lives_target_rows() -> Array:
	var rows: Array = [_all_players_row()]
	rows.append_array(_game_target_rows())
	for row in target_rows():
		var player_id_text: String = str(row.get("player_id", ""))
		var infinite_lives_on: bool = _player_feature_enabled(player_id_text, "infinite_lives")
		var label: String = "%s: %s" % [player_id_text, "Active" if infinite_lives_on else "Inactive"]

		rows.append({
			"player_id": player_id_text,
			"status": str(row.get("status", "DEAD")),
			"alive": bool(row.get("alive", false)),
			"is_self": bool(row.get("is_self", false)),
			"label": label,
		})

	return rows


func player_frozen_target_rows() -> Array:
	var rows: Array = [_all_players_row()]
	rows.append_array(_game_target_rows())
	for row in target_rows():
		var player_id_text: String = str(row.get("player_id", ""))
		var player_frozen_on: bool = _player_feature_enabled(player_id_text, "player_frozen")
		var label: String = "%s: %s" % [player_id_text, "Active" if player_frozen_on else "Inactive"]

		rows.append({
			"player_id": player_id_text,
			"status": str(row.get("status", "DEAD")),
			"alive": bool(row.get("alive", false)),
			"is_self": bool(row.get("is_self", false)),
			"label": label,
		})

	return rows


func respawn_player_target_rows() -> Array:
	var rows: Array = [_all_players_row()]
	rows.append_array(target_rows())
	return rows


func kill_player_target_rows() -> Array:
	var rows: Array = [_all_players_row()]
	rows.append_array(_game_target_rows())
	rows.append_array(target_rows())
	return rows


func _feature_target_rows(feature_key: String) -> Array:
	var rows: Array = []
	var feature_label: String = _feature_label(feature_key)
	for row in target_rows():
		var player_id_text: String = str(row.get("player_id", ""))
		var status_text: String = str(row.get("status", "DEAD"))
		var feature_on: bool = _player_feature_enabled(player_id_text, feature_key)
		var feature_status: String = "ON" if feature_on else "OFF"
		var label: String = "%s: %s (%s %s)" % [player_id_text, status_text, feature_label, feature_status]

		rows.append({
			"player_id": player_id_text,
			"status": status_text,
			"alive": bool(row.get("alive", false)),
			"is_self": bool(row.get("is_self", false)),
			"label": label,
		})

	return rows


func _all_players_row() -> Dictionary:
	return {
		"player_id": DevtoolsTargetResolver.TARGET_ALL_PLAYERS,
		"label": DevtoolsTargetResolver.TARGET_ALL_PLAYERS_LABEL,
		"is_self": false,
	}


func _player_feature_enabled(player_id: String, feature_key: String) -> bool:
	var player_status_value = debug_statuses.get(player_id, {})
	if !(player_status_value is Dictionary):
		return false
	return bool(player_status_value.get(feature_key, false))


func _feature_label(feature_key: String) -> String:
	match feature_key:
		"invincible":
			return "INVINCIBLE"
		"infinite_lives":
			return "INFINITE LIVES"
		"player_frozen":
			return "PLAYER FROZEN"
		_:
			return feature_key.to_upper()


func _game_target_rows() -> Array:
	if game_target_kind != "player" or game_target_player_id == "":
		return []

	return [{
		"player_id": DevtoolsTargetResolver.TARGET_GAME,
		"label": _compact_game_target_label(game_target_player_id),
		"is_self": false,
	}]


func _compact_game_target_label(player_id: String) -> String:
	var lower_player_id := player_id.to_lower()
	var numeric_suffix := lower_player_id.trim_prefix("player-")
	if numeric_suffix != "" and numeric_suffix.is_valid_int():
		return "Target : P%s" % numeric_suffix
	return "Target : %s" % player_id
