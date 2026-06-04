extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")

var connection_service = null
var target_candidate_provider: Callable = Callable()
var mouse_visual_position_provider: Callable = Callable()
var server_position_converter: Callable = Callable()

func configure(
	next_connection_service,
	next_target_candidate_provider: Callable,
	next_mouse_visual_position_provider: Callable,
	next_server_position_converter: Callable
) -> void:
	connection_service = next_connection_service
	target_candidate_provider = next_target_candidate_provider
	mouse_visual_position_provider = next_mouse_visual_position_provider
	server_position_converter = next_server_position_converter

func select_target() -> bool:
	if connection_service == null:
		return false
	if target_candidate_provider.is_null():
		return false
	if mouse_visual_position_provider.is_null():
		return false
	if server_position_converter.is_null():
		return false

	var candidates = target_candidate_provider.call()
	if candidates == null:
		return false
	if candidates is Array and candidates.is_empty():
		return false

	var mouse_visual_position = mouse_visual_position_provider.call()
	var selected_candidate = TargetVisualPicker.pick(candidates, mouse_visual_position)
	if selected_candidate == null:
		return false

	var server_position = server_position_converter.call(mouse_visual_position)
	connection_service.send_packet(
		Packets.select_target_at_position_request_packet(
			server_position.x,
			server_position.y,
			selected_candidate.target_kind,
			selected_candidate.target_id
		)
	)

	return true

func deselect_target() -> void:
	if connection_service == null:
		return

	connection_service.send_packet(Packets.clear_target_request_packet())

