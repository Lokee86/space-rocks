extends GutTest

const EffectsScript := preload("res://scripts/gameplay/effects.gd")
const CameraFollowScript := preload("res://scripts/camera/camera_follow.gd")
const GameScript := preload("res://scripts/gameplay/game.gd")
const HudControllerScript := preload("res://scripts/ui/hud/hud_controller.gd")
const HudScene := preload("res://scenes/ui/hud.tscn")
const NetworkClientScript := preload("res://scripts/networking/network_client.gd")
const Packets := preload("res://scripts/networking/packets.gd")
const PlayerScene := preload("res://scenes/player.tscn")
const WorldStateFixture := preload("res://tests/fixtures/world_state_fixture.gd")
const WorldSyncScript := preload("res://scripts/networking/world_sync.gd")

var game
var hud_scene: Control
var hud_controller: HudController
var effects_owner: Node2D


func before_each() -> void:
	game = GameScript.new()
	game.player = PlayerScene.instantiate()
	game.bullets = Node2D.new()
	game.asteroids = Node2D.new()
	game.add_child(game.player)
	game.add_child(game.bullets)
	game.add_child(game.asteroids)

	hud_scene = HudScene.instantiate()
	add_child(hud_scene)
	hud_controller = HudControllerScript.new()
	hud_controller.configure(hud_scene)
	game.hud_controller = hud_controller

	game.world_sync = WorldSyncScript.new()
	game.world_sync.configure(game, game.player, game.bullets, game.asteroids)
	game.gameplay_camera = game.player.get_node_or_null("Camera2D") as Camera2D
	game.camera_follow = CameraFollowScript.new()
	game.camera_follow.configure(game.gameplay_camera)
	game.network_client = NetworkClientScript.new()
	game.add_child(game.network_client)
	effects_owner = Node2D.new()
	add_child(effects_owner)
	game.effects = EffectsScript.new()
	game.effects.configure(effects_owner, null)


func after_each() -> void:
	hud_controller = null
	if hud_scene != null:
		hud_scene.free()
		hud_scene = null
	if effects_owner != null:
		effects_owner.free()
		effects_owner = null
	if game != null:
		game.free()
		game = null


func test_apply_state_missing_lives_keeps_current_hud_lives_and_does_not_crash() -> void:
	hud_controller.set_lives(7)
	var state := WorldStateFixture.state()
	state.erase(Packets.FIELD_LIVES)
	state[Packets.FIELD_ASTEROIDS] = {}
	state[Packets.FIELD_BULLETS] = {}

	game._apply_state(state)

	assert_eq(_lives_label().text, "7 x ")
	assert_true(game.has_received_state)


func test_apply_state_stores_player_lifecycle() -> void:
	var state := WorldStateFixture.state()
	state[Packets.FIELD_PLAYER_LIFECYCLE] = {
		WorldStateFixture.LOCAL_PLAYER_ID: "active",
		WorldStateFixture.REMOTE_PLAYER_ID: "unexpected_status",
	}
	state[Packets.FIELD_ASTEROIDS] = {}
	state[Packets.FIELD_BULLETS] = {}

	game._apply_state(state)

	assert_eq(game._player_lifecycle()[WorldStateFixture.LOCAL_PLAYER_ID], "active")
	assert_eq(game._player_lifecycle()[WorldStateFixture.REMOTE_PLAYER_ID], "unexpected_status")


func test_apply_state_missing_player_lifecycle_uses_empty_dictionary() -> void:
	game._set_player_lifecycle({"stale-player": "active"})
	var state := WorldStateFixture.state()
	state.erase(Packets.FIELD_PLAYER_LIFECYCLE)
	state[Packets.FIELD_ASTEROIDS] = {}
	state[Packets.FIELD_BULLETS] = {}

	game._apply_state(state)

	assert_true(game._player_lifecycle().is_empty())


func test_apply_state_invalid_player_lifecycle_uses_empty_dictionary() -> void:
	game._set_player_lifecycle({"stale-player": "active"})
	var state := WorldStateFixture.state()
	state[Packets.FIELD_PLAYER_LIFECYCLE] = "active"
	state[Packets.FIELD_ASTEROIDS] = {}
	state[Packets.FIELD_BULLETS] = {}

	game._apply_state(state)

	assert_true(game._player_lifecycle().is_empty())


