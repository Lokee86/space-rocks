extends RefCounted
class_name GameplayDebugFlow

const Packets = preload("res://scripts/networking/packets/packets.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")
const DevtoolsTargetResolverScript = preload("res://scripts/devtools/devtools_target_resolver.gd")

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


func toggle_invincible(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = ""
) -> void:
	if connection_service == null:
		return
	debug_invincible_enabled = !debug_invincible_enabled
	connection_service.send_packet(_build_player_toggle_packet(Packets.TYPE_TOGGLE_DEBUG_INVINCIBLE, target_scope, target_player_id))
	ClientLogger.game_info("Devtools invincibility toggle sent")


func toggle_infinite_lives(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = ""
) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(_build_player_toggle_packet(Packets.TYPE_TOGGLE_DEBUG_INFINITE_LIVES, target_scope, target_player_id))
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


func toggle_freeze_player(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = ""
) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(_build_player_toggle_packet(Packets.TYPE_TOGGLE_DEBUG_FREEZE_PLAYER, target_scope, target_player_id))
	ClientLogger.game_info("Devtools player freeze toggle sent")


func set_score(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = "",
	score: int = 0
) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(_build_counter_packet(Packets.TYPE_DEBUG_SET_SCORE, target_scope, target_player_id, Packets.FIELD_SCORE, score))
	ClientLogger.game_info("Devtools set score sent")


func add_score(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = "",
	amount: int = 0
) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(_build_counter_packet(Packets.TYPE_DEBUG_ADD_SCORE, target_scope, target_player_id, Packets.FIELD_AMOUNT, amount))
	ClientLogger.game_info("Devtools add score sent")


func set_lives(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = "",
	lives: int = 0
) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(_build_counter_packet(Packets.TYPE_DEBUG_SET_LIVES, target_scope, target_player_id, Packets.FIELD_LIVES, lives))
	ClientLogger.game_info("Devtools set lives sent")


func add_lives(
	target_scope: String = DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER,
	target_player_id: String = "",
	amount: int = 0
) -> void:
	if connection_service == null:
		return
	connection_service.send_packet(_build_counter_packet(Packets.TYPE_DEBUG_ADD_LIVES, target_scope, target_player_id, Packets.FIELD_AMOUNT, amount))
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


func _build_player_toggle_packet(packet_type: String, target_scope: String, target_player_id: String) -> Dictionary:
	var packet := {
		Packets.FIELD_TYPE: packet_type,
		Packets.FIELD_TARGET_SCOPE: target_scope,
	}
	if target_scope == DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER and target_player_id != "":
		packet[Packets.FIELD_TARGET_PLAYER_ID] = target_player_id
	return packet


func _build_counter_packet(
	packet_type: String,
	target_scope: String,
	target_player_id: String,
	value_field: String,
	value: int
) -> Dictionary:
	var packet := {
		Packets.FIELD_TYPE: packet_type,
		Packets.FIELD_TARGET_SCOPE: target_scope,
		value_field: value,
	}
	if target_scope == DevtoolsTargetResolverScript.TARGET_SCOPE_SINGLE_PLAYER and target_player_id != "":
		packet[Packets.FIELD_TARGET_PLAYER_ID] = target_player_id
	return packet
