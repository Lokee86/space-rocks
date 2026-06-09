require "test_helper"

class Auth::OauthStateIssuerTest < ActiveSupport::TestCase
  test "issuer returns a raw state token" do
    result = Auth::OauthStateIssuer.call(provider: "discord")

    assert_predicate result[:state], :present?
  end

  test "issuer stores only a digest, not the raw token" do
    result = Auth::OauthStateIssuer.call(provider: "discord")

    assert_equal OauthState.digest_for(result[:state]), result[:oauth_state].state_digest
    assert_not_equal result[:state], result[:oauth_state].state_digest
  end
end
