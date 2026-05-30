extends RefCounted

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")
const Constants := preload("res://scripts/constants/constants.gd")


static func status_text(
	room_state: String,
	local_player_id: String,
	owner_id: String,
	members: Array,
	can_start: bool
) -> String:
	if room_state == Constants.ROOM_STATE_STARTING:
		return Constants.STATUS_STARTING_GAME
	if room_state == Constants.ROOM_STATE_IN_GAME:
		return Constants.STATUS_GAME_IN_PROGRESS
	if room_state == Constants.ROOM_STATE_GAME_OVER:
		return Constants.STATUS_GAME_OVER
	if room_state != Constants.ROOM_STATE_LOBBY:
		return room_state

	if local_player_id == owner_id:
		if can_start:
			return Constants.STATUS_READY_TO_START
		return Constants.STATUS_WAITING_FOR_PLAYERS_READY

	if LobbyMemberViewModel.is_local_ready(local_player_id, members):
		return Constants.STATUS_WAITING_FOR_HOST_START
	return Constants.STATUS_PRESS_READY_WHEN_READY
