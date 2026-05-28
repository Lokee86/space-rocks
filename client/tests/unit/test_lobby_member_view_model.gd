extends GutTest

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")
const Packets := preload("res://scripts/networking/packets/packets.gd")


func test_is_owner_uses_member_player_id() -> void:
	var member := {
		Packets.FIELD_MEMBER_ID: "session-1",
		Packets.FIELD_PLAYER_ID: "Player-1",
	}

	assert_true(LobbyMemberViewModel.is_owner(member, "Player-1"))


func test_display_name_prefers_player_id_and_marks_local_member() -> void:
	var member := {
		Packets.FIELD_MEMBER_ID: "session-1",
		Packets.FIELD_PLAYER_ID: "Player-1",
	}

	assert_eq(LobbyMemberViewModel.display_name(member, "session-1"), "Player-1 (You)")
