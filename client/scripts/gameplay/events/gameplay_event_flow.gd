extends RefCounted
class_name GameplayEventFlow

signal self_death_event(event: Dictionary)

const EffectsScript = preload("res://scripts/gameplay/effects/gameplay_effects.gd")
const GameplayEventController = preload("res://scripts/gameplay/events/gameplay_event_controller.gd")

var effects
var gameplay_event_controller


func configure(
	game_owner: Node2D,
	hud: Control,
	visual_position_for_server_position: Callable
) -> void:
	effects = EffectsScript.new()
	effects.configure(game_owner, hud)
	gameplay_event_controller = GameplayEventController.new()
	gameplay_event_controller.configure(effects, visual_position_for_server_position)


func apply_server_events(server_events: Array, self_id: String) -> void:
	gameplay_event_controller.apply_server_events(
		server_events,
		self_id,
		Callable(self, "_on_self_death_event")
	)


func play_game_over_sound_after_delay() -> void:
	if effects != null:
		effects.play_game_over_sound_after_delay()


func reset() -> void:
	if effects != null:
		effects.reset_game_over_sound()


func _on_self_death_event(event: Dictionary) -> void:
	self_death_event.emit(event)
