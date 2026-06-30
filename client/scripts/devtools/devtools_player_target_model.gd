extends RefCounted

const TELEMETRY_SOURCE_PLAYERS = "players"
const TELEMETRY_SOURCE_PLAYER_WORLD_STATES = "player_world_states"


var self_id := ""
var world_ships: Dictionary = {}
var world_asteroids: Dictionary = {}
var world_bullets: Dictionary = {}
var world_pickups: Dictionary = {}
var session_players: Dictionary = {}
var session_player_lifecycle: Dictionary = {}
var overlay_self_id := ""
var server_enemies: Dictionary = {}
var has_server_enemies := false
var debug_statuses: Dictionary = {}
var game_target_kind := ""
var game_target_id := ""
var game_target_player_id := ""


func reset() -> void:
	self_id = ""
	world_ships = {}
	world_asteroids = {}
	world_bullets = {}
	world_pickups = {}
	session_players = {}
	session_player_lifecycle = {}
	overlay_self_id = ""
	server_enemies = {}
	has_server_enemies = false
	debug_statuses = {}
	game_target_kind = ""
	game_target_id = ""
	game_target_player_id = ""


func apply_gameplay_state(state: Dictionary) -> void:
	var world: Dictionary = _dict_or_empty(state.get("world", null))
	world_ships = _dict_or_empty(world.get("ships", null))
	world_asteroids = _dict_or_empty(world.get("asteroids", null))
	world_bullets = _dict_or_empty(world.get("bullets", null))
	world_pickups = _dict_or_empty(world.get("pickups", null))


	var session: Dictionary = _dict_or_empty(state.get("session", null))
	session_players = _dict_or_empty(session.get("players", null))
	session_player_lifecycle = _dict_or_empty(session.get("player_lifecycle", null))


	var overlay: Dictionary = _dict_or_empty(state.get("overlay", null))
	overlay_self_id = str(overlay.get("self_id", ""))
	self_id = overlay_self_id

	has_server_enemies = false
	server_enemies = {}
	if state.has("server_enemies"):
		var server_enemies_value: Variant = state.get("server_enemies", {})
		if server_enemies_value is Dictionary:
			server_enemies = server_enemies_value
		has_server_enemies = true
	elif state.has("enemies"):
		var enemies_value: Variant = state.get("enemies", {})
		if enemies_value is Dictionary:
			server_enemies = enemies_value
		has_server_enemies = true
	game_target_kind = ""
	game_target_id = ""
	game_target_player_id = ""
	if self_id != "":
		var local_player_value = world_ships.get(self_id, {})
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


func apply_debug_statuses(statuses: Dictionary) -> void:
	debug_statuses = statuses if statuses is Dictionary else {}


func target_rows() -> Array:
	var union_ids: Dictionary = {}
	for player_id in session_player_lifecycle.keys():
		union_ids[str(player_id)] = true
	for player_id in world_ships.keys():
		union_ids[str(player_id)] = true

	var rows: Array = []
	for player_id in union_ids.keys():
		var player_id_text: String = str(player_id)
		var lifecycle_status: String = _lifecycle_status_for_player(player_id_text)
		var alive: bool = lifecycle_status == "active"
		if lifecycle_status == "":
			alive = world_ships.has(player_id_text)
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


func _lifecycle_status_for_player(player_id: String) -> String:
	var value = session_player_lifecycle.get(player_id, "")
	if value is Dictionary:
		return str(value.get("status", ""))
	return str(value)


func active_player_target_rows() -> Array:
	var rows: Array = [_all_players_row()]
	rows.append_array(_game_target_rows())
	for player_id in world_ships.keys():
		var player_id_text: String = str(player_id)
		rows.append({
			"player_id": player_id_text,
			"label": player_id_text,
			"is_self": player_id_text == self_id,
		})

	return rows


func local_player_state() -> Dictionary:
	return local_player_state_for_source(TELEMETRY_SOURCE_PLAYERS)


func local_player_state_for_source(source: String) -> Dictionary:
	if self_id == "":
		return {}

	match source:
		TELEMETRY_SOURCE_PLAYERS:
			var local_state = world_ships.get(self_id, null)
			if local_state is Dictionary:
				return local_state
		TELEMETRY_SOURCE_PLAYER_WORLD_STATES:
			var local_session_state = session_players.get(self_id, null)
			if local_session_state is Dictionary:
				return local_session_state

	return {}


func target_state() -> Dictionary:
	return target_state_for_source(TELEMETRY_SOURCE_PLAYERS)


func target_state_for_source(source: String) -> Dictionary:
	if game_target_kind == "" or game_target_id == "":
		return {}

	var value = null
	match source:
		TELEMETRY_SOURCE_PLAYERS:
			match game_target_kind:
				"player":
					value = world_ships.get(game_target_id, null)
				"asteroid":
					value = world_asteroids.get(game_target_id, null)
				"bullet":
					value = world_bullets.get(game_target_id, null)
				"pickup":
					value = world_pickups.get(game_target_id, null)
				"enemy":
					if has_server_enemies:
						value = server_enemies.get(game_target_id, null)
				_:
					return {}
		TELEMETRY_SOURCE_PLAYER_WORLD_STATES:
			if game_target_kind == "player":
				value = session_players.get(game_target_id, null)
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
	var player_status = debug_statuses.get(player_id, null)
	if not (player_status is Dictionary):
		return false

	return bool(player_status.get(feature_key, false))


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


func _dict_or_empty(value) -> Dictionary:
	return value if value is Dictionary else {}

