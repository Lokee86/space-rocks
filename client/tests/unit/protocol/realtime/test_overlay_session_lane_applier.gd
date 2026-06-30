extends GutTest

const OverlayLaneApplier := preload("res://scripts/protocol/realtime/overlay_lane_applier.gd")
const OverlayLaneState := preload("res://scripts/protocol/realtime/overlay_lane_state.gd")
const SessionLaneApplier := preload("res://scripts/protocol/realtime/session_lane_applier.gd")
const SessionLaneState := preload("res://scripts/protocol/realtime/session_lane_state.gd")
const BaselineTracker := preload("res://scripts/protocol/realtime/baseline_tracker.gd")
const LaneMetadata := preload("res://scripts/protocol/realtime/lane_metadata.gd")
const PresentationAdapter := preload("res://scripts/protocol/realtime/presentation_adapter.gd")
const GameplayReadiness := preload("res://scripts/protocol/realtime/gameplay_readiness.gd")


func test_overlay_full_updates_readout_cache() -> void:
	var applier := OverlayLaneApplier.new()
	var overlay_lane_state := OverlayLaneState.new()
	var baseline_tracker := BaselineTracker.new()

	applier.apply_overlay_full(
		overlay_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_OVERLAY,
		{
			"baseline_id": "overlay-baseline-1",
			"sequence": 1,
			"snapshot_id": "overlay-snapshot-1",
			"self_id": "player-1",
			"lives": 3,
			"score": 120,
			"respawn_cooldown": 2.0,
			"primary_weapon_id": "laser",
			"secondary_weapon_id": "burst",
			"primary_ammo_policy": "finite",
			"secondary_ammo_policy": "infinite",
			"primary_cooldown_remaining": 1.5,
			"secondary_cooldown_remaining": 0.5,
			"primary_ammo_remaining": 9,
			"secondary_ammo_remaining": 99,
			"is_final_chunk": true,
		}
	)

	assert_eq(overlay_lane_state.self_id, "player-1")
	assert_eq(overlay_lane_state.lives, 3)
	assert_eq(overlay_lane_state.score, 120)
	assert_eq(overlay_lane_state.respawn_cooldown, 2.0)
	assert_eq(overlay_lane_state.primary_weapon_id, "laser")
	assert_eq(overlay_lane_state.secondary_weapon_id, "burst")
	assert_eq(overlay_lane_state.primary_ammo_policy, "finite")
	assert_eq(overlay_lane_state.secondary_ammo_policy, "infinite")
	assert_eq(overlay_lane_state.primary_cooldown_remaining, 1.5)
	assert_eq(overlay_lane_state.secondary_cooldown_remaining, 0.5)
	assert_eq(overlay_lane_state.primary_ammo_remaining, 9)
	assert_eq(overlay_lane_state.secondary_ammo_remaining, 99)


func test_overlay_delta_updates_only_provided_fields() -> void:
	var applier := OverlayLaneApplier.new()
	var overlay_lane_state := OverlayLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_overlay_full(
		overlay_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_OVERLAY,
		{
			"baseline_id": "overlay-baseline-1",
			"sequence": 1,
			"self_id": "player-1",
			"lives": 3,
			"score": 120,
			"primary_weapon_id": "laser",
			"secondary_weapon_id": "burst",
			"primary_ammo_policy": "finite",
			"secondary_ammo_policy": "infinite",
			"primary_cooldown_remaining": 1.5,
			"secondary_cooldown_remaining": 0.5,
			"primary_ammo_remaining": 9,
			"secondary_ammo_remaining": 99,
			"is_final_chunk": true,
		}
	)

	var applied := applier.apply_overlay_delta(
		overlay_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_OVERLAY,
		{
			"baseline_id": "overlay-baseline-1",
			"sequence": 2,
			"lives": 2,
			"primary_cooldown_remaining": 1.0,
		}
	)

	assert_true(applied)
	assert_eq(overlay_lane_state.self_id, "player-1")
	assert_eq(overlay_lane_state.lives, 2)
	assert_eq(overlay_lane_state.score, 120)
	assert_eq(overlay_lane_state.primary_weapon_id, "laser")
	assert_eq(overlay_lane_state.secondary_weapon_id, "burst")
	assert_eq(overlay_lane_state.primary_cooldown_remaining, 1.0)
	assert_eq(overlay_lane_state.secondary_cooldown_remaining, 0.5)
	assert_eq(overlay_lane_state.primary_ammo_remaining, 9)
	assert_eq(overlay_lane_state.secondary_ammo_remaining, 99)


