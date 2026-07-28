extends RefCounted
class_name GameplaySpectateFlow

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")

var menu_flow
var spectate_menu_state
var world_sync
var connection_service
var is_spectating := false
var target_player_id := ""


func configure(menu_flow_ref, spectate_menu_state_ref, world_sync_ref, connection_service_ref = null) -> void:
	menu_flow = menu_flow_ref
	spectate_menu_state = spectate_menu_state_ref
	world_sync = world_sync_ref
	connection_service = connection_service_ref


func reset() -> void:
	if connection_service != null and connection_service.has_method("send_packet") and not target_player_id.is_empty():
		connection_service.send_packet(Packets.clear_view_target_request_packet())
	is_spectating = false
	target_player_id = ""
	if world_sync != null and world_sync.has_method("clear_view_target_player"):
		world_sync.clear_view_target_player()


func process() -> void:
	if not is_spectating or target_player_id.is_empty() or world_sync == null:
		return
	if world_sync.has_method("set_view_target_player"):
		world_sync.set_view_target_player(target_player_id)
	world_sync.focus_camera_on_player(target_player_id)


func request_open_spectate_menu() -> void:
	if !is_spectating || menu_flow == null:
		return
	menu_flow.show_spectating_menu()


func request_cycle_target() -> void:
	if !is_spectating || spectate_menu_state == null || world_sync == null:
		return
	_set_target(spectate_menu_state.cycle_next_target())


func begin_spectating() -> void:
	if spectate_menu_state == null || world_sync == null:
		return
	_set_target(spectate_menu_state.begin_spectating())


func request_target(next_target_player_id: String) -> void:
	_set_target(next_target_player_id)


func _set_target(next_target_player_id: String) -> void:
	if next_target_player_id.is_empty():
		return
	target_player_id = next_target_player_id
	is_spectating = true
	if world_sync.has_method("set_view_target_player"):
		world_sync.set_view_target_player(target_player_id)
	world_sync.focus_camera_on_player(target_player_id)
	if connection_service != null and connection_service.has_method("send_packet"):
		connection_service.send_packet(Packets.set_view_target_request_packet(target_player_id))
