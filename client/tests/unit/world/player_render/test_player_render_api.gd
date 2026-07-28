extends GutTest

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const PlayerRenderApiScript := preload("res://scripts/world/player_render/player_render_api.gd")
const PlayerScene := preload("res://scenes/player.tscn")

var game_owner: Node2D
var player: Player
var view_anchor: Node2D
var api


func before_each() -> void:
	game_owner = Node2D.new()
	add_child(game_owner)

	player = PlayerScene.instantiate()
	view_anchor = Node2D.new()
	view_anchor.add_child(Camera2D.new())

	game_owner.add_child(player)
	game_owner.add_child(view_anchor)

	api = PlayerRenderApiScript.new()
	api.configure(game_owner, player, view_anchor)


func after_each() -> void:
	api = null
	if game_owner != null:
		game_owner.free()
		game_owner = null


func test_apply_state_uses_self_position_when_no_view_target_is_set() -> void:
	var self_id := "player-1"
	var server_players := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}

	api.apply_state(self_id, server_players)

	assert_eq(api.server_position(), Vector2(100.0, 200.0))
	api.interpolate(1.0, self_id)
	assert_eq(view_anchor.global_position, api.visual_position())


func test_apply_state_uses_view_target_position_when_view_target_is_set() -> void:
	var self_id := "player-1"
	var target_id := "player-2"
	var server_players := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		},
		target_id: {
			Packets.FIELD_X: 400.0,
			Packets.FIELD_Y: 500.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}

	api.set_view_target_player(target_id)
	api.apply_state(self_id, server_players)

	assert_eq(api.server_position(), Vector2(400.0, 500.0))


func test_apply_state_updates_remote_players_from_last_anchor_when_self_is_missing() -> void:
	var self_id := "player-1"
	var remote_id := "player-2"
	api.apply_state(self_id, {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		},
		remote_id: {
			Packets.FIELD_X: 300.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	})
	api.interpolate(1.0, self_id)

	api.remove_missing({
		remote_id: {
			Packets.FIELD_X: 500.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}, self_id)
	api.apply_state(self_id, {
		remote_id: {
			Packets.FIELD_X: 500.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	})
	api.interpolate(1.0, self_id)

	var remote_nodes: Dictionary = api.remote_player_nodes(self_id)
	assert_true(remote_nodes.has(remote_id))
	assert_eq(remote_nodes[remote_id].position, Vector2(500.0, 200.0))
	assert_eq(api.server_position(), Vector2(100.0, 200.0))


func test_remove_missing_preserves_remote_hue_for_locator_indicators() -> void:
	var self_id := "player-1"
	var remote_id := "player-2"
	var self_state := {
		Packets.FIELD_X: 100.0,
		Packets.FIELD_Y: 200.0,
		Packets.FIELD_ROTATION: 0.0,
	}
	api.apply_state(self_id, {
		self_id: self_state,
		remote_id: {
			Packets.FIELD_X: 300.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
			Packets.FIELD_TEAM_ID: "team_2",
		}
	})
	var hue_before_interest_loss: float = api.get_remote_player_hues(self_id)[remote_id]

	api.remove_missing({self_id: self_state}, self_id)

	var hues_after_interest_loss: Dictionary = api.get_remote_player_hues(self_id)
	assert_true(hues_after_interest_loss.has(remote_id))
	assert_eq(hues_after_interest_loss[remote_id], hue_before_interest_loss)


func test_focus_camera_on_player_selects_existing_remote_render_anchor() -> void:
	var self_id := "player-1"
	var target_id := "player-2"
	var server_players := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		},
		target_id: {
			Packets.FIELD_X: 400.0,
			Packets.FIELD_Y: 500.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}

	api.apply_state(self_id, server_players)

	assert_true(api.focus_camera_on_player(target_id))
	api.apply_state(self_id, server_players)

	assert_eq(api.server_position(), Vector2(400.0, 500.0))


func test_focus_camera_on_player_rejects_missing_player_node() -> void:
	assert_false(api.focus_camera_on_player("missing-player"))


func test_apply_state_falls_back_to_self_position_when_view_target_missing() -> void:
	var self_id := "player-1"
	var server_players := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}

	api.set_view_target_player("missing-player")
	api.apply_state(self_id, server_players)

	assert_eq(api.server_position(), Vector2(100.0, 200.0))


func test_apply_state_sets_view_anchor_target_without_snapping_after_initial_position() -> void:
	var self_id := "player-1"
	var first_state := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}
	var second_state := {
		self_id: {
			Packets.FIELD_X: 300.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}

	api.apply_state(self_id, first_state)
	api.interpolate(1.0, self_id)

	assert_eq(view_anchor.global_position, Vector2(100.0, 200.0))

	api.apply_state(self_id, second_state)

	assert_ne(view_anchor.global_position, Vector2(300.0, 200.0))
	assert_eq(api.server_position(), Vector2(300.0, 200.0))

	api.interpolate(1.0, self_id)

	assert_eq(view_anchor.global_position, Vector2(300.0, 200.0))


func test_view_anchor_rotation_smooths_through_interpolate() -> void:
	var self_id := "player-1"
	var first_state := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 0.0,
		}
	}
	var second_state := {
		self_id: {
			Packets.FIELD_X: 100.0,
			Packets.FIELD_Y: 200.0,
			Packets.FIELD_ROTATION: 1.0,
		}
	}

	api.apply_state(self_id, first_state)
	api.interpolate(1.0, self_id)

	api.apply_state(self_id, second_state)

	assert_ne(view_anchor.rotation, 1.0)

	api.interpolate(1.0, self_id)

	assert_eq(view_anchor.rotation, 1.0)

