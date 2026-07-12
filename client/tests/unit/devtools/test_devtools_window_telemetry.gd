extends GutTest

const DevtoolsWindowTelemetry := preload("res://scripts/devtools/devtools_window_telemetry.gd")


func _configured_telemetry() -> Array:
	var local_select := OptionButton.new()
	var target_select := OptionButton.new()
	var local_text := Label.new()
	var target_text := Label.new()
	var telemetry := DevtoolsWindowTelemetry.new()
	telemetry.configure(local_select, target_select, local_text, target_text)
	return [telemetry, local_select, target_select, local_text, target_text]


func test_configure_initializes_sources_and_defaults_to_players() -> void:
	var controls := _configured_telemetry()
	var telemetry: DevtoolsWindowTelemetry = controls[0]
	var local_select: OptionButton = controls[1]
	var target_select: OptionButton = controls[2]

	assert_eq(local_select.get_item_count(), 2)
	assert_eq(target_select.get_item_count(), 2)
	assert_eq(telemetry.local_source(), "players")
	assert_eq(telemetry.target_source(), "players")
	assert_eq(str(local_select.get_item_text(1)), "player_world_states")
	assert_eq(str(target_select.get_item_metadata(1)), "player_world_states")


func test_source_lookup_and_selection_use_metadata() -> void:
	var controls := _configured_telemetry()
	var telemetry: DevtoolsWindowTelemetry = controls[0]

	telemetry.set_sources("player_world_states", "missing")

	assert_eq(telemetry.local_source(), "player_world_states")
	assert_eq(telemetry.target_source(), "players")


func test_local_empty_state_renders_placeholder() -> void:
	var controls := _configured_telemetry()
	var telemetry: DevtoolsWindowTelemetry = controls[0]
	var local_text: Label = controls[3]

	telemetry.refresh_local_state({})

	assert_eq(local_text.text, "—")


func test_local_state_sorts_keys_and_formats_json_and_floats() -> void:
	var controls := _configured_telemetry()
	var telemetry: DevtoolsWindowTelemetry = controls[0]
	var local_text: Label = controls[3]

	telemetry.refresh_local_state({
		"zeta": 1.234567,
		"alpha": {"nested": [1, 2]},
	})

	assert_eq(local_text.text, "alpha: {\"nested\":[1,2]}\nzeta: 1.2346")


func test_target_missing_identity_renders_placeholder() -> void:
	var controls := _configured_telemetry()
	var telemetry: DevtoolsWindowTelemetry = controls[0]
	var target_text: Label = controls[4]

	telemetry.refresh_target_state("", "", {"score": 7})

	assert_eq(target_text.text, "—")


func test_target_empty_state_renders_identity_and_state_placeholder() -> void:
	var controls := _configured_telemetry()
	var telemetry: DevtoolsWindowTelemetry = controls[0]
	var target_text: Label = controls[4]

	telemetry.refresh_target_state("player", "Player-2", {})

	assert_eq(target_text.text, "target_kind: player\ntarget_id: Player-2\nstate: —")


func test_target_state_sorts_keys_and_formats_values() -> void:
	var controls := _configured_telemetry()
	var telemetry: DevtoolsWindowTelemetry = controls[0]
	var target_text: Label = controls[4]

	telemetry.refresh_target_state("asteroid", "asteroid-1", {
		"velocity": [1, 2],
		"angle": 2.0 / 3.0,
		"active": true,
	})

	assert_eq(
		target_text.text,
		"target_kind: asteroid\ntarget_id: asteroid-1\n\nactive: true\nangle: 0.6667\nvelocity: [1,2]"
	)