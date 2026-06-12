extends GutTest

const SignInFlow := preload("res://scripts/ui/sign_in/sign_in_flow.gd")


class FakeLoginWindow:
	extends Control

	signal back_requested
	signal discord_login_requested


class Probe:
	extends RefCounted

	var calls := 0

	func mark_called() -> void:
		calls += 1


func test_back_requested_calls_show_main_menu() -> void:
	var fake := FakeLoginWindow.new()
	var flow := SignInFlow.new()
	var show_main_probe := Probe.new()
	var discord_probe := Probe.new()

	add_child_autofree(fake)
	flow.configure(fake, Callable(show_main_probe, "mark_called"), Callable(discord_probe, "mark_called"))

	fake.back_requested.emit()

	assert_eq(show_main_probe.calls, 1)
	assert_eq(discord_probe.calls, 0)


func test_discord_login_requested_calls_discord_sign_in() -> void:
	var fake := FakeLoginWindow.new()
	var flow := SignInFlow.new()
	var show_main_probe := Probe.new()
	var discord_probe := Probe.new()

	add_child_autofree(fake)
	flow.configure(fake, Callable(show_main_probe, "mark_called"), Callable(discord_probe, "mark_called"))

	fake.discord_login_requested.emit()

	assert_eq(discord_probe.calls, 1)
	assert_eq(show_main_probe.calls, 0)
