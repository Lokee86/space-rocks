extends RefCounted

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")


static func render(
	container: Container,
	row_scene: PackedScene,
	local_member_id: String,
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
		var row := row_scene.instantiate()
		container.add_child(row)
		if row.has_method("set_member"):
			row.set_member(
				LobbyMemberViewModel.display_name(member, local_member_id),
				LobbyMemberViewModel.member_ready(member),
				LobbyMemberViewModel.member_connected(member)
			)
