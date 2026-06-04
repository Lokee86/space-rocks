extends RefCounted
class_name PacketEncodeResult

var ok: bool = false
var wire_message: String = ""
var error: String = ""


static func success(message: String):
	var result := PacketEncodeResult.new()
	result.ok = true
	result.wire_message = message
	return result


static func failure(message: String):
	var result := PacketEncodeResult.new()
	result.error = message
	return result
