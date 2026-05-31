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
	var toggle_pressed := Input.is_action_just_pressed("DevToggle1")
	var infinite_lives_toggle_pressed := Input.is_action_just_pressed("DevToggle2")
	var world_freeze_toggle_pressed := Input.is_action_just_pressed("DevToggle3")
	var player_freeze_toggle_pressed := Input.is_action_just_pressed("DevToggle4")
	if !has_received_state || connection_service == null:
		debug_invincible_toggle_was_pressed = toggle_pressed
		return

	if toggle_pressed:
		if !debug_invincible_toggle_was_pressed:
			debug_invincible_toggle_was_pressed = true
			toggle_invincible()
	else:
		debug_invincible_toggle_was_pressed = false

	if infinite_lives_toggle_pressed:
		toggle_infinite_lives()

	if world_freeze_toggle_pressed:
		toggle_freeze_world()

	if player_freeze_toggle_pressed:
		toggle_freeze_player()


func toggle_invincible(target_player_id := "") -> void:
	if connection_service == null:
		return
	debug_invincible_enabled = !debug_invincible_enabled
	if target_player_id == "":
		connection_service.send_packet(Packets.toggle_debug_invincible_packet())
	else:
		connection_service.send_packet(Packets.toggle_debug_invincible_target_player_packet(target_player_id))
	ClientLogger.game_info("Devtools invincibility toggle sent")


func toggle_infinite_lives(target_player_id := "") -> void:
	if connection_service == null:
		return
	if target_player_id == "":
		connection_service.send_packet(Packets.toggle_debug_infinite_lives_packet())
	else:
		connection_service.send_packet(Packets.toggle_debug_infinite_lives_target_player_packet(target_player_id))
	ClientLogger.game_info("Devtools infinite lives toggle sent")


func toggle_freeze_world(freeze_target := "") -> void:
	if connection_service == null:
		return
	if freeze_target == "" || freeze_target == "all":
		connection_service.send_packet(Packets.toggle_debug_freeze_world_packet())
		ClientLogger.game_info("Devtools world freeze toggle sent")
	else:
		connection_service.send_packet(Packets.toggle_debug_freeze_world_target_packet(freeze_target))
		ClientLogger.game_info("Devtools world freeze toggle sent (freeze_target='%s')" % freeze_target)


func toggle_freeze_player(target_player_id := "") -> void:
	if connection_service == null:
		return
	if target_player_id == "":
		connection_service.send_packet(Packets.toggle_debug_freeze_player_packet())
	else:
		connection_service.send_packet(Packets.toggle_debug_freeze_player_target_player_packet(target_player_id))
	ClientLogger.game_info("Devtools player freeze toggle sent")


func set_score(target_player_id: String, score: int) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(Packets.debug_set_score_packet(target_player_id, score))
	ClientLogger.game_info("Devtools set score sent")


func add_score(target_player_id: String, amount: int) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(Packets.debug_add_score_packet(target_player_id, amount))
	ClientLogger.game_info("Devtools add score sent")


func set_lives(target_player_id: String, lives: int) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(Packets.debug_set_lives_packet(target_player_id, lives))
	ClientLogger.game_info("Devtools set lives sent")


func add_lives(target_player_id: String, amount: int) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(Packets.debug_add_lives_packet(target_player_id, amount))
	ClientLogger.game_info("Devtools add lives sent")


func clear_bullets() -> void:
	if connection_service == null:
		return
	connection_service.send_packet(Packets.debug_clear_bullets_packet())
	ClientLogger.game_info("Devtools clear bullets sent")


func clear_asteroids() -> void:
	if connection_service == null:
		return
	connection_service.send_packet(Packets.debug_clear_asteroids_packet())
	ClientLogger.game_info("Devtools clear asteroids sent")
