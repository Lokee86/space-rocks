extends GutTest

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")


func test_is_owner_uses_member_player_id() -> void:
	var member := {
		"player_id": "Player-1",
	}

	assert_true(LobbyMemberViewModel.is_owner(member, "Player-1"))


func test_display_name_prefers_player_id_and_marks_local_member() -> void:
	var member := {
		"player_id": "Player-1",
	}

	assert_eq(LobbyMemberViewModel.display_name(member, "Player-1"), "Player-1 (You)")


func test_display_name_marks_bot_member() -> void:
	var member := {
		"player_id": "Player-2",
		"is_bot": true,
	}

	assert_true(LobbyMemberViewModel.member_is_bot(member))
	assert_eq(LobbyMemberViewModel.display_name(member, "Player-1"), "Player-2 (Bot)")