func test_overlay_delta_applies_nested_receiver_updates() -> void:
	var applier := OverlayLaneApplier.new()
	var overlay_lane_state := OverlayLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_overlay_full(
		overlay_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_OVERLAY,
		{
			"baseline_id": "overlay-baseline-1",
			"sequence": 1,
			"self_id": "player-1",
			"lives": 3,
			"score": 120,
			"primary_weapon_id": "laser",
			"secondary_weapon_id": "burst",
			"primary_ammo_policy": "finite",
			"secondary_ammo_policy": "infinite",
			"primary_cooldown_remaining": 1.5,
			"secondary_cooldown_remaining": 0.5,
			"primary_ammo_remaining": 9,
			"secondary_ammo_remaining": 99,
			"is_final_chunk": true,
		}
	)

	var applied := applier.apply_overlay_delta(
		overlay_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_OVERLAY,
		{
			"baseline_id": "overlay-baseline-1",
			"sequence": 2,
			"receiver_updates": [
				{
					"primary_weapon_id": "laser",
					"secondary_weapon_id": "mine",
					"primary_ammo_policy": "infinite",
				},
			],
		}
	)

	assert_true(applied)
	assert_eq(overlay_lane_state.primary_weapon_id, "laser")
	assert_eq(overlay_lane_state.secondary_weapon_id, "mine")
	assert_eq(overlay_lane_state.primary_ammo_policy, "infinite")
	assert_eq(overlay_lane_state.secondary_ammo_policy, "infinite")


func test_session_full_updates_score_lives_and_lifecycle_cache() -> void:
	var applier := SessionLaneApplier.new()
	var session_lane_state := SessionLaneState.new()
	var baseline_tracker := BaselineTracker.new()

	applier.apply_session_full(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 1,
			"snapshot_id": "session-snapshot-1",
			"total_asteroids": 4,
			"players": [
				{"id": "player-1", "score": 120, "lives": 3},
				{"id": "player-2", "score": 90, "lives": 2},
			],
			"player_lifecycle": [
				{"player_id": "player-1", "state": "active"},
				{"player_id": "player-2", "state": "pending_respawn"},
			],
			"is_final_chunk": true,
		}
	)

	assert_eq(session_lane_state.total_asteroids, 4)
	assert_eq(session_lane_state.player_sessions["player-1"]["score"], 120)
	assert_eq(session_lane_state.player_sessions["player-1"]["lives"], 3)
	assert_eq(session_lane_state.player_lifecycle["player-2"]["state"], "pending_respawn")


func test_session_delta_updates_and_deletes_records_without_clearing_missing_records() -> void:
	var applier := SessionLaneApplier.new()
	var session_lane_state := SessionLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_session_full(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 1,
			"players": [
				{"id": "player-1", "score": 120, "lives": 3},
				{"id": "player-2", "score": 90, "lives": 2},
			],
			"player_lifecycle": [
				{"player_id": "player-1", "state": "active"},
				{"player_id": "player-2", "state": "pending_respawn"},
			],
			"is_final_chunk": true,
		}
	)

	var applied := applier.apply_session_delta(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 2,
			"player_session_updates": [
				{"id": "player-1", "score": 150},
			],
			"player_lifecycle_updates": [
				{"player_id": "player-2", "state": "active"},
			],
		}
	)

	assert_true(applied)
	assert_eq(session_lane_state.player_sessions["player-1"]["score"], 150)
	assert_eq(session_lane_state.player_sessions["player-2"]["score"], 90)
	assert_eq(session_lane_state.player_lifecycle["player-2"]["state"], "active")


