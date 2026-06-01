extends RefCounted
class_name DevtoolsTargetResolver

const TARGET_GAME := "__game_target__"
const TARGET_GAME_LABEL := "Game Target"
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
