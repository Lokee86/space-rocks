extends GutTest

const LongMatchStoreMetrics := preload("res://scripts/devtools/telemetry/long_match_store_metrics.gd")


class FakePipeline:
	var presentation_state

	func get_presentation_state():
		return presentation_state


func test_snapshot_counts_store_dictionaries_and_total_entries() -> void:
	var pipeline := FakePipeline.new()
	pipeline.presentation_state = {
		"event_batch_applier": {
			"_applied_batch_ids": {"batch-1": true, "batch-2": true},
			"_applied_event_ids": {"event-1": true, "event-2": true, "event-3": true},
		},
		"world_lane_state": {
			"deleted_bullet_ids": {"bullet-1": true, "bullet-2": true},
			"pending_bullet_updates": {"bullet-3": {}, "bullet-4": {}, "bullet-5": {}},
		},
	}
	var world_sync := {
		"projectile_sync": {"deleted_projectile_ids": {"projectile-1": true}},
		"asteroid_sync": {"deleted_asteroid_ids": {"asteroid-1": true, "asteroid-2": true}},
	}

	var result := LongMatchStoreMetrics.snapshot(pipeline, world_sync)

	assert_eq(result["applied_event_batch_ids"], 2)
	assert_eq(result["applied_event_ids"], 3)
	assert_eq(result["world_lane_deleted_bullet_ids"], 2)
	assert_eq(result["world_lane_pending_bullet_updates"], 3)
	assert_eq(result["projectile_sync_deleted_projectile_ids"], 1)
	assert_eq(result["asteroid_sync_deleted_asteroid_ids"], 2)
	assert_eq(result["total_entries"], 13)


func test_snapshot_returns_zero_for_missing_owners() -> void:
	var result := LongMatchStoreMetrics.snapshot(null, null)

	assert_eq(result["total_entries"], 0)
	for key in result:
		assert_eq(result[key], 0)
