extends RefCounted

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")
const PlayerRowType := preload("res://scripts/ui/lobby/player_row.gd")


static func render(
	container: Container,
	row_scene: PackedScene,
	local_player_id: String,
	owner_id: String,
	members: Array
) -> void:
	if container == null:
		return

	for child in container.get_children():
		container.remove_child(child)
		child.queue_free()

	if row_scene == null:
		return

	for member in members:
		var row_instance: Node = row_scene.instantiate()
		var row := row_instance as PlayerRowType
		if row == null:
			push_error("Lobby player row scene must instantiate PlayerRow; got %s" % row_instance.get_class())
			row_instance.queue_free()
			continue
		container.add_child(row)
		row.set_member(
			LobbyMemberViewModel.display_name(member, local_player_id),
			LobbyMemberViewModel.member_ready(member),
			LobbyMemberViewModel.member_connected(member),
			LobbyMemberViewModel.is_owner(member, owner_id)
		)
