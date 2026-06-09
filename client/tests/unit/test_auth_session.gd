extends GutTest

const AuthSession := preload("res://scripts/auth/auth_session.gd")


func test_auth_session_starts_signed_out() -> void:
	var session := AuthSession.new()

	assert_false(session.is_signed_in())
	assert_false(session.signed_in)
	assert_eq(session.token, "")
	assert_null(session.user_id)
	assert_eq(session.display_name, "")
	assert_null(session.email)


func test_set_signed_in_stores_token_and_user_display_name() -> void:
	var session := AuthSession.new()

	session.set_signed_in("bearer-token", {
		"id": 7,
		"display_name": "Ada Lovelace",
		"email": "ada@example.com",
	})

	assert_true(session.is_signed_in())
	assert_eq(session.token, "bearer-token")
	assert_eq(session.user_id, 7)
	assert_eq(session.display_name, "Ada Lovelace")
	assert_eq(session.email, "ada@example.com")


func test_clear_resets_signed_in_state_and_user_data() -> void:
	var session := AuthSession.new()
	session.set_signed_in("bearer-token", {
		"id": 7,
		"display_name": "Ada Lovelace",
		"email": "ada@example.com",
	})

	session.clear()

	assert_false(session.is_signed_in())
	assert_eq(session.token, "")
	assert_null(session.user_id)
	assert_eq(session.display_name, "")
	assert_null(session.email)
