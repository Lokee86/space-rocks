extends RefCounted
class_name GameplayWorldStateApplyFlow

var world_sync


func configure(world_sync_ref) -> void:
	world_sync = world_sync_ref


func apply_world_lane_state(world_lane_state) -> void:
	if world_sync == null or world_lane_state == null:
		return
	world_sync.apply_world_lane_state(world_lane_state)


func apply_world_state(state: Dictionary, _required_lane_baselines_synced: bool) -> void:
	if world_sync == null:
		return

	world_sync.apply_state(
		str(state.get("self_id", "")),
		state.get("server_players", {}),
		state.get("server_bullets", {}),
		state.get("server_asteroids", {}),
		state.get("server_pickups", {})
	)
