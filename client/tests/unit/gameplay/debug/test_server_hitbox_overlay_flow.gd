extends GutTest

const ServerHitboxOverlayFlow := preload("res://scripts/gameplay/debug/server_hitbox_overlay_flow.gd")


class FakeOverlay:
	var enabled := true
	var last_entries: Array = []

	func is_enabled() -> bool:
		return enabled

	func set_hitbox_entries(entries: Array) -> void:
		last_entries = entries


class FakeWorldSync:
	func visual_position_for_server_position(value: Vector2) -> Vector2:
		return value


func test_process_draws_player_hitbox_from_catalog_and_gameplay_state() -> void:
	var flow := ServerHitboxOverlayFlow.new()
	var overlay := FakeOverlay.new()
	var world_sync := FakeWorldSync.new()
	flow.configure(null, world_sync)
	flow.overlay = overlay
	flow.shape_catalog_store.apply_catalog_state({
		"shapes": {
			"player:v_wing": {
				"id": "player:v_wing",
				"kind": "player",
				"shape_type": "polygon",
				"points": [
					{"x": -1.0, "y": 0.0},
					{"x": 1.0, "y": 0.0},
					{"x": 0.0, "y": 1.0},
				],
			}
		}
	})
	flow.apply_gameplay_state({
		"server_players": {
			"player-1": {
				"ship_type": "v_wing",
				"x": 10.0,
				"y": 20.0,
				"rotation": 0.0,
			}
		}
	})

	flow.process()

	assert_eq(overlay.last_entries.size(), 1)
	assert_eq(overlay.last_entries[0]["kind"], "player")
	assert_eq(overlay.last_entries[0]["id"], "player:v_wing")
	assert_true(overlay.last_entries[0]["points"] is PackedVector2Array)
	assert_false((overlay.last_entries[0]["points"] as PackedVector2Array).is_empty())


func test_process_uses_no_entries_when_catalog_missing() -> void:
	var flow := ServerHitboxOverlayFlow.new()
	var overlay := FakeOverlay.new()
	var world_sync := FakeWorldSync.new()
	flow.configure(null, world_sync)
	flow.overlay = overlay
	flow.apply_gameplay_state({
		"server_players": {
			"player-1": {
				"ship_type": "v_wing",
				"x": 10.0,
				"y": 20.0,
				"rotation": 0.0,
			}
		}
	})

	flow.process()

	assert_eq(overlay.last_entries.size(), 0)
