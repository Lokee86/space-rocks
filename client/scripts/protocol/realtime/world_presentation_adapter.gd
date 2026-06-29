extends RefCounted

func apply_world_lane_state(world_sync, world_lane_state, self_id = "") -> void:
	if world_sync == null or world_lane_state == null:
		return
	if self_id != "" and world_sync.has_method("set_current_self_id"):
		world_sync.set_current_self_id(self_id)
	if world_sync.has_method("apply_world_lane_state"):
		world_sync.apply_world_lane_state(world_lane_state)
