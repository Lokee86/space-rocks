extends GutTest

const PlayerLifecycle = preload("res://scripts/gameplay/lifecycle/player_lifecycle.gd")


func test_status_for_accepts_string_value() -> void:
	assert_eq(PlayerLifecycle.status_for({"player-1": "active"}, "player-1"), PlayerLifecycle.STATUS_ACTIVE)


func test_status_for_accepts_state_record() -> void:
	assert_eq(PlayerLifecycle.status_for({"player-1": {"state": "pending_respawn"}}, "player-1"), PlayerLifecycle.STATUS_PENDING_RESPAWN)


func test_status_for_accepts_status_record() -> void:
	assert_eq(PlayerLifecycle.status_for({"player-1": {"status": "eliminated"}}, "player-1"), PlayerLifecycle.STATUS_ELIMINATED)


func test_status_for_returns_empty_for_missing_player() -> void:
	assert_eq(PlayerLifecycle.status_for({}, "player-1"), "")


func test_status_for_returns_empty_for_empty_player_id() -> void:
	assert_eq(PlayerLifecycle.status_for({"player-1": "active"}, ""), "")


func test_known_status_constants() -> void:
	assert_eq(PlayerLifecycle.STATUS_ACTIVE, "active")
	assert_eq(PlayerLifecycle.STATUS_PENDING_RESPAWN, "pending_respawn")
	assert_eq(PlayerLifecycle.STATUS_ELIMINATED, "eliminated")


func test_is_player_active_uses_canonical_status_reader() -> void:
	assert_true(PlayerLifecycle.is_player_active({"player-1": {"status": "active"}}, "player-1"))
	assert_false(PlayerLifecycle.is_player_active({"player-1": "pending_respawn"}, "player-1"))