extends RefCounted
class_name GameplaySpectateContext


var spectate_flow
var menu_flow
var spectate_menu_state
var world_sync


func configure(menu_flow_ref, spectate_menu_state_ref, world_sync_ref) -> void:
	menu_flow = menu_flow_ref
	spectate_menu_state = spectate_menu_state_ref
	world_sync = world_sync_ref
	if spectate_flow == null:
		spectate_flow = GameplaySpectateFlow.new()
	if menu_flow != null && menu_flow.has_signal("spectate_requested"):
		var spectate_callable := Callable(self, "_on_spectate_requested")
		if !menu_flow.spectate_requested.is_connected(spectate_callable):
			menu_flow.spectate_requested.connect(spectate_callable)
	spectate_flow.configure(menu_flow, spectate_menu_state, world_sync)


func configure_menu_state(spectate_menu_state_ref) -> void:
	spectate_menu_state = spectate_menu_state_ref
	if spectate_flow != null:
		spectate_flow.configure(menu_flow, spectate_menu_state, world_sync)


func reset() -> void:
	if spectate_flow != null:
		spectate_flow.reset()


func process() -> void:
	if spectate_flow != null:
		spectate_flow.process()


func begin_spectating() -> void:
	if spectate_flow != null:
		spectate_flow.begin_spectating()


func request_open_spectate_menu() -> void:
	if spectate_flow != null:
		spectate_flow.request_open_spectate_menu()


func request_cycle_target() -> void:
	if spectate_flow != null:
		spectate_flow.request_cycle_target()


func _on_spectate_requested() -> void:
	begin_spectating()
