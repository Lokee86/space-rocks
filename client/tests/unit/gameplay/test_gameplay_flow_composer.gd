extends GutTest

const GameplayFlowComposer = preload("res://scripts/gameplay/runtime/gameplay_flow_composer.gd")
const GameplayHudFlow = preload("res://scripts/shell/gameplay_hud_flow.gd")

var nodes_to_free: Array[Node] = []


class FakeInputContext:
	var handle_unhandled_input_call_count := 0
	var last_event = null
	var last_required_lane_baselines_synced = null
	var configure_spectate_routes_call_count := 0

	func handle_unhandled_input(event, required_lane_baselines_synced) -> bool:
		handle_unhandled_input_call_count += 1
		last_event = event
		last_required_lane_baselines_synced = required_lane_baselines_synced
		return true

	func configure_spectate_routes(_open_spectate_menu, _cycle_target) -> void:
		configure_spectate_routes_call_count += 1


class FakeGameplayStateApplyFlow:
	var apply_state_call_count := 0
	var last_state = null
	var last_required_lane_baselines_synced = null
	var return_result := GameplayStateApplyResult.new()

	func apply_state(state: Dictionary, required_lane_baselines_synced: bool) -> GameplayStateApplyResult:
		apply_state_call_count += 1
		last_state = state
		last_required_lane_baselines_synced = required_lane_baselines_synced
		return return_result


class FakeProcessFlow:
	var process_call_count := 0
	var last_delta := -1.0
	var last_required_lane_baselines_synced = null

	func process(delta: float, required_lane_baselines_synced: bool) -> void:
		process_call_count += 1
		last_delta = delta
		last_required_lane_baselines_synced = required_lane_baselines_synced


class FakeResettableFlow:
	var reset_call_count := 0

	func reset() -> void:
		reset_call_count += 1


class FakeDevtoolsContext:
	var configure_call_count := 0

	func configure(_connection_service_ref) -> void:
		configure_call_count += 1


class FakeWorldSync:
	func visual_position_for_server_position(value):
		return value

	func server_position_for_visual_position(value):
		return value

	func target_source():
		return RefCounted.new()

	func remote_player_nodes() -> Dictionary:
		return {}


class FakeRuntimeContext:
	var world_sync := FakeWorldSync.new()
	var respawn_flow := RefCounted.new()

	func request_respawn(_required_lane_baselines_synced: bool) -> void:
		pass

	func remote_player_nodes() -> Dictionary:
		return {}


func after_each() -> void:
	for node in nodes_to_free:
		if is_instance_valid(node):
			node.free()
	nodes_to_free.clear()


func _tracked(node: Node) -> Node:
	nodes_to_free.append(node)
	return node


func test_apply_gameplay_state_delegates_to_injected_flow() -> void:
	var composer = GameplayFlowComposer.new()
	var fake_state_apply_flow = FakeGameplayStateApplyFlow.new()
	fake_state_apply_flow.return_result.gameplay_ready = true
	fake_state_apply_flow.return_result.started_gameplay = true
	composer.configure(
		null,
		_tracked(Node2D.new()),
		_tracked(Player.new()),
		null,
		null,
		FakeRuntimeContext.new(),
		null,
		null,
		null,
		fake_state_apply_flow,
		null
	)

	var state := {"phase": 9}
	var result: GameplayStateApplyResult = composer.apply_gameplay_state(state, false)

	assert_eq(fake_state_apply_flow.apply_state_call_count, 1)
	assert_eq(fake_state_apply_flow.last_state, state)
	assert_eq(fake_state_apply_flow.last_required_lane_baselines_synced, false)
	assert_eq(result, fake_state_apply_flow.return_result)
	assert_true(result.gameplay_ready)
	assert_true(result.started_gameplay)


func test_handle_unhandled_input_delegates_to_injected_input_context() -> void:
	var composer = GameplayFlowComposer.new()
	var fake_input_context = FakeInputContext.new()
	composer.configure(
		null,
		_tracked(Node2D.new()),
		_tracked(Player.new()),
		null,
		null,
		FakeRuntimeContext.new(),
		null,
		fake_input_context,
		null,
		null,
		null
	)

	var event := InputEventKey.new()
	var handled := composer.handle_unhandled_input(event, true)

	assert_true(handled)
	assert_eq(fake_input_context.handle_unhandled_input_call_count, 1)
	assert_eq(fake_input_context.last_event, event)
	assert_eq(fake_input_context.last_required_lane_baselines_synced, true)


