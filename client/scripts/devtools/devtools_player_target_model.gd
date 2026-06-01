extends RefCounted

const DevtoolsTargetResolver := preload("res://scripts/devtools/devtools_target_resolver.gd")


var self_id := ""
var server_players: Dictionary = {}
var player_lifecycle: Dictionary = {}
var debug_statuses: Dictionary = {}
var game_target_kind := ""
var game_target_id := ""
var game_target_player_id := ""


func reset() -> void:
	self_id = ""
	server_players = {}
	player_lifecycle = {}
	debug_statuses = {}
	game_target_kind = ""
	game_target_id = ""
	game_target_player_id = ""


func apply_gameplay_state(state: Dictionary) -> void:
	self_id = str(state.get("self_id", ""))

	var players_value = state.get("server_players", {})
	server_players = players_value if players_value is Dictionary else {}
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
	var rows: Array = [_game_target_row()]
	for player_id in server_players.keys():
		var player_id_text: String = str(player_id)
		rows.append({
			"player_id": player_id_text,
			"label": player_id_text,
			"is_self": player_id_text == self_id,
		})

	return rows


func invincible_target_rows() -> Array:
	var rows: Array = [_game_target_row()]
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
	var rows: Array = [_game_target_row()]
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
	var rows: Array = [_game_target_row()]
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


func _game_target_row() -> Dictionary:
	return {
		"player_id": DevtoolsTargetResolver.TARGET_GAME,
		"label": DevtoolsTargetResolver.TARGET_GAME_LABEL,
		"is_self": false,
	}
