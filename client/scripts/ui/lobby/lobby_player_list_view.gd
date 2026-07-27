extends RefCounted

const LobbyMemberViewModel := preload("res://scripts/ui/lobby/lobby_member_view_model.gd")
const PlayerRowType := preload("res://scripts/ui/lobby/player_row.gd")
const TeamPresentation := preload("res://scripts/teams/team_presentation.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


static func render(
	container: Container,
	row_scene: PackedScene,
	local_player_id: String,
	owner_id: String,
	members: Array,
	team_structure: String = "ffa",
	team_assignment_mode: String = "",
	team_count: int = 0,
	team_assignments_locked: bool = false,
	remove_callback: Callable = Callable(),
	team_callback: Callable = Callable()
) -> void:
	if container == null:
		return
	for child in container.get_children():
		container.remove_child(child)
		child.queue_free()
	if row_scene == null:
		return

	var local_is_owner := not local_player_id.is_empty() and local_player_id == owner_id
	var show_team := team_structure != "ffa"
	var available_team_ids := _available_team_ids(team_structure, team_count)
	for member in members:
		var row_instance: Node = row_scene.instantiate()
		var row := row_instance as PlayerRowType
		if row == null:
			_report_wrong_row(row_instance, row_scene)
			continue
		container.add_child(row)
		var player_id := LobbyMemberViewModel.member_player_id(member)
		var can_edit_team := _can_edit_team(player_id, local_player_id, local_is_owner, team_structure, team_assignment_mode, team_assignments_locked)
		row.set_member(
			player_id,
			LobbyMemberViewModel.display_name(member, local_player_id),
			LobbyMemberViewModel.member_ready(member),
			LobbyMemberViewModel.member_connected(member),
			LobbyMemberViewModel.is_owner(member, owner_id),
			local_is_owner and player_id != owner_id,
			str(member.get(Packets.FIELD_TEAM_ID, "")),
			available_team_ids,
			can_edit_team,
			show_team
		)
		if not remove_callback.is_null() and not row.remove_requested.is_connected(remove_callback):
			row.remove_requested.connect(remove_callback)
		if not team_callback.is_null() and not row.team_assignment_requested.is_connected(team_callback):
			row.team_assignment_requested.connect(team_callback)


static func _available_team_ids(structure: String, team_count: int) -> Array:
	if structure == "co_op":
		return TeamPresentation.ids_for_count(1)
	if structure == "auto_balanced":
		return TeamPresentation.ids_for_count(team_count)
	if structure == "custom":
		return TeamPresentation.ids_for_count(8)
	return []


static func _can_edit_team(player_id: String, local_player_id: String, local_is_owner: bool, structure: String, assignment_mode: String, locked: bool) -> bool:
	if locked or structure != "custom":
		return false
	if assignment_mode == "owner_assigned":
		return local_is_owner
	if assignment_mode == "player_selected":
		return player_id == local_player_id
	return false


static func _report_wrong_row(row_instance: Node, row_scene: PackedScene) -> void:
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
