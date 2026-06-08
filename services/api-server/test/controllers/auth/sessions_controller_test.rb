require "test_helper"

class Auth::SessionsControllerTest < ActionDispatch::IntegrationTest
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

  test "POST /auth/login succeeds with valid credentials" do
    post "/auth/login", params: {
      email: "ADA@EXAMPLE.COM",
      password: "secret123"
    }

    assert_response :ok

    body = JSON.parse(response.body)

    assert_predicate body["token"], :present?
    assert_equal @user.id, body["user"]["id"]
    assert_equal "Ada", body["user"]["display_name"]
    assert_equal "ada@example.com", body["user"]["email"]
  end

  test "POST /auth/login fails with wrong password" do
    post "/auth/login", params: {
      email: "ada@example.com",
      password: "wrong-password"
    }

    assert_response :unauthorized

    body = JSON.parse(response.body)
    assert_equal "invalid_credentials", body["error"]
  end

  test "DELETE /auth/logout with a valid bearer token revokes only that token" do
    delete "/auth/logout", headers: auth_headers(@raw_token)

    assert_response :no_content

    assert_predicate @access_token.reload, :revoked?
    other_access_token = @other_access_token.reload
    assert_not_predicate other_access_token, :revoked?
    assert_predicate other_access_token, :active?
  end

  test "DELETE /auth/logout without a token returns 401" do
    delete "/auth/logout"

    assert_response :unauthorized
  end

  private

  def auth_headers(raw_token)
    { "Authorization" => "Bearer #{raw_token}" }
  end
end
