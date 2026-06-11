require "test_helper"

class Api::Auth::MeControllerTest < ActionDispatch::IntegrationTest
  setup do
    @user = User.create!(display_name: "Ada")
    PasswordCredential.create!(
      user: @user,
      email: "ada@example.com",
      password: "secret123",
      password_confirmation: "secret123"
    )
    @raw_token, @access_token = AccessToken.issue_for(@user)
    @revoked_raw_token, @revoked_access_token = AccessToken.issue_for(@user)
    @revoked_access_token.update!(revoked_at: Time.current)
    @expired_raw_token = "expired-token"
    AccessToken.create!(
      user: @user,
      token_digest: AccessToken.digest_for(@expired_raw_token),
      audience: "api",
      expires_at: 1.minute.ago
    )
  end

  test "GET /api/auth/me without a token returns 401" do
    get "/api/auth/me"

    assert_response :unauthorized
    assert_openapi_response!
  end

  test "GET /api/auth/me with a malformed Authorization header returns 401" do
    get "/api/auth/me", headers: { "Authorization" => "Token #{@raw_token}" }

    assert_response :unauthorized
    assert_openapi_response!
  end

  test "GET /api/auth/me with an unknown bearer token returns 401" do
    get "/api/auth/me", headers: auth_headers("unknown-token")

    assert_response :unauthorized
    assert_openapi_response!
  end

  test "GET /api/auth/me with a valid bearer token returns 200" do
    before_last_used_at = @access_token.last_used_at

    get "/api/auth/me", headers: auth_headers(@raw_token)

    assert_response :success
    assert_openapi_response!

    body = JSON.parse(response.body)
    assert_equal @user.id, body["user"]["id"]
    assert_equal @user.account_id, body["user"]["account_id"]
    assert_equal "Ada", body["user"]["display_name"]
    assert_equal "ada@example.com", body["user"]["email"]
    assert_nil body["user"]["password_digest"]
    assert_nil body["user"]["token_digest"]
    assert_nil before_last_used_at
    assert_predicate @access_token.reload.last_used_at, :present?
  end

  test "GET /api/auth/me with a revoked token returns 401" do
    get "/api/auth/me", headers: auth_headers(@revoked_raw_token)

    assert_response :unauthorized
    assert_openapi_response!
  end

  test "GET /api/auth/me with an expired token returns 401" do
    get "/api/auth/me", headers: auth_headers(@expired_raw_token)

    assert_response :unauthorized
    assert_openapi_response!
  end

  private

  def auth_headers(raw_token)
    { "Authorization" => "Bearer #{raw_token}" }
  end
end
