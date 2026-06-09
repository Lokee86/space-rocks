module Auth
  class VerifyAccessToken
    INVALID_TOKEN_ERROR = :invalid_token

    def self.call(raw_token:)
      access_token = AccessToken.find_active_by_raw_token(raw_token)
      return Auth::Result.new(error: INVALID_TOKEN_ERROR) unless access_token

      access_token.update!(last_used_at: Time.current)

      Auth::Result.new(user: access_token.user, token: access_token)
    end
  end
end
