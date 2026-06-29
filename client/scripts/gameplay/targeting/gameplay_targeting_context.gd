extends RefCounted
class_name GameplayTargetingContext

const GameplayTargetCandidateFlowScript = preload("res://scripts/gameplay/targeting/gameplay_target_candidate_flow.gd")
const TargetRequestFlowScript = preload("res://scripts/gameplay/targeting/target_request_flow.gd")

var target_candidate_flow
var target_request_flow


func configure(
	connection_service_ref,
	target_position_source_ref,
	mouse_visual_position_provider_ref: Callable,
	server_position_converter_ref: Callable
) -> void:
	target_candidate_flow = GameplayTargetCandidateFlowScript.new()
	target_candidate_flow.configure(target_position_source_ref)
	target_request_flow = TargetRequestFlowScript.new()
	target_request_flow.configure(
		connection_service_ref,
		Callable(target_candidate_flow, "target_visual_candidates"),
		mouse_visual_position_provider_ref,
		server_position_converter_ref
	)


func target_visual_candidates() -> Array:
	if target_candidate_flow == null:
		return []
	return target_candidate_flow.target_visual_candidates()


func select_target() -> bool:
	if target_request_flow == null:
		return false
	return target_request_flow.select_target()


func deselect_target() -> void:
	if target_request_flow != null:
		target_request_flow.deselect_target()

