extends RefCounted
class_name TeamPresentation

const Constants := preload("res://scripts/generated/constants/constants.gd")

const TEAM_IDS := [
	"team_1", "team_2", "team_3", "team_4",
	"team_5", "team_6", "team_7", "team_8",
]
const TEAM_HUES := [
	Constants.REMOTE_PLAYER_HUE_ZERO,
	Constants.REMOTE_PLAYER_HUE_ONE,
	Constants.REMOTE_PLAYER_HUE_TWO,
	Constants.REMOTE_PLAYER_HUE_THREE,
	Constants.REMOTE_PLAYER_HUE_FOUR,
	Constants.REMOTE_PLAYER_HUE_FIVE,
	Constants.REMOTE_PLAYER_HUE_SIX,
	Constants.REMOTE_PLAYER_HUE_SEVEN,
]

static func display_name(team_id: String) -> String:
	var index := TEAM_IDS.find(team_id)
	if index < 0:
		return "NO TEAM"
	return "TEAM %d" % (index + 1)


static func hue(team_id: String, fallback := Constants.PLAYER_DEFAULT_HUE) -> float:
	var index := TEAM_IDS.find(team_id)
	if index < 0:
		return float(fallback)
	return float(TEAM_HUES[index])


static func color(team_id: String) -> Color:
	return Color.from_hsv(hue(team_id), 0.72, 1.0)


static func ids_for_count(team_count: int) -> Array:
	var count := clampi(team_count, 0, TEAM_IDS.size())
	return TEAM_IDS.slice(0, count)


static func structure_name(structure: String) -> String:
	match structure:
		"co_op":
			return "CO-OP"
		"custom":
			return "CUSTOM TEAMS"
		"auto_balanced":
			return "AUTO-BALANCED"
		_:
			return "FREE-FOR-ALL"


static func assignment_name(mode: String) -> String:
	match mode:
		"owner_assigned":
			return "OWNER ASSIGNED"
		"player_selected":
			return "PLAYER SELECTED"
		_:
			return "AUTOMATIC"
