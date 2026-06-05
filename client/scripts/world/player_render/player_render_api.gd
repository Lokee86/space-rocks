extends RefCounted
class_name PlayerRenderApi

# This API coordinates player meaning and ViewAnchor/render-anchor state.

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const PlayerMeaningApiScript = preload("res://scripts/world/player_render/player_meaning_api.gd")
const ViewAnchorSyncScript = preload("res://scripts/world/player_render/view_anchor_sync.gd")

var player_meaning
var view_anchor_sync
var view_anchor_node: Node2D
var anchor_rotation := 0.0


func _init() -> void:
	player_meaning = PlayerMeaningApiScript.new()
	view_anchor_sync = ViewAnchorSyncScript.new()


func configure(game_owner: Node2D, player: Player, view_anchor: Node2D, pause_state_tracker = null) -> void:
	view_anchor_node = view_anchor
	player_meaning.configure(game_owner, player, pause_state_tracker)


func reset() -> void:
	player_meaning.reset()
	view_anchor_sync.reset()
	anchor_rotation = 0.0


func remove_missing(server_players: Dictionary, self_id: String) -> void:
	player_meaning.remove_missing(server_players, self_id)


func interpolate(weight: float, current_self_id: String) -> void:
	player_meaning.interpolate(weight, current_self_id)


func apply_state(self_id: String, server_players: Dictionary) -> void:
	var anchor_state: Dictionary = {}
	if server_players.has(self_id):
		anchor_state = server_players[self_id]

	var anchor_id := self_id
	if player_meaning.get_view_target_player_id() != "":
		anchor_id = player_meaning.get_view_target_player_id()
		if server_players.has(anchor_id):
			anchor_state = server_players[anchor_id]

	if anchor_state.is_empty():
		return

	view_anchor_sync.update_from_anchor_server_position(Vector2(anchor_state[Packets.FIELD_X], anchor_state[Packets.FIELD_Y]))
	anchor_rotation = float(anchor_state[Packets.FIELD_ROTATION])
	if view_anchor_node != null:
		view_anchor_node.global_position = view_anchor_sync.visual_position()
		view_anchor_node.rotation = anchor_rotation

	player_meaning.apply_with_anchor(
		self_id,
		anchor_id,
		server_players,
		view_anchor_sync.visual_position(),
		view_anchor_sync.server_position()
	)


func get_remote_player_visual_positions(current_self_id: String) -> Dictionary:
	return player_meaning.get_remote_player_visual_positions(current_self_id)


func get_remote_player_hues(current_self_id: String) -> Dictionary:
	return player_meaning.get_remote_player_hues(current_self_id)


func remote_player_nodes(current_self_id: String) -> Dictionary:
	return player_meaning.remote_player_nodes(current_self_id)


func player_nodes() -> Dictionary:
	return player_meaning.player_nodes()


func focus_camera_on_player(player_id: String) -> bool:
	return player_meaning.focus_camera_on_player(player_id)


func set_view_target_player(player_id: String) -> void:
	player_meaning.set_view_target_player(player_id)


func clear_view_target_player() -> void:
	player_meaning.clear_view_target_player()


func visual_position() -> Vector2:
	return view_anchor_sync.visual_position()


func server_position() -> Vector2:
	return view_anchor_sync.server_position()


func visual_position_for_server_position(server_authoritative_position: Vector2) -> Vector2:
	return view_anchor_sync.visual_position_for_server_position(server_authoritative_position)


func server_position_for_visual_position(client_visual_position: Vector2) -> Vector2:
	return view_anchor_sync.server_position_for_visual_position(client_visual_position)

