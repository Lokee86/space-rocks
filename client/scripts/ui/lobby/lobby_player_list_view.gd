extends RefCounted

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")
const PlayerRowType := preload("res://scripts/ui/lobby/player_row.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")


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
			ClientLogger.emit_canonical(
		ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
		"Lobby player row scene must instantiate its presentation root",
		{},
		{
			"subsystem": "lobby",
			"failure_mode": "wrong_row_type",
			"resource_kind": "scene",
			"expected_type": "PlayerRow",
			"actual_type": row_instance.get_class(),
			"resource_path": row_scene.resource_path,
		}
	)
			row_instance.queue_free()
			continue
		container.add_child(row)
		row.set_member(
			LobbyMemberViewModel.display_name(member, local_player_id),
			LobbyMemberViewModel.member_ready(member),
			LobbyMemberViewModel.member_connected(member),
			LobbyMemberViewModel.is_owner(member, owner_id)
		)
