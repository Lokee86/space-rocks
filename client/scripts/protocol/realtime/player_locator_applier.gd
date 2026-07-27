extends RefCounted
class_name PlayerLocatorApplier

const PlayerLocatorStateScript := preload("res://scripts/protocol/realtime/player_locator_state.gd")

func apply_player_locator(player_locator_state: PlayerLocatorStateScript, packet: Dictionary) -> bool:
	var sequence := int(packet.get("sequence", 0))
	if sequence <= player_locator_state.sequence:
		return false
	var next_player_locators: Dictionary = {}
	var packet_player_locators = packet.get("player_locators", [])
	if packet_player_locators is Array:
		for locator in packet_player_locators:
			if not (locator is Dictionary):
				continue
			var player_id = locator.get("id")
			if player_id == null or str(player_id).is_empty():
				continue
			next_player_locators[player_id] = locator.duplicate(true)
	return player_locator_state.replace_player_locators(
		next_player_locators,
		sequence,
		int(packet.get("server_sent_msec", 0))
	)
