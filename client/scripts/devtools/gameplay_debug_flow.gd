extends RefCounted
class_name GameplayDebugFlow

const Packets = preload("res://scripts/networking/packets/packets.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var connection_service
var debug_invincible_enabled := false
var debug_invincible_toggle_was_pressed := false


func configure(connection_service_ref) -> void:
	connection_service = connection_service_ref


func reset() -> void:
	debug_invincible_enabled = false
	debug_invincible_toggle_was_pressed = false


func process(has_received_state: bool) -> void:
	var toggle_pressed := Input.is_key_pressed(KEY_1)
	if !has_received_state || connection_service == null:
		debug_invincible_toggle_was_pressed = toggle_pressed
		return
	if !toggle_pressed:
		debug_invincible_toggle_was_pressed = false
		return
	if debug_invincible_toggle_was_pressed:
		return

	debug_invincible_toggle_was_pressed = true
	debug_invincible_enabled = !debug_invincible_enabled
	connection_service.send_packet(Packets.toggle_debug_invincible_packet())
	ClientLogger.game_debug("Debug invincibility toggled: %s" % debug_invincible_enabled)
