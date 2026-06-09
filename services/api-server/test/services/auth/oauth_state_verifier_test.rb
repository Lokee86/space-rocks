require "test_helper"

class Auth::OauthStateVerifierTest < ActiveSupport::TestCase
  test "verifier accepts a valid state" do
    issued = Auth::OauthStateIssuer.call(provider: "discord")

    verified = Auth::OauthStateVerifier.call(provider: "discord", state: issued[:state])

    assert_predicate verified, :success?
    assert_equal issued[:oauth_state].id, verified.token.id
  end

  test "verifier consumes state after successful validation" do
    issued = Auth::OauthStateIssuer.call(provider: "discord")

    verified = Auth::OauthStateVerifier.call(provider: "discord", state: issued[:state])

    assert_predicate verified.token, :consumed?
    assert_predicate verified.token.consumed_at, :present?
  end

  test "verifier rejects missing state" do
    result = Auth::OauthStateVerifier.call(provider: "discord", state: nil)

    assert_not result.success?
  end

  test "verifier rejects expired state" do
    oauth_state = OauthState.create!(
      provider: "discord",
      state_digest: OauthState.digest_for("expired-state"),
      expires_at: 1.minute.ago
    )

    result = Auth::OauthStateVerifier.call(provider: "discord", state: "expired-state")

    assert_not result.success?
    assert_not oauth_state.usable?
  end

  test "verifier rejects consumed state" do
    issued = Auth::OauthStateIssuer.call(provider: "discord")
    issued[:oauth_state].consume!

    result = Auth::OauthStateVerifier.call(provider: "discord", state: issued[:state])

    assert_not result.success?
  end

  test "verifier rejects wrong-provider state" do
    issued = Auth::OauthStateIssuer.call(provider: "google")

    result = Auth::OauthStateVerifier.call(provider: "discord", state: issued[:state])

    assert_not result.success?
  end
end
