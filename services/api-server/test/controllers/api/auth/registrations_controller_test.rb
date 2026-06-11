require "test_helper"

class Api::Auth::RegistrationsControllerTest < ActionDispatch::IntegrationTest
  test "POST /api/auth/register creates a user and password credential" do
    post "/api/auth/register", params: {
      display_name: "Ada",
      email: "Ada@Example.com",
      password: "secret123"
    }, as: :json

    assert_response :created
    assert_openapi_contract!

    body = JSON.parse(response.body)

    assert_predicate body["token"], :present?
    assert_equal "Ada", body["user"]["display_name"]
    assert_equal "ada@example.com", body["user"]["email"]
    assert_predicate body["user"]["id"], :present?
    assert_nil body["password_digest"]
    assert_nil body["token_digest"]
    assert_nil body["user"]["password_digest"]
    assert_nil body["user"]["token_digest"]

    user = User.find(body["user"]["id"])
    assert_equal "Ada", user.display_name
    assert_equal "ada@example.com", user.password_credential.email
    assert user.password_credential.authenticate_password("secret123")
  end

  test "POST /api/auth/register returns an error for duplicate email" do
    existing_user = User.create!(display_name: "Existing")
    PasswordCredential.create!(
      user: existing_user,
      email: "duplicate@example.com",
      password: "secret123",
      password_confirmation: "secret123"
    )

    post "/api/auth/register", params: {
      display_name: "Another",
      email: "duplicate@example.com",
      password: "secret456"
    }, as: :json

    assert_response :unprocessable_entity
    assert_openapi_response!

    body = JSON.parse(response.body)
    assert_equal "invalid", body["error"]
  end
end
