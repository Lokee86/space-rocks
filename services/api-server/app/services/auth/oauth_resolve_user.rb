module Auth
  class OauthResolveUser
    def self.call(profile:)
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
