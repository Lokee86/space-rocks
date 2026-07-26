extends RefCounted
class_name GameplayInputFlow

const INPUT_HEARTBEAT_MSEC := 250

var connection_service
var player
var menu_flow
var _clock: Callable
var _last_sent_input := {}
var _last_sent_msec := -1


func _init(clock: Callable = Callable(Time, "get_ticks_msec")) -> void:
	_clock = clock


func configure(connection_service_ref, player_ref, menu_flow_ref) -> void:
	connection_service = connection_service_ref
	player = player_ref
	menu_flow = menu_flow_ref


func reset() -> void:
	_last_sent_input.clear()
	_last_sent_msec = -1


func process(required_lane_baselines_synced: bool) -> void:
	if !required_lane_baselines_synced:
		return
	if player == null:
		return
	if connection_service == null:
		return
	if menu_flow != null && menu_flow.is_gameplay_paused:
		return

	var input_packet: Dictionary = player.get_input_packet()
	var input_state: Dictionary = input_packet.get("input", {})
	var now_msec := int(_clock.call())
	var input_changed := _last_sent_msec < 0 or input_state != _last_sent_input
	var heartbeat_due := _last_sent_msec < 0 or now_msec - _last_sent_msec >= INPUT_HEARTBEAT_MSEC
	if !input_changed and !heartbeat_due:
		return

	connection_service.send_input_packet(input_packet)
	_last_sent_input = input_state.duplicate(true)
	_last_sent_msec = now_msec
