extends GutTest

const HudScene := preload("res://scenes/ui/hud.tscn")
const GameplayHudFlow := preload("res://scripts/shell/gameplay_hud_flow.gd")


func test_match_timer_replaces_room_id_and_uses_hud_label_settings() -> void:
	var hud := await _create_hud()
	var timer_label := hud.get_node_or_null("MatchTimer") as Label

	assert_not_null(timer_label)
	assert_null(hud.get_node_or_null("RoomID"))
	assert_eq(timer_label.text, "TIME: 00:00.00")
	assert_same(timer_label.label_settings, (hud.get_node("MarginContainer/HBoxContainer/MarginContainer/Score") as Label).label_settings)


func test_match_timer_starts_with_gameplay_and_advances_from_delta() -> void:
	var hud := await _create_hud()
	var flow := GameplayHudFlow.new()
	flow.configure(hud)

	flow.show_gameplay()
	flow.update(65.437)

	assert_eq((hud.get_node("MatchTimer") as Label).text, "TIME: 01:05.43")


func test_match_timer_stops_at_game_over_and_resets_for_next_match() -> void:
	var hud := await _create_hud()
	var flow := GameplayHudFlow.new()
	flow.configure(hud)
	flow.show_gameplay()
	flow.update(12.34)

	flow.set_game_over()
	flow.update(5.0)
	assert_eq((hud.get_node("MatchTimer") as Label).text, "TIME: 00:12.34")

	flow.reset()
	assert_eq((hud.get_node("MatchTimer") as Label).text, "TIME: 00:00.00")
	flow.show_gameplay()
	flow.update(1.25)
	assert_eq((hud.get_node("MatchTimer") as Label).text, "TIME: 00:01.25")


func _create_hud() -> Control:
	var hud := HudScene.instantiate()
	add_child_autofree(hud)
	await get_tree().process_frame
	return hud
