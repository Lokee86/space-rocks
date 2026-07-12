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