func test_self_death_event_missing_respawn_delay_uses_safe_zero_delay() -> void:
	game.self_id = WorldStateFixture.LOCAL_PLAYER_ID
	hud_controller.set_lives(2)
	var event := {
		Packets.FIELD_TYPE: Packets.TYPE_SHIP_DEATH,
		Packets.FIELD_PLAYER_ID: WorldStateFixture.LOCAL_PLAYER_ID,
		Packets.FIELD_X: 100.0,
		Packets.FIELD_Y: 120.0,
		Packets.FIELD_LIVES: 1,
	}

	game._gameplay_lifecycle_controller().apply_self_death_event(event)

	assert_eq(_lives_label().text, "1 x ")
	assert_true(hud_controller.is_dead)
	assert_eq(hud_controller.respawn_countdown_remaining, 0.0)


func test_open_game_menu_during_gameplay_shows_menu_without_game_over() -> void:
	game._gameplay_menu_controller().open_game_menu(game.has_initial_spawn)

	assert_true(hud_controller.is_game_menu_visible())
	assert_false(hud_controller.game_over_overlay.visible)
	assert_true(_primary_label("Resume").visible)
	assert_false(_primary_label("Lobby").visible)


func test_resume_hides_game_menu_and_clears_paused_state() -> void:
	game._gameplay_menu_controller().is_gameplay_paused = true
	hud_controller.show_game_menu()

	game._gameplay_menu_controller().on_resume_requested()

	assert_false(game._gameplay_menu_controller().is_gameplay_paused)
	assert_false(hud_controller.is_game_menu_visible())


func test_single_player_game_over_shows_game_over_and_disabled_resume_menu() -> void:
	game.session_mode = "SinglePlayer"

	game._gameplay_lifecycle_controller().set_game_over_state()

	assert_true(hud_controller.game_over_overlay.visible)
	assert_true(hud_controller.is_game_menu_visible())
	assert_true(_primary_label("Resume").visible)
	assert_false(_primary_label("Lobby").visible)
	assert_true(hud_controller.get_game_menu().primary_action_button.disabled)
	assert_false(_cycle_view_label().visible)
	assert_eq(_game_menu_nodes().size(), 1)


func test_multiplayer_local_game_over_without_targets_shows_disabled_waiting_until_room_game_over() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"

	game._gameplay_lifecycle_controller().set_game_over_state()

	assert_true(hud_controller.game_over_overlay.visible)
	assert_true(hud_controller.is_game_menu_visible())
	assert_false(_primary_label("Resume").visible)
	assert_false(_primary_label("Lobby").visible)
	assert_false(_primary_label("Spectate").visible)
	assert_true(_primary_label("Waiting").visible)
	assert_true(hud_controller.get_game_menu().primary_action_button.disabled)
	assert_false(_cycle_view_label().visible)
	assert_eq(_game_menu_nodes().size(), 1)


func test_multiplayer_local_game_over_with_targets_shows_spectate() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target()

	game._gameplay_lifecycle_controller().set_game_over_state()

	assert_true(hud_controller.game_over_overlay.visible)
	assert_true(hud_controller.is_game_menu_visible())
	assert_false(_primary_label("Lobby").visible)
	assert_true(_primary_label("Spectate").visible)
	assert_false(_primary_label("Waiting").visible)
	assert_false(hud_controller.get_game_menu().primary_action_button.disabled)


func test_multiplayer_local_game_over_with_pending_target_shows_waiting() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-player", Vector2(42.0, 24.0), "pending_respawn")

	game._gameplay_lifecycle_controller().set_game_over_state()

	assert_false(_primary_label("Spectate").visible)
	assert_true(_primary_label("Waiting").visible)
	assert_true(hud_controller.get_game_menu().primary_action_button.disabled)


func test_multiplayer_room_game_over_enables_lobby_menu() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "GameOver"

	game._gameplay_lifecycle_controller().set_game_over_state()

	assert_true(hud_controller.game_over_overlay.visible)
	assert_true(hud_controller.is_game_menu_visible())
	assert_false(_primary_label("Resume").visible)
	assert_true(_primary_label("Lobby").visible)
	assert_false(hud_controller.get_game_menu().primary_action_button.disabled)
	assert_eq(_game_menu_nodes().size(), 1)


func test_multiplayer_room_game_over_packet_enables_open_lobby_menu() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	game._gameplay_lifecycle_controller().set_game_over_state()
	assert_true(hud_controller.get_game_menu().primary_action_button.disabled)
	assert_true(_primary_label("Waiting").visible)

	game._store_room_state({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: "GameOver",
	})

	assert_false(hud_controller.get_game_menu().primary_action_button.disabled)
	assert_true(_primary_label("Lobby").visible)


