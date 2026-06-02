extends RefCounted
class_name DevtoolsTargetResolver

const TARGET_GAME := "__game_target__"
const TARGET_GAME_LABEL := "Game Target"
const TARGET_ALL_PLAYERS := "__all_players__"
const TARGET_ALL_PLAYERS_LABEL := "All Players"
const TARGET_SCOPE_ALL_PLAYERS := "all_players"
const TARGET_SCOPE_SINGLE_PLAYER := "single_player"
const TARGET_KIND_PLAYER := "player"
const TARGET_KIND_ENEMY := "enemy"
const TARGET_KIND_ASTEROID := "asteroid"
const TARGET_KIND_BULLET := "bullet"


static func resolve(selected_tool_target: String, game_target: String, local_player_id: String) -> String:
	if selected_tool_target != "" and selected_tool_target != TARGET_GAME:
		return selected_tool_target
	if game_target != "":
		return game_target
	if local_player_id != "":
		return local_player_id
	return ""


static func resolve_player_target(
	selected_tool_target: String,
	game_target_kind: String,
	game_target_id: String,
	local_player_id: String
) -> String:
	if selected_tool_target != "" and selected_tool_target != TARGET_GAME:
		return selected_tool_target
	if selected_tool_target == TARGET_GAME:
		if game_target_kind == TARGET_KIND_PLAYER and game_target_id != "":
			return game_target_id
		return ""

	if game_target_kind == TARGET_KIND_PLAYER:
		if game_target_id != "":
			return game_target_id
		if local_player_id != "":
			return local_player_id
		return ""

	if game_target_kind != "" and game_target_id != "":
		return ""

	if local_player_id != "":
		return local_player_id
	return ""


static func resolve_player_target_scope(
	selected_tool_target: String,
	game_target_kind: String,
	game_target_id: String,
	local_player_id: String
) -> Dictionary:
	if selected_tool_target == TARGET_ALL_PLAYERS:
		return {
			"target_scope": TARGET_SCOPE_ALL_PLAYERS,
			"target_player_id": "",
		}

	return {
		"target_scope": TARGET_SCOPE_SINGLE_PLAYER,
		"target_player_id": resolve_player_target(
			selected_tool_target,
			game_target_kind,
			game_target_id,
			local_player_id
		),
	}
