extends RefCounted
class_name PacketDecodeResult

var ok: bool = false
var packet: Dictionary = {}
var error: String = ""
var raw: String = ""


static func success(decoded_packet: Dictionary):
	var result := PacketDecodeResult.new()
	result.ok = true
	result.packet = decoded_packet
	return result


static func failure(message: String, raw_message: String = ""):
	var result := PacketDecodeResult.new()
	result.error = message
	result.raw = raw_message
	return result
