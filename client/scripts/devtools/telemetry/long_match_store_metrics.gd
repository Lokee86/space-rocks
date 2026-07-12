extends RefCounted


static func snapshot(realtime_packet_pipeline, world_sync) -> Dictionary:
	var counts := {
		"applied_event_batch_ids": 0,
		"applied_event_ids": 0,
		"logged_applied_batch_ids": 0,
		"world_lane_deleted_bullet_ids": 0,
		"world_lane_pending_bullet_updates": 0,
		"projectile_sync_deleted_projectile_ids": 0,
		"asteroid_sync_deleted_asteroid_ids": 0,
		"total_entries": 0,
	}

	var presentation_state = _call_or_null(realtime_packet_pipeline, "get_presentation_state")
	var event_batch_applier = _owner_value(presentation_state, "event_batch_applier")
	counts["applied_event_batch_ids"] = _dictionary_size(_owner_value(event_batch_applier, "_applied_batch_ids"))
	counts["applied_event_ids"] = _dictionary_size(_owner_value(event_batch_applier, "_applied_event_ids"))
	counts["logged_applied_batch_ids"] = _dictionary_size(_owner_value(event_batch_applier, "_logged_applied_batch_ids"))

	var world_lane_state = _owner_value(presentation_state, "world_lane_state")
	counts["world_lane_deleted_bullet_ids"] = _dictionary_size(_owner_value(world_lane_state, "deleted_bullet_ids"))
	counts["world_lane_pending_bullet_updates"] = _dictionary_size(_owner_value(world_lane_state, "pending_bullet_updates"))

	var projectile_sync = _owner_value(world_sync, "projectile_sync")
	counts["projectile_sync_deleted_projectile_ids"] = _dictionary_size(_owner_value(projectile_sync, "deleted_projectile_ids"))

	var asteroid_sync = _owner_value(world_sync, "asteroid_sync")
	counts["asteroid_sync_deleted_asteroid_ids"] = _dictionary_size(_owner_value(asteroid_sync, "deleted_asteroid_ids"))

	counts["total_entries"] = (
		counts["applied_event_batch_ids"]
		+ counts["applied_event_ids"]
		+ counts["logged_applied_batch_ids"]
		+ counts["world_lane_deleted_bullet_ids"]
		+ counts["world_lane_pending_bullet_updates"]
		+ counts["projectile_sync_deleted_projectile_ids"]
		+ counts["asteroid_sync_deleted_asteroid_ids"]
	)
	return counts


static func _call_or_null(owner, method_name: String):
	if owner == null or not owner.has_method(method_name):
		return null
	return owner.call(method_name)


static func _owner_value(owner, property_name: String):
	if owner == null:
		return null
	if owner is Dictionary:
		return owner.get(property_name, null)
	return owner.get(property_name)


static func _dictionary_size(value) -> int:
	return value.size() if value is Dictionary else 0
