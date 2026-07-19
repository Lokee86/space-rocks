extends GutTest

const TimingSummary := preload("res://scripts/devtools/measurement/timing_summary.gd")
const PeriodicSampleStore := preload("res://scripts/devtools/measurement/periodic_sample_store.gd")
const ClientMeasurementSession := preload("res://scripts/devtools/measurement/client_measurement_session.gd")


class FakeClock:
	var now_msec := 1000

	func read() -> int:
		return now_msec


var finalized_results: Array = []


func before_each() -> void:
	finalized_results.clear()


func _capture_finalized(result: Dictionary) -> void:
	finalized_results.append(result)


func test_timing_summary_is_bounded_and_reports_statistics() -> void:
	var summary := TimingSummary.new()
	for duration in [1.0, 2.0, 16.0, 50.0, 100.0]:
		summary.record(duration)

	var result := summary.snapshot()

	assert_eq(result["count"], 5)
	assert_eq(result["total"], 169.0)
	assert_eq(result["minimum"], 1.0)
	assert_eq(result["maximum"], 100.0)
	assert_eq(result["p95"], 100.0)
	assert_eq(result["p99"], 100.0)
	assert_false(result.has("samples"))


func test_periodic_sample_store_keeps_a_bounded_window() -> void:
	var store := PeriodicSampleStore.new(2)
	store.append({"second": 1})
	store.append({"second": 2})
	store.append({"second": 3})

	var result := store.snapshot()

	assert_eq(result["count"], 2)
	assert_eq(result["dropped"], 1)
	assert_eq(result["samples"][0]["second"], 2)
	assert_eq(result["samples"][1]["second"], 3)


func test_session_lifecycle_and_partial_reset() -> void:
	var clock := FakeClock.new()
	var session := ClientMeasurementSession.new(Callable(clock, "read"))
	session.finalized.connect(Callable(self, "_capture_finalized"))
	session.start({"scenario": "test"})
	session.record_lane_application("world", 4.0)
	session.record_presentation(2.0)
	session.record_lifecycle("players", "created")
	session.process_frame(0.25)
	clock.now_msec += 250

	var result := session.reset()

	assert_false(session.is_recording())
	assert_eq(result["status"], "partial")
	assert_eq(result["partial_reason"], "reset")
	assert_eq(result["metadata"]["scenario"], "test")
	assert_eq(result["frame_timing"]["count"], 1)
	assert_eq(result["lane_application_timing"]["world"]["count"], 1)
	assert_eq(result["presentation_timing"]["count"], 1)
	assert_eq(result["lifecycle_churn"]["cumulative"]["players"]["created"], 1)
	assert_eq(finalized_results.size(), 1)


func test_disabled_session_does_not_accumulate_measurements() -> void:
	var session := ClientMeasurementSession.new()
	session.process_frame(1.0)
	session.record_lane_application("world", 4.0)
	session.record_periodic_sample({"ignored": true})

	assert_false(session.is_recording())
	assert_true(session.snapshot().is_empty())


func test_sampling_occurs_once_per_second_and_does_not_store_raw_frames() -> void:
	var session := ClientMeasurementSession.new()
	var sample_state := {"count": 0}
	session.set_periodic_sample_provider(func() -> Dictionary:
		sample_state["count"] += 1
		return {"second": sample_state["count"], "network_metrics": {"rtt_ms": 12}}
	)
	session.start()
	session.process_frame(0.4)
	session.process_frame(0.4)
	session.process_frame(0.2)
	session.process_frame(0.1)

	var result := session.stop()

	assert_eq(sample_state["count"], 1)
	assert_eq(result["resource_samples"]["count"], 1)
	assert_eq(result["network_metrics"]["rtt_ms"], 12)
	assert_eq(result["frame_timing"]["count"], 4)
	assert_false(result.has("frames"))
	assert_false(result.has("packets"))
