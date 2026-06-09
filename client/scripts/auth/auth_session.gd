extends RefCounted
class_name AuthSession

var signed_in: bool = false
var token: String = ""
var user_id = null
var display_name: String = ""
var email = null


func clear() -> void:
	signed_in = false
	token = ""
	user_id = null
	display_name = ""
	email = null


func set_signed_in(token_value: String, user_payload: Dictionary) -> void:
	signed_in = true
	token = token_value
	user_id = user_payload.get("id")
	display_name = user_payload.get("display_name", "")
	email = user_payload.get("email")


func is_signed_in() -> bool:
	return signed_in
