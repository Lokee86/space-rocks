extends RefCounted
class_name RealtimePresentationState

var world_lane_state = null
var player_locator_state = null
var overlay_lane_state = null
var session_lane_state = null
var event_batch_applier = null

func update_from_router(router) -> void:
	if router == null:
		clear()
		return
	world_lane_state = router.world_lane_state
	player_locator_state = router.player_locator_state
	overlay_lane_state = router.overlay_lane_state
	session_lane_state = router.session_lane_state
	event_batch_applier = router.event_batch_applier

func clear() -> void:
	world_lane_state = null
	player_locator_state = null
	overlay_lane_state = null
	session_lane_state = null
	event_batch_applier = null
