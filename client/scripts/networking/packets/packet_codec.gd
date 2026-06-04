extends RefCounted

const PacketDecodeResult = preload("res://scripts/networking/packets/packet_decode_result.gd")
const PacketEncodeResult = preload("res://scripts/networking/packets/packet_encode_result.gd")

# PacketCodec owns wire parsing and envelope checks only; packet readers validate payload details.
static func encode(packet: Dictionary) -> PacketEncodeResult:
	return PacketEncodeResult.success(JSON.stringify(packet))


static func decode(text: String) -> PacketDecodeResult:
	var parser := JSON.new()
	var parse_error := parser.parse(text)
	if parse_error != OK:
		return PacketDecodeResult.failure("Invalid packet JSON: %s" % parser.get_error_message(), text)

	var decoded = parser.data
	if typeof(decoded) != TYPE_DICTIONARY:
		return PacketDecodeResult.failure("Packet JSON must decode to a Dictionary", text)

	var packet: Dictionary = decoded
	if !packet.has("type"):
		return PacketDecodeResult.failure("Packet envelope is missing required 'type' field", text)
	if typeof(packet["type"]) != TYPE_STRING:
		return PacketDecodeResult.failure("Packet envelope field 'type' must be a String", text)
	if str(packet["type"]).strip_edges().is_empty():
		return PacketDecodeResult.failure("Packet envelope field 'type' must not be empty", text)
	if packet.has("payload") && typeof(packet["payload"]) != TYPE_DICTIONARY:
		return PacketDecodeResult.failure("Packet envelope field 'payload' must be a Dictionary when present", text)

	return PacketDecodeResult.success(packet)
