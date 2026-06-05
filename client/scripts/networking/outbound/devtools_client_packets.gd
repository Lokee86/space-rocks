extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func debug_kill_player_packet(target_scope: String = "", target_player_id: String = "") -> Dictionary:
	var packet := Packets.debug_kill_player_packet()
	if target_scope != "":
		packet[Packets.FIELD_TARGET_SCOPE] = target_scope
	if target_player_id != "":
		packet[Packets.FIELD_TARGET_PLAYER_ID] = target_player_id
	return packet


static func debug_kill_target_player_packet(target_player_id: String, target_scope: String = "") -> Dictionary:
	var packet := Packets.debug_kill_target_player_packet(target_player_id)
	if target_scope != "":
		packet[Packets.FIELD_TARGET_SCOPE] = target_scope
	return packet


static func set_target_player_request_packet(target_player_id: String) -> Dictionary:
	return Packets.set_target_player_request_packet("player", target_player_id)


static func clear_target_request_packet() -> Dictionary:
	return Packets.clear_target_request_packet()


static func toggle_debug_invincible_packet() -> Dictionary:
	return Packets.toggle_debug_invincible_packet()


static func toggle_debug_invincible_target_player_packet(target_player_id: String) -> Dictionary:
	return Packets.toggle_debug_invincible_target_player_packet(target_player_id)


static func toggle_debug_infinite_lives_packet() -> Dictionary:
	return Packets.toggle_debug_infinite_lives_packet()


static func toggle_debug_infinite_lives_target_player_packet(target_player_id: String) -> Dictionary:
	return Packets.toggle_debug_infinite_lives_target_player_packet(target_player_id)


static func toggle_debug_freeze_world_packet() -> Dictionary:
	return Packets.toggle_debug_freeze_world_packet()


static func toggle_debug_freeze_world_target_packet(freeze_target) -> Dictionary:
	return Packets.toggle_debug_freeze_world_target_packet(freeze_target)


static func toggle_debug_freeze_player_packet() -> Dictionary:
	return Packets.toggle_debug_freeze_player_packet()


static func toggle_debug_freeze_player_target_player_packet(target_player_id: String) -> Dictionary:
	return Packets.toggle_debug_freeze_player_target_player_packet(target_player_id)


static func debug_set_score_packet(target_player_id: String, score) -> Dictionary:
	return Packets.debug_set_score_packet(target_player_id, score)


static func debug_add_score_packet(target_player_id: String, amount) -> Dictionary:
	return Packets.debug_add_score_packet(target_player_id, amount)


static func debug_set_lives_packet(target_player_id: String, lives) -> Dictionary:
	return Packets.debug_set_lives_packet(target_player_id, lives)


static func debug_add_lives_packet(target_player_id: String, amount) -> Dictionary:
	return Packets.debug_add_lives_packet(target_player_id, amount)


static func debug_clear_bullets_packet() -> Dictionary:
	return Packets.debug_clear_bullets_packet()


static func debug_clear_asteroids_packet() -> Dictionary:
	return Packets.debug_clear_asteroids_packet()
