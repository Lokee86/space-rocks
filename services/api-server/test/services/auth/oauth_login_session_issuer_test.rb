require "test_helper"

class Auth::OauthLoginSessionIssuerTest < ActiveSupport::TestCase
  test "issuer returns a raw poll secret and stores only its digest" do
    result = Auth::OauthLoginSessionIssuer.call

    assert_predicate result[:poll_secret], :present?
    assert_predicate result[:oauth_login_session], :present?
    assert_equal "discord", result[:oauth_login_session].provider
    assert_equal OauthLoginSession::STATUS_PENDING, result[:oauth_login_session].status
    assert_equal Digest::SHA256.hexdigest(result[:poll_secret]), result[:oauth_login_session].poll_secret_digest
    assert_not_equal result[:poll_secret], result[:oauth_login_session].poll_secret_digest
    assert_operator result[:oauth_login_session].expires_at, :>, Time.current
  end
end
