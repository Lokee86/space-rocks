extends RefCounted
class_name PacketDecodeResult

var ok: bool = false
var packet: Dictionary = {}
var error_code: String = ""
var error: String = ""
var raw: String = ""


static func success(decoded_packet: Dictionary):
	var result := PacketDecodeResult.new()
	result.ok = true
	result.packet = decoded_packet
	return result


static func failure(code: String, message: String, raw_message: String = ""):
	var result := PacketDecodeResult.new()
	result.error_code = code
	result.error = message
	result.raw = raw_message
	return result
