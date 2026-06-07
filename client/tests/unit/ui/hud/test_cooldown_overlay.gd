extends GutTest

const CooldownOverlayScript = preload("res://scripts/ui/hud/cooldown_overlay.gd")

var _overlay: Control
var _cooldown_finished_count := 0


func before_each() -> void:
	_overlay = Control.new()
	_overlay.set_script(CooldownOverlayScript)

	var cooldown_label := Label.new()
	cooldown_label.name = "CooldownLabel"
	_overlay.add_child(cooldown_label)

	add_child_autofree(_overlay)
	_overlay._ready()
	_overlay.cooldown_finished.connect(_on_cooldown_finished)


func test_apply_cooldown_with_remaining_time_makes_overlay_visible() -> void:
	_overlay.apply_cooldown(5.0, 15.0)

	assert_true(_overlay.visible)


func test_apply_cooldown_formats_label_with_one_decimal_place() -> void:
	_overlay.apply_cooldown(5.25, 15.0)

	assert_eq(_cooldown_label().text, "5.3")


func test_apply_cooldown_label_is_not_integer_only_string() -> void:
	_overlay.apply_cooldown(5.25, 15.0)

	assert_ne(_cooldown_label().text, "5")


func test_apply_cooldown_with_zero_remaining_hides_overlay() -> void:
	_overlay.apply_cooldown(0.0, 15.0)

	assert_false(_overlay.visible)


func test_apply_cooldown_with_zero_total_hides_overlay() -> void:
	_overlay.apply_cooldown(5.0, 0.0)

	assert_false(_overlay.visible)


func test_clear_countdown_hides_overlay_and_clears_label_text() -> void:
	_overlay.apply_cooldown(5.0, 15.0)

	_overlay.clear_countdown()

	assert_false(_overlay.visible)
	assert_eq(_cooldown_label().text, "")


func test_cooldown_finished_is_emitted_when_process_naturally_reaches_zero() -> void:
	_overlay.apply_cooldown(0.1, 15.0)

	_overlay._process(0.1)

	assert_eq(_cooldown_finished_count, 1)


func test_cooldown_finished_is_not_emitted_by_direct_clear_countdown() -> void:
	_overlay.apply_cooldown(5.0, 15.0)

	_overlay.clear_countdown()

	assert_eq(_cooldown_finished_count, 0)


func _cooldown_label() -> Label:
	return _overlay.get_node("CooldownLabel") as Label


func _on_cooldown_finished() -> void:
	_cooldown_finished_count += 1
