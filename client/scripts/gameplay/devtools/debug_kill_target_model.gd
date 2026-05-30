extends RefCounted


var self_id := ""
var server_players: Dictionary = {}
var player_lifecycle: Dictionary = {}


func apply_gameplay_state(state: Dictionary) -> void:
	self_id = str(state.get("self_id", ""))

	var players_value = state.get("server_players", {})
	server_players = players_value if players_value is Dictionary else {}

	var lifecycle_value = state.get("player_lifecycle", {})
	player_lifecycle = lifecycle_value if lifecycle_value is Dictionary else {}


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
