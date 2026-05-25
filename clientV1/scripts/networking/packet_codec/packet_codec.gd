extends RefCounted


static func encode(packet: Dictionary) -> String:
	return JSON.stringify(packet)


static func decode(text: String):
	return JSON.parse_string(text)
