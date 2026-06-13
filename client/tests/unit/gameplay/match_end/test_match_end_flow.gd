extends GutTest

const MatchEndFlow := preload("res://scripts/gameplay/match_end/match_end_flow.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


class FakeSessionContext:
	extends RefCounted

	var active_mode := ""


class FakeRoomStateProvider:
	extends RefCounted

	var room_state := ""

	func current_room_state() -> String:
		return room_state


class FakeMatchResultProvider:
	extends RefCounted

	var match_result := {}

	func current_match_result() -> Dictionary:
		return match_result


class FakeHudFlow:
	extends RefCounted

	var hud := Control.new()
	var last_lives := -1
	var game_over_calls := 0

	func apply_lives(lives) -> void:
		last_lives = lives

	func set_game_over() -> void:
		game_over_calls += 1


class FakeMenuFlow:
	extends RefCounted

	var game_over_calls := 0

	func set_game_over() -> void:
		game_over_calls += 1


class FakeEventFlow:
	extends RefCounted

	var play_game_over_sound_after_delay_calls := 0

	func play_game_over_sound_after_delay() -> void:
		play_game_over_sound_after_delay_calls += 1


class FakeMatchResultsFlow:
	extends RefCounted

	var show_results_calls := 0
	var last_session_mode := ""
	var last_rows := []

	func show_results(session_mode: String, rows: Array = []) -> Control:
		show_results_calls += 1
		last_session_mode = session_mode
		last_rows = rows
		return null


func test_handle_local_player_eliminated_applies_game_over_orchestration() -> void:
	var flow := MatchEndFlow.new()
	var hud_flow := FakeHudFlow.new()
	var menu_flow := FakeMenuFlow.new()
	var event_flow := FakeEventFlow.new()
	var match_results_flow := FakeMatchResultsFlow.new()

	flow.configure(hud_flow, menu_flow)
	flow.configure_event_flow(event_flow)
	flow.configure_match_results_flow(match_results_flow)

	flow.handle_local_player_eliminated({Packets.FIELD_LIVES: 0})

	assert_eq(hud_flow.last_lives, 0)
	assert_eq(hud_flow.game_over_calls, 1)
	assert_eq(menu_flow.game_over_calls, 1)
	assert_eq(event_flow.play_game_over_sound_after_delay_calls, 1)
	assert_eq(match_results_flow.show_results_calls, 0)


func test_handle_room_match_over_hides_hud_and_passes_rows_to_results() -> void:
	var flow := MatchEndFlow.new()
	var hud_flow := FakeHudFlow.new()
	var menu_flow := FakeMenuFlow.new()
	var event_flow := FakeEventFlow.new()
	var match_results_flow := FakeMatchResultsFlow.new()
	var session_context := FakeSessionContext.new()
	session_context.active_mode = Constants.SESSION_MODE_MULTIPLAYER
	var match_result_provider := FakeMatchResultProvider.new()
	match_result_provider.match_result = {
		"players": [
			{
				"game_player_id": "player-1",
				"score": 125,
				"ship_deaths": 3,
				"won": true,
			}
		]
	}

	add_child_autofree(hud_flow.hud)
	flow.configure(hud_flow, menu_flow, session_context)
	flow.configure_event_flow(event_flow)
	flow.configure_match_results_flow(match_results_flow)
	flow.configure_match_result_provider(Callable(match_result_provider, "current_match_result"))

	flow.handle_room_match_over()

	assert_false(hud_flow.hud.visible)
	assert_eq(menu_flow.game_over_calls, 1)
	assert_eq(event_flow.play_game_over_sound_after_delay_calls, 1)
	assert_eq(match_results_flow.show_results_calls, 1)
	assert_eq(match_results_flow.last_session_mode, "multiplayer")
	assert_eq(match_results_flow.last_rows, [
		{
			"game_player_id": "player-1",
			"score": 125,
			"ship_deaths": 3,
			"won": true,
		}
	])


func test_handle_room_match_over_passes_empty_rows_when_provider_returns_empty_dictionary() -> void:
	var flow := MatchEndFlow.new()
	var hud_flow := FakeHudFlow.new()
	var menu_flow := FakeMenuFlow.new()
	var event_flow := FakeEventFlow.new()
	var match_results_flow := FakeMatchResultsFlow.new()
	var session_context := FakeSessionContext.new()
	session_context.active_mode = Constants.SESSION_MODE_MULTIPLAYER
	var match_result_provider := FakeMatchResultProvider.new()
	match_result_provider.match_result = {}

	add_child_autofree(hud_flow.hud)
	flow.configure(hud_flow, menu_flow, session_context)
	flow.configure_event_flow(event_flow)
	flow.configure_match_results_flow(match_results_flow)
	flow.configure_match_result_provider(Callable(match_result_provider, "current_match_result"))

	flow.handle_room_match_over()

	assert_eq(match_results_flow.show_results_calls, 1)
	assert_eq(match_results_flow.last_rows, [])


func test_refresh_match_end_state_ignores_repeated_room_match_over() -> void:
	var flow := MatchEndFlow.new()
	var hud_flow := FakeHudFlow.new()
	var menu_flow := FakeMenuFlow.new()
	var event_flow := FakeEventFlow.new()
	var match_results_flow := FakeMatchResultsFlow.new()
	var session_context := FakeSessionContext.new()
	session_context.active_mode = Constants.SESSION_MODE_MULTIPLAYER
	var room_state_provider := FakeRoomStateProvider.new()
	room_state_provider.room_state = Constants.ROOM_STATE_GAME_OVER

	add_child_autofree(hud_flow.hud)
	flow.configure(hud_flow, menu_flow, session_context)
	flow.configure_event_flow(event_flow)
	flow.configure_match_results_flow(match_results_flow)
	flow.configure_room_state_provider(Callable(room_state_provider, "current_room_state"))

	flow.refresh_match_end_state()
	flow.refresh_match_end_state()

	assert_eq(match_results_flow.show_results_calls, 1)
	assert_false(hud_flow.hud.visible)