func test_multiplayer_local_game_over_allows_open_menu_while_room_in_game() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	game._gameplay_menu_controller().open_menu_input_armed = true
	_add_spectate_target()
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.hide_game_menu()

	assert_false(game._gameplay_menu_controller().should_block_open_menu_for_game_over())
	game._gameplay_menu_controller().open_game_menu(game.has_initial_spawn)
	assert_true(hud_controller.is_game_menu_visible())
	assert_true(_primary_label("Spectate").visible)
	assert_false(hud_controller.get_game_menu().primary_action_button.disabled)


func test_multiplayer_spectate_primary_action_emits_only_spectate_signal() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target()
	game._gameplay_lifecycle_controller().set_game_over_state()
	var menu := hud_controller.get_game_menu()
	var emitted := {
		"lobby": false,
		"spectate": false,
	}
	menu.lobby_requested.connect(func() -> void:
		emitted["lobby"] = true
	)
	menu.spectate_requested.connect(func() -> void:
		emitted["spectate"] = true
	)

	menu._on_primary_action_pressed()

	assert_false(emitted["lobby"])
	assert_true(emitted["spectate"])


func test_multiplayer_spectate_action_hides_menu_and_follows_remote_player() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-player", Vector2(42.0, 24.0))
	game._gameplay_lifecycle_controller().set_game_over_state()

	hud_controller.get_game_menu()._on_primary_action_pressed()

	assert_true(game._spectate_controller().is_active())
	assert_eq(game._spectate_controller().current_target_id(), "remote-player")
	assert_false(hud_controller.is_game_menu_visible())
	assert_true(_cycle_view_label().visible)
	assert_eq(game.gameplay_camera.global_position, Vector2(42.0, 24.0))


func test_multiplayer_open_menu_while_spectating_hides_cycle_view() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-player", Vector2(42.0, 24.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	assert_true(_cycle_view_label().visible)

	game._gameplay_menu_controller().open_game_menu(game.has_initial_spawn)

	assert_true(hud_controller.is_game_menu_visible())
	assert_false(_cycle_view_label().visible)


func test_multiplayer_closing_menu_while_spectating_shows_cycle_view() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-player", Vector2(42.0, 24.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	game._gameplay_menu_controller().open_game_menu(game.has_initial_spawn)
	assert_false(_cycle_view_label().visible)

	game._gameplay_menu_controller().close_game_menu()

	assert_false(hud_controller.is_game_menu_visible())
	assert_true(_cycle_view_label().visible)


