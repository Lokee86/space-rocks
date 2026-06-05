extends RefCounted
class_name GameplayWorldStateApplyFlow

var world_sync


func configure(world_sync_ref) -> void:
	world_sync = world_sync_ref


func apply_world_state(state: Dictionary, _has_received_state: bool) -> void:
	if world_sync == null:
		return

	world_sync.apply_state(
		state["self_id"],
		state.get("server_players", {}),
		state.get("server_bullets", {}),
		state.get("server_asteroids", {}),
		state.get("server_pickups", {})
	)
