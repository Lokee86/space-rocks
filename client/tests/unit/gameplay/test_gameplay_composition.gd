extends GutTest

const GameplayComposition := preload("res://scripts/gameplay/gameplay_composition.gd")
const MatchResultsFlow := preload("res://scripts/ui/match_results/match_results_flow.gd")


class FakeConnectionService:
	extends RefCounted

	signal realtime_replay_availability_changed(available: bool)

	var replay_available := false

	func is_realtime_replay_available() -> bool:
		return replay_available


class FakeMatchResultsFlow extends MatchResultsFlow:

	var availability_values: Array = []

	func set_replay_available(available: bool) -> void:
		availability_values.append(available)


class FakeSpectateSessionFlow:
	extends RefCounted

	var applied_states: Array[Dictionary] = []

	func apply_gameplay_state(state: Dictionary) -> bool:
		applied_states.append(state)
		return true


class FakeMenuFlow:
	extends RefCounted

	var refresh_count := 0

	func refresh_game_over_menu_state() -> void:
		refresh_count += 1


class FakeShellFlow:
	extends RefCounted

	var applied_states: Array[Dictionary] = []
	var refreshed_slot_counts: Array[int] = []

	func apply_devtools_gameplay_state(state: Dictionary) -> void:
		applied_states.append(state)

	func refresh_devtools_spawn_player_slots(max_players: int) -> void:
		refreshed_slot_counts.append(max_players)


func test_gameplay_composition_initializes_and_propagates_replay_availability() -> void:
	var composition := GameplayComposition.new()
	var connection_service := FakeConnectionService.new()
	var match_results_flow := FakeMatchResultsFlow.new()
	composition.connection_service = connection_service
	composition.match_results_flow = match_results_flow

	composition._configure_realtime_replay_availability()
	connection_service.replay_available = true
	connection_service.realtime_replay_availability_changed.emit(true)

	assert_eq(match_results_flow.availability_values, [false, true])


func test_gameplay_state_updates_spectate_targets_before_refreshing_game_over_menu() -> void:
	var composition := GameplayComposition.new()
	var spectate_flow := FakeSpectateSessionFlow.new()
	var menu_flow := FakeMenuFlow.new()
	var shell_flow := FakeShellFlow.new()
	composition.spectate_session_flow = spectate_flow
	composition.gameplay_menu_flow = menu_flow
	composition.gameplay_shell_flow = shell_flow
	composition.room_max_players_provider = func() -> int: return 8
	var state := {
		"overlay": {"self_id": "player-1"},
		"session": {"player_lifecycle": {"player-2": "active"}},
	}

	composition.apply_devtools_gameplay_state(state)

	assert_eq(spectate_flow.applied_states, [state])
	assert_eq(menu_flow.refresh_count, 1)
	assert_eq(shell_flow.applied_states, [state])
	assert_eq(shell_flow.refreshed_slot_counts, [8])
