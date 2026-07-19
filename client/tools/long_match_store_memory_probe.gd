extends SceneTree





const THRESHOLDS := [1000, 5000, 10000, 25000, 50000, 100000, 250000]

var event_batch_applier := EventBatchApplier.new()
var world_lane_state := WorldLaneState.new()
var projectile_sync := ProjectileSync.new()
var asteroid_sync := AsteroidSync.new()


func _initialize() -> void:
	var baseline_bytes := _static_memory_bytes()
	var previous_entry_count := 0
	var started_msec := Time.get_ticks_msec()

	for threshold in THRESHOLDS:
		for entry_index in range(previous_entry_count, threshold):
			_add_entry(entry_index)

		var current_bytes := _static_memory_bytes()
		var delta_bytes := current_bytes - baseline_bytes
		var entry_count: int = threshold
		var store_sizes := _store_sizes()
		print(JSON.stringify({
			"entry_count": entry_count,
			"static_memory_bytes": current_bytes,
			"delta_bytes": delta_bytes,
			"bytes_per_entry": float(delta_bytes) / float(entry_count) if entry_count > 0 else 0.0,
			"elapsed_insertion_msec": Time.get_ticks_msec() - started_msec,
			"store_sizes": store_sizes,
			"store_total_entries": store_sizes["total_entries"],
		}))
		previous_entry_count = threshold

	print(JSON.stringify({
		"record": "completion",
		"entry_count": previous_entry_count,
		"static_memory_bytes": _static_memory_bytes(),
		"store_sizes": _store_sizes(),
		"store_total_entries": _store_sizes()["total_entries"],
	}))
	quit()


func _add_entry(entry_index: int) -> void:
	var bucket := entry_index % 6
	match bucket:
		0:
			event_batch_applier._record_applied_batch_id("batch-%08d" % entry_index)
		1:
			event_batch_applier._record_applied_event_id("event-%08d" % entry_index)
		2:
			world_lane_state.delete_bullet("world-bullet-%08d" % entry_index)
		3:
			world_lane_state.merge_or_buffer_bullet_update({"id": "world-pending-bullet-%08d" % entry_index, "x": entry_index})
		4:
			projectile_sync.remove_projectile("projectile-%08d" % entry_index)
		5:
			asteroid_sync.remove_asteroid("asteroid-%08d" % entry_index)


func _store_sizes() -> Dictionary:
	var sizes := {
		"applied_batch_ids": event_batch_applier._applied_batch_ids.size(),
		"applied_event_ids": event_batch_applier._applied_event_ids.size(),
		"world_lane_deleted_bullet_ids": world_lane_state.deleted_bullet_ids.size(),
		"world_lane_pending_bullet_updates": world_lane_state.pending_bullet_updates.size(),
		"projectile_sync_deleted_projectile_ids": projectile_sync.deleted_projectile_ids.size(),
		"asteroid_sync_deleted_asteroid_ids": asteroid_sync.deleted_asteroid_ids.size(),
	}
	sizes["total_entries"] = 0
	for size in sizes.values():
		sizes["total_entries"] += size
	return sizes


func _static_memory_bytes() -> int:
	return int(Performance.get_monitor(Performance.MEMORY_STATIC))
