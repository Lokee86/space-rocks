extends GutTest

const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")

func _world_packet(match_id: String = "") -> Dictionary:
	return {"type": "world_full", "match_id": match_id, "baseline_id": "world-baseline-1", "sequence": 1, "snapshot_id": "world-snapshot-1", "is_final_chunk": true, "ships": [], "bullets": [], "asteroids": [], "pickups": []}

func test_packets_are_buffered_without_active_match_and_replayed_on_activation() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var applied := [0]
	pipeline.gameplay_packet_applied.connect(func(_packet): applied[0] += 1)
	pipeline.apply_packet(_world_packet("match-1"))
	assert_eq(applied[0], 0)
	assert_eq(pipeline.active_match_id(), "")
	pipeline.begin_match("match-1")
	assert_eq(applied[0], 1)

func test_packets_for_unrelated_match_are_not_replayed_when_new_match_activates() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var applied := []
	pipeline.gameplay_packet_applied.connect(func(packet): applied.append(packet.get("match_id", "")))
	pipeline.apply_packet(_world_packet("old-match"))
	pipeline.apply_packet(_world_packet("match-2"))
	pipeline.begin_match("match-2")
	assert_eq(applied, ["match-2"])

func test_pending_packets_are_cleared_when_realtime_session_resets() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.apply_packet(_world_packet("match-1"))
	pipeline.reset()
	pipeline.begin_match("match-1")
	assert_false(pipeline.is_gameplay_ready())

func test_matching_readable_and_compact_packets_are_accepted() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.begin_match("match-1")
	var applied := [0]
	pipeline.gameplay_packet_applied.connect(func(_packet): applied[0] += 1)
	pipeline.apply_packet(_world_packet("match-1"))
	var compact := _world_packet("")
	compact.erase("type")
	compact.erase("match_id")
	compact["t"] = "wf"
	compact["mid"] = "match-1"
	pipeline.apply_packet(compact)
	assert_eq(applied[0], 2)

func test_missing_and_mismatched_packets_are_rejected() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.begin_match("match-1")
	var applied := [0]
	pipeline.gameplay_packet_applied.connect(func(_packet): applied[0] += 1)
	pipeline.apply_packet(_world_packet())
	pipeline.apply_packet(_world_packet("old-match"))
	assert_eq(applied[0], 0)

func test_changed_match_rejects_old_unordered_packet_and_resets_state() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.begin_match("match-1")
	pipeline.apply_packet(_world_packet("match-1"))
	pipeline.begin_match("match-2")
	var applied := [0]
	pipeline.gameplay_packet_applied.connect(func(_packet): applied[0] += 1)
	pipeline.apply_packet(_world_packet("match-1"))
	assert_eq(applied[0], 0)
	assert_eq(pipeline.active_match_id(), "match-2")

func test_same_match_begin_is_idempotent() -> void:
	var pipeline := RealtimePacketPipeline.new()
	pipeline.begin_match("match-1")
	pipeline.apply_packet(_world_packet("match-1"))
	var router := pipeline.get_router()
	pipeline.begin_match("match-1")
	assert_true(pipeline.get_router() == router)
	assert_eq(pipeline.active_match_id(), "match-1")

func _clock(value: Array) -> Callable:
	return func(): return value[0]

func _resync_counts(pipeline: RealtimePacketPipeline) -> Dictionary:
	var counts := {"world": 0, "overlay": 0, "session": 0}
	pipeline.resync_request_required.connect(func(lane, _baseline_id, _sequence, _reason):
		if counts.has(lane):
			counts[lane] += 1
	)
	return counts

func test_packet_count_overflow_discards_bucket_and_requests_each_lane_once() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var counts := _resync_counts(pipeline)
	for index in range(129):
		var packet := _world_packet("overflow")
		packet["sequence"] = index + 1
		pipeline.apply_packet(packet)
	pipeline.begin_match("overflow")
	assert_eq(counts, {"world": 1, "overlay": 1, "session": 1})

