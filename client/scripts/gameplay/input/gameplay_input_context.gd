extends RefCounted
class_name GameplayInputContext

const GameplayInputFlow = preload("res://scripts/gameplay/input/gameplay_input_flow.gd")
const GameplayPauseInputFlow = preload("res://scripts/gameplay/input/gameplay_pause_input_flow.gd")
const GameplayDevtoolsContext = preload("res://scripts/devtools/gameplay_devtools_context.gd")

var input_flow
var pause_input_flow
var devtools_context


func configure(connection_service_ref, player_ref, menu_flow_ref) -> void:
	input_flow = GameplayInputFlow.new()
	input_flow.configure(connection_service_ref, player_ref, menu_flow_ref)
	devtools_context = GameplayDevtoolsContext.new()
	devtools_context.configure(connection_service_ref)


func configure_pause_input(menu_flow_ref) -> void:
	pause_input_flow = GameplayPauseInputFlow.new()
	pause_input_flow.configure(menu_flow_ref)


func reset() -> void:
	if input_flow != null:
		input_flow.reset()
	if pause_input_flow != null:
		pause_input_flow.reset()
	if devtools_context != null:
		devtools_context.reset()


func mark_gameplay_state_received() -> void:
	if input_flow != null:
		input_flow.mark_gameplay_state_received()


func process(has_received_state: bool) -> void:
	if pause_input_flow != null:
		pause_input_flow.process(has_received_state)
	if devtools_context != null:
		devtools_context.process(has_received_state)
	if input_flow != null:
		input_flow.process()
