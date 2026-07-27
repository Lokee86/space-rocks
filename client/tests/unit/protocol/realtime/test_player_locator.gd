extends GutTest

const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")


func test_player_locator_packet_replaces_coarse_player_snapshot() -> void:
	var router := RealtimeRouter.new()
	router.route_lane_packet({
		"type": "player_locator",
		"lane": "ships",
		"match_id": "match-1",
		"sequence": 1,
		"snapshot_id": "ships-locator-1",
		"snapshot_kind": "delta",
		"server_sent_msec": 1234,
		"chunk_index": 0,
		"chunk_count": 1,
		"is_final_chunk": true,
		"player_locators": [
			{"id": "player-2", "x": 10.0, "y": 20.0, "velocity_x": 3.0, "velocity_y": 4.0, "active": true},
		],
	})

	assert_eq(router.player_locator_state.sequence, 1)
	assert_eq(router.player_locator_state.server_sent_msec, 1234)
	assert_true(router.player_locator_state.received_msec > 0)
	assert_eq(router.player_locator_state.player_locators.size(), 1)
	assert_eq(router.player_locator_state.player_locators["player-2"]["velocity_y"], 4.0)

	router.route_lane_packet({
		"type": "player_locator",
		"lane": "ships",
		"match_id": "match-1",
		"sequence": 1,
		"snapshot_id": "ships-locator-stale",
		"snapshot_kind": "delta",
		"server_sent_msec": 1200,
		"chunk_index": 0,
		"chunk_count": 1,
		"is_final_chunk": true,
		"player_locators": [{"id": "player-stale", "active": true}],
	})
	assert_true(router.player_locator_state.player_locators.has("player-2"))
	assert_false(router.player_locator_state.player_locators.has("player-stale"))

	router.reset()
	assert_true(router.player_locator_state.player_locators.is_empty())
	assert_eq(router.player_locator_state.sequence, 0)
	assert_eq(router.player_locator_state.server_sent_msec, 0)
	assert_eq(router.player_locator_state.received_msec, 0)
