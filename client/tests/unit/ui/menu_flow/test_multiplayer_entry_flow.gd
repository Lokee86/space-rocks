extends GutTest

const MultiplayerEntryFlow := preload("res://scripts/ui/menu_flow/multiplayer_entry_flow.gd")
const MenuRoute := preload("res://scripts/ui/menu_flow/menu_route.gd")


class FakeSession:
	extends RefCounted

	var signed_in := false

	func is_signed_in() -> bool:
		return signed_in


class FakeAuthSessionController:
	extends RefCounted

	var session = FakeSession.new()

	func get_session():
		return session


class FakeMenuFlowController:
	extends RefCounted

	var current_route := ""
	var sign_in_calls := 0
	var multiplayer_pregame_calls := 0

	func get_current_route() -> String:
		return current_route

	func show_sign_in_screen() -> void:
		sign_in_calls += 1
		current_route = MenuRoute.SIGN_IN_SCREEN

	func show_multiplayer_pregame() -> void:
		multiplayer_pregame_calls += 1
		current_route = MenuRoute.PREGAME_MENU


func test_request_multiplayer_signed_out_opens_sign_in() -> void:
	var flow = MultiplayerEntryFlow.new()
	var menu_flow_controller = FakeMenuFlowController.new()
	var auth_session_controller = FakeAuthSessionController.new()

	auth_session_controller.session.signed_in = false
	flow.configure(menu_flow_controller, auth_session_controller)

	flow.request_multiplayer()

	assert_eq(menu_flow_controller.sign_in_calls, 1)
	assert_eq(menu_flow_controller.multiplayer_pregame_calls, 0)


func test_request_multiplayer_signed_in_opens_multiplayer_pregame() -> void:
	var flow = MultiplayerEntryFlow.new()
	var menu_flow_controller = FakeMenuFlowController.new()
	var auth_session_controller = FakeAuthSessionController.new()

	auth_session_controller.session.signed_in = true
	flow.configure(menu_flow_controller, auth_session_controller)

	flow.request_multiplayer()

	assert_eq(menu_flow_controller.multiplayer_pregame_calls, 1)
	assert_eq(menu_flow_controller.sign_in_calls, 0)


func test_auth_state_changed_on_sign_in_screen_routes_signed_in_to_pregame() -> void:
	var flow = MultiplayerEntryFlow.new()
	var menu_flow_controller = FakeMenuFlowController.new()
	var auth_session_controller = FakeAuthSessionController.new()

	menu_flow_controller.current_route = MenuRoute.SIGN_IN_SCREEN
	auth_session_controller.session.signed_in = true
	flow.configure(menu_flow_controller, auth_session_controller)

	flow.handle_auth_state_changed()

	assert_eq(menu_flow_controller.multiplayer_pregame_calls, 1)


func test_auth_state_changed_outside_sign_in_screen_does_not_route() -> void:
	var flow = MultiplayerEntryFlow.new()
	var menu_flow_controller = FakeMenuFlowController.new()
	var auth_session_controller = FakeAuthSessionController.new()

	menu_flow_controller.current_route = MenuRoute.MAIN_MENU
	auth_session_controller.session.signed_in = true
	flow.configure(menu_flow_controller, auth_session_controller)

	flow.handle_auth_state_changed()

	assert_eq(menu_flow_controller.multiplayer_pregame_calls, 0)


func test_auth_state_changed_signed_out_on_sign_in_screen_does_not_route() -> void:
	var flow = MultiplayerEntryFlow.new()
	var menu_flow_controller = FakeMenuFlowController.new()
	var auth_session_controller = FakeAuthSessionController.new()

	menu_flow_controller.current_route = MenuRoute.SIGN_IN_SCREEN
	auth_session_controller.session.signed_in = false
	flow.configure(menu_flow_controller, auth_session_controller)

	flow.handle_auth_state_changed()

	assert_eq(menu_flow_controller.multiplayer_pregame_calls, 0)
