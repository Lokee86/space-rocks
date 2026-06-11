require "test_helper"

class Api::Auth::SessionsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @user = User.create!(display_name: "Ada")
    @password_credential = PasswordCredential.create!(
      user: @user,
      email: "ada@example.com",
      password: "secret123",
      password_confirmation: "secret123"
    )
    @raw_token, @access_token = AccessToken.issue_for(@user)
    @other_raw_token, @other_access_token = AccessToken.issue_for(@user)
  end

  test "POST /api/auth/login succeeds with valid credentials" do
    post "/api/auth/login", params: {
      email: "ADA@EXAMPLE.COM",
      password: "secret123"
    }, as: :json

    assert_response :ok
    assert_openapi_contract!

    body = JSON.parse(response.body)

    assert_predicate body["token"], :present?
    assert_equal @user.id, body["user"]["id"]
    assert_equal "Ada", body["user"]["display_name"]
    assert_equal "ada@example.com", body["user"]["email"]
  end

  test "POST /api/auth/login fails with wrong password" do
    post "/api/auth/login", params: {
      email: "ada@example.com",
      password: "wrong-password"
    }, as: :json

    assert_response :unauthorized
    assert_openapi_response!

    body = JSON.parse(response.body)
    assert_equal "invalid_credentials", body["error"]
  end

  test "DELETE /api/auth/logout with a valid bearer token revokes only that token" do
    delete "/api/auth/logout", headers: auth_headers(@raw_token)

    assert_response :no_content
    assert_openapi_response!

    assert_predicate @access_token.reload, :revoked?
    other_access_token = @other_access_token.reload
    assert_not_predicate other_access_token, :revoked?
    assert_predicate other_access_token, :active?
  end

  test "DELETE /api/auth/logout without a token returns 401" do
    delete "/api/auth/logout"

    assert_response :unauthorized
    assert_openapi_response!
  end

  private

  def auth_headers(raw_token)
    { "Authorization" => "Bearer #{raw_token}" }
  end
end
