extends SceneTree

const THRESHOLDS := [1000, 5000, 10000, 25000, 50000, 100000, 250000]

var applied_batch_ids := {}
var applied_event_ids := {}
var logged_applied_batch_ids := {}
var world_lane_deleted_bullet_ids := {}
var projectile_sync_deleted_projectile_ids := {}
var asteroid_sync_deleted_asteroid_ids := {}


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
		print(JSON.stringify({
			"entry_count": entry_count,
			"static_memory_bytes": current_bytes,
			"delta_bytes": delta_bytes,
			"bytes_per_entry": float(delta_bytes) / float(entry_count) if entry_count > 0 else 0.0,
			"elapsed_insertion_msec": Time.get_ticks_msec() - started_msec,
		}))
		previous_entry_count = threshold

	print(JSON.stringify({
		"record": "completion",
		"entry_count": previous_entry_count,
		"static_memory_bytes": _static_memory_bytes(),
	}))
	quit()


func _add_entry(entry_index: int) -> void:
	var bucket := entry_index % 6
	match bucket:
		0:
			applied_batch_ids["batch-%08d" % entry_index] = true
		1:
			applied_event_ids["event-%08d" % entry_index] = true
		2:
			logged_applied_batch_ids["logged-batch-%08d" % entry_index] = true
		3:
			world_lane_deleted_bullet_ids["world-bullet-%08d" % entry_index] = true
		4:
			projectile_sync_deleted_projectile_ids["projectile-%08d" % entry_index] = true
		5:
			asteroid_sync_deleted_asteroid_ids["asteroid-%08d" % entry_index] = true


func _static_memory_bytes() -> int:
	return int(Performance.get_monitor(Performance.MEMORY_STATIC))
