require "test_helper"

class Auth::VerifyAccessTokenTest < ActiveSupport::TestCase
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

  test "valid raw token returns the user and access token" do
    result = Auth::VerifyAccessToken.call(raw_token: @raw_token)

    assert_predicate result, :success?
    assert_equal @user, result.user
    assert_equal @access_token.id, result.token.id
  end

  test "valid raw token updates last_used_at" do
    before_last_used_at = @access_token.last_used_at

    result = Auth::VerifyAccessToken.call(raw_token: @raw_token)

    assert_nil before_last_used_at
    assert_predicate result.token.reload.last_used_at, :present?
  end

  test "unknown token returns invalid_token" do
    result = Auth::VerifyAccessToken.call(raw_token: "unknown-token")

    assert_not result.success?
    assert_equal :invalid_token, result.error
  end

  test "nil token returns invalid_token" do
    result = Auth::VerifyAccessToken.call(raw_token: nil)

    assert_not result.success?
    assert_equal :invalid_token, result.error
  end

  test "revoked token returns invalid_token" do
    result = Auth::VerifyAccessToken.call(raw_token: @revoked_raw_token)

    assert_not result.success?
    assert_equal :invalid_token, result.error
  end

  test "expired token returns invalid_token" do
    result = Auth::VerifyAccessToken.call(raw_token: @expired_raw_token)

    assert_not result.success?
    assert_equal :invalid_token, result.error
  end
end