func test_byte_overflow_discards_bucket_and_requests_each_lane_once() -> void:
	var pipeline := RealtimePacketPipeline.new()
	var counts := _resync_counts(pipeline)
	var packet := _world_packet("byte-overflow")
	packet["padding"] = "x".repeat(256 * 1024)
	pipeline.apply_packet(packet)
	pipeline.begin_match("byte-overflow")
	assert_eq(counts, {"world": 1, "overlay": 1, "session": 1})

func test_expired_bucket_discards_and_requests_each_lane_once() -> void:
	var now := [0]
	var pipeline := RealtimePacketPipeline.new(_clock(now))
	var counts := _resync_counts(pipeline)
	pipeline.apply_packet(_world_packet("expired"))
	now[0] = 5000
	pipeline.apply_packet(_world_packet("other"))
	pipeline.begin_match("expired")
	assert_eq(counts, {"world": 1, "overlay": 1, "session": 1})

func test_fifth_match_evicts_oldest_bucket_and_recovers_on_activation() -> void:
	var now := [0]
	var pipeline := RealtimePacketPipeline.new(_clock(now))
	for match_id in ["oldest", "match-2", "match-3", "match-4"]:
		pipeline.apply_packet(_world_packet(match_id))
		now[0] += 1
	var counts := _resync_counts(pipeline)
	pipeline.apply_packet(_world_packet("match-5"))
	pipeline.begin_match("oldest")
	assert_eq(counts, {"world": 1, "overlay": 1, "session": 1})

func test_activation_clears_unrelated_pending_and_recovery_state() -> void:
	var pipeline := RealtimePacketPipeline.new()
	for index in range(129):
		pipeline.apply_packet(_world_packet("recovery"))
	pipeline.apply_packet(_world_packet("valid"))
	var counts := _resync_counts(pipeline)
	var applied := [0]
	pipeline.gameplay_packet_applied.connect(func(_packet): applied[0] += 1)
	pipeline.begin_match("valid")
	assert_eq(applied[0], 1)
	pipeline.begin_match("recovery")
	assert_eq(counts, {"world": 0, "overlay": 0, "session": 0})

func _discard_match(pipeline: RealtimePacketPipeline, match_id: String) -> void:
	for index in range(129):
		var packet := _world_packet(match_id)
		packet["sequence"] = index + 1
		pipeline.apply_packet(packet)

func test_recovery_marker_overflow_keeps_forgotten_match_fail_closed() -> void:
	var pipeline := RealtimePacketPipeline.new()
	for match_id in ["forgotten", "discard-2", "discard-3", "discard-4", "discard-5"]:
		_discard_match(pipeline, match_id)
	var counts := _resync_counts(pipeline)
	pipeline.begin_match("forgotten")
	assert_eq(counts, {"world": 1, "overlay": 1, "session": 1})

func test_valid_selected_bucket_replays_despite_unrelated_recovery_uncertainty() -> void:
	var pipeline := RealtimePacketPipeline.new()
	for match_id in ["discard-1", "discard-2", "discard-3", "discard-4", "discard-5"]:
		_discard_match(pipeline, match_id)
	pipeline.apply_packet(_world_packet("valid"))
	var counts := _resync_counts(pipeline)
	var applied := [0]
	pipeline.gameplay_packet_applied.connect(func(_packet): applied[0] += 1)
	pipeline.begin_match("valid")
	assert_eq(applied[0], 1)
	assert_eq(counts, {"world": 0, "overlay": 0, "session": 0})

func test_direct_activation_after_bucket_expiry_requests_recovery() -> void:
	var now := [0]
	var pipeline := RealtimePacketPipeline.new(_clock(now))
	var counts := _resync_counts(pipeline)
	pipeline.apply_packet(_world_packet("expiring"))
	now[0] = 5000
	pipeline.begin_match("expiring")
	assert_eq(counts, {"world": 1, "overlay": 1, "session": 1})
