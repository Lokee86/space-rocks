extends RefCounted


static func read(packet: Dictionary) -> Dictionary:
	var shapes_value: Variant = packet.get("shapes", {})
	if shapes_value is Dictionary:
		return {
			"shapes": shapes_value,
		}
	return {
		"shapes": {},
	}
