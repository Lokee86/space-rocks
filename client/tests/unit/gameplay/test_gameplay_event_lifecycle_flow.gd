extends GutTest

const GameplayEventLifecycleFlow = preload("res://scripts/gameplay/events/gameplay_event_lifecycle_flow.gd")
const GameplayDeathFlow = preload("res://scripts/gameplay/events/gameplay_death_flow.gd")

var _game_owner: FakeNode2D
var _hud: FakeControl


class FakeEventFlow:
	signal self_death_event

	var configure_call_count := 0
	var reset_call_count := 0
	var apply_server_events_call_count := 0
	var last_game_owner
	var last_hud
	var last_visual_position_callable
	var last_server_events
	var last_self_id

	func configure(game_owner: Node2D, hud: Control, visual_position_for_server_position: Callable) -> void:
		configure_call_count += 1
		last_game_owner = game_owner
		last_hud = hud
		last_visual_position_callable = visual_position_for_server_position

	func apply_server_events(server_events: Array, self_id: String) -> void:
		apply_server_events_call_count += 1
		last_server_events = server_events
		last_self_id = self_id

	func reset() -> void:
		reset_call_count += 1


class FakeDeathFlow:
	var configure_call_count := 0
	var last_hud_flow
	var last_menu_flow
	var last_player
	var apply_self_death_event_call_count := 0

	func configure(hud_flow_ref, menu_flow_ref, player_ref = null) -> void:
		configure_call_count += 1
		last_hud_flow = hud_flow_ref
		last_menu_flow = menu_flow_ref
		last_player = player_ref

	func apply_self_death_event(_event) -> void:
		apply_self_death_event_call_count += 1


class FakeMatchEndFlow:
	var local_player_eliminated_call_count := 0
	var last_event: Dictionary = {}

	func handle_local_player_eliminated(event: Dictionary) -> void:
		local_player_eliminated_call_count += 1
		last_event = event


class FakeHudFlow:
	var last_lives := -1
	var game_over_calls := 0
	var dead_calls := 0

	func apply_lives(lives) -> void:
		last_lives = lives

	func set_dead(respawn_delay) -> void:
		dead_calls += 1

	func set_game_over() -> void:
		game_over_calls += 1


class FakeMenuFlow:
	var game_over_calls := 0

	func set_game_over() -> void:
		game_over_calls += 1


class FakeNode2D:
	extends Node2D


class FakeControl:
	extends Control


class FakeCallableTarget:
	func visual_position_for_server_position(_server_position):
		return Vector2.ZERO


func after_each() -> void:
	if is_instance_valid(_game_owner):
		_game_owner.queue_free()
		_game_owner = null
	if is_instance_valid(_hud):
		_hud.queue_free()
		_hud = null


func test_configure_creates_event_and_death_flows() -> void:
	var event_flow := FakeEventFlow.new()
	var death_flow := FakeDeathFlow.new()
	var flow := GameplayEventLifecycleFlow.new()
	_game_owner = FakeNode2D.new()
	add_child_autofree(_game_owner)
	_hud = FakeControl.new()
	add_child_autofree(_hud)
	var hud_flow := Object.new()
	var menu_flow := Object.new()
	var player := Object.new()
	var callable_target := FakeCallableTarget.new()

	flow.configure(
		_game_owner,
		_hud,
		hud_flow,
		menu_flow,
		player,
		Callable(callable_target, "visual_position_for_server_position"),
		event_flow,
		death_flow
	)

	assert_eq(flow.event_flow, event_flow)
	assert_eq(flow.death_flow, death_flow)
	assert_eq(event_flow.configure_call_count, 1)
	assert_eq(death_flow.configure_call_count, 1)
	assert_eq(death_flow.last_player, player)


func test_apply_server_events_forwards_state_fields() -> void:
	var event_flow := FakeEventFlow.new()
	var death_flow := FakeDeathFlow.new()
	var flow := GameplayEventLifecycleFlow.new()
	var callable_target := FakeCallableTarget.new()
	_game_owner = FakeNode2D.new()
	add_child_autofree(_game_owner)
	_hud = FakeControl.new()
	add_child_autofree(_hud)

	flow.configure(
		_game_owner,
		_hud,
		Object.new(),
		Object.new(),
		Object.new(),
		Callable(callable_target, "visual_position_for_server_position"),
		event_flow,
		death_flow
	)

	var state := {
		"server_events": [{"type": "test_event"}],
		"self_id": "player-1",
	}
	flow.apply_server_events_from_state(state)

	assert_eq(event_flow.apply_server_events_call_count, 1)
	assert_eq(event_flow.last_server_events, state["server_events"])
	assert_eq(event_flow.last_self_id, "player-1")


func test_reset_calls_owned_event_flow_reset() -> void:
	var event_flow := FakeEventFlow.new()
	var death_flow := FakeDeathFlow.new()
	var flow := GameplayEventLifecycleFlow.new()
	var callable_target := FakeCallableTarget.new()
	_game_owner = FakeNode2D.new()
	add_child_autofree(_game_owner)
	_hud = FakeControl.new()
	add_child_autofree(_hud)

	flow.configure(
		_game_owner,
		_hud,
		Object.new(),
		Object.new(),
		Object.new(),
		Callable(callable_target, "visual_position_for_server_position"),
		event_flow,
		death_flow
	)

	flow.reset()

	assert_eq(event_flow.reset_call_count, 1)


func test_apply_self_death_event_final_death_uses_game_end_handoff() -> void:
	var death_flow := GameplayDeathFlow.new()
	var hud_flow := FakeHudFlow.new()
	var match_end_flow := FakeMatchEndFlow.new()

	death_flow.configure(hud_flow, match_end_flow, null)

	death_flow.apply_self_death_event({"lives": 0})

	assert_eq(hud_flow.last_lives, 0)
	assert_eq(match_end_flow.local_player_eliminated_call_count, 1)
	assert_eq(match_end_flow.last_event["lives"], 0)


func test_apply_self_death_event_non_final_death_uses_dead_presentation() -> void:
	var death_flow := GameplayDeathFlow.new()
	var hud_flow := FakeHudFlow.new()

	death_flow.configure(hud_flow, null)

	death_flow.apply_self_death_event({"lives": 2})

	assert_eq(hud_flow.last_lives, 2)
	assert_eq(hud_flow.dead_calls, 1)
