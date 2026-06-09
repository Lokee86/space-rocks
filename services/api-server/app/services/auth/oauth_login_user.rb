module Auth
  class OauthLoginUser
    def self.call(profile:)
      user = Auth::OauthResolveUser.call(profile: profile)
      return Auth::Result.new(error: :invalid) unless user

      token_result = Auth::IssueAccessToken.call(user: user)

      Auth::Result.new(user: user, token: token_result[:token])
    end
  end
end
