extends GutTest

const HotDeltaChunkAssembler := preload("res://scripts/protocol/realtime/hot_delta_chunk_assembler.gd")
const WorldLaneApplier := preload("res://scripts/protocol/realtime/world_lane_applier.gd")
const WorldLaneState := preload("res://scripts/protocol/realtime/world_lane_state.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")


func test_assembler_completes_out_of_order_chunks_atomically() -> void:
	var assembler := HotDeltaChunkAssembler.new()
	var second := assembler.accept(_asteroid_chunk(7, 1, 2, "asteroid-2", 30), "asteroid_updates")
	var first := assembler.accept(_asteroid_chunk(7, 0, 2, "asteroid-1", 10), "asteroid_updates")

	assert_eq(second.status, HotDeltaChunkAssembler.PENDING)
	assert_eq(first.status, HotDeltaChunkAssembler.COMPLETE)
	assert_eq(first.packet.asteroid_updates.size(), 2)
	assert_eq(first.packet.asteroid_updates[0].id, "asteroid-1")
	assert_eq(first.packet.asteroid_updates[1].id, "asteroid-2")
	assert_eq(first.packet.chunk_count, 1)


func test_newer_sequence_supersedes_incomplete_sequence() -> void:
	var assembler := HotDeltaChunkAssembler.new()
	assert_eq(
		assembler.accept(_asteroid_chunk(7, 0, 2, "asteroid-1", 10), "asteroid_updates").status,
		HotDeltaChunkAssembler.PENDING
	)
	var replacement_first := assembler.accept(
		_asteroid_chunk(8, 0, 2, "asteroid-1", 20),
		"asteroid_updates"
	)
	var replacement_final := assembler.accept(
		_asteroid_chunk(8, 1, 2, "asteroid-2", 40),
		"asteroid_updates"
	)

	assert_eq(replacement_first.status, HotDeltaChunkAssembler.PENDING)
	assert_eq(replacement_first.superseded.sequence, 7)
	assert_eq(replacement_first.superseded.received_chunks, 1)
	assert_eq(replacement_final.status, HotDeltaChunkAssembler.COMPLETE)
	assert_eq(
		assembler.accept(_asteroid_chunk(7, 1, 2, "asteroid-2", 30), "asteroid_updates").status,
		HotDeltaChunkAssembler.REJECTED
	)


func test_world_lane_applier_does_not_mutate_until_hot_sequence_is_complete() -> void:
	var applier := WorldLaneApplier.new()
	var state := WorldLaneState.new()
	state.upsert_asteroid({"id": "asteroid-1", "x": 0.0, "y": 0.0})
	state.upsert_asteroid({"id": "asteroid-2", "x": 0.0, "y": 0.0})
	state.clear_asteroid_change_sets()

	applier.apply_asteroid_delta(
		state,
		LaneMetadata.LANE_ASTEROIDS,
		_asteroid_chunk(7, 0, 2, "asteroid-1", 100)
	)

	assert_eq(state.asteroids["asteroid-1"].x, 0.0)
	assert_true(state.dirty_asteroid_ids.is_empty())

	applier.apply_asteroid_delta(
		state,
		LaneMetadata.LANE_ASTEROIDS,
		_asteroid_chunk(7, 1, 2, "asteroid-2", 300)
	)

	assert_eq(state.asteroids["asteroid-1"].x, 10.0)
	assert_eq(state.asteroids["asteroid-2"].x, 30.0)
	assert_true(state.dirty_asteroid_ids.has("asteroid-1"))
	assert_true(state.dirty_asteroid_ids.has("asteroid-2"))


func test_world_lane_applier_applies_newer_complete_sequence_not_partial_old_sequence() -> void:
	var applier := WorldLaneApplier.new()
	var state := WorldLaneState.new()
	state.upsert_asteroid({"id": "asteroid-1", "x": 0.0, "y": 0.0})
	state.upsert_asteroid({"id": "asteroid-2", "x": 0.0, "y": 0.0})
	state.clear_asteroid_change_sets()

	applier.apply_asteroid_delta(state, LaneMetadata.LANE_ASTEROIDS, _asteroid_chunk(7, 0, 2, "asteroid-1", 100))
	applier.apply_asteroid_delta(state, LaneMetadata.LANE_ASTEROIDS, _asteroid_chunk(8, 0, 2, "asteroid-1", 200))
	applier.apply_asteroid_delta(state, LaneMetadata.LANE_ASTEROIDS, _asteroid_chunk(8, 1, 2, "asteroid-2", 400))
	applier.apply_asteroid_delta(state, LaneMetadata.LANE_ASTEROIDS, _asteroid_chunk(7, 1, 2, "asteroid-2", 300))

	assert_eq(state.asteroids["asteroid-1"].x, 20.0)
	assert_eq(state.asteroids["asteroid-2"].x, 40.0)
	assert_eq(state.latest_asteroid_delta_sequence, 8)


func _asteroid_chunk(
	sequence: int,
	chunk_index: int,
	chunk_count: int,
	asteroid_id: String,
	x: int
) -> Dictionary:
	return {
		"type": "asteroid_delta",
		"lane": LaneMetadata.LANE_ASTEROIDS,
		"sequence": sequence,
		"baseline_id": "world-baseline-1",
		"snapshot_id": "asteroids-delta-%d" % sequence,
		"chunk_index": chunk_index,
		"chunk_count": chunk_count,
		"is_final_chunk": chunk_index == chunk_count - 1,
		"asteroid_updates": [{"id": asteroid_id, "x": x, "y": 0}],
	}
