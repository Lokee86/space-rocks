extends RefCounted


static func read(packet: Dictionary) -> Dictionary:
	var debug_status = packet.get("debug_status", {})
	if typeof(debug_status) != TYPE_DICTIONARY:
		debug_status = {}

	var debug_statuses = packet.get("debug_statuses", {})
	if typeof(debug_statuses) != TYPE_DICTIONARY:
		debug_statuses = {}

	return {
		"debug_status": debug_status,
		"debug_statuses": debug_statuses,
	}
