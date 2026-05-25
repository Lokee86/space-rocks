extends RefCounted

const LobbyConstants := preload("res://scripts/ui/lobby/constants.gd")
const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")
const ShellConstants := preload("res://scripts/shell/constants.gd")


static func status_text(
	room_state: String,
	local_member_id: String,
	owner_id: String,
	members: Array,
	can_start: bool
) -> String:
	if room_state == ShellConstants.ROOM_STATE_STARTING:
		return LobbyConstants.STATUS_STARTING_GAME
	if room_state == ShellConstants.ROOM_STATE_IN_GAME:
		return LobbyConstants.STATUS_GAME_IN_PROGRESS
	if room_state == ShellConstants.ROOM_STATE_GAME_OVER:
		return LobbyConstants.STATUS_GAME_OVER
	if room_state != ShellConstants.ROOM_STATE_LOBBY:
		return room_state

	if local_member_id == owner_id:
		if can_start:
			return LobbyConstants.STATUS_READY_TO_START
		return LobbyConstants.STATUS_WAITING_FOR_PLAYERS_READY

	if LobbyMemberViewModel.is_local_ready(local_member_id, members):
		return LobbyConstants.STATUS_WAITING_FOR_HOST_START
	return LobbyConstants.STATUS_PRESS_READY_WHEN_READY
