require "test_helper"

class OauthStateTest < ActiveSupport::TestCase
  test "expired? returns true after expires_at is in the past" do
    oauth_state = OauthState.new(
      provider: "discord",
      state_digest: "digest",
      expires_at: 1.minute.ago
    )

    assert_predicate oauth_state, :expired?
  end

  test "consumed? returns true when consumed_at is set" do
    oauth_state = OauthState.new(
      provider: "discord",
      state_digest: "digest",
      expires_at: 1.minute.from_now,
      consumed_at: Time.current
    )

    assert_predicate oauth_state, :consumed?
  end

  test "usable? is true only when unexpired and unconsumed" do
    usable_state = OauthState.new(
      provider: "discord",
      state_digest: "digest-usable",
      expires_at: 1.minute.from_now
    )

    expired_state = OauthState.new(
      provider: "discord",
      state_digest: "digest-expired",
      expires_at: 1.minute.ago
    )

    consumed_state = OauthState.new(
      provider: "discord",
      state_digest: "digest-consumed",
      expires_at: 1.minute.from_now,
      consumed_at: Time.current
    )

    assert_predicate usable_state, :usable?
    assert_not expired_state.usable?
    assert_not consumed_state.usable?
  end

  test "consume! sets consumed_at" do
    oauth_state = OauthState.create!(
      provider: "discord",
      state_digest: "digest",
      expires_at: 1.minute.from_now
    )

    oauth_state.consume!

    assert_predicate oauth_state.consumed_at, :present?
  end
end