func test_handle_unhandled_input_returns_false_without_input_context() -> void:
	var composer = GameplayFlowComposer.new()
	composer.configure(
		null,
		_tracked(Node2D.new()),
		_tracked(Player.new()),
		null,
		null,
		FakeRuntimeContext.new()
	)
	composer.input_context = null

	assert_false(composer.handle_unhandled_input(InputEventKey.new(), false))


func test_process_delegates_to_injected_process_flow() -> void:
	var composer = GameplayFlowComposer.new()
	var fake_process_flow = FakeProcessFlow.new()
	composer.configure(
		null,
		_tracked(Node2D.new()),
		_tracked(Player.new()),
		null,
		null,
		FakeRuntimeContext.new(),
		null,
		null,
		null,
		null,
		fake_process_flow
	)

	composer.process(0.25, true)

	assert_eq(fake_process_flow.process_call_count, 1)
	assert_eq(fake_process_flow.last_delta, 0.25)
	assert_true(fake_process_flow.last_required_lane_baselines_synced)


func test_reset_calls_owned_flow_resets() -> void:
	var composer = GameplayFlowComposer.new()
	composer.configure(
		null,
		_tracked(Node2D.new()),
		_tracked(Player.new()),
		null,
		null,
		FakeRuntimeContext.new(),
		null,
		null,
		null,
		null,
		null
	)
	composer.input_context = FakeResettableFlow.new()
	composer.event_lifecycle_flow = FakeResettableFlow.new()
	composer.alive_restore_flow = FakeResettableFlow.new()
	composer.runtime_tick_flow = FakeResettableFlow.new()
	composer.spectate_context = FakeResettableFlow.new()
	composer.server_hitbox_overlay_flow = FakeResettableFlow.new()

	composer.reset()

	assert_eq(composer.input_context.reset_call_count, 1)
	assert_eq(composer.event_lifecycle_flow.reset_call_count, 1)
	assert_eq(composer.alive_restore_flow.reset_call_count, 1)
	assert_eq(composer.runtime_tick_flow.reset_call_count, 1)
	assert_eq(composer.spectate_context.reset_call_count, 1)
	assert_eq(composer.server_hitbox_overlay_flow.reset_call_count, 1)


func test_configure_creates_core_owned_flows() -> void:
	var composer = GameplayFlowComposer.new()
	composer.configure(
		null,
		_tracked(Node2D.new()),
		_tracked(Player.new()),
		null,
		null,
		FakeRuntimeContext.new(),
		null,
		FakeInputContext.new(),
		FakeDevtoolsContext.new(),
		FakeGameplayStateApplyFlow.new(),
		FakeProcessFlow.new()
	)

	assert_not_null(composer.event_lifecycle_flow)
	assert_not_null(composer.alive_restore_flow)
	assert_not_null(composer.targeting_context)
	assert_not_null(composer.pointer_position_provider)
	assert_not_null(composer.input_context)
	assert_not_null(composer.devtools_context)
	assert_not_null(composer.gameplay_state_apply_flow)
func test_configure_accepts_gameplay_hud_flow_in_shell_argument_slot() -> void:
	var composer = GameplayFlowComposer.new()
	var hud := Control.new()
	var gameplay_ui := Control.new()
	var scene_root := Node2D.new()
	var player := Player.new()
	var view_anchor := Node2D.new()
	var bullets := Node2D.new()
	var asteroids := Node2D.new()
	var pickups := Node2D.new()
	var session_context := RefCounted.new()

	composer.configure(
		null,
		_tracked(scene_root),
		_tracked(player),
		_tracked(view_anchor),
		_tracked(bullets),
		_tracked(asteroids),
		_tracked(pickups),
		hud,
		gameplay_ui,
		session_context,
		Callable()
	)

	assert_not_null(composer.gameplay_shell_flow)
	assert_true(composer.gameplay_shell_flow.gameplay_hud_flow is GameplayHudFlow)
