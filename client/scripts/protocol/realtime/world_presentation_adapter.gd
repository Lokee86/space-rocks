extends RefCounted
class_name WorldPresentationAdapter

func apply_world_lane_state(world_sync: WorldSync, world_lane_state, self_id := "") -> void:
	if world_sync == null or world_lane_state == null:
		return
	if self_id != "":
		world_sync.set_current_self_id(self_id)
	world_sync.apply_world_lane_state(world_lane_state)
