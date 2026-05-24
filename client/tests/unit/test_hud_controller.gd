extends GutTest

const HudControllerScript := preload("res://scripts/ui/hud_controller.gd")
const HudScene := preload("res://scenes/ui/hud.tscn")

var hud_scene: Control
var hud_controller: HudController


func before_each() -> void:
	hud_scene = HudScene.instantiate()
	add_child(hud_scene)
	hud_controller = HudControllerScript.new()
	hud_controller.configure(hud_scene)


func after_each() -> void:
	hud_controller = null
	if hud_scene != null:
		hud_scene.free()
		hud_scene = null


func test_set_lives_shows_three_lives() -> void:
	hud_controller.set_lives(3)

	assert_eq(_lives_label().text, "3 x ")


func test_set_lives_shows_two_lives() -> void:
	hud_controller.set_lives(2)

	assert_eq(_lives_label().text, "2 x ")


func test_set_lives_shows_zero_lives() -> void:
	hud_controller.set_lives(0)

	assert_eq(_lives_label().text, "0 x ")


func test_set_dead_shows_death_overlay_and_countdown() -> void:
	hud_controller.set_dead(3.2)

	assert_true(hud_controller.death_overlay.visible)
	assert_false(hud_controller.game_over_overlay.visible)
	assert_eq(hud_controller.respawn_timer_label.text, "Respawn in 4")
	assert_true(hud_controller.respawn_timer_label.visible)
	assert_false(hud_controller.respawn_tell_label.visible)


func test_set_alive_hides_death_overlay() -> void:
	hud_controller.set_dead(3.0)
	hud_controller.set_alive()

	assert_false(hud_controller.death_overlay.visible)
	assert_false(hud_controller.game_over_overlay.visible)
	assert_false(hud_controller.is_dead)
	assert_false(hud_controller.can_respawn)


func test_configure_finds_and_hides_hud_owned_game_menu() -> void:
	assert_not_null(hud_controller.get_game_menu())
	assert_false(hud_controller.is_game_menu_visible())
	assert_false(_cycle_view_label().visible)


func test_game_over_overlay_can_hide_without_hiding_game_menu_parent() -> void:
	hud_controller.set_alive()
	hud_controller.show_game_menu()

	assert_false(hud_controller.game_over_overlay.visible)
	assert_true(hud_controller.is_game_menu_visible())
	assert_true(hud_controller.get_game_menu().is_visible_in_tree())


func test_update_shows_manual_respawn_prompt_when_countdown_finishes() -> void:
	hud_controller.set_dead(1.0)
	hud_controller.update(1.0)

	assert_true(hud_controller.death_overlay.visible)
	assert_true(hud_controller.can_respawn)
	assert_eq(hud_controller.respawn_timer_label.text, "")
	assert_true(hud_controller.respawn_tell_label.visible)


func test_room_id_hidden_for_single_player_even_when_room_id_exists() -> void:
	hud_controller.set_session_mode("SinglePlayer")
	hud_controller.set_room_id("ABC123")

	assert_false(_room_id_label().visible)
	assert_eq(_room_id_label().text, "ROOMID: ABC123")


func test_room_id_visible_for_multiplayer_when_room_id_exists() -> void:
	hud_controller.set_session_mode("Multiplayer")
	hud_controller.set_room_id("ABC123")

	assert_true(_room_id_label().visible)
	assert_eq(_room_id_label().text, "ROOMID: ABC123")


func test_room_id_hidden_for_multiplayer_when_room_id_is_empty() -> void:
	hud_controller.set_session_mode("Multiplayer")
	hud_controller.set_room_id("")

	assert_false(_room_id_label().visible)


func test_cycle_view_hidden_for_single_player_even_when_available() -> void:
	hud_controller.set_session_mode("SinglePlayer")
	hud_controller.set_game_over()
	hud_controller.hide_game_menu()
	hud_controller.set_cycle_view_available(true)

	assert_false(_cycle_view_label().visible)


func test_cycle_view_visible_for_multiplayer_game_over_when_menu_hidden() -> void:
	hud_controller.set_session_mode("Multiplayer")
	hud_controller.set_game_over()
	hud_controller.hide_game_menu()
	hud_controller.set_cycle_view_available(true)

	assert_true(_cycle_view_label().visible)


func test_cycle_view_hidden_when_game_menu_visible() -> void:
	hud_controller.set_session_mode("Multiplayer")
	hud_controller.set_game_over()
	hud_controller.hide_game_menu()
	hud_controller.set_cycle_view_available(true)

	hud_controller.show_game_menu()

	assert_false(_cycle_view_label().visible)


func test_cycle_view_hidden_for_alive_gameplay() -> void:
	hud_controller.set_session_mode("Multiplayer")
	hud_controller.set_game_over()
	hud_controller.hide_game_menu()
	hud_controller.set_cycle_view_available(true)

	hud_controller.set_alive()

	assert_false(_cycle_view_label().visible)


func _lives_label() -> Label:
	return hud_scene.find_child("LivesCount", true, false) as Label


func _room_id_label() -> Label:
	return hud_scene.find_child("RoomID", true, false) as Label


func _cycle_view_label() -> Label:
	return hud_scene.find_child("CycleView", true, false) as Label
