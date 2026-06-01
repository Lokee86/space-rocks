extends RefCounted
class_name DevtoolsTargetResolver

const TARGET_GAME := "__game_target__"
const TARGET_GAME_LABEL := "Game Target"


static func resolve(selected_tool_target: String, game_target: String, local_player_id: String) -> String:
	if selected_tool_target != "" and selected_tool_target != TARGET_GAME:
		return selected_tool_target
	if game_target != "":
		return game_target
	if local_player_id != "":
		return local_player_id
	return ""