func test_session_delta_delete_removes_record() -> void:
	var applier := SessionLaneApplier.new()
	var session_lane_state := SessionLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_session_full(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 1,
			"players": [
				{"id": "player-1", "score": 120, "lives": 3},
			],
			"player_lifecycle": [
				{"player_id": "player-1", "state": "active"},
			],
			"is_final_chunk": true,
		}
	)

	applier.apply_session_delta(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 2,
			"player_session_deletes": [
				{"id": "player-1"},
			],
			"player_lifecycle_deletes": [
				{"player_id": "player-1"},
			],
		}
	)

	assert_false(session_lane_state.player_sessions.has("player-1"))
	assert_false(session_lane_state.player_lifecycle.has("player-1"))


func test_session_full_rejects_legacy_player_sessions_packet_input() -> void:
	var applier := SessionLaneApplier.new()
	var session_lane_state := SessionLaneState.new()
	var baseline_tracker := BaselineTracker.new()

	applier.apply_session_full(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 1,
			"player_sessions": [
				{"id": "player-1", "score": 120, "lives": 3},
			],
			"player_lifecycle": [
				{"player_id": "player-1", "state": "active"},
			],
			"is_final_chunk": true,
		}
	)

	assert_false(session_lane_state.player_sessions.has("player-1"))
	assert_eq(session_lane_state.player_lifecycle["player-1"]["state"], "active")


func test_session_delta_accepts_players_and_player_lifecycle_keys() -> void:
	var applier := SessionLaneApplier.new()
	var session_lane_state := SessionLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	applier.apply_session_full(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 1,
			"players": [
				{"id": "player-1", "score": 120, "lives": 3},
			],
			"player_lifecycle": [
				{"player_id": "player-1", "state": "pending_respawn"},
			],
			"is_final_chunk": true,
		}
	)

	var applied := applier.apply_session_delta(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 2,
			"players": [
				{"id": "player-1", "score": 150, "lives": 2},
			],
			"player_lifecycle": [
				{"player_id": "player-1", "state": "active"},
			],
		}
	)

	assert_true(applied)
	assert_eq(session_lane_state.player_sessions["player-1"]["score"], 150)
	assert_eq(session_lane_state.player_sessions["player-1"]["lives"], 2)
	assert_eq(session_lane_state.player_lifecycle["player-1"]["state"], "active")


func test_overlay_only_does_not_mark_gameplay_ready() -> void:
	var overlay_lane_state := OverlayLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	var readiness := GameplayReadiness.new()
	var presentation_adapter := PresentationAdapter.new()
	presentation_adapter.bind_gameplay_readiness(readiness)

	OverlayLaneApplier.new().apply_overlay_full(
		overlay_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_OVERLAY,
		{
			"baseline_id": "overlay-baseline-1",
			"sequence": 1,
			"self_id": "player-1",
			"lives": 3,
			"is_final_chunk": true,
		}
	)
	readiness.mark_overlay_baseline_synced()

	assert_false(presentation_adapter.is_presentable())


func test_session_only_does_not_mark_gameplay_ready() -> void:
	var session_lane_state := SessionLaneState.new()
	var baseline_tracker := BaselineTracker.new()
	var readiness := GameplayReadiness.new()
	var presentation_adapter := PresentationAdapter.new()
	presentation_adapter.bind_gameplay_readiness(readiness)

	SessionLaneApplier.new().apply_session_full(
		session_lane_state,
		baseline_tracker,
		LaneMetadata.LANE_SESSION,
		{
			"baseline_id": "session-baseline-1",
			"sequence": 1,
			"players": [
				{"id": "player-1", "score": 120, "lives": 3},
			],
			"player_lifecycle": [
				{"player_id": "player-1", "state": "active"},
			],
			"is_final_chunk": true,
		}
	)
	readiness.mark_session_baseline_synced()

	assert_false(presentation_adapter.is_presentable())

