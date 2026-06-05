extends RefCounted
class_name PlayerMeaningApi

# This API exposes player meaning facts from quarantined player/render legacy code.

const LegacyPlayerSyncScript = preload("res://legacy/player_render/player_sync.gd")

var legacy_player_sync


func _init() -> void:
	legacy_player_sync = LegacyPlayerSyncScript.new()


func configure(game_owner: Node2D, player: Player, pause_state_tracker = null) -> void:
	legacy_player_sync.configure(game_owner, player, pause_state_tracker)


func reset() -> void:
	legacy_player_sync.reset()


func remove_missing(server_players: Dictionary, self_id: String) -> void:
	legacy_player_sync.remove_missing(server_players, self_id)


func apply(self_id: String, server_players: Dictionary, anchor_visual_position: Vector2, anchor_server_position: Vector2) -> void:
	legacy_player_sync.apply(self_id, server_players, anchor_visual_position, anchor_server_position)


func interpolate(weight: float, current_self_id: String) -> void:
	legacy_player_sync.interpolate(weight, current_self_id)


func remote_player_nodes(current_self_id: String) -> Dictionary:
	return legacy_player_sync.remote_player_nodes(current_self_id)


func player_nodes() -> Dictionary:
	return legacy_player_sync.player_nodes()


func get_remote_player_visual_positions(current_self_id: String) -> Dictionary:
	return legacy_player_sync.get_remote_player_visual_positions(current_self_id)


func get_remote_player_hues(current_self_id: String) -> Dictionary:
	return legacy_player_sync.get_remote_player_hues(current_self_id)


func set_view_target_player(player_id: String) -> void:
	legacy_player_sync.set_view_target_player(player_id)


func clear_view_target_player() -> void:
	legacy_player_sync.clear_view_target_player()


func focus_camera_on_player(player_id: String) -> bool:
	return legacy_player_sync.focus_camera_on_player(player_id)


func get_view_target_player_id() -> String:
	return legacy_player_sync.get_view_target_player_id()


func apply_with_anchor(self_id: String, anchor_player_id: String, server_players: Dictionary, anchor_visual_position: Vector2, anchor_server_position: Vector2) -> void:
	legacy_player_sync.apply_with_anchor(self_id, anchor_player_id, server_players, anchor_visual_position, anchor_server_position)
