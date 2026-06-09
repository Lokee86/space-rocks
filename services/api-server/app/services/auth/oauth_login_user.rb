module Auth
  class OauthLoginUser
    def self.call(profile:)
      user = find_or_create_user(profile)
      return Auth::Result.new(error: :invalid) unless user

      token_result = Auth::IssueAccessToken.call(user: user)

      Auth::Result.new(user: user, token: token_result[:token])
    end

    def self.find_or_create_user(profile)
      identity = UserIdentity.find_by(
        provider: profile.provider,
        provider_uid: profile.provider_user_id
      )
      return identity.user if identity

      user = nil

      ActiveRecord::Base.transaction do
        user = User.create!(display_name: profile.display_name)
        user.user_identities.create!(
          provider: profile.provider,
          provider_uid: profile.provider_user_id,
          email: profile.email
        )
      end

      user
    rescue ActiveRecord::RecordInvalid
      nil
    end
  end
end
