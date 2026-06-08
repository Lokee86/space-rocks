require "test_helper"

class AccessTokenTest < ActiveSupport::TestCase
  test "issue_for returns a raw token" do
    user = User.create!(display_name: "One")

    raw_token, = AccessToken.issue_for(user)

    assert_predicate raw_token, :present?
  end

  test "database stores token_digest and not the raw token" do
    user = User.create!(display_name: "One")

    raw_token, access_token = AccessToken.issue_for(user)

    assert_equal AccessToken.digest_for(raw_token), access_token.token_digest
    assert_not_equal raw_token, access_token.token_digest
  end

  test "find_active_by_raw_token finds a valid token" do
    user = User.create!(display_name: "One")
    raw_token, = AccessToken.issue_for(user)

    found = AccessToken.find_active_by_raw_token(raw_token)

    assert_equal user, found.user
  end

  test "expired tokens are not active" do
    raw_token = "expired-token"
    access_token = AccessToken.create!(
      user: User.create!(display_name: "One"),
      token_digest: AccessToken.digest_for(raw_token),
      audience: "api",
      expires_at: 1.minute.ago
    )

    assert_not access_token.active?
    assert_not AccessToken.find_active_by_raw_token(raw_token)
  end

  test "revoked tokens are not active" do
    raw_token = "revoked-token"
    access_token = AccessToken.create!(
      user: User.create!(display_name: "One"),
      token_digest: AccessToken.digest_for(raw_token),
      audience: "api",
      expires_at: 1.hour.from_now,
      revoked_at: Time.current
    )

    assert_not access_token.active?
    assert_not AccessToken.find_active_by_raw_token(raw_token)
  end

  test "active? returns false for expired or revoked tokens" do
    expired_token = AccessToken.create!(
      user: User.create!(display_name: "One"),
      token_digest: "digest3",
      audience: "api",
      expires_at: 1.minute.ago
    )

    revoked_token = AccessToken.create!(
      user: User.create!(display_name: "Two"),
      token_digest: "digest4",
      audience: "api",
      expires_at: 1.hour.from_now,
      revoked_at: Time.current
    )

    assert_not expired_token.active?
    assert_not revoked_token.active?
  end
end
