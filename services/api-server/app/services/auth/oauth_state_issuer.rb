module Auth
  class OauthStateIssuer
    STATE_BYTES = 32

    def self.call(provider:, redirect_after: nil, expires_at: 10.minutes.from_now)
      raw_state = SecureRandom.hex(STATE_BYTES)
      oauth_state = OauthState.create!(
        provider: provider,
        state_digest: OauthState.digest_for(raw_state),
        redirect_after: redirect_after,
        expires_at: expires_at
      )

      { state: raw_state, oauth_state: oauth_state }
    end
  end
end
