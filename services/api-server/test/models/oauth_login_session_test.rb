require "test_helper"

class OauthLoginSessionTest < ActiveSupport::TestCase
  test "status predicates reflect the current status" do
    pending_session = OauthLoginSession.new(
      public_id: "login-session-pending",
      provider: "discord",
      poll_secret_digest: "digest-pending",
      status: OauthLoginSession::STATUS_PENDING,
      expires_at: 1.minute.from_now
    )

    authenticated_session = OauthLoginSession.new(
      public_id: "login-session-authenticated",
      provider: "discord",
      poll_secret_digest: "digest-authenticated",
      status: OauthLoginSession::STATUS_AUTHENTICATED,
      expires_at: 1.minute.from_now
    )

    assert_predicate pending_session, :pending?
    assert_predicate authenticated_session, :authenticated?
    assert_predicate pending_session, :usable_for_poll?
    assert_not authenticated_session.pending?
  end

  test "expired? and consumed? reflect timestamp state" do
    expired_session = OauthLoginSession.new(
      public_id: "login-session-expired",
      provider: "discord",
      poll_secret_digest: "digest-expired",
      status: OauthLoginSession::STATUS_PENDING,
      expires_at: 1.minute.ago
    )

    consumed_session = OauthLoginSession.new(
      public_id: "login-session-consumed",
      provider: "discord",
      poll_secret_digest: "digest-consumed",
      status: OauthLoginSession::STATUS_CONSUMED,
      expires_at: 1.minute.from_now,
      consumed_at: Time.current
    )

    assert_predicate expired_session, :expired?
    assert_predicate consumed_session, :consumed?
    assert_not consumed_session.usable_for_poll?
  end

  test "authenticate! stores the user and sets authenticated status" do
    user = User.create!(display_name: "Ada")
    oauth_login_session = OauthLoginSession.create!(
      public_id: "login-session-authenticate",
      provider: "discord",
      poll_secret_digest: "digest-authenticate",
      status: OauthLoginSession::STATUS_PENDING,
      expires_at: 1.minute.from_now
    )

    oauth_login_session.authenticate!(user)

    assert_equal user, oauth_login_session.reload.user
    assert_equal OauthLoginSession::STATUS_AUTHENTICATED, oauth_login_session.status
  end

  test "consume! marks the session consumed" do
    oauth_login_session = OauthLoginSession.create!(
      public_id: "login-session-consume",
      provider: "discord",
      poll_secret_digest: "digest-consume",
      status: OauthLoginSession::STATUS_PENDING,
      expires_at: 1.minute.from_now
    )

    oauth_login_session.consume!

    assert_equal OauthLoginSession::STATUS_CONSUMED, oauth_login_session.status
    assert_predicate oauth_login_session.consumed_at, :present?
  end
end
