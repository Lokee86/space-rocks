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
