extends GutTest

const ClientMeasurementSnapshotScheduler := preload("res://scripts/devtools/measurement/client_measurement_snapshot_scheduler.gd")


class FakeCoordinator extends RefCounted:
	var recording := false
	var pending_request_ids: Dictionary = {}
	var snapshot_calls := 0

	func is_recording() -> bool:
		return recording

	func snapshot() -> String:
		snapshot_calls += 1
		pending_request_ids["snapshot"] = "request-%d" % snapshot_calls
		return pending_request_ids["snapshot"]

	func complete_snapshot() -> void:
		pending_request_ids.erase("snapshot")


func test_requests_snapshots_at_the_configured_cadence() -> void:
	var coordinator := FakeCoordinator.new()
	coordinator.recording = true
	var scheduler := ClientMeasurementSnapshotScheduler.new(coordinator)

	scheduler.process(0.1)
	scheduler.process(0.14)
	assert_eq(coordinator.snapshot_calls, 0)
	scheduler.process(0.01)
	assert_eq(coordinator.snapshot_calls, 1)

	coordinator.complete_snapshot()
	scheduler.process(0.25)
	assert_eq(coordinator.snapshot_calls, 2)


func test_does_not_request_another_snapshot_while_one_is_pending() -> void:
	var coordinator := FakeCoordinator.new()
	coordinator.recording = true
	var scheduler := ClientMeasurementSnapshotScheduler.new(coordinator)

	scheduler.process(0.25)
	scheduler.process(0.25)
	assert_eq(coordinator.snapshot_calls, 1)

	coordinator.complete_snapshot()
	scheduler.process(0.25)
	assert_eq(coordinator.snapshot_calls, 2)


func test_inactive_processing_resets_elapsed_state_and_sends_no_snapshot() -> void:
	var coordinator := FakeCoordinator.new()
	coordinator.recording = true
	var scheduler := ClientMeasurementSnapshotScheduler.new(coordinator)

	scheduler.process(0.2)
	assert_almost_eq(scheduler.elapsed_seconds, 0.2, 0.0001)
	coordinator.recording = false
	scheduler.process(0.5)
	assert_eq(scheduler.elapsed_seconds, 0.0)
	assert_eq(coordinator.snapshot_calls, 0)

	coordinator.recording = true
	scheduler.process(0.24)
	assert_eq(coordinator.snapshot_calls, 0)
	scheduler.process(0.01)
	assert_eq(coordinator.snapshot_calls, 1)
