extends GutTest

const WorldPresentationAdapter := preload("res://scripts/protocol/realtime/world_presentation_adapter.gd")
const OverlayPresentationAdapter := preload("res://scripts/protocol/realtime/overlay_presentation_adapter.gd")
const SessionPresentationAdapter := preload("res://scripts/protocol/realtime/session_presentation_adapter.gd")
const EventPresentationAdapter := preload("res://scripts/protocol/realtime/event_presentation_adapter.gd")
const DebugPresentationAdapter := preload("res://scripts/protocol/realtime/debug_presentation_adapter.gd")
const GameplayHudFlow := preload("res://scripts/shell/gameplay_hud_flow.gd")
const HudScene := preload("res://scenes/ui/hud.tscn")
const WorldLaneState := preload("res://scripts/protocol/realtime/world_lane_state.gd")
const OverlayLaneState := preload("res://scripts/protocol/realtime/overlay_lane_state.gd")
const SessionLaneState := preload("res://scripts/protocol/realtime/session_lane_state.gd")
const EventBatchApplier := preload("res://scripts/protocol/realtime/event_batch_applier.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


class FakeWorldSync:
	var applied_world_lane_state = null
	var current_self_id := ""

	func set_current_self_id(self_id: String) -> void:
		current_self_id = self_id

	func apply_world_lane_state(world_lane_state) -> void:
		applied_world_lane_state = world_lane_state


class FakeHudFlow:
	var overlay_lane_state = null
	var session_lane_state = null
	var applied_events: Array = []
	var applied_self_id := ""

	func apply_overlay_lane_state(state) -> void:
		overlay_lane_state = state

	func apply_session_lane_state(state, self_id := "") -> void:
		session_lane_state = state
		applied_self_id = self_id

	func apply_server_events(events: Array, self_id: String) -> void:
		applied_events.append({"events": events, "self_id": self_id})


func test_world_adapter_applies_lane_state_without_full_read_model() -> void:
	var adapter := WorldPresentationAdapter.new()
	var world_sync := FakeWorldSync.new()
	var world_lane_state := WorldLaneState.new()
	world_lane_state.ships = {"ship-1": {"id": "ship-1"}}
	world_lane_state.bullets = {"bullet-1": {"id": "bullet-1"}}
	world_lane_state.asteroids = {"asteroid-1": {"id": "asteroid-1"}}
	world_lane_state.pickups = {"pickup-1": {"id": "pickup-1"}}

	adapter.apply_world_lane_state(world_sync, world_lane_state, "player-1")

	assert_eq(world_sync.current_self_id, "player-1")
	assert_eq(world_sync.applied_world_lane_state, world_lane_state)


func test_overlay_adapter_updates_hud_from_overlay_lane() -> void:
	var adapter := OverlayPresentationAdapter.new()
	var hud_flow := FakeHudFlow.new()
	var overlay_lane_state := OverlayLaneState.new()
	overlay_lane_state.self_id = "player-1"
	overlay_lane_state.lives = 3
	overlay_lane_state.score = 120
	overlay_lane_state.primary_weapon_id = "laser"
	overlay_lane_state.secondary_weapon_id = "burst"

	adapter.apply_overlay_lane_state(hud_flow, overlay_lane_state)

	assert_eq(hud_flow.overlay_lane_state, overlay_lane_state)


func test_session_adapter_updates_hud_from_session_lane() -> void:
	var adapter := SessionPresentationAdapter.new()
	var hud_flow := FakeHudFlow.new()
	var session_lane_state := SessionLaneState.new()
	session_lane_state.total_asteroids = 4
	session_lane_state.player_sessions = {"player-1": {"score": 120, "lives": 3}}
	session_lane_state.player_lifecycle = {"player-1": {"state": "active"}}

	adapter.apply_session_lane_state(hud_flow, session_lane_state, "player-1")

	assert_eq(hud_flow.session_lane_state, session_lane_state)
	assert_eq(hud_flow.applied_self_id, "player-1")


func test_gameplay_hud_flow_session_lane_zero_cooldown_keeps_dead_presentation_and_makes_respawn_available_by_countdown() -> void:
	var hud := HudScene.instantiate()
	add_child_autofree(hud)
	var hud_flow := GameplayHudFlow.new()
	hud_flow.configure(hud)
	hud_flow.apply_score(120)
	hud_flow.apply_lives(2)
	hud_flow.set_dead(0.5)

	var session_lane_state := SessionLaneState.new()
	session_lane_state.player_sessions = {
		"player-1": {
			"score": 120,
			"lives": 2,
			"respawn_cooldown": 0.0,
		}
	}
	session_lane_state.player_lifecycle = {
		"player-1": {
			"status": "active",
		}
	}

	hud_flow.apply_session_lane_state(session_lane_state, "player-1")
	assert_true(hud_flow.is_dead)
	assert_false(hud_flow.can_respawn)
	assert_eq(hud_flow.respawn_countdown_remaining, 0.5)
	assert_true((hud.get_node("CenterContainer/VBoxContainer2") as CanvasItem).visible)
	assert_eq(hud_flow.score(), 120)

	hud_flow.update(0.5)

	assert_true(hud_flow.is_dead)
	assert_true(hud_flow.can_respawn)
	assert_eq(hud_flow.respawn_countdown_remaining, 0.0)


func test_gameplay_hud_flow_overlay_lane_shows_torpedo_loadout_with_cooldown() -> void:
	var hud := HudScene.instantiate()
	add_child_autofree(hud)
	var hud_flow := GameplayHudFlow.new()
	hud_flow.configure(hud)
	var overlay_lane_state := OverlayLaneState.new()
	overlay_lane_state.self_id = "player-1"
	overlay_lane_state.secondary_weapon_id = "torpedo"
	overlay_lane_state.secondary_ammo_policy = "limited"
	overlay_lane_state.secondary_ammo_remaining = 3
	overlay_lane_state.respawn_cooldown = 4.0

	hud_flow.apply_overlay_lane_state(overlay_lane_state)

	var loadout_container := hud.get_node("%LoadoutContainer") as HBoxContainer
	assert_eq(loadout_container.get_child_count(), 1)
	var display := loadout_container.get_child(0)
	assert_true((display.get_node("%CooldownOverlay") as CanvasItem).visible)
	assert_eq((display.get_node("%CooldownOverlay/CooldownLabel") as Label).text, "4.0")
	assert_true((display.get_node("%AmmoLabel") as CanvasItem).visible)


func test_gameplay_hud_flow_session_lane_does_not_overwrite_overlay_owned_torpedo_loadout() -> void:
	var hud := HudScene.instantiate()
	add_child_autofree(hud)
	var hud_flow := GameplayHudFlow.new()
	hud_flow.configure(hud)
	var overlay_lane_state := OverlayLaneState.new()
	overlay_lane_state.self_id = "player-1"
	overlay_lane_state.secondary_weapon_id = "torpedo"
	overlay_lane_state.secondary_ammo_policy = "limited"
	overlay_lane_state.secondary_ammo_remaining = 2
	overlay_lane_state.respawn_cooldown = 3.0
	hud_flow.apply_overlay_lane_state(overlay_lane_state)

	var session_lane_state := SessionLaneState.new()
	session_lane_state.player_sessions = {
		"player-1": {
			Packets.FIELD_SCORE: 120,
			Packets.FIELD_LIVES: 2,
			Packets.FIELD_SECONDARY_WEAPON_ID: "mine",
			Packets.FIELD_SECONDARY_AMMO_POLICY: "infinite",
		}
	}
	hud_flow.apply_session_lane_state(session_lane_state, "player-1")

	var loadout_container := hud.get_node("%LoadoutContainer") as HBoxContainer
	assert_eq(loadout_container.get_child_count(), 1)
	var display := loadout_container.get_child(0)
	assert_true((display.get_node("%CooldownOverlay") as CanvasItem).visible)
	assert_eq((display.get_node("%CooldownOverlay/CooldownLabel") as Label).text, "3.0")
	assert_eq((display.get_node("%AmmoLabel") as Label).text, "x2")


func test_event_adapter_displays_once_and_suppresses_duplicates() -> void:
	var adapter := EventPresentationAdapter.new()
	var hud_flow := FakeHudFlow.new()
	var applier := EventBatchApplier.new()

	applier.apply_event_batch(
		{
			"batch_id": "batch-1",
			"events": [
				{"event_id": "event-1", "type": "spark", "payload": {"value": 1}},
			],
		},
		null
	)
	adapter.apply_event_batch_output(hud_flow, applier, "player-1")
	adapter.apply_event_batch_output(hud_flow, applier, "player-1")

	assert_eq(hud_flow.applied_events.size(), 1)
	assert_eq(hud_flow.applied_events[0]["self_id"], "player-1")
	assert_eq(hud_flow.applied_events[0]["events"].size(), 1)


func test_debug_adapter_does_not_mark_gameplay_ready() -> void:
	var adapter := DebugPresentationAdapter.new()
	var presentation_adapter := preload("res://scripts/protocol/realtime/presentation_adapter.gd").new()

	adapter.apply_debug_packet({"type": "debug_telemetry", "payload": {"value": 1}})

	assert_false(presentation_adapter.is_presentable())