func test_multiplayer_spectate_selects_another_target_when_current_disappears() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-a", Vector2(10.0, 20.0))
	_add_spectate_target("remote-b", Vector2(30.0, 40.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	assert_eq(game._spectate_controller().current_target_id(), "remote-a")

	game.world_sync.player_sync.remote_player_visual_positions.erase("remote-a")
	game._update_spectate_camera()

	assert_true(game._spectate_controller().is_active())
	assert_eq(game._spectate_controller().current_target_id(), "remote-b")
	assert_eq(game.gameplay_camera.global_position, Vector2(30.0, 40.0))


func test_switch_camera_does_nothing_before_spectating() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-a", Vector2(10.0, 20.0))
	_add_spectate_target("remote-b", Vector2(30.0, 40.0))

	game._cycle_spectate_target()

	assert_false(game._spectate_controller().is_active())
	assert_eq(game._spectate_controller().current_target_id(), "")
	assert_eq(game.gameplay_camera.global_position, game.player.global_position)


func test_switch_camera_cycles_to_next_spectate_target() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-a", Vector2(10.0, 20.0))
	_add_spectate_target("remote-b", Vector2(30.0, 40.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	assert_eq(game._spectate_controller().current_target_id(), "remote-a")

	game._cycle_spectate_target()

	assert_true(game._spectate_controller().is_active())
	assert_eq(game._spectate_controller().current_target_id(), "remote-b")
	assert_eq(game.gameplay_camera.global_position, Vector2(30.0, 40.0))


func test_switch_camera_skips_eliminated_spectate_target() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-a", Vector2(10.0, 20.0))
	_add_spectate_target("remote-b", Vector2(30.0, 40.0), "eliminated")
	_add_spectate_target("remote-c", Vector2(50.0, 60.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	assert_eq(game._spectate_controller().current_target_id(), "remote-a")

	game._cycle_spectate_target()

	assert_true(game._spectate_controller().is_active())
	assert_eq(game._spectate_controller().current_target_id(), "remote-c")
	assert_eq(game.gameplay_camera.global_position, Vector2(50.0, 60.0))


func test_switch_camera_wraps_to_first_spectate_target() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-a", Vector2(10.0, 20.0))
	_add_spectate_target("remote-b", Vector2(30.0, 40.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	game._cycle_spectate_target()
	assert_eq(game._spectate_controller().current_target_id(), "remote-b")

	game._cycle_spectate_target()

	assert_eq(game._spectate_controller().current_target_id(), "remote-a")
	assert_eq(game.gameplay_camera.global_position, Vector2(10.0, 20.0))


func test_switch_camera_recovers_when_current_target_disappeared() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-a", Vector2(10.0, 20.0))
	_add_spectate_target("remote-b", Vector2(30.0, 40.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	assert_eq(game._spectate_controller().current_target_id(), "remote-a")

	game.world_sync.player_sync.remote_player_visual_positions.erase("remote-a")
	game._cycle_spectate_target()

	assert_true(game._spectate_controller().is_active())
	assert_eq(game._spectate_controller().current_target_id(), "remote-b")
	assert_eq(game.gameplay_camera.global_position, Vector2(30.0, 40.0))


func test_switch_camera_does_nothing_after_room_game_over_exits_spectate() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-a", Vector2(10.0, 20.0))
	_add_spectate_target("remote-b", Vector2(30.0, 40.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	game._store_room_state({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: "GameOver",
	})
	assert_false(game._spectate_controller().is_active())

	game._cycle_spectate_target()

	assert_false(game._spectate_controller().is_active())
	assert_eq(game._spectate_controller().current_target_id(), "")
	assert_true(_primary_label("Lobby").visible)


func test_multiplayer_spectate_falls_back_to_waiting_menu_when_targets_disappear() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-player", Vector2(42.0, 24.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	assert_true(game._spectate_controller().is_active())

	game.world_sync.player_sync.remote_player_visual_positions.clear()
	game._update_spectate_camera()

	assert_false(game._spectate_controller().is_active())
	assert_eq(game._spectate_controller().current_target_id(), "")
	assert_true(hud_controller.is_game_menu_visible())
	assert_true(_primary_label("Waiting").visible)
	assert_true(hud_controller.get_game_menu().primary_action_button.disabled)
	assert_false(_cycle_view_label().visible)


func test_multiplayer_room_game_over_exits_spectate_and_shows_lobby() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "InGame"
	_add_spectate_target("remote-player", Vector2(42.0, 24.0))
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.get_game_menu()._on_primary_action_pressed()
	assert_true(game._spectate_controller().is_active())

	game._store_room_state({
		Packets.FIELD_TYPE: Packets.TYPE_ROOM_SNAPSHOT,
		Packets.FIELD_ROOM_STATE: "GameOver",
	})

	assert_false(game._spectate_controller().is_active())
	assert_true(hud_controller.is_game_menu_visible())
	assert_true(_primary_label("Lobby").visible)
	assert_false(hud_controller.get_game_menu().primary_action_button.disabled)
	assert_false(_cycle_view_label().visible)


func test_multiplayer_room_game_over_primary_action_still_emits_lobby_signal() -> void:
	game.session_mode = "Multiplayer"
	game.current_room_state = "GameOver"
	game._gameplay_lifecycle_controller().set_game_over_state()
	var menu := hud_controller.get_game_menu()
	var emitted := {
		"lobby": false,
		"spectate": false,
	}
	menu.lobby_requested.connect(func() -> void:
		emitted["lobby"] = true
	)
	menu.spectate_requested.connect(func() -> void:
		emitted["spectate"] = true
	)

	menu._on_primary_action_pressed()

	assert_true(emitted["lobby"])
	assert_false(emitted["spectate"])


func test_single_player_game_over_still_blocks_open_menu_toggle() -> void:
	game.session_mode = "SinglePlayer"
	game._gameplay_menu_controller().open_menu_input_armed = true
	game._gameplay_lifecycle_controller().set_game_over_state()
	hud_controller.hide_game_menu()

	assert_true(game._gameplay_menu_controller().should_block_open_menu_for_game_over())
	assert_false(hud_controller.is_game_menu_visible())


func _lives_label() -> Label:
	return hud_scene.find_child("LivesCount", true, false) as Label


func _primary_label(label_name: String) -> Label:
	return hud_controller.get_game_menu().primary_action_button.find_child(label_name, true, false) as Label


func _cycle_view_label() -> Label:
	return hud_scene.find_child("CycleView", true, false) as Label


func _game_menu_nodes() -> Array[Node]:
	return hud_scene.find_children("GameMenu", "", true, false)


func _add_spectate_target(
	player_id := "remote-player",
	position := Vector2(42.0, 24.0),
	lifecycle_status := "active"
) -> void:
	game.world_sync.player_sync.remote_player_visual_positions[player_id] = position
	game.set_player_lifecycle_status(player_id, lifecycle_status)
