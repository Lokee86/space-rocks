require "test_helper"

class Internal::Auth::VerifyTokensControllerTest < ActionDispatch::IntegrationTest
  setup do
    @user = User.create!(display_name: "Ada")
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

  test "POST /internal/auth/verify-token without an Authorization header returns 401" do
    with_internal_token_env do
      post "/internal/auth/verify-token"
    end

    assert_response :unauthorized
  end

  test "POST /internal/auth/verify-token with a malformed Authorization header returns 401" do
    with_internal_token_env do
      post "/internal/auth/verify-token", headers: { "Authorization" => "Token test-internal-token" }
    end

    assert_response :unauthorized
  end

  test "POST /internal/auth/verify-token with the wrong bearer token returns 401" do
    with_internal_token_env do
      post "/internal/auth/verify-token", headers: internal_headers("wrong-token")
    end

    assert_response :unauthorized
  end

  test "POST /internal/auth/verify-token with a valid internal secret and valid user token returns 200" do
    with_internal_token_env do
      post "/internal/auth/verify-token", params: { token: @raw_token }, headers: internal_headers
    end

    assert_response :success

    body = JSON.parse(response.body)

    assert_equal true, body["valid"]
    assert_equal @user.id, body["user"]["id"]
    assert_equal "Ada", body["user"]["display_name"]
    refute_includes response.body, "email"
    refute_includes response.body, "token_digest"
    refute_includes response.body, "password_digest"
    refute_includes response.body, "audience"
    refute_includes response.body, "expires_at"
    refute_includes response.body, "revoked_at"
    refute_includes response.body, "last_used_at"
  end

  test "POST /internal/auth/verify-token with a valid internal secret and unknown user token returns valid false" do
    with_internal_token_env do
      post "/internal/auth/verify-token", params: { token: "unknown-token" }, headers: internal_headers
    end

    assert_response :success
    assert_equal({ "valid" => false }, JSON.parse(response.body))
  end

  test "POST /internal/auth/verify-token with a valid internal secret and missing token param returns valid false" do
    with_internal_token_env do
      post "/internal/auth/verify-token", headers: internal_headers
    end

    assert_response :success
    assert_equal({ "valid" => false }, JSON.parse(response.body))
  end

  test "POST /internal/auth/verify-token with a valid internal secret and revoked user token returns valid false" do
    with_internal_token_env do
      post "/internal/auth/verify-token", params: { token: @revoked_raw_token }, headers: internal_headers
    end

    assert_response :success
    assert_equal({ "valid" => false }, JSON.parse(response.body))
  end

  test "POST /internal/auth/verify-token with a valid internal secret and expired user token returns valid false" do
    with_internal_token_env do
      post "/internal/auth/verify-token", params: { token: @expired_raw_token }, headers: internal_headers
    end

    assert_response :success
    assert_equal({ "valid" => false }, JSON.parse(response.body))
  end

  private

  def internal_headers(token = "test-internal-token")
    { "Authorization" => "Bearer #{token}" }
  end

  def with_internal_token_env
    previous_value = ENV["GAME_SERVER_INTERNAL_TOKEN"]
    ENV["GAME_SERVER_INTERNAL_TOKEN"] = "test-internal-token"
    yield
  ensure
    if previous_value.nil?
      ENV.delete("GAME_SERVER_INTERNAL_TOKEN")
    else
      ENV["GAME_SERVER_INTERNAL_TOKEN"] = previous_value
    end
  end
end
