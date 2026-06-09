module Auth
  class OauthStateVerifier
    INVALID_STATE_ERROR = :invalid_oauth_state

    def self.call(provider:, state:)
      oauth_state = OauthState.find_by(state_digest: OauthState.digest_for(state))
      return Auth::Result.new(error: INVALID_STATE_ERROR) unless oauth_state&.provider == provider
      return Auth::Result.new(error: INVALID_STATE_ERROR) unless oauth_state.usable?

      oauth_state.consume!
      Auth::Result.new(token: oauth_state)
    end
  end
end
